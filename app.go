package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/install"
	"github.com/shinerio/skillflow/core/notify"
	"github.com/shinerio/skillflow/core/registry"
	"github.com/shinerio/skillflow/core/skill"
	toolsync "github.com/shinerio/skillflow/core/sync"
	"github.com/shinerio/skillflow/core/update"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	hub         *notify.Hub
	storage     *skill.Storage
	config      *config.Service
	starStorage *coregit.StarStorage
	cacheDir    string

	// Git sync state
	gitConflictMu      sync.Mutex
	gitConflictPending bool
	stopAutoSync       chan struct{}
}

func NewApp() *App {
	return &App{hub: notify.NewHub()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dataDir := config.AppDataDir()
	a.config = config.NewService(dataDir)
	cfg, _ := a.config.Load()
	a.storage = skill.NewStorage(cfg.SkillsStorageDir)
	a.cacheDir = filepath.Join(dataDir, "cache")
	a.starStorage = coregit.NewStarStorage(filepath.Join(dataDir, "star_repos.json"))
	registerAdapters()
	registerProviders()
	go forwardEvents(ctx, a.hub)
	go a.checkUpdatesOnStartup()
	go a.updateStarredReposOnStartup()
	go a.gitPullOnStartup()
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
}

// proxyHTTPClient builds an *http.Client configured according to the saved proxy settings.
// Falls back to http.DefaultClient on any error.
func (a *App) proxyHTTPClient() *http.Client {
	cfg, err := a.config.Load()
	if err != nil {
		return http.DefaultClient
	}
	switch cfg.Proxy.Mode {
	case config.ProxyModeSystem:
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	case config.ProxyModeManual:
		if cfg.Proxy.URL == "" {
			return http.DefaultClient
		}
		proxyURL, err := url.Parse(cfg.Proxy.URL)
		if err != nil {
			return http.DefaultClient
		}
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	default: // ProxyModeNone or empty
		return &http.Client{Transport: &http.Transport{Proxy: nil}}
	}
}

func (a *App) domReady(_ context.Context)         {}
func (a *App) beforeClose(_ context.Context) bool { return false }
func (a *App) shutdown(_ context.Context)         {}

// autoBackup triggers cloud backup after any mutating operation if cloud is enabled.
func (a *App) autoBackup() {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider == "" {
		return
	}
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return
	}
	a.hub.Publish(notify.Event{Type: notify.EventBackupStarted})
	err = provider.Sync(a.ctx, cfg.SkillsStorageDir, cfg.Cloud.BucketName, cfg.Cloud.RemotePath,
		func(file string) {
			a.hub.Publish(notify.Event{
				Type:    notify.EventBackupProgress,
				Payload: notify.BackupProgressPayload{CurrentFile: file},
			})
		})
	if err != nil {
		a.hub.Publish(notify.Event{Type: notify.EventBackupFailed, Payload: err.Error()})
	} else {
		a.hub.Publish(notify.Event{Type: notify.EventBackupCompleted})
	}
}

// gitPullOnStartup pulls from the remote git repo at startup when the git provider is enabled.
func (a *App) gitPullOnStartup() {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider != backup.GitProviderName {
		return
	}
	p, ok := registry.GetCloudProvider(backup.GitProviderName)
	if !ok {
		return
	}
	if err := p.Init(cfg.Cloud.Credentials); err != nil {
		return
	}
	gitP := p.(*backup.GitProvider)
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	if err := gitP.Restore(a.ctx, "", "", cfg.SkillsStorageDir); err != nil {
		var conflictErr *backup.GitConflictError
		if errors.As(err, &conflictErr) {
			a.gitConflictMu.Lock()
			a.gitConflictPending = true
			a.gitConflictMu.Unlock()
			a.hub.Publish(notify.Event{
				Type:    notify.EventGitConflict,
				Payload: notify.GitConflictPayload{Message: conflictErr.Output},
			})
		} else {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return
	}
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted})
	// Reload storage so newly pulled skills are visible
	if cfg2, err := a.config.Load(); err == nil {
		a.storage = skill.NewStorage(cfg2.SkillsStorageDir)
	}
}

// startAutoSyncTimer starts (or restarts) a periodic auto-backup ticker.
// intervalMinutes <= 0 disables the timer.
func (a *App) startAutoSyncTimer(intervalMinutes int) {
	if a.stopAutoSync != nil {
		close(a.stopAutoSync)
		a.stopAutoSync = nil
	}
	if intervalMinutes <= 0 {
		return
	}
	stop := make(chan struct{})
	a.stopAutoSync = stop
	go func() {
		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.autoBackup()
			case <-stop:
				return
			}
		}
	}()
}

// GetGitConflictPending returns true when a git conflict from startup pull is waiting to be resolved.
func (a *App) GetGitConflictPending() bool {
	a.gitConflictMu.Lock()
	defer a.gitConflictMu.Unlock()
	return a.gitConflictPending
}

// ResolveGitConflict resolves a pending git merge conflict.
// useLocal=true  → keep local changes, force-push to remote.
// useLocal=false → discard local changes, reset to remote state.
func (a *App) ResolveGitConflict(useLocal bool) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	p, ok := registry.GetCloudProvider(backup.GitProviderName)
	if !ok {
		return fmt.Errorf("git provider 未注册")
	}
	if err := p.Init(cfg.Cloud.Credentials); err != nil {
		return err
	}
	gitP := p.(*backup.GitProvider)
	if useLocal {
		err = gitP.ResolveConflictUseLocal(cfg.SkillsStorageDir)
	} else {
		err = gitP.ResolveConflictUseRemote(cfg.SkillsStorageDir)
	}
	if err != nil {
		return err
	}
	a.gitConflictMu.Lock()
	a.gitConflictPending = false
	a.gitConflictMu.Unlock()
	// Reload storage so the dashboard reflects the resolved state
	if cfg2, err2 := a.config.Load(); err2 == nil {
		a.storage = skill.NewStorage(cfg2.SkillsStorageDir)
	}
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted})
	return nil
}

func (a *App) gitProxyURL() string {
	cfg, err := a.config.Load()
	if err != nil {
		return ""
	}
	if cfg.Proxy.Mode == config.ProxyModeManual {
		return cfg.Proxy.URL
	}
	return ""
}

// --- Skills ---

func (a *App) ListSkills() ([]*skill.Skill, error) {
	skills, err := a.storage.ListAll()
	if err != nil {
		return nil, err
	}
	cfg, _ := a.config.Load()
	defaultCat := cfg.DefaultCategory
	if defaultCat == "" {
		defaultCat = "Imported"
	}
	for _, sk := range skills {
		if sk.Category == "" {
			sk.Category = defaultCat
		}
	}
	return skills, nil
}

func (a *App) ListCategories() ([]string, error) {
	cats, err := a.storage.ListCategories()
	if err != nil {
		return nil, err
	}
	cfg, _ := a.config.Load()
	defaultCat := cfg.DefaultCategory
	if defaultCat == "" {
		defaultCat = "Imported"
	}
	// 检查 defaultCat 是否已在列表中
	hasDefault := false
	for _, c := range cats {
		if c == defaultCat {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		// 将 defaultCat 加到列表最前面
		cats = append([]string{defaultCat}, cats...)
	}
	return cats, nil
}

func (a *App) CreateCategory(name string) error {
	return a.storage.CreateCategory(name)
}

func (a *App) RenameCategory(oldName, newName string) error {
	return a.storage.RenameCategory(oldName, newName)
}

func (a *App) DeleteCategory(name string) error {
	cfg, _ := a.config.Load()
	defaultCat := cfg.DefaultCategory
	if defaultCat == "" {
		defaultCat = "Imported"
	}
	skills, err := a.storage.ListAll()
	if err != nil {
		return err
	}
	for _, sk := range skills {
		if sk.Category == name {
			if err := a.storage.MoveCategory(sk.ID, defaultCat); err != nil {
				return err
			}
		}
	}
	return os.Remove(filepath.Join(cfg.SkillsStorageDir, name))
}

func (a *App) MoveSkillCategory(skillID, category string) error {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	return a.storage.MoveCategory(skillID, category)
}

func (a *App) DeleteSkill(skillID string) error {
	if err := a.storage.Delete(skillID); err != nil {
		return err
	}
	go a.autoBackup()
	return nil
}

func (a *App) DeleteSkills(skillIDs []string) error {
	for _, id := range skillIDs {
		if err := a.storage.Delete(id); err != nil {
			return err
		}
	}
	go a.autoBackup()
	return nil
}

func (a *App) GetSkillMeta(skillID string) (*skill.SkillMeta, error) {
	sk, err := a.storage.Get(skillID)
	if err != nil {
		return nil, err
	}
	return skill.ReadMeta(sk.Path)
}

// GetSkillMetaByPath reads skill.md frontmatter from a skill directory path (no ID required).
func (a *App) GetSkillMetaByPath(path string) (*skill.SkillMeta, error) {
	return skill.ReadMeta(path)
}

// ReadSkillFileContent returns the full text content of skill.md inside the given skill directory.
func (a *App) ReadSkillFileContent(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(e.Name()) == "skill.md" {
			data, err := os.ReadFile(filepath.Join(path, e.Name()))
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("skill.md not found in %s", path)
}

// OpenURL opens the given URL in the system default browser.
// Non-HTTP URLs (e.g. SSH git remotes) are first converted to HTTPS.
func (a *App) OpenURL(rawURL string) error {
	target := rawURL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		if canonical, err := coregit.CanonicalRepoURL(rawURL); err == nil {
			target = canonical
		}
	}
	runtime.BrowserOpenURL(a.ctx, target)
	return nil
}

// PushStarSkillsToTools copies starred skill directories directly to the push directory of each
// specified tool, skipping skills that already exist. Returns a list of conflict descriptions.
func (a *App) PushStarSkillsToTools(skillPaths []string, toolNames []string) ([]string, error) {
	cfg, _ := a.config.Load()
	var conflicts []string
	for _, toolName := range toolNames {
		for _, t := range cfg.Tools {
			if t.Name != toolName {
				continue
			}
			if t.PushDir == "" {
				return nil, fmt.Errorf("工具 %s 未配置推送路径", toolName)
			}
			if err := os.MkdirAll(t.PushDir, 0755); err != nil {
				return nil, err
			}
			adapter := getAdapter(t)
			for _, skillPath := range skillPaths {
				name := filepath.Base(skillPath)
				dst := filepath.Join(t.PushDir, name)
				if _, err := os.Stat(dst); err == nil {
					conflicts = append(conflicts, fmt.Sprintf("%s → %s", name, toolName))
					continue
				}
				sk := []*skill.Skill{{Name: name, Path: skillPath}}
				if err := adapter.Push(a.ctx, sk, t.PushDir); err != nil {
					return nil, err
				}
			}
		}
	}
	return conflicts, nil
}

// PushStarSkillsToToolsForce copies starred skill directories to tool push directories,
// overwriting any existing skills.
func (a *App) PushStarSkillsToToolsForce(skillPaths []string, toolNames []string) error {
	cfg, _ := a.config.Load()
	for _, toolName := range toolNames {
		for _, t := range cfg.Tools {
			if t.Name != toolName {
				continue
			}
			if t.PushDir == "" {
				return fmt.Errorf("工具 %s 未配置推送路径", toolName)
			}
			var tempSkills []*skill.Skill
			for _, skillPath := range skillPaths {
				name := filepath.Base(skillPath)
				_ = os.RemoveAll(filepath.Join(t.PushDir, name))
				tempSkills = append(tempSkills, &skill.Skill{Name: name, Path: skillPath})
			}
			if err := getAdapter(t).Push(a.ctx, tempSkills, t.PushDir); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- Install ---

// ScanGitHub scans a remote git repo for valid skills, marking already-installed ones.
func (a *App) ScanGitHub(repoURL string) ([]install.SkillCandidate, error) {
	dataDir := config.AppDataDir()
	cacheDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		return nil, err
	}
	if err := coregit.CloneOrUpdate(a.ctx, repoURL, cacheDir, a.gitProxyURL()); err != nil {
		return nil, err
	}
	repoName, _ := coregit.ParseRepoName(repoURL)
	repoSource, err := coregit.RepoSource(repoURL)
	if err != nil {
		return nil, err
	}
	starSkills, err := coregit.ScanSkills(cacheDir, repoURL, repoName, repoSource)
	if err != nil {
		return nil, err
	}
	existing, _ := a.storage.ListAll()
	existingNames := map[string]bool{}
	for _, sk := range existing {
		existingNames[sk.Name] = true
	}
	var candidates []install.SkillCandidate
	for _, ss := range starSkills {
		candidates = append(candidates, install.SkillCandidate{
			Name:      ss.Name,
			Path:      ss.SubPath,
			Installed: existingNames[ss.Name],
		})
	}
	return candidates, nil
}

// InstallFromGitHub imports selected skills from a scanned remote git repo into storage.
func (a *App) InstallFromGitHub(repoURL string, candidates []install.SkillCandidate, category string) error {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	dataDir := config.AppDataDir()
	cacheDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		return err
	}
	canonicalRepoURL, err := coregit.CanonicalRepoURL(repoURL)
	if err != nil {
		return err
	}
	for _, c := range candidates {
		skillDir := filepath.Join(cacheDir, filepath.FromSlash(c.Path))
		sha, _ := coregit.GetSubPathSHA(a.ctx, cacheDir, c.Path)
		sk, err := a.storage.Import(skillDir, category, skill.SourceGitHub, canonicalRepoURL, c.Path)
		if err != nil {
			return fmt.Errorf("import %s: %w", c.Name, err)
		}
		sk.SourceSHA = sha
		_ = a.storage.UpdateMeta(sk)
	}
	go a.autoBackup()
	return nil
}

func (a *App) ImportLocal(dir, category string) (*skill.Skill, error) {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	sk, err := a.storage.Import(dir, category, skill.SourceManual, "", "")
	if err != nil {
		return nil, err
	}
	go a.autoBackup()
	return sk, nil
}

// --- Sync ---

func (a *App) GetEnabledTools() ([]config.ToolConfig, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	var enabled []config.ToolConfig
	for _, t := range cfg.Tools {
		if t.Enabled {
			enabled = append(enabled, t)
		}
	}
	return enabled, nil
}

// ScanToolSkills lists all skills in a tool's configured scan directories for the pull page.
func (a *App) ScanToolSkills(toolName string) ([]*skill.Skill, error) {
	cfg, _ := a.config.Load()
	for _, t := range cfg.Tools {
		if t.Name == toolName {
			return scanToolSkills(a.ctx, getAdapter(t), t.ScanDirs)
		}
	}
	return nil, nil
}

// CheckMissingPushDirs returns tool names and paths whose push directory does not yet exist.
// Each element is map{"name": toolName, "dir": pushDir}.
func (a *App) CheckMissingPushDirs(toolNames []string) ([]map[string]string, error) {
	cfg, _ := a.config.Load()
	var missing []map[string]string
	for _, toolName := range toolNames {
		for _, t := range cfg.Tools {
			if t.Name != toolName || t.PushDir == "" {
				continue
			}
			if _, err := os.Stat(t.PushDir); os.IsNotExist(err) {
				missing = append(missing, map[string]string{"name": t.Name, "dir": t.PushDir})
			}
		}
	}
	return missing, nil
}

// PushToTools pushes selected skills to target tools.
// Returns list of conflict descriptions that were skipped.
func (a *App) PushToTools(skillIDs []string, toolNames []string) ([]string, error) {
	cfg, _ := a.config.Load()
	skills, err := a.storage.ListAll()
	if err != nil {
		return nil, err
	}
	idSet := map[string]bool{}
	for _, id := range skillIDs {
		idSet[id] = true
	}
	var selected []*skill.Skill
	for _, sk := range skills {
		if idSet[sk.ID] {
			selected = append(selected, sk)
		}
	}

	var conflicts []string
	for _, toolName := range toolNames {
		for _, t := range cfg.Tools {
			if t.Name != toolName {
				continue
			}
			if t.PushDir == "" {
				return nil, fmt.Errorf("工具 %s 未配置推送路径", toolName)
			}
			adapter := getAdapter(t)
			for _, sk := range selected {
				dst := filepath.Join(t.PushDir, sk.Name)
				if _, err := os.Stat(dst); err == nil {
					conflicts = append(conflicts, fmt.Sprintf("%s -> %s", sk.Name, toolName))
					continue
				}
			}
			_ = adapter.Push(a.ctx, selected, t.PushDir)
		}
	}
	return conflicts, nil
}

// PushToToolsForce pushes and overwrites conflicts.
func (a *App) PushToToolsForce(skillIDs []string, toolNames []string) error {
	cfg, _ := a.config.Load()
	skills, _ := a.storage.ListAll()
	idSet := map[string]bool{}
	for _, id := range skillIDs {
		idSet[id] = true
	}
	var selected []*skill.Skill
	for _, sk := range skills {
		if idSet[sk.ID] {
			selected = append(selected, sk)
		}
	}
	for _, toolName := range toolNames {
		for _, t := range cfg.Tools {
			if t.Name == toolName {
				if t.PushDir == "" {
					return fmt.Errorf("工具 %s 未配置推送路径", toolName)
				}
				_ = getAdapter(t).Push(a.ctx, selected, t.PushDir)
			}
		}
	}
	return nil
}

// PullFromTool imports selected skills from a tool into SkillFlow storage.
func (a *App) PullFromTool(toolName string, skillNames []string, category string) ([]string, error) {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	cfg, _ := a.config.Load()
	nameSet := map[string]bool{}
	for _, n := range skillNames {
		nameSet[n] = true
	}
	for _, t := range cfg.Tools {
		if t.Name != toolName {
			continue
		}
		candidates, err := scanToolSkills(a.ctx, getAdapter(t), t.ScanDirs)
		if err != nil {
			return nil, err
		}
		var conflicts []string
		for _, sk := range candidates {
			if !nameSet[sk.Name] {
				continue
			}
			if _, err := a.storage.Import(sk.Path, category, skill.SourceManual, "", ""); err == skill.ErrSkillExists {
				conflicts = append(conflicts, sk.Name)
			}
		}
		go a.autoBackup()
		return conflicts, nil
	}
	return nil, nil
}

// PullFromToolForce imports selected skills, overwriting existing ones.
func (a *App) PullFromToolForce(toolName string, skillNames []string, category string) error {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	cfg, _ := a.config.Load()
	nameSet := map[string]bool{}
	for _, n := range skillNames {
		nameSet[n] = true
	}
	for _, t := range cfg.Tools {
		if t.Name != toolName {
			continue
		}
		candidates, err := scanToolSkills(a.ctx, getAdapter(t), t.ScanDirs)
		if err != nil {
			return err
		}
		for _, sk := range candidates {
			if !nameSet[sk.Name] {
				continue
			}
			existing, _ := a.storage.ListAll()
			for _, e := range existing {
				if e.Name == sk.Name {
					_ = a.storage.Delete(e.ID)
					break
				}
			}
			_, _ = a.storage.Import(sk.Path, category, skill.SourceManual, "", "")
		}
		go a.autoBackup()
	}
	return nil
}

func scanToolSkills(ctx context.Context, adapter toolsync.ToolAdapter, scanDirs []string) ([]*skill.Skill, error) {
	var result []*skill.Skill
	seen := map[string]struct{}{}
	for _, dir := range scanDirs {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		skills, err := adapter.Pull(ctx, dir)
		if err != nil {
			return nil, err
		}
		for _, sk := range skills {
			if _, ok := seen[sk.Name]; ok {
				continue
			}
			seen[sk.Name] = struct{}{}
			result = append(result, sk)
		}
	}
	return result, nil
}

func getAdapter(t config.ToolConfig) toolsync.ToolAdapter {
	if a, ok := registry.GetAdapter(t.Name); ok {
		return a
	}
	return toolsync.NewFilesystemAdapter(t.Name, t.PushDir)
}

// --- Config ---

func (a *App) GetConfig() (config.AppConfig, error) {
	return a.config.Load()
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
	if err := a.config.Save(cfg); err != nil {
		return err
	}
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
	return nil
}

func (a *App) AddCustomTool(name, pushDir string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	cfg.Tools = append(cfg.Tools, config.ToolConfig{
		Name:     name,
		ScanDirs: []string{pushDir},
		PushDir:  pushDir,
		Enabled:  true,
		Custom:   true,
	})
	return a.config.Save(cfg)
}

func (a *App) RemoveCustomTool(name string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	var filtered []config.ToolConfig
	for _, t := range cfg.Tools {
		if !(t.Custom && t.Name == name) {
			filtered = append(filtered, t)
		}
	}
	cfg.Tools = filtered
	return a.config.Save(cfg)
}

// --- Backup ---

func (a *App) BackupNow() error {
	a.autoBackup()
	return nil
}

func (a *App) ListCloudFiles() ([]backup.RemoteFile, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return nil, err
	}
	return provider.List(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath)
}

func (a *App) RestoreFromCloud() error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return err
	}
	return provider.Restore(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath, cfg.SkillsStorageDir)
}

// ListCloudProviders returns all registered provider names and their required credential fields.
func (a *App) ListCloudProviders() []map[string]any {
	var result []map[string]any
	for _, p := range registry.AllCloudProviders() {
		result = append(result, map[string]any{
			"name":   p.Name(),
			"fields": p.RequiredCredentials(),
		})
	}
	return result
}

// --- Updates ---

func (a *App) CheckUpdates() error {
	skills, err := a.storage.ListAll()
	if err != nil {
		return err
	}
	checker := update.NewChecker("", a.proxyHTTPClient())
	for _, sk := range skills {
		result, err := checker.Check(a.ctx, sk)
		if err != nil {
			continue
		}
		if result.HasUpdate {
			sk.LatestSHA = result.LatestSHA
			_ = a.storage.UpdateMeta(sk)
			a.hub.Publish(notify.Event{
				Type: notify.EventUpdateAvailable,
				Payload: notify.UpdateAvailablePayload{
					SkillID:    sk.ID,
					SkillName:  sk.Name,
					CurrentSHA: sk.SourceSHA,
					LatestSHA:  result.LatestSHA,
				},
			})
		}
	}
	return nil
}

// UpdateSkill re-downloads a GitHub skill and updates local files and SHA.
func (a *App) UpdateSkill(skillID string) error {
	sk, err := a.storage.Get(skillID)
	if err != nil {
		return err
	}
	inst := install.NewGitHubInstaller("", a.proxyHTTPClient())
	tmpDir := filepath.Join(os.TempDir(), "skillflow-update", sk.Name)
	defer os.RemoveAll(tmpDir)

	c := install.SkillCandidate{Name: sk.Name, Path: sk.SourceSubPath}
	if err := inst.DownloadTo(a.ctx, install.InstallSource{Type: "github", URI: sk.SourceURL}, c, tmpDir); err != nil {
		return err
	}
	if err := a.storage.OverwriteFromDir(skillID, tmpDir); err != nil {
		return err
	}
	sk.SourceSHA = sk.LatestSHA
	sk.LatestSHA = ""
	_ = a.storage.UpdateMeta(sk)
	go a.autoBackup()
	return nil
}

func (a *App) checkUpdatesOnStartup() {
	_ = a.CheckUpdates()
}

func (a *App) updateStarredReposOnStartup() {
	_ = a.UpdateAllStarredRepos()
}

// OpenFolderDialog wraps Wails file dialog for frontend use.
func (a *App) OpenFolderDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Skill 目录",
	})
}

// OpenPath opens the given filesystem path in the OS default file manager.
func (a *App) OpenPath(path string) error {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// Greet is kept for Wails template compatibility.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// --- Starred Repos ---

func (a *App) AddStarredRepo(repoURL string) (*coregit.StarredRepo, error) {
	if err := coregit.CheckGitInstalled(); err != nil {
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	for i, r := range repos {
		if coregit.SameRepo(r.URL, repoURL) {
			if repos[i].Source == "" {
				if source, err := coregit.RepoSource(repos[i].URL); err == nil {
					repos[i].Source = source
					_ = a.starStorage.Save(repos)
				}
			}
			return &repos[i], nil // already starred
		}
	}
	name, err := coregit.ParseRepoName(repoURL)
	if err != nil {
		return nil, err
	}
	dataDir := filepath.Dir(a.cacheDir)
	localDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		return nil, err
	}
	repo := coregit.StarredRepo{URL: repoURL, Name: name, Source: source, LocalDir: localDir}
	if cloneErr := coregit.CloneOrUpdate(a.ctx, repoURL, localDir, a.gitProxyURL()); cloneErr != nil {
		// Return typed errors for auth failures so the frontend can show the right dialog.
		if coregit.IsSSHAuthError(cloneErr) {
			return nil, fmt.Errorf("AUTH_SSH:%s", cloneErr.Error())
		}
		if coregit.IsAuthError(cloneErr) {
			return nil, fmt.Errorf("AUTH_HTTP:%s", cloneErr.Error())
		}
		repo.SyncError = cloneErr.Error()
	} else {
		repo.LastSync = time.Now()
	}
	repos = append(repos, repo)
	if err := a.starStorage.Save(repos); err != nil {
		return nil, err
	}
	return &repos[len(repos)-1], nil
}

// AddStarredRepoWithCredentials clones a repo using the provided HTTP username/password,
// removing any previously failed entry for the same URL first.
func (a *App) AddStarredRepoWithCredentials(repoURL, username, password string) (*coregit.StarredRepo, error) {
	if err := coregit.CheckGitInstalled(); err != nil {
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	// Remove any existing (possibly failed) entry for this URL.
	filtered := repos[:0]
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			filtered = append(filtered, r)
		}
	}
	name, err := coregit.ParseRepoName(repoURL)
	if err != nil {
		return nil, err
	}
	dataDir := filepath.Dir(a.cacheDir)
	localDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		return nil, err
	}
	repo := coregit.StarredRepo{URL: repoURL, Name: name, Source: source, LocalDir: localDir}
	if cloneErr := coregit.CloneOrUpdateWithCreds(a.ctx, repoURL, localDir, a.gitProxyURL(), username, password); cloneErr != nil {
		return nil, cloneErr
	}
	repo.LastSync = time.Now()
	filtered = append(filtered, repo)
	if err := a.starStorage.Save(filtered); err != nil {
		return nil, err
	}
	return &filtered[len(filtered)-1], nil
}

func (a *App) RemoveStarredRepo(repoURL string) error {
	repos, err := a.starStorage.Load()
	if err != nil {
		return err
	}
	filtered := make([]coregit.StarredRepo, 0, len(repos))
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			filtered = append(filtered, r)
		}
	}
	return a.starStorage.Save(filtered)
}

func (a *App) ListStarredRepos() ([]coregit.StarredRepo, error) {
	repos, err := a.starStorage.Load()
	if repos == nil {
		return []coregit.StarredRepo{}, err
	}
	changed := false
	for i := range repos {
		if repos[i].Source != "" {
			continue
		}
		if source, parseErr := coregit.RepoSource(repos[i].URL); parseErr == nil {
			repos[i].Source = source
			changed = true
		}
	}
	if changed {
		_ = a.starStorage.Save(repos)
	}
	return repos, err
}

func (a *App) ListAllStarSkills() ([]coregit.StarSkill, error) {
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	existing, _ := a.storage.ListAll()
	importedNames := map[string]bool{}
	for _, sk := range existing {
		importedNames[sk.Name] = true
	}
	var all []coregit.StarSkill
	for _, r := range repos {
		source := r.Source
		if source == "" {
			source, _ = coregit.RepoSource(r.URL)
		}
		skills, _ := coregit.ScanSkills(r.LocalDir, r.URL, r.Name, source)
		for i := range skills {
			skills[i].Imported = importedNames[skills[i].Name]
		}
		all = append(all, skills...)
	}
	if all == nil {
		return []coregit.StarSkill{}, nil
	}
	return all, nil
}

func (a *App) ListRepoStarSkills(repoURL string) ([]coregit.StarSkill, error) {
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	existing, _ := a.storage.ListAll()
	importedNames := map[string]bool{}
	for _, sk := range existing {
		importedNames[sk.Name] = true
	}
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			continue
		}
		source := r.Source
		if source == "" {
			source, _ = coregit.RepoSource(r.URL)
		}
		skills, err := coregit.ScanSkills(r.LocalDir, r.URL, r.Name, source)
		if err != nil {
			return nil, err
		}
		for i := range skills {
			skills[i].Imported = importedNames[skills[i].Name]
		}
		if skills == nil {
			return []coregit.StarSkill{}, nil
		}
		return skills, nil
	}
	return []coregit.StarSkill{}, nil
}

func (a *App) UpdateStarredRepo(repoURL string) error {
	repos, err := a.starStorage.Load()
	if err != nil {
		return err
	}
	for i, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			continue
		}
		syncErr := coregit.CloneOrUpdate(a.ctx, r.URL, r.LocalDir, a.gitProxyURL())
		if syncErr != nil {
			repos[i].SyncError = syncErr.Error()
		} else {
			repos[i].SyncError = ""
			repos[i].LastSync = time.Now()
		}
		return a.starStorage.Save(repos)
	}
	return nil
}

func (a *App) UpdateAllStarredRepos() error {
	repos, err := a.starStorage.Load()
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	mu := &sync.Mutex{}
	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r := repos[idx]
			syncErr := coregit.CloneOrUpdate(a.ctx, r.URL, r.LocalDir, a.gitProxyURL())
			mu.Lock()
			if syncErr != nil {
				repos[idx].SyncError = syncErr.Error()
			} else {
				repos[idx].SyncError = ""
				repos[idx].LastSync = time.Now()
			}
			mu.Unlock()
			a.hub.Publish(notify.Event{
				Type: notify.EventStarSyncProgress,
				Payload: notify.StarSyncProgressPayload{
					RepoURL:   repos[idx].URL,
					RepoName:  repos[idx].Name,
					SyncError: repos[idx].SyncError,
				},
			})
		}(i)
	}
	wg.Wait()
	a.hub.Publish(notify.Event{Type: notify.EventStarSyncDone})
	return a.starStorage.Save(repos)
}

func (a *App) ImportStarSkills(skillPaths []string, repoURL, category string) error {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	repos, _ := a.starStorage.Load()
	var repoLocalDir string
	canonicalRepoURL := repoURL
	if normalized, err := coregit.CanonicalRepoURL(repoURL); err == nil {
		canonicalRepoURL = normalized
	}
	for _, r := range repos {
		if coregit.SameRepo(r.URL, repoURL) {
			repoLocalDir = r.LocalDir
			if normalized, err := coregit.CanonicalRepoURL(r.URL); err == nil {
				canonicalRepoURL = normalized
			}
			break
		}
	}
	if repoLocalDir == "" {
		return fmt.Errorf("starred repo not found: %s", repoURL)
	}
	for _, skillPath := range skillPaths {
		subPath, _ := filepath.Rel(repoLocalDir, skillPath)
		subPath = filepath.ToSlash(subPath)
		sk, err := a.storage.Import(skillPath, category, skill.SourceGitHub, canonicalRepoURL, subPath)
		if err == skill.ErrSkillExists {
			continue
		}
		if err != nil {
			return err
		}
		sha, _ := coregit.GetSubPathSHA(a.ctx, repoLocalDir, subPath)
		sk.SourceSHA = sha
		_ = a.storage.UpdateMeta(sk)
	}
	go a.autoBackup()
	return nil
}

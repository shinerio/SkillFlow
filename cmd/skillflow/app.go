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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/applog"
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

type maxDepthPuller interface {
	PullWithMaxDepth(ctx context.Context, sourceDir string, maxDepth int) ([]*skill.Skill, error)
}

type App struct {
	ctx         context.Context
	hub         *notify.Hub
	sysLog      *applog.Logger
	storage     *skill.Storage
	config      *config.Service
	starStorage *coregit.StarStorage
	cacheDir    string
	startupOnce sync.Once

	// Git sync state
	gitConflictMu      sync.Mutex
	gitConflictPending bool
	stopAutoSync       chan struct{}
}

const defaultCategoryName = "Default"

func normalizeCategoryName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || strings.EqualFold(trimmed, defaultCategoryName) {
		return defaultCategoryName
	}
	return trimmed
}

func NewApp() *App {
	return &App{hub: notify.NewHub()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dataDir := config.AppDataDir()
	a.config = config.NewService(dataDir)
	cfg, err := a.config.Load()
	if err != nil {
		runtime.LogErrorf(ctx, "load config failed: %v", err)
		cfg = config.DefaultConfig(dataDir)
	}
	a.initLogger(cfg.LogLevel)
	a.logInfof("application startup, version=%s, dataDir=%s", Version, dataDir)
	a.storage = skill.NewStorage(cfg.SkillsStorageDir)
	a.cacheDir = filepath.Join(dataDir, "cache")
	a.starStorage = coregit.NewStarStorage(filepath.Join(dataDir, "star_repos.json"))
	registerAdapters()
	registerProviders()
	go forwardEvents(ctx, a.hub)
	a.logDebugf("startup background tasks deferred until ui ready")
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

func (a *App) domReady(ctx context.Context) {
	if err := setupTray(a); err != nil {
		runtime.LogWarningf(ctx, "tray init failed: %v", err)
	}
	a.startBackgroundStartupTasks()
}

func (a *App) startBackgroundStartupTasks() {
	a.startupOnce.Do(func() {
		a.logDebugf("startup background tasks scheduled, delay=750ms")
		time.AfterFunc(750*time.Millisecond, func() {
			a.logDebugf("startup background tasks started")
			go a.checkUpdatesOnStartup()
			go a.updateStarredReposOnStartup()
			go a.checkAppUpdateOnStartup()
			go a.gitPullOnStartup()
		})
	})
}

func (a *App) beforeClose(_ context.Context) bool {
	a.logInfof("application quit started")
	return false
}

func (a *App) shutdown(_ context.Context) {
	a.logInfof("application shutdown completed")
	teardownTray()
}

func (a *App) showMainWindow() {
	a.logInfof("main window show started")
	if err := showMainWindowNative(a.ctx); err != nil {
		a.logErrorf("main window show failed: %v", err)
		return
	}
	a.logInfof("main window show completed")
}

func (a *App) hideMainWindow() {
	if goruntime.GOOS != "darwin" {
		a.logInfof("main window hide started")
	}
	if err := hideMainWindowNative(a.ctx); err != nil {
		a.logErrorf("main window hide failed: %v", err)
		return
	}
	if goruntime.GOOS != "darwin" {
		a.logInfof("main window hide completed")
	}
}

func (a *App) quitApp() {
	runtime.Quit(a.ctx)
}

// autoBackup triggers cloud backup after any mutating operation if cloud is enabled.
func (a *App) autoBackup() {
	a.logDebugf("auto backup triggered")
	_ = a.runBackup()
}

func (a *App) runBackup() error {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider == "" {
		if err != nil {
			a.logErrorf("backup aborted: load config failed: %v", err)
		}
		return err
	}
	a.logInfof("backup started (provider=%s)", cfg.Cloud.Provider)
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return err
	}
	isGit := cfg.Cloud.Provider == backup.GitProviderName
	backupDir := a.backupRootDir(cfg)
	if isGit {
		backupDir, err = a.prepareGitBackupRoot(cfg)
		if err != nil {
			return err
		}
	}
	if isGit {
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	}
	a.hub.Publish(notify.Event{Type: notify.EventBackupStarted})
	err = provider.Sync(a.ctx, backupDir, cfg.Cloud.BucketName, cfg.Cloud.RemotePath,
		func(file string) {
			a.hub.Publish(notify.Event{
				Type:    notify.EventBackupProgress,
				Payload: notify.BackupProgressPayload{CurrentFile: file},
			})
		})
	if err != nil {
		a.logErrorf("backup failed: %v", err)
		var conflictErr *backup.GitConflictError
		if isGit && errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
		}
		if isGit {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		a.hub.Publish(notify.Event{Type: notify.EventBackupFailed, Payload: err.Error()})
		return err
	} else {
		a.logInfof("backup completed")
		if isGit {
			a.clearGitConflictPending()
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted})
		}
		a.hub.Publish(notify.Event{Type: notify.EventBackupCompleted})
		return nil
	}
}

func (a *App) publishGitConflict(conflictErr *backup.GitConflictError) {
	a.gitConflictMu.Lock()
	a.gitConflictPending = true
	a.gitConflictMu.Unlock()
	a.hub.Publish(notify.Event{
		Type: notify.EventGitConflict,
		Payload: notify.GitConflictPayload{
			Message: conflictErr.Output,
			Files:   conflictErr.Files,
		},
	})
}

func (a *App) clearGitConflictPending() {
	a.gitConflictMu.Lock()
	a.gitConflictPending = false
	a.gitConflictMu.Unlock()
}

func (a *App) reloadStateFromDisk() {
	cfg, err := a.config.Load()
	if err != nil {
		return
	}
	a.storage = skill.NewStorage(cfg.SkillsStorageDir)
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
}

// gitPullOnStartup pulls from the remote git repo at startup when the git provider is enabled.
func (a *App) gitPullOnStartup() {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider != backup.GitProviderName {
		return
	}
	a.logInfof("startup git pull started")
	p, ok := registry.GetCloudProvider(backup.GitProviderName)
	if !ok {
		return
	}
	if err := p.Init(cfg.Cloud.Credentials); err != nil {
		return
	}
	gitP := p.(*backup.GitProvider)
	backupDir, prepErr := a.prepareGitBackupRoot(cfg)
	if prepErr != nil {
		a.logErrorf("startup git pull failed: prepare git backup root failed: %v", prepErr)
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: prepErr.Error()})
		return
	}
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	if err := gitP.Restore(a.ctx, "", "", backupDir); err != nil {
		a.logErrorf("startup git pull failed: %v", err)
		var conflictErr *backup.GitConflictError
		if errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		} else {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return
	}
	a.logInfof("startup git pull completed")
	a.clearGitConflictPending()
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted})
	a.reloadStateFromDisk()
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
	backupDir := a.backupRootDir(cfg)
	if useLocal {
		err = gitP.ResolveConflictUseLocal(backupDir)
	} else {
		err = gitP.ResolveConflictUseRemote(backupDir)
	}
	if err != nil {
		return err
	}
	a.gitConflictMu.Lock()
	a.gitConflictPending = false
	a.gitConflictMu.Unlock()
	a.reloadStateFromDisk()
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted})
	return nil
}

func (a *App) backupRootDir(cfg config.AppConfig) string {
	appDataDir := filepath.Clean(config.AppDataDir())
	skillsDir := filepath.Clean(cfg.SkillsStorageDir)

	// Prefer app data dir so cloud backup includes config/meta/skills.
	if rel, err := filepath.Rel(appDataDir, skillsDir); err == nil &&
		rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return appDataDir
	}
	// If skills dir is outside app data dir, use its parent as the git root.
	return filepath.Dir(skillsDir)
}

func (a *App) prepareGitBackupRoot(cfg config.AppConfig) (string, error) {
	backupDir := a.backupRootDir(cfg)
	migratedTo, migrated, err := backup.MigrateLegacyNestedGitDir(cfg.SkillsStorageDir, backupDir)
	if err != nil {
		a.logErrorf("git backup root preparation failed: skillsDir=%s, backupDir=%s, err=%v", cfg.SkillsStorageDir, backupDir, err)
		return "", err
	}
	if migrated {
		a.logInfof("git backup root preparation completed: moved legacy nested git dir from skills storage to %s", migratedTo)
	}
	return backupDir, nil
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
	for _, sk := range skills {
		if sk.Category == "" {
			sk.Category = defaultCategoryName
		}
	}
	return skills, nil
}

func (a *App) ListCategories() ([]string, error) {
	cats, err := a.storage.ListCategories()
	if err != nil {
		return nil, err
	}
	// 检查默认分类是否已在列表中
	hasDefault := false
	for _, c := range cats {
		if normalizeCategoryName(c) == defaultCategoryName {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		// 将默认分类加到列表最前面
		cats = append([]string{defaultCategoryName}, cats...)
	}
	return cats, nil
}

func (a *App) CreateCategory(name string) error {
	return a.storage.CreateCategory(name)
}

func (a *App) RenameCategory(oldName, newName string) error {
	if normalizeCategoryName(oldName) == defaultCategoryName {
		return fmt.Errorf("默认分类不可重命名")
	}
	if normalizeCategoryName(newName) == defaultCategoryName {
		return fmt.Errorf("不能重命名为默认分类")
	}
	return a.storage.RenameCategory(strings.TrimSpace(oldName), strings.TrimSpace(newName))
}

func (a *App) DeleteCategory(name string) error {
	name = strings.TrimSpace(name)
	a.logInfof("delete category started: category=%s", name)
	if normalizeCategoryName(name) == defaultCategoryName {
		err := fmt.Errorf("默认分类不可删除")
		a.logErrorf("delete category failed: category=%s, err=%v", name, err)
		return err
	}
	if err := a.storage.DeleteCategory(name); err != nil {
		if errors.Is(err, skill.ErrCategoryNotEmpty) {
			wrapped := fmt.Errorf("分类下仍有 Skill，请先清空后再删除")
			a.logErrorf("delete category failed: category=%s, err=%v", name, wrapped)
			return wrapped
		}
		a.logErrorf("delete category failed: category=%s, err=%v", name, err)
		return err
	}
	a.logInfof("delete category completed: category=%s", name)
	return nil
}

func (a *App) MoveSkillCategory(skillID, category string) error {
	category = normalizeCategoryName(category)
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
	starSkills, err := coregit.ScanSkillsWithMaxDepth(cacheDir, repoURL, repoName, repoSource, a.repoScanMaxDepth())
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
	category = normalizeCategoryName(category)
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
	category = normalizeCategoryName(category)
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
			return scanToolSkills(a.ctx, getAdapter(t), t.ScanDirs, a.repoScanMaxDepth())
		}
	}
	return nil, nil
}

// ToolSkillEntry describes a skill found in a tool's configured directories.
type ToolSkillEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	InPush bool   `json:"inPush"`
	InScan bool   `json:"inScan"`
}

// ListToolSkills returns all skills for a tool, annotated with whether each
// skill lives in the push directory and/or the scan directories.
func (a *App) ListToolSkills(toolName string) ([]ToolSkillEntry, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	var tc *config.ToolConfig
	for i := range cfg.Tools {
		if cfg.Tools[i].Name == toolName {
			tc = &cfg.Tools[i]
			break
		}
	}
	if tc == nil {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}
	adapter := getAdapter(*tc)

	type entryState struct {
		path   string
		inPush bool
		inScan bool
	}
	byName := map[string]*entryState{}

	if tc.PushDir != "" {
		if _, statErr := os.Stat(tc.PushDir); statErr == nil {
			if pushSkills, pullErr := pullToolSkills(a.ctx, adapter, tc.PushDir, a.repoScanMaxDepth()); pullErr == nil {
				for _, sk := range pushSkills {
					byName[sk.Name] = &entryState{path: sk.Path, inPush: true}
				}
			}
		}
	}

	if scanSkills, _ := scanToolSkills(a.ctx, adapter, tc.ScanDirs, a.repoScanMaxDepth()); scanSkills != nil {
		for _, sk := range scanSkills {
			if e, ok := byName[sk.Name]; ok {
				e.inScan = true
			} else {
				byName[sk.Name] = &entryState{path: sk.Path, inScan: true}
			}
		}
	}

	result := make([]ToolSkillEntry, 0, len(byName))
	for name, e := range byName {
		result = append(result, ToolSkillEntry{Name: name, Path: e.path, InPush: e.inPush, InScan: e.inScan})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// DeleteToolSkill removes a skill directory from a tool's push directory.
// Returns an error if skillPath is not within the tool's configured push directory.
func (a *App) DeleteToolSkill(toolName string, skillPath string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	var tc *config.ToolConfig
	for i := range cfg.Tools {
		if cfg.Tools[i].Name == toolName {
			tc = &cfg.Tools[i]
			break
		}
	}
	if tc == nil {
		return fmt.Errorf("tool %s not found", toolName)
	}
	if tc.PushDir == "" {
		return fmt.Errorf("工具 %s 未配置推送路径", toolName)
	}
	rel, relErr := filepath.Rel(tc.PushDir, skillPath)
	if relErr != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("无法删除不在推送路径下的 Skill")
	}
	a.logInfof("DeleteToolSkill: deleting %s from tool %s push dir started", filepath.Base(skillPath), toolName)
	if err := os.RemoveAll(skillPath); err != nil {
		a.logErrorf("DeleteToolSkill: delete %s failed: %v", skillPath, err)
		return fmt.Errorf("删除失败: %w", err)
	}
	a.logInfof("DeleteToolSkill: deleted %s from tool %s push dir completed", filepath.Base(skillPath), toolName)
	return nil
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
	category = normalizeCategoryName(category)
	cfg, _ := a.config.Load()
	nameSet := map[string]bool{}
	for _, n := range skillNames {
		nameSet[n] = true
	}
	for _, t := range cfg.Tools {
		if t.Name != toolName {
			continue
		}
		candidates, err := scanToolSkills(a.ctx, getAdapter(t), t.ScanDirs, a.repoScanMaxDepth())
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
	category = normalizeCategoryName(category)
	cfg, _ := a.config.Load()
	nameSet := map[string]bool{}
	for _, n := range skillNames {
		nameSet[n] = true
	}
	for _, t := range cfg.Tools {
		if t.Name != toolName {
			continue
		}
		candidates, err := scanToolSkills(a.ctx, getAdapter(t), t.ScanDirs, a.repoScanMaxDepth())
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

func pullToolSkills(ctx context.Context, adapter toolsync.ToolAdapter, dir string, maxDepth int) ([]*skill.Skill, error) {
	if depthAware, ok := adapter.(maxDepthPuller); ok {
		return depthAware.PullWithMaxDepth(ctx, dir, maxDepth)
	}
	return adapter.Pull(ctx, dir)
}

func scanToolSkills(ctx context.Context, adapter toolsync.ToolAdapter, scanDirs []string, maxDepth int) ([]*skill.Skill, error) {
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
		skills, err := pullToolSkills(ctx, adapter, dir, maxDepth)
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

func (a *App) repoScanMaxDepth() int {
	cfg, err := a.config.Load()
	if err != nil {
		return config.DefaultRepoScanMaxDepth
	}
	return config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
}

// --- Config ---

func (a *App) GetConfig() (config.AppConfig, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return cfg, err
	}
	cfg.DefaultCategory = defaultCategoryName
	cfg.LogLevel = config.NormalizeLogLevel(cfg.LogLevel)
	cfg.RepoScanMaxDepth = config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	return cfg, nil
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
	a.logInfof("save config requested")
	cfg.DefaultCategory = defaultCategoryName
	cfg.LogLevel = config.NormalizeLogLevel(cfg.LogLevel)
	cfg.RepoScanMaxDepth = config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("save config failed: %v", err)
		return err
	}
	a.setLoggerLevel(cfg.LogLevel)
	a.logInfof("save config completed: logLevel=%s repoScanMaxDepth=%d", cfg.LogLevel, cfg.RepoScanMaxDepth)
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
	return nil
}

func (a *App) AddCustomTool(name, pushDir string) error {
	a.logInfof("add custom tool requested: name=%s", name)
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("add custom tool failed: %v", err)
		return err
	}
	cfg.Tools = append(cfg.Tools, config.ToolConfig{
		Name:     name,
		ScanDirs: []string{pushDir},
		PushDir:  pushDir,
		Enabled:  true,
		Custom:   true,
	})
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("add custom tool failed: %v", err)
		return err
	}
	a.logInfof("add custom tool done: name=%s", name)
	return nil
}

func (a *App) RemoveCustomTool(name string) error {
	a.logInfof("remove custom tool requested: name=%s", name)
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("remove custom tool failed: %v", err)
		return err
	}
	var filtered []config.ToolConfig
	for _, t := range cfg.Tools {
		if !(t.Custom && t.Name == name) {
			filtered = append(filtered, t)
		}
	}
	cfg.Tools = filtered
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("remove custom tool failed: %v", err)
		return err
	}
	a.logInfof("remove custom tool done: name=%s", name)
	return nil
}

// --- Backup ---

func (a *App) BackupNow() error {
	a.logInfof("manual backup requested")
	return a.runBackup()
}

func (a *App) ListCloudFiles() ([]backup.RemoteFile, error) {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("list cloud files failed: load config failed: %v", err)
		return nil, err
	}
	a.logInfof("list cloud files started (provider=%s, remotePath=%s)", cfg.Cloud.Provider, cfg.Cloud.RemotePath)
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		err := fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
		a.logErrorf("list cloud files failed: %v", err)
		return nil, err
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		a.logErrorf("list cloud files failed: init provider %s failed: %v", cfg.Cloud.Provider, err)
		return nil, err
	}
	files, err := provider.List(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath)
	if err != nil {
		a.logErrorf("list cloud files failed: provider=%s, remotePath=%s, err=%v", cfg.Cloud.Provider, cfg.Cloud.RemotePath, err)
		return nil, err
	}
	a.logInfof("list cloud files completed (provider=%s, remotePath=%s, count=%d)", cfg.Cloud.Provider, cfg.Cloud.RemotePath, len(files))
	return files, nil
}

func (a *App) RestoreFromCloud() error {
	a.logInfof("restore from cloud requested")
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("restore from cloud failed: %v", err)
		return err
	}
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return err
	}
	isGit := cfg.Cloud.Provider == backup.GitProviderName
	restoreDir := a.backupRootDir(cfg)
	if isGit {
		restoreDir, err = a.prepareGitBackupRoot(cfg)
		if err != nil {
			a.logErrorf("restore from cloud failed: prepare git backup root failed: %v", err)
			return err
		}
	}
	if isGit {
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	}
	if err := provider.Restore(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath, restoreDir); err != nil {
		a.logErrorf("restore from cloud failed: %v", err)
		var conflictErr *backup.GitConflictError
		if isGit && errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
		}
		if isGit {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return err
	}
	a.logInfof("restore from cloud completed")
	a.reloadStateFromDisk()
	if isGit {
		a.clearGitConflictPending()
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted})
	}
	return nil
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
	a.logDebugf("check skill updates started")
	skills, err := a.storage.ListAll()
	if err != nil {
		a.logErrorf("check skill updates failed: %v", err)
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
	a.logDebugf("check skill updates completed")
	return nil
}

// UpdateSkill re-downloads a GitHub skill and updates local files and SHA.
func (a *App) UpdateSkill(skillID string) error {
	a.logInfof("update skill requested: id=%s", skillID)
	sk, err := a.storage.Get(skillID)
	if err != nil {
		a.logErrorf("update skill failed: %v", err)
		return err
	}
	inst := install.NewGitHubInstaller("", a.proxyHTTPClient())
	tmpDir := filepath.Join(os.TempDir(), "skillflow-update", sk.Name)
	defer os.RemoveAll(tmpDir)

	c := install.SkillCandidate{Name: sk.Name, Path: sk.SourceSubPath}
	if err := inst.DownloadTo(a.ctx, install.InstallSource{Type: "github", URI: sk.SourceURL}, c, tmpDir); err != nil {
		a.logErrorf("update skill download failed: %v", err)
		return err
	}
	if err := a.storage.OverwriteFromDir(skillID, tmpDir); err != nil {
		a.logErrorf("update skill overwrite failed: %v", err)
		return err
	}
	sk.SourceSHA = sk.LatestSHA
	sk.LatestSHA = ""
	_ = a.storage.UpdateMeta(sk)
	go a.autoBackup()
	a.logInfof("update skill completed: id=%s name=%s", skillID, sk.Name)
	return nil
}

func (a *App) checkUpdatesOnStartup() {
	_ = a.CheckUpdates()
}

func (a *App) updateStarredReposOnStartup() {
	_ = a.UpdateAllStarredRepos()
}

// OpenFolderDialog wraps Wails file dialog for frontend use.
func (a *App) OpenFolderDialog(defaultDir string) (string, error) {
	options := runtime.OpenDialogOptions{Title: "选择 Skill 目录"}
	if dir := nearestExistingDirectory(defaultDir); dir != "" {
		options.DefaultDirectory = dir
	}
	return runtime.OpenDirectoryDialog(a.ctx, options)
}

// OpenPath opens the given filesystem path in the OS default file manager.
func (a *App) OpenPath(path string) error {
	target, err := resolveOpenPathTarget(path)
	if err != nil {
		a.logErrorf("open path failed: path=%s, err=%v", path, err)
		return err
	}
	a.logInfof("open path started: requested=%s target=%s", path, target)
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "--", target)
	case "windows":
		cmd = exec.Command("explorer.exe", filepath.Clean(target))
	default:
		cmd = exec.Command("xdg-open", target)
	}
	if err := cmd.Start(); err != nil {
		a.logErrorf("open path failed: requested=%s target=%s err=%v", path, target, err)
		return err
	}
	a.logInfof("open path completed: requested=%s target=%s", path, target)
	return nil
}

// GetLogDir returns the app log directory path.
func (a *App) GetLogDir() string {
	return a.logDir()
}

// OpenLogDir opens the app log directory in the OS file manager.
func (a *App) OpenLogDir() error {
	return a.OpenPath(a.logDir())
}

// Greet is kept for Wails template compatibility.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// --- Starred Repos ---

func (a *App) AddStarredRepo(repoURL string) (*coregit.StarredRepo, error) {
	a.logInfof("add starred repo requested: %s", repoURL)
	if err := coregit.CheckGitInstalled(); err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
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
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	dataDir := filepath.Dir(a.cacheDir)
	localDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
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
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	a.logInfof("add starred repo completed: %s", repoURL)
	return &repos[len(repos)-1], nil
}

// AddStarredRepoWithCredentials clones a repo using the provided HTTP username/password,
// removing any previously failed entry for the same URL first.
func (a *App) AddStarredRepoWithCredentials(repoURL, username, password string) (*coregit.StarredRepo, error) {
	a.logInfof("add starred repo with credentials requested: %s", repoURL)
	if err := coregit.CheckGitInstalled(); err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
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
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	dataDir := filepath.Dir(a.cacheDir)
	localDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	repo := coregit.StarredRepo{URL: repoURL, Name: name, Source: source, LocalDir: localDir}
	if cloneErr := coregit.CloneOrUpdateWithCreds(a.ctx, repoURL, localDir, a.gitProxyURL(), username, password); cloneErr != nil {
		a.logErrorf("add starred repo with credentials failed: %v", cloneErr)
		return nil, cloneErr
	}
	repo.LastSync = time.Now()
	filtered = append(filtered, repo)
	if err := a.starStorage.Save(filtered); err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	a.logInfof("add starred repo with credentials completed: %s", repoURL)
	return &filtered[len(filtered)-1], nil
}

func (a *App) RemoveStarredRepo(repoURL string) error {
	a.logInfof("remove starred repo requested: %s", repoURL)
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("remove starred repo failed: %v", err)
		return err
	}
	filtered := make([]coregit.StarredRepo, 0, len(repos))
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			filtered = append(filtered, r)
		}
	}
	if err := a.starStorage.Save(filtered); err != nil {
		a.logErrorf("remove starred repo failed: %v", err)
		return err
	}
	a.logInfof("remove starred repo completed: %s", repoURL)
	return nil
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
	maxDepth := a.repoScanMaxDepth()
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
		skills, _ := coregit.ScanSkillsWithMaxDepth(r.LocalDir, r.URL, r.Name, source, maxDepth)
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
	maxDepth := a.repoScanMaxDepth()
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
		skills, err := coregit.ScanSkillsWithMaxDepth(r.LocalDir, r.URL, r.Name, source, maxDepth)
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
	a.logInfof("update all starred repos requested")
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("update all starred repos failed: %v", err)
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
	if err := a.starStorage.Save(repos); err != nil {
		a.logErrorf("update all starred repos failed: %v", err)
		return err
	}
	a.logInfof("update all starred repos completed")
	return nil
}

func (a *App) ImportStarSkills(skillPaths []string, repoURL, category string) error {
	category = normalizeCategoryName(category)
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

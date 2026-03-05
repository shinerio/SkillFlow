package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/install"
	"github.com/shinerio/skillflow/core/notify"
	"github.com/shinerio/skillflow/core/registry"
	"github.com/shinerio/skillflow/core/skill"
	toolsync "github.com/shinerio/skillflow/core/sync"
	"github.com/shinerio/skillflow/core/update"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	hub     *notify.Hub
	storage *skill.Storage
	config  *config.Service
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
	registerAdapters()
	registerProviders()
	go forwardEvents(ctx, a.hub)
	go a.checkUpdatesOnStartup()
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

func (a *App) domReady(_ context.Context)          {}
func (a *App) beforeClose(_ context.Context) bool  { return false }
func (a *App) shutdown(_ context.Context)          {}

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

// --- Install ---

// ScanGitHub scans a GitHub repo for valid skills, marking already-installed ones.
func (a *App) ScanGitHub(repoURL string) ([]install.SkillCandidate, error) {
	inst := install.NewGitHubInstaller("", a.proxyHTTPClient())
	candidates, err := inst.Scan(a.ctx, install.InstallSource{Type: "github", URI: repoURL})
	if err != nil {
		return nil, err
	}
	existing, _ := a.storage.ListAll()
	existingNames := map[string]bool{}
	for _, sk := range existing {
		existingNames[sk.Name] = true
	}
	for i := range candidates {
		candidates[i].Installed = existingNames[candidates[i].Name]
	}
	return candidates, nil
}

// InstallFromGitHub downloads selected skills from GitHub and imports them into storage.
func (a *App) InstallFromGitHub(repoURL string, candidates []install.SkillCandidate, category string) error {
	if category == "" {
		cfg, _ := a.config.Load()
		category = cfg.DefaultCategory
		if category == "" {
			category = "Imported"
		}
	}
	inst := install.NewGitHubInstaller("", a.proxyHTTPClient())
	source := install.InstallSource{Type: "github", URI: repoURL}

	for _, c := range candidates {
		tmpDir := filepath.Join(os.TempDir(), "skillflow-install", c.Name)
		defer os.RemoveAll(tmpDir)

		if err := inst.DownloadTo(a.ctx, source, c, tmpDir); err != nil {
			return fmt.Errorf("download %s: %w", c.Name, err)
		}
		sha, _ := inst.GetLatestSHA(a.ctx, repoURL, c.Path)

		_, err := a.storage.Import(tmpDir, category, skill.SourceGitHub, repoURL, c.Path)
		if err != nil {
			return err
		}
		skills, _ := a.storage.ListAll()
		for _, sk := range skills {
			if sk.Name == c.Name && sk.SourceURL == repoURL {
				sk.SourceSHA = sha
				_ = a.storage.UpdateMeta(sk)
				break
			}
		}
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

// ScanToolSkills lists all skills in a tool's directory for the pull page.
func (a *App) ScanToolSkills(toolName string) ([]*skill.Skill, error) {
	cfg, _ := a.config.Load()
	for _, t := range cfg.Tools {
		if t.Name == toolName {
			return getAdapter(t).Pull(a.ctx, t.SkillsDir)
		}
	}
	return nil, nil
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
			adapter := getAdapter(t)
			for _, sk := range selected {
				dst := filepath.Join(t.SkillsDir, sk.Name)
				if _, err := os.Stat(dst); err == nil {
					conflicts = append(conflicts, fmt.Sprintf("%s -> %s", sk.Name, toolName))
					continue
				}
			}
			_ = adapter.Push(a.ctx, selected, t.SkillsDir)
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
				_ = getAdapter(t).Push(a.ctx, selected, t.SkillsDir)
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
		candidates, err := getAdapter(t).Pull(a.ctx, t.SkillsDir)
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
		candidates, _ := getAdapter(t).Pull(a.ctx, t.SkillsDir)
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

func getAdapter(t config.ToolConfig) toolsync.ToolAdapter {
	if a, ok := registry.GetAdapter(t.Name); ok {
		return a
	}
	return toolsync.NewFilesystemAdapter(t.Name, t.SkillsDir)
}

// --- Config ---

func (a *App) GetConfig() (config.AppConfig, error) {
	return a.config.Load()
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
	return a.config.Save(cfg)
}

func (a *App) AddCustomTool(name, skillsDir string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	cfg.Tools = append(cfg.Tools, config.ToolConfig{
		Name:      name,
		SkillsDir: skillsDir,
		Enabled:   true,
		Custom:    true,
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

// --- Favorite Repos ---

// ListFavoriteRepos returns all saved GitHub repository favorites.
func (a *App) ListFavoriteRepos() ([]config.FavoriteRepo, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	if cfg.FavoriteRepos == nil {
		return []config.FavoriteRepo{}, nil
	}
	return cfg.FavoriteRepos, nil
}

// AddFavoriteRepo adds a GitHub repo URL to the favorites list.
func (a *App) AddFavoriteRepo(repoURL, description string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	for _, r := range cfg.FavoriteRepos {
		if r.URL == repoURL {
			return fmt.Errorf("该仓库已在收藏列表中")
		}
	}
	// Derive display name (owner/repo) from URL.
	name := repoURL
	uri := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(repoURL, "/"), ".git"), "/")
	parts := strings.Split(uri, "/")
	if len(parts) >= 2 {
		name = parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	cfg.FavoriteRepos = append(cfg.FavoriteRepos, config.FavoriteRepo{
		URL:         repoURL,
		Name:        name,
		Description: description,
		AddedAt:     time.Now(),
	})
	return a.config.Save(cfg)
}

// RemoveFavoriteRepo removes a GitHub repo URL from the favorites list.
func (a *App) RemoveFavoriteRepo(repoURL string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	var filtered []config.FavoriteRepo
	for _, r := range cfg.FavoriteRepos {
		if r.URL != repoURL {
			filtered = append(filtered, r)
		}
	}
	cfg.FavoriteRepos = filtered
	return a.config.Save(cfg)
}

// InstallFromGitHubToTool installs selected skills directly into a tool's skills directory,
// bypassing SkillFlow storage (no metadata tracking or update checks).
func (a *App) InstallFromGitHubToTool(repoURL string, candidates []install.SkillCandidate, toolName string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	var targetDir string
	for _, t := range cfg.Tools {
		if t.Name == toolName && t.Enabled {
			targetDir = t.SkillsDir
			break
		}
	}
	if targetDir == "" {
		return fmt.Errorf("工具 %q 未找到或未启用", toolName)
	}
	inst := install.NewGitHubInstaller("", a.proxyHTTPClient())
	source := install.InstallSource{Type: "github", URI: repoURL}
	for _, c := range candidates {
		skillDir := filepath.Join(targetDir, c.Name)
		if err := inst.DownloadTo(a.ctx, source, c, skillDir); err != nil {
			return fmt.Errorf("下载 %s 失败: %w", c.Name, err)
		}
	}
	return nil
}

// Greet is kept for Wails template compatibility.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

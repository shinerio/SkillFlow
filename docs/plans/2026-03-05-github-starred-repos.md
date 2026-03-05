# GitHub Starred Repos Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the GitHub API-based skill installer with git clone/pull, and add a "GitHub 收藏" page for managing starred repos with auto-sync, dual-view browsing, and batch import.

**Architecture:** New `core/git` package wraps git subprocess calls (clone/update/scan/storage). `App.ScanGitHub` and `App.InstallFromGitHub` are refactored to use local cache. New App methods serve the starred repos frontend. A new `StarredRepos.tsx` page provides folder and flat views with batch import.

**Tech Stack:** Go `os/exec` (git), `core/git` package, React/TypeScript, Wails v2 bindings, lucide-react icons.

---

### Task 1: core/git — model and StarStorage

**Files:**
- Create: `core/git/model.go`
- Create: `core/git/storage.go`
- Create: `core/git/storage_test.go`

**Step 1: Create model.go**

```go
package git

import "time"

type StarredRepo struct {
	URL       string    `json:"url"`
	Name      string    `json:"name"`     // "owner/repo"
	LocalDir  string    `json:"localDir"` // absolute path under cache/
	LastSync  time.Time `json:"lastSync"`
	SyncError string    `json:"syncError,omitempty"`
}

type StarSkill struct {
	Name     string `json:"name"`
	Path     string `json:"path"`     // absolute local path to skill directory
	SubPath  string `json:"subPath"`  // relative path within repo, e.g. "skills/my-skill"
	RepoURL  string `json:"repoUrl"`
	RepoName string `json:"repoName"` // "owner/repo"
	Imported bool   `json:"imported"` // already exists in My Skills
}
```

**Step 2: Create storage.go**

```go
package git

import (
	"encoding/json"
	"os"
	"sync"
)

type StarStorage struct {
	path string
	mu   sync.Mutex
}

func NewStarStorage(path string) *StarStorage {
	return &StarStorage{path: path}
}

func (s *StarStorage) Load() ([]StarredRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []StarredRepo
	return repos, json.Unmarshal(data, &repos)
}

func (s *StarStorage) Save(repos []StarredRepo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
```

**Step 3: Write storage_test.go**

```go
package git

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStarStorageLoadEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorage(path)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("expected nil, got %v", repos)
	}
}

func TestStarStorageSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorage(path)
	want := []StarredRepo{
		{URL: "https://github.com/a/b", Name: "a/b", LocalDir: "/tmp/a/b", LastSync: time.Time{}},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].URL != want[0].URL || got[0].Name != want[0].Name {
		t.Fatalf("mismatch: %+v", got)
	}
}
```

**Step 4: Run tests**

```bash
go test ./core/git/... -v
```
Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add core/git/
git commit -m "feat: add core/git model and StarStorage"
```

---

### Task 2: core/git — git client

**Files:**
- Create: `core/git/client.go`
- Create: `core/git/client_test.go`

**Step 1: Create client.go**

```go
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckGitInstalled returns nil if git is in PATH, or a user-friendly error.
func CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git 未安装，请先安装 git（https://git-scm.com）再使用此功能")
	}
	return nil
}

// ParseRepoName extracts "owner/repo" from a GitHub URL.
func ParseRepoName(repoURL string) (string, error) {
	u := strings.TrimSuffix(strings.TrimSuffix(repoURL, "/"), ".git")
	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效的 GitHub URL: %s", repoURL)
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1], nil
}

// CacheDir returns the local clone directory for a repo URL under dataDir/cache/.
func CacheDir(dataDir, repoURL string) (string, error) {
	name, err := ParseRepoName(repoURL)
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "cache", filepath.FromSlash(name)), nil
}

// CloneOrUpdate clones repoURL into dir, or force-updates it if already present.
// proxyURL is optional (empty = inherit environment).
func CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error {
	if err := CheckGitInstalled(); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		// already cloned — force-pull (handles force-push on remote)
		if err := runGit(ctx, dir, proxyURL, "fetch", "origin"); err != nil {
			return fmt.Errorf("git fetch: %w", err)
		}
		return runGit(ctx, dir, proxyURL, "reset", "--hard", "origin/HEAD")
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return err
	}
	return runGit(ctx, "", proxyURL, "clone", repoURL, dir)
}

// GetSubPathSHA returns the latest commit SHA for a path within a local git repo.
func GetSubPathSHA(ctx context.Context, repoDir, subPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-n", "1", "--format=%H", "--", subPath)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func runGit(ctx context.Context, dir, proxyURL string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if proxyURL != "" {
		cmd.Env = append(os.Environ(),
			"HTTP_PROXY="+proxyURL,
			"HTTPS_PROXY="+proxyURL,
			"http_proxy="+proxyURL,
			"https_proxy="+proxyURL,
		)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
```

**Step 2: Create client_test.go**

```go
package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// makeLocalRepo creates a bare-minimum local git repo in dir with one commit.
func makeLocalRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		if err := runGit(context.Background(), dir, "", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		if err := runGit(context.Background(), dir, "", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
}

func TestParseRepoName(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://github.com/owner/repo", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo/", "owner/repo"},
	}
	for _, c := range cases {
		got, err := ParseRepoName(c.url)
		if err != nil || got != c.want {
			t.Errorf("ParseRepoName(%q) = %q, %v; want %q", c.url, got, err, c.want)
		}
	}
}

func TestCloneOrUpdate(t *testing.T) {
	if err := CheckGitInstalled(); err != nil {
		t.Skip("git not installed")
	}
	src := t.TempDir()
	makeLocalRepo(t, src)

	dst := filepath.Join(t.TempDir(), "clone")

	// First call: clone
	if err := CloneOrUpdate(context.Background(), src, dst, ""); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Fatal("README.md missing after clone")
	}

	// Add a new file to source
	if err := os.WriteFile(filepath.Join(src, "NEW.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "add file"}} {
		if err := runGit(context.Background(), src, "", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}

	// Second call: update
	if err := CloneOrUpdate(context.Background(), src, dst, ""); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "NEW.md")); err != nil {
		t.Fatal("NEW.md missing after update")
	}
}
```

**Step 3: Run tests**

```bash
go test ./core/git/... -v -run TestParseRepoName
go test ./core/git/... -v -run TestCloneOrUpdate
```
Expected: PASS

**Step 4: Commit**

```bash
git add core/git/
git commit -m "feat: add core/git client (CheckGitInstalled, CloneOrUpdate)"
```

---

### Task 3: core/git — skill scanner

**Files:**
- Create: `core/git/scanner.go`
- Create: `core/git/scanner_test.go`

**Step 1: Create scanner.go**

Repos follow the convention: `<repo>/skills/<skill-name>/SKILLS.md`

```go
package git

import (
	"os"
	"path/filepath"
)

// ScanSkills walks <repoDir>/skills/ and returns entries that contain a SKILLS.md file.
func ScanSkills(repoDir, repoURL, repoName string) ([]StarSkill, error) {
	skillsRoot := filepath.Join(repoDir, "skills")
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result []StarSkill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillDir := filepath.Join(skillsRoot, e.Name())
		if _, err := os.Stat(filepath.Join(skillDir, "SKILLS.md")); err == nil {
			result = append(result, StarSkill{
				Name:     e.Name(),
				Path:     skillDir,
				SubPath:  "skills/" + e.Name(),
				RepoURL:  repoURL,
				RepoName: repoName,
			})
		}
	}
	return result, nil
}
```

**Step 2: Create scanner_test.go**

```go
package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSkillsEmpty(t *testing.T) {
	dir := t.TempDir()
	skills, err := ScanSkills(dir, "https://github.com/a/b", "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(skills))
	}
}

func TestScanSkills(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	// create two valid skills and one invalid (no SKILLS.md)
	for _, name := range []string{"alpha", "beta"} {
		d := filepath.Join(skillsDir, name)
		os.MkdirAll(d, 0755)
		os.WriteFile(filepath.Join(d, "SKILLS.md"), []byte("# "+name), 0644)
	}
	os.MkdirAll(filepath.Join(skillsDir, "no-skills-md"), 0755)

	skills, err := ScanSkills(dir, "https://github.com/a/b", "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(skills), skills)
	}
	for _, sk := range skills {
		if sk.RepoURL != "https://github.com/a/b" {
			t.Errorf("RepoURL wrong: %s", sk.RepoURL)
		}
		if sk.SubPath != "skills/"+sk.Name {
			t.Errorf("SubPath wrong: %s", sk.SubPath)
		}
	}
}
```

**Step 3: Run tests**

```bash
go test ./core/git/... -v -run TestScanSkills
```
Expected: PASS

**Step 4: Commit**

```bash
git add core/git/
git commit -m "feat: add core/git scanner (ScanSkills)"
```

---

### Task 4: notify — add star sync events

**Files:**
- Modify: `core/notify/model.go`

**Step 1: Add two new event constants**

After `EventUpdateAvailable`:
```go
EventStarSyncProgress EventType = "star.sync.progress" // one repo finished syncing
EventStarSyncDone     EventType = "star.sync.done"      // all repos finished
```

Add a payload type:
```go
type StarSyncProgressPayload struct {
	RepoURL   string `json:"repoUrl"`
	RepoName  string `json:"repoName"`
	SyncError string `json:"syncError,omitempty"`
}
```

**Step 2: Verify build**

```bash
go build ./...
```

**Step 3: Commit**

```bash
git add core/notify/model.go
git commit -m "feat: add star sync events to notify package"
```

---

### Task 5: Refactor App.ScanGitHub and App.InstallFromGitHub

**Files:**
- Modify: `app.go`

The two methods are refactored to use local git cache instead of the GitHub Contents API.

`ScanGitHub` now:
1. Derives `cacheDir` from `config.AppDataDir()` + `c.Path`
2. Calls `git.CloneOrUpdate`
3. Calls `git.ScanSkills` on the local dir
4. Marks `Installed` flag

`InstallFromGitHub` now:
1. Derives `cacheDir` for the repo
2. For each candidate, `skillDir = filepath.Join(cacheDir, c.Path)` (local path)
3. Gets SHA via `git.GetSubPathSHA` from the local repo
4. Calls `a.storage.Import(skillDir, category, skill.SourceGitHub, repoURL, c.Path)`

Add a helper `gitProxyURL()` on `App`:

```go
func (a *App) gitProxyURL() string {
	cfg, err := a.config.Load()
	if err != nil {
		return ""
	}
	if cfg.Proxy.Mode == config.ProxyModeManual {
		return cfg.Proxy.URL
	}
	return "" // system or none: git handles env vars naturally
}
```

**Step 1: Add import for core/git in app.go**

```go
coregit "github.com/shinerio/skillflow/core/git"
```

**Step 2: Replace ScanGitHub**

```go
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
	starSkills, err := coregit.ScanSkills(cacheDir, repoURL, repoName)
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
```

**Step 3: Replace InstallFromGitHub**

```go
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
	for _, c := range candidates {
		skillDir := filepath.Join(cacheDir, filepath.FromSlash(c.Path))
		sha, _ := coregit.GetSubPathSHA(a.ctx, cacheDir, c.Path)
		sk, err := a.storage.Import(skillDir, category, skill.SourceGitHub, repoURL, c.Path)
		if err != nil {
			return fmt.Errorf("import %s: %w", c.Name, err)
		}
		sk.SourceSHA = sha
		_ = a.storage.UpdateMeta(sk)
	}
	go a.autoBackup()
	return nil
}
```

**Step 4: Add gitProxyURL helper**

```go
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
```

**Step 5: Build and verify**

```bash
go build ./...
```

**Step 6: Commit**

```bash
git add app.go
git commit -m "refactor: ScanGitHub and InstallFromGitHub use git clone/pull"
```

---

### Task 6: App — starred repo methods

**Files:**
- Modify: `app.go`

Add `starStorage` and `cacheDir` fields to `App`, initialize in `startup()`, then add all starred repo methods.

**Step 1: Add fields to App struct**

```go
type App struct {
	ctx         context.Context
	hub         *notify.Hub
	storage     *skill.Storage
	config      *config.Service
	starStorage *coregit.StarStorage // new
	cacheDir    string               // new
}
```

**Step 2: Initialize in startup()**

After `a.storage = skill.NewStorage(cfg.SkillsStorageDir)`:
```go
dataDir := config.AppDataDir()
a.cacheDir = filepath.Join(dataDir, "cache")
a.starStorage = coregit.NewStarStorage(filepath.Join(dataDir, "star_repos.json"))
```

**Step 3: Add starred repo methods**

```go
func (a *App) AddStarredRepo(repoURL string) (*coregit.StarredRepo, error) {
	if err := coregit.CheckGitInstalled(); err != nil {
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	for _, r := range repos {
		if r.URL == repoURL {
			return &r, nil // already starred
		}
	}
	name, err := coregit.ParseRepoName(repoURL)
	if err != nil {
		return nil, err
	}
	localDir, err := coregit.CacheDir(filepath.Dir(a.cacheDir), repoURL)
	if err != nil {
		return nil, err
	}
	repo := coregit.StarredRepo{URL: repoURL, Name: name, LocalDir: localDir}
	if cloneErr := coregit.CloneOrUpdate(a.ctx, repoURL, localDir, a.gitProxyURL()); cloneErr != nil {
		repo.SyncError = cloneErr.Error()
	} else {
		repo.LastSync = time.Now()
	}
	repos = append(repos, repo)
	if err := a.starStorage.Save(repos); err != nil {
		return nil, err
	}
	return &repo, nil
}

func (a *App) RemoveStarredRepo(repoURL string) error {
	repos, err := a.starStorage.Load()
	if err != nil {
		return err
	}
	filtered := repos[:0]
	for _, r := range repos {
		if r.URL != repoURL {
			filtered = append(filtered, r)
		}
	}
	return a.starStorage.Save(filtered)
}

func (a *App) ListStarredRepos() ([]coregit.StarredRepo, error) {
	return a.starStorage.Load()
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
		skills, _ := coregit.ScanSkills(r.LocalDir, r.URL, r.Name)
		for i := range skills {
			skills[i].Imported = importedNames[skills[i].Name]
		}
		all = append(all, skills...)
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
		if r.URL != repoURL {
			continue
		}
		skills, err := coregit.ScanSkills(r.LocalDir, r.URL, r.Name)
		if err != nil {
			return nil, err
		}
		for i := range skills {
			skills[i].Imported = importedNames[skills[i].Name]
		}
		return skills, nil
	}
	return nil, nil
}

func (a *App) UpdateStarredRepo(repoURL string) error {
	repos, err := a.starStorage.Load()
	if err != nil {
		return err
	}
	for i, r := range repos {
		if r.URL != repoURL {
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
	var wg sync.WaitGroup
	mu := &sync.Mutex{}
	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r := &repos[idx]
			syncErr := coregit.CloneOrUpdate(a.ctx, r.URL, r.LocalDir, a.gitProxyURL())
			mu.Lock()
			if syncErr != nil {
				r.SyncError = syncErr.Error()
			} else {
				r.SyncError = ""
				r.LastSync = time.Now()
			}
			mu.Unlock()
			a.hub.Publish(notify.Event{
				Type: notify.EventStarSyncProgress,
				Payload: notify.StarSyncProgressPayload{
					RepoURL:   r.URL,
					RepoName:  r.Name,
					SyncError: r.SyncError,
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
	for _, r := range repos {
		if r.URL == repoURL {
			repoLocalDir = r.LocalDir
			break
		}
	}
	for _, skillPath := range skillPaths {
		// derive SubPath relative to repoLocalDir
		subPath, _ := filepath.Rel(repoLocalDir, skillPath)
		subPath = filepath.ToSlash(subPath)
		sk, err := a.storage.Import(skillPath, category, skill.SourceGitHub, repoURL, subPath)
		if err == skill.ErrSkillExists {
			continue // skip duplicates silently
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
```

Also add `"sync"` and `"time"` to imports in app.go.

**Step 4: Build**

```bash
go build ./...
```

**Step 5: Commit**

```bash
git add app.go
git commit -m "feat: add starred repo App methods (Add, Remove, List, Update, Import)"
```

---

### Task 7: App — startup auto-sync

**Files:**
- Modify: `app.go`

**Step 1: Add updateStarredReposOnStartup to startup()**

After `go a.checkUpdatesOnStartup()`:
```go
go a.updateStarredReposOnStartup()
```

**Step 2: Add the method**

```go
func (a *App) updateStarredReposOnStartup() {
	_ = a.UpdateAllStarredRepos()
}
```

**Step 3: Build**

```bash
go build ./...
```

**Step 4: Commit**

```bash
git add app.go
git commit -m "feat: auto-sync starred repos on startup"
```

---

### Task 8: Regenerate Wails bindings

**Files:**
- Auto-generated: `frontend/wailsjs/go/main/App.js`
- Auto-generated: `frontend/wailsjs/go/main/App.d.ts`
- Auto-generated: `frontend/wailsjs/go/models.ts`

**Step 1: Regenerate**

```bash
make generate
# or: ~/go/bin/wails generate module
```

**Step 2: Verify new methods appear in App.d.ts**

Check that `AddStarredRepo`, `RemoveStarredRepo`, `ListStarredRepos`, `ListAllStarSkills`, `ListRepoStarSkills`, `UpdateStarredRepo`, `UpdateAllStarredRepos`, `ImportStarSkills` are all present.

**Step 3: Commit**

```bash
git add frontend/wailsjs/
git commit -m "chore: regenerate Wails bindings for starred repo methods"
```

---

### Task 9: Frontend — sidebar entry and StarredRepos page (folder view)

**Files:**
- Modify: `frontend/src/App.tsx`
- Create: `frontend/src/pages/StarredRepos.tsx`

**Step 1: Add import and route to App.tsx**

Add import:
```tsx
import { Star } from 'lucide-react'
import StarredRepos from './pages/StarredRepos'
```

Add NavItem after "从工具拉取":
```tsx
<NavItem to="/starred" icon={<Star size={16} />} label="GitHub 收藏" />
```

Add Route inside `<Routes>`:
```tsx
<Route path="/starred" element={<StarredRepos />} />
<Route path="/starred/:repoEncoded" element={<StarredRepos />} />
```

**Step 2: Create StarredRepos.tsx — folder view skeleton**

```tsx
import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ListStarredRepos, AddStarredRepo, RemoveStarredRepo,
  UpdateStarredRepo, UpdateAllStarredRepos,
  ListAllStarSkills, ListRepoStarSkills,
  ImportStarSkills, ListCategories,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  Star, RefreshCw, Plus, Trash2, LayoutGrid, Folder,
  ChevronLeft, CheckSquare, Download, AlertCircle, X
} from 'lucide-react'
```

**Step 3: Implement folder view**

Full component (folder view renders repo cards, flat view renders skill cards):

```tsx
export default function StarredRepos() {
  const { repoEncoded } = useParams()
  const navigate = useNavigate()
  const currentRepo = repoEncoded ? decodeURIComponent(repoEncoded) : null

  const [repos, setRepos] = useState<any[]>([])
  const [repoSkills, setRepoSkills] = useState<any[]>([])
  const [allSkills, setAllSkills] = useState<any[]>([])
  const [view, setView] = useState<'folder' | 'flat'>('folder')
  const [syncing, setSyncing] = useState(false)
  const [addUrl, setAddUrl] = useState('')
  const [showAdd, setShowAdd] = useState(false)
  const [adding, setAdding] = useState(false)
  const [addError, setAddError] = useState('')
  const [selectMode, setSelectMode] = useState(false)
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [importCategory, setImportCategory] = useState('')
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [importing, setImporting] = useState(false)

  const loadRepos = async () => {
    const r = await ListStarredRepos()
    setRepos(r ?? [])
  }

  const loadAllSkills = async () => {
    const s = await ListAllStarSkills()
    setAllSkills(s ?? [])
  }

  const loadRepoSkills = async (url: string) => {
    const s = await ListRepoStarSkills(url)
    setRepoSkills(s ?? [])
  }

  useEffect(() => {
    loadRepos()
    loadAllSkills()
    ListCategories().then(c => {
      setCategories(c ?? [])
      if (c && c.length > 0) setImportCategory(c[0])
    })
    EventsOn('star.sync.progress', loadRepos)
    EventsOn('star.sync.done', () => { loadRepos(); loadAllSkills(); setSyncing(false) })
  }, [])

  useEffect(() => {
    if (currentRepo) loadRepoSkills(currentRepo)
  }, [currentRepo])

  const handleAddRepo = async () => {
    setAdding(true); setAddError('')
    try {
      await AddStarredRepo(addUrl)
      setShowAdd(false); setAddUrl('')
      await Promise.all([loadRepos(), loadAllSkills()])
    } catch (e: any) {
      setAddError(String(e?.message ?? e ?? '添加失败'))
    } finally { setAdding(false) }
  }

  const handleUpdateAll = async () => {
    setSyncing(true)
    await UpdateAllStarredRepos()
  }

  const handleUpdateOne = async (url: string) => {
    await UpdateStarredRepo(url)
    await Promise.all([loadRepos(), loadAllSkills()])
  }

  const handleRemove = async (url: string) => {
    await RemoveStarredRepo(url)
    await Promise.all([loadRepos(), loadAllSkills()])
  }

  const toggleSelectPath = (path: string) => {
    setSelectedPaths(prev => {
      const next = new Set(prev)
      next.has(path) ? next.delete(path) : next.add(path)
      return next
    })
  }

  const toggleSelectAll = (skills: any[]) => {
    if (selectedPaths.size === skills.length) setSelectedPaths(new Set())
    else setSelectedPaths(new Set(skills.map((s: any) => s.Path)))
  }

  const openImportDialog = () => setShowImportDialog(true)

  const handleBatchImport = async () => {
    setImporting(true)
    try {
      const skills = currentRepo ? repoSkills : allSkills
      const repoURL = currentRepo ?? skills.find((s: any) => selectedPaths.has(s.Path))?.RepoUrl ?? ''
      // group by repoUrl for multi-repo flat import
      const byRepo = new Map<string, string[]>()
      for (const path of selectedPaths) {
        const sk = skills.find((s: any) => s.Path === path)
        if (!sk) continue
        const arr = byRepo.get(sk.RepoUrl) ?? []
        arr.push(path)
        byRepo.set(sk.RepoUrl, arr)
      }
      for (const [rURL, paths] of byRepo) {
        await ImportStarSkills(paths, rURL, importCategory)
      }
      setShowImportDialog(false)
      setSelectMode(false)
      setSelectedPaths(new Set())
      if (currentRepo) loadRepoSkills(currentRepo); else loadAllSkills()
    } finally { setImporting(false) }
  }

  // ── Render ──────────────────────────────────────────────────────────────

  const skills = currentRepo ? repoSkills : allSkills

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-3 px-6 py-4 border-b border-gray-800 flex-wrap">
        {currentRepo ? (
          <button onClick={() => { navigate('/starred'); setSelectMode(false); setSelectedPaths(new Set()) }}
            className="flex items-center gap-1 text-sm text-gray-400 hover:text-white">
            <ChevronLeft size={14} />{decodeURIComponent(currentRepo).split('/').slice(-2).join('/')}
          </button>
        ) : (
          <h2 className="text-sm font-medium flex items-center gap-2"><Star size={14} /> GitHub 收藏</h2>
        )}

        <div className="flex-1" />

        {selectMode ? (
          <>
            <button onClick={() => toggleSelectAll(skills)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">
              <CheckSquare size={14} />{selectedPaths.size === skills.length ? '取消全选' : '全选'}
            </button>
            <button onClick={openImportDialog} disabled={selectedPaths.size === 0}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 rounded-lg">
              <Download size={14} /> 导入 {selectedPaths.size > 0 ? `(${selectedPaths.size})` : ''}
            </button>
            <button onClick={() => { setSelectMode(false); setSelectedPaths(new Set()) }}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">取消</button>
          </>
        ) : (
          <>
            {!currentRepo && (
              <>
                <button onClick={() => setView('folder')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg ${view === 'folder' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:bg-gray-800'}`}>
                  <Folder size={14} /> 文件夹
                </button>
                <button onClick={() => setView('flat')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg ${view === 'flat' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:bg-gray-800'}`}>
                  <LayoutGrid size={14} /> 平铺
                </button>
              </>
            )}
            <button onClick={() => setSelectMode(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">
              <CheckSquare size={14} /> 批量导入
            </button>
            <button onClick={handleUpdateAll} disabled={syncing}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-400 hover:text-white rounded-lg hover:bg-gray-800">
              <RefreshCw size={14} className={syncing ? 'animate-spin' : ''} /> 全部更新
            </button>
            {!currentRepo && (
              <button onClick={() => setShowAdd(true)}
                className="flex items-center gap-1.5 px-4 py-1.5 text-sm bg-indigo-600 hover:bg-indigo-500 rounded-lg">
                <Plus size={14} /> 添加仓库
              </button>
            )}
          </>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {currentRepo ? (
          <SkillGrid skills={repoSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} />
        ) : view === 'folder' ? (
          <RepoGrid repos={repos} onEnter={url => navigate(`/starred/${encodeURIComponent(url)}`)}
            onUpdate={handleUpdateOne} onRemove={handleRemove} />
        ) : (
          <SkillGrid skills={allSkills} selectMode={selectMode} selectedPaths={selectedPaths} onToggle={toggleSelectPath} showRepo />
        )}
      </div>

      {/* Add repo dialog */}
      {showAdd && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[460px] border border-gray-700">
            <div className="flex justify-between items-center mb-4">
              <h3 className="font-semibold flex items-center gap-2"><Star size={16} /> 添加 GitHub 仓库</h3>
              <button onClick={() => { setShowAdd(false); setAddError('') }}><X size={16} className="text-gray-400" /></button>
            </div>
            <div className="flex gap-2 mb-3">
              <input value={addUrl} onChange={e => setAddUrl(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && !adding && addUrl && handleAddRepo()}
                placeholder="https://github.com/user/repo"
                className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500" />
              <button onClick={handleAddRepo} disabled={adding || !addUrl}
                className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50 min-w-[72px]">
                {adding ? '克隆中...' : '添加'}
              </button>
            </div>
            {addError && (
              <div className="flex items-start gap-2 bg-red-950 border border-red-700 text-red-300 rounded-lg px-4 py-3 text-sm">
                <AlertCircle size={15} className="mt-0.5 shrink-0" />
                <span>{addError}</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Import category dialog */}
      {showImportDialog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[380px] border border-gray-700">
            <h3 className="font-semibold mb-4">选择导入分类</h3>
            <select value={importCategory} onChange={e => setImportCategory(e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm mb-4">
              {categories.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
            <div className="flex gap-3">
              <button onClick={handleBatchImport} disabled={importing}
                className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
                {importing ? '导入中...' : `导入 ${selectedPaths.size} 个`}
              </button>
              <button onClick={() => setShowImportDialog(false)}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
```

**Step 4: Add RepoGrid sub-component**

```tsx
function RepoGrid({ repos, onEnter, onUpdate, onRemove }: {
  repos: any[]
  onEnter: (url: string) => void
  onUpdate: (url: string) => void
  onRemove: (url: string) => void
}) {
  if (repos.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48 text-gray-500">
        <Star size={32} className="mb-2 opacity-30" />
        <p className="text-sm">还没有收藏的仓库</p>
        <p className="text-xs mt-1">点击「添加仓库」开始收藏</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
      {repos.map(r => (
        <div key={r.URL} onClick={() => onEnter(r.URL)}
          className="bg-gray-800 rounded-xl p-4 border border-gray-700 hover:border-indigo-500 cursor-pointer transition-colors">
          <div className="flex justify-between items-start mb-2">
            <span className="font-medium text-sm">{r.Name}</span>
            <div className="flex gap-1" onClick={e => e.stopPropagation()}>
              <button onClick={() => onUpdate(r.URL)}
                className="p-1 text-gray-400 hover:text-white rounded"><RefreshCw size={12} /></button>
              <button onClick={() => onRemove(r.URL)}
                className="p-1 text-gray-400 hover:text-red-400 rounded"><Trash2 size={12} /></button>
            </div>
          </div>
          {r.SyncError ? (
            <p className="text-xs text-red-400 truncate">{r.SyncError}</p>
          ) : (
            <p className="text-xs text-gray-500">
              {r.LastSync ? `同步于 ${new Date(r.LastSync).toLocaleDateString()}` : '未同步'}
            </p>
          )}
        </div>
      ))}
    </div>
  )
}
```

**Step 5: Add SkillGrid sub-component**

```tsx
function SkillGrid({ skills, selectMode, selectedPaths, onToggle, showRepo = false }: {
  skills: any[]
  selectMode: boolean
  selectedPaths: Set<string>
  onToggle: (path: string) => void
  showRepo?: boolean
}) {
  if (skills.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48 text-gray-500">
        <p className="text-sm">没有找到 Skills</p>
      </div>
    )
  }
  return (
    <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
      {skills.map((sk: any) => (
        <div key={sk.Path}
          onClick={() => selectMode && onToggle(sk.Path)}
          className={`bg-gray-800 rounded-xl p-4 border transition-colors cursor-pointer
            ${selectMode ? 'hover:border-indigo-400' : ''}
            ${selectedPaths.has(sk.Path) ? 'border-indigo-500 bg-indigo-900/20' : 'border-gray-700'}`}>
          {selectMode && (
            <input type="checkbox" checked={selectedPaths.has(sk.Path)}
              onChange={() => onToggle(sk.Path)}
              className="accent-indigo-500 mb-2" onClick={e => e.stopPropagation()} />
          )}
          <p className="text-sm font-medium truncate">{sk.Name}</p>
          {showRepo && <p className="text-xs text-gray-500 truncate mt-1">{sk.RepoName}</p>}
          {sk.Imported && (
            <span className="text-xs bg-blue-900/50 text-blue-300 px-2 py-0.5 rounded mt-1 inline-block">已导入</span>
          )}
        </div>
      ))}
    </div>
  )
}
```

**Step 6: Verify the app compiles**

```bash
cd frontend && npm run build
```

**Step 7: Commit**

```bash
git add frontend/src/App.tsx frontend/src/pages/StarredRepos.tsx
git commit -m "feat: add StarredRepos page with folder/flat view and batch import"
```

---

### Task 10: Frontend — update GitHubInstallDialog for git clone UX

**Files:**
- Modify: `frontend/src/components/GitHubInstallDialog.tsx`

The scan now triggers a git clone (slow on first run), so the UX needs to reflect this.

**Step 1: Update the scanning loading message**

Find the scanning button label and update it:
```tsx
// Change:
{scanning ? (
  <span className="flex items-center gap-1.5">
    <span className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block" />
    扫描中
  </span>
) : '扫描'}

// To:
{scanning ? (
  <span className="flex items-center gap-1.5">
    <span className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block" />
    克隆/更新中...
  </span>
) : '扫描'}
```

**Step 2: Add a hint below the input**

After the input row (`<div className="flex gap-2 mb-4">`), add:
```tsx
<p className="text-xs text-gray-500 mb-3">首次扫描会 git clone 仓库，后续自动 git pull 更新</p>
```

**Step 3: Build**

```bash
cd frontend && npm run build
```

**Step 4: Commit**

```bash
git add frontend/src/components/GitHubInstallDialog.tsx
git commit -m "feat: update GitHubInstallDialog to reflect git clone UX"
```

---

## Summary of changed files

| File | Change |
|---|---|
| `core/git/model.go` | New — StarredRepo, StarSkill |
| `core/git/storage.go` | New — StarStorage |
| `core/git/storage_test.go` | New |
| `core/git/client.go` | New — CheckGitInstalled, CloneOrUpdate, GetSubPathSHA, ParseRepoName, CacheDir |
| `core/git/client_test.go` | New |
| `core/git/scanner.go` | New — ScanSkills |
| `core/git/scanner_test.go` | New |
| `core/notify/model.go` | Add EventStarSyncProgress, EventStarSyncDone, StarSyncProgressPayload |
| `app.go` | Refactor ScanGitHub/InstallFromGitHub; add starStorage field, startup init, all starred repo methods |
| `frontend/wailsjs/…` | Regenerated |
| `frontend/src/App.tsx` | Add sidebar entry + route |
| `frontend/src/pages/StarredRepos.tsx` | New page |
| `frontend/src/components/GitHubInstallDialog.tsx` | Update scan UX copy |

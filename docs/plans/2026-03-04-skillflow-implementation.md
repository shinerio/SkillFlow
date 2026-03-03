# SkillFlow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build SkillFlow, a cross-platform desktop app (macOS + Windows) for managing LLM SKILLS across multiple tools, with GitHub install, cloud backup, and cross-tool sync.

**Architecture:** Go 1.26 core library (no UI deps) with interface-based extensibility for cloud providers/tool adapters/installers. Wails v2 bridges core to a React+TypeScript frontend via method bindings and channel-based event forwarding.

**Tech Stack:** Go 1.26, Wails v2, React 18, TypeScript, Zustand, Tailwind CSS, testify, zalando/go-keyring

---

## Phase 1: Foundation

### Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `wails.json`
- Create: `app/wails/main.go`
- Create: `frontend/package.json`

**Step 1: Initialize Go module**

```bash
cd /Users/shinerio/Workspace/code/SkillFlow
go mod init github.com/shinerio/skillflow
```

**Step 2: Install Wails CLI and init project**

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails init -n SkillFlow -t react-ts -d .
```

**Step 3: Create core directory structure**

```bash
mkdir -p core/{skill,sync,backup,install,update,config,registry,notify}
mkdir -p app/wails
mkdir -p docs/plans
```

**Step 4: Add Go dependencies**

```bash
go get github.com/stretchr/testify@latest
go get github.com/zalando/go-keyring@latest
go get github.com/google/uuid@latest
go get github.com/aliyun/aliyun-oss-go-sdk/v3@latest
go get github.com/tencentyun/cos-go-sdk-v5@latest
go get github.com/huaweicloud/huaweicloud-sdk-go-obs@latest
```

**Step 5: Verify Wails project runs**

```bash
wails dev
```
Expected: Browser opens with default Wails React template.

**Step 6: Commit**

```bash
git init
git add .
git commit -m "chore: initialize SkillFlow project with Wails + Go 1.26"
```

---

### Task 2: Core Data Models

**Files:**
- Create: `core/skill/model.go`
- Create: `core/config/model.go`
- Create: `core/notify/model.go`
- Create: `core/skill/model_test.go`

**Step 1: Write model tests**

Create `core/skill/model_test.go`:

```go
package skill_test

import (
    "testing"
    "time"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/stretchr/testify/assert"
)

func TestSkillSourceTypes(t *testing.T) {
    s := skill.Skill{
        ID:       "test-id",
        Name:     "my-skill",
        Source:   skill.SourceGitHub,
        Category: "coding",
    }
    assert.Equal(t, skill.SourceType("github"), s.Source)
    assert.True(t, s.IsGitHub())
    assert.False(t, s.IsManual())
}

func TestSkillIsManual(t *testing.T) {
    s := skill.Skill{Source: skill.SourceManual}
    assert.True(t, s.IsManual())
    assert.False(t, s.IsGitHub())
}

func TestSkillHasUpdate(t *testing.T) {
    s := skill.Skill{
        Source:       skill.SourceGitHub,
        SourceSHA:    "abc123",
        LatestSHA:    "def456",
    }
    assert.True(t, s.HasUpdate())

    s.LatestSHA = "abc123"
    assert.False(t, s.HasUpdate())
}
```

**Step 2: Run test to verify it fails**

```bash
cd core/skill && go test ./... -v
```
Expected: FAIL — package not defined.

**Step 3: Implement `core/skill/model.go`**

```go
package skill

import "time"

type SourceType string

const (
    SourceGitHub SourceType = "github"
    SourceManual SourceType = "manual"
)

type Skill struct {
    ID            string
    Name          string
    Path          string
    Category      string
    Source        SourceType
    SourceURL     string
    SourceSubPath string
    SourceSHA     string
    LatestSHA     string
    InstalledAt   time.Time
    UpdatedAt     time.Time
    LastCheckedAt time.Time
}

func (s *Skill) IsGitHub() bool { return s.Source == SourceGitHub }
func (s *Skill) IsManual() bool { return s.Source == SourceManual }
func (s *Skill) HasUpdate() bool {
    return s.IsGitHub() && s.LatestSHA != "" && s.LatestSHA != s.SourceSHA
}
```

**Step 4: Implement `core/config/model.go`**

```go
package config

type ToolConfig struct {
    Name      string `json:"name"`
    SkillsDir string `json:"skillsDir"`
    Enabled   bool   `json:"enabled"`
    Custom    bool   `json:"custom"`
}

type CloudConfig struct {
    Provider    string            `json:"provider"`
    Enabled     bool              `json:"enabled"`
    BucketName  string            `json:"bucketName"`
    RemotePath  string            `json:"remotePath"`
    Credentials map[string]string `json:"credentials"`
}

type AppConfig struct {
    SkillsStorageDir string       `json:"skillsStorageDir"`
    DefaultCategory  string       `json:"defaultCategory"`
    Tools            []ToolConfig `json:"tools"`
    Cloud            CloudConfig  `json:"cloud"`
}
```

**Step 5: Implement `core/notify/model.go`**

```go
package notify

type EventType string

const (
    EventBackupStarted   EventType = "backup.started"
    EventBackupProgress  EventType = "backup.progress"
    EventBackupCompleted EventType = "backup.completed"
    EventBackupFailed    EventType = "backup.failed"
    EventSyncCompleted   EventType = "sync.completed"
    EventUpdateAvailable EventType = "update.available"
    EventSkillConflict   EventType = "skill.conflict"
)

type Event struct {
    Type    EventType `json:"type"`
    Payload any       `json:"payload"`
}

type BackupProgressPayload struct {
    FilesTotal    int    `json:"filesTotal"`
    FilesUploaded int    `json:"filesUploaded"`
    CurrentFile   string `json:"currentFile"`
}

type UpdateAvailablePayload struct {
    SkillID   string `json:"skillId"`
    SkillName string `json:"skillName"`
    CurrentSHA string `json:"currentSha"`
    LatestSHA  string `json:"latestSha"`
}

type ConflictPayload struct {
    SkillName  string `json:"skillName"`
    TargetPath string `json:"targetPath"`
}
```

**Step 6: Run tests**

```bash
go test ./core/... -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add core/
git commit -m "feat: add core data models (skill, config, notify)"
```

---

### Task 3: Notify Hub

**Files:**
- Create: `core/notify/hub.go`
- Create: `core/notify/hub_test.go`

**Step 1: Write failing test**

Create `core/notify/hub_test.go`:

```go
package notify_test

import (
    "testing"
    "time"
    "github.com/shinerio/skillflow/core/notify"
    "github.com/stretchr/testify/assert"
)

func TestHubPublishSubscribe(t *testing.T) {
    hub := notify.NewHub()
    ch := hub.Subscribe()
    defer hub.Unsubscribe(ch)

    hub.Publish(notify.Event{Type: notify.EventBackupStarted, Payload: nil})

    select {
    case evt := <-ch:
        assert.Equal(t, notify.EventBackupStarted, evt.Type)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected event, got timeout")
    }
}

func TestHubMultipleSubscribers(t *testing.T) {
    hub := notify.NewHub()
    ch1 := hub.Subscribe()
    ch2 := hub.Subscribe()
    defer hub.Unsubscribe(ch1)
    defer hub.Unsubscribe(ch2)

    hub.Publish(notify.Event{Type: notify.EventSyncCompleted})

    for _, ch := range []<-chan notify.Event{ch1, ch2} {
        select {
        case evt := <-ch:
            assert.Equal(t, notify.EventSyncCompleted, evt.Type)
        case <-time.After(100 * time.Millisecond):
            t.Fatal("subscriber did not receive event")
        }
    }
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/notify/... -v
```
Expected: FAIL

**Step 3: Implement `core/notify/hub.go`**

```go
package notify

import "sync"

type Hub struct {
    mu          sync.RWMutex
    subscribers map[chan Event]struct{}
}

func NewHub() *Hub {
    return &Hub{subscribers: make(map[chan Event]struct{})}
}

func (h *Hub) Subscribe() <-chan Event {
    ch := make(chan Event, 32)
    h.mu.Lock()
    h.subscribers[ch] = struct{}{}
    h.mu.Unlock()
    return ch
}

func (h *Hub) Unsubscribe(ch <-chan Event) {
    h.mu.Lock()
    defer h.mu.Unlock()
    for sub := range h.subscribers {
        if sub == ch {
            delete(h.subscribers, sub)
            close(sub)
            return
        }
    }
}

func (h *Hub) Publish(evt Event) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for sub := range h.subscribers {
        select {
        case sub <- evt:
        default: // drop if subscriber is slow
        }
    }
}
```

**Step 4: Run tests**

```bash
go test ./core/notify/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/notify/
git commit -m "feat: add channel-based notify hub"
```

---

### Task 4: Config Service

**Files:**
- Create: `core/config/service.go`
- Create: `core/config/service_test.go`
- Create: `core/config/defaults.go`

**Step 1: Write failing tests**

Create `core/config/service_test.go`:

```go
package config_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/config"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
    dir := t.TempDir()
    svc := config.NewService(dir)
    cfg, err := svc.Load()
    require.NoError(t, err)
    assert.NotEmpty(t, cfg.SkillsStorageDir)
    assert.Equal(t, "Imported", cfg.DefaultCategory)
    assert.NotEmpty(t, cfg.Tools)
}

func TestSaveAndLoadConfig(t *testing.T) {
    dir := t.TempDir()
    svc := config.NewService(dir)
    cfg := config.DefaultConfig(dir)
    cfg.DefaultCategory = "MyCategory"
    err := svc.Save(cfg)
    require.NoError(t, err)

    loaded, err := svc.Load()
    require.NoError(t, err)
    assert.Equal(t, "MyCategory", loaded.DefaultCategory)
}

func TestConfigFileCreatedOnFirstLoad(t *testing.T) {
    dir := t.TempDir()
    svc := config.NewService(dir)
    _, err := svc.Load()
    require.NoError(t, err)
    _, err = os.Stat(filepath.Join(dir, "config.json"))
    assert.NoError(t, err)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/config/... -v
```
Expected: FAIL

**Step 3: Implement `core/config/defaults.go`**

```go
package config

import (
    "os"
    "path/filepath"
    "runtime"
)

func AppDataDir() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "SkillFlow")
    default: // darwin
        home, _ := os.UserHomeDir()
        return filepath.Join(home, "Library", "Application Support", "SkillFlow")
    }
}

func DefaultToolsDir(toolName string) string {
    home, _ := os.UserHomeDir()
    dirs := map[string]map[string]string{
        "darwin": {
            "claude-code": filepath.Join(home, ".claude", "skills"),
            "opencode":    filepath.Join(home, ".opencode", "skills"),
            "codex":       filepath.Join(home, ".codex", "skills"),
            "gemini-cli":  filepath.Join(home, ".gemini", "skills"),
            "openclaw":    filepath.Join(home, ".openclaw", "skills"),
        },
        "windows": {
            "claude-code": filepath.Join(os.Getenv("APPDATA"), "claude", "skills"),
            "opencode":    filepath.Join(os.Getenv("APPDATA"), "opencode", "skills"),
            "codex":       filepath.Join(os.Getenv("APPDATA"), "codex", "skills"),
            "gemini-cli":  filepath.Join(os.Getenv("APPDATA"), "gemini", "skills"),
            "openclaw":    filepath.Join(os.Getenv("APPDATA"), "openclaw", "skills"),
        },
    }
    goos := runtime.GOOS
    if goos != "windows" {
        goos = "darwin"
    }
    return dirs[goos][toolName]
}

var builtinTools = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
    tools := make([]ToolConfig, 0, len(builtinTools))
    for _, name := range builtinTools {
        dir := DefaultToolsDir(name)
        _, err := os.Stat(dir)
        tools = append(tools, ToolConfig{
            Name:      name,
            SkillsDir: dir,
            Enabled:   err == nil,
            Custom:    false,
        })
    }
    return AppConfig{
        SkillsStorageDir: filepath.Join(dataDir, "skills"),
        DefaultCategory:  "Imported",
        Tools:            tools,
        Cloud:            CloudConfig{RemotePath: "skillflow/"},
    }
}
```

**Step 4: Implement `core/config/service.go`**

```go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Service struct {
    dataDir    string
    configPath string
}

func NewService(dataDir string) *Service {
    return &Service{
        dataDir:    dataDir,
        configPath: filepath.Join(dataDir, "config.json"),
    }
}

func (s *Service) Load() (AppConfig, error) {
    if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
        cfg := DefaultConfig(s.dataDir)
        if err := s.Save(cfg); err != nil {
            return AppConfig{}, err
        }
        return cfg, nil
    }
    data, err := os.ReadFile(s.configPath)
    if err != nil {
        return AppConfig{}, err
    }
    var cfg AppConfig
    return cfg, json.Unmarshal(data, &cfg)
}

func (s *Service) Save(cfg AppConfig) error {
    if err := os.MkdirAll(s.dataDir, 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(s.configPath, data, 0644)
}
```

**Step 5: Run tests**

```bash
go test ./core/config/... -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add core/config/
git commit -m "feat: add config service with cross-platform defaults"
```

---

### Task 5: Registry

**Files:**
- Create: `core/registry/registry.go`

**Step 1: Implement registry (no tests needed — pure registration plumbing)**

```go
package registry

import (
    "github.com/shinerio/skillflow/core/backup"
    "github.com/shinerio/skillflow/core/install"
    "github.com/shinerio/skillflow/core/sync"
)

var (
    installers     = map[string]install.Installer{}
    adapters       = map[string]sync.ToolAdapter{}
    cloudProviders = map[string]backup.CloudProvider{}
)

func RegisterInstaller(i install.Installer)         { installers[i.Type()] = i }
func RegisterAdapter(a sync.ToolAdapter)             { adapters[a.Name()] = a }
func RegisterCloudProvider(p backup.CloudProvider)   { cloudProviders[p.Name()] = p }

func GetInstaller(t string) (install.Installer, bool) {
    i, ok := installers[t]
    return i, ok
}

func GetAdapter(name string) (sync.ToolAdapter, bool) {
    a, ok := adapters[name]
    return a, ok
}

func GetCloudProvider(name string) (backup.CloudProvider, bool) {
    p, ok := cloudProviders[name]
    return p, ok
}

func AllAdapters() []sync.ToolAdapter {
    result := make([]sync.ToolAdapter, 0, len(adapters))
    for _, a := range adapters {
        result = append(result, a)
    }
    return result
}

func AllCloudProviders() []backup.CloudProvider {
    result := make([]backup.CloudProvider, 0, len(cloudProviders))
    for _, p := range cloudProviders {
        result = append(result, p)
    }
    return result
}
```

**Step 2: Commit**

```bash
git add core/registry/
git commit -m "feat: add extensible registry for installers/adapters/providers"
```

---

## Phase 2: Core Skill Management

### Task 6: Skill Validator

**Files:**
- Create: `core/skill/validator.go`
- Create: `core/skill/validator_test.go`

**Step 1: Write failing tests**

```go
package skill_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestValidatorAcceptsDirectoryWithSKILLSmd(t *testing.T) {
    dir := t.TempDir()
    skillDir := filepath.Join(dir, "my-skill")
    require.NoError(t, os.MkdirAll(skillDir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILLS.md"), []byte("# skill"), 0644))

    v := skill.NewValidator()
    err := v.Validate(skillDir)
    assert.NoError(t, err)
}

func TestValidatorRejectsDirectoryWithoutSKILLSmd(t *testing.T) {
    dir := t.TempDir()
    skillDir := filepath.Join(dir, "not-a-skill")
    require.NoError(t, os.MkdirAll(skillDir, 0755))

    v := skill.NewValidator()
    err := v.Validate(skillDir)
    assert.ErrorIs(t, err, skill.ErrNoSKILLSmd)
}

func TestValidatorRejectsNonDirectory(t *testing.T) {
    v := skill.NewValidator()
    err := v.Validate("/nonexistent/path")
    assert.Error(t, err)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/skill/... -run TestValidator -v
```

**Step 3: Implement `core/skill/validator.go`**

```go
package skill

import (
    "errors"
    "os"
    "path/filepath"
)

var ErrNoSKILLSmd = errors.New("SKILLS.md not found in skill directory")

// ValidationRule is the extension point for future complex validators.
type ValidationRule func(dir string) error

type Validator struct {
    rules []ValidationRule
}

func NewValidator(extraRules ...ValidationRule) *Validator {
    rules := []ValidationRule{requireSKILLSmd}
    return &Validator{rules: append(rules, extraRules...)}
}

func (v *Validator) Validate(dir string) error {
    for _, rule := range v.rules {
        if err := rule(dir); err != nil {
            return err
        }
    }
    return nil
}

func requireSKILLSmd(dir string) error {
    if _, err := os.Stat(dir); err != nil {
        return err
    }
    mdPath := filepath.Join(dir, "SKILLS.md")
    if _, err := os.Stat(mdPath); os.IsNotExist(err) {
        return ErrNoSKILLSmd
    }
    return nil
}
```

**Step 4: Run tests**

```bash
go test ./core/skill/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/skill/
git commit -m "feat: add extensible skill validator (SKILLS.md check)"
```

---

### Task 7: Skill Storage Service

**Files:**
- Create: `core/skill/storage.go`
- Create: `core/skill/storage_test.go`

**Step 1: Write failing tests**

```go
package skill_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func makeTestSkillDir(t *testing.T, baseDir, name string) string {
    t.Helper()
    dir := filepath.Join(baseDir, name)
    require.NoError(t, os.MkdirAll(dir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
    return dir
}

func TestStorageListCategories(t *testing.T) {
    root := t.TempDir()
    svc := skill.NewStorage(root)
    require.NoError(t, svc.CreateCategory("coding"))
    require.NoError(t, svc.CreateCategory("writing"))
    cats, err := svc.ListCategories()
    require.NoError(t, err)
    assert.ElementsMatch(t, []string{"coding", "writing"}, cats)
}

func TestStorageImportSkill(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "my-skill")
    svc := skill.NewStorage(root)

    imported, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
    require.NoError(t, err)
    assert.Equal(t, "my-skill", imported.Name)
    assert.Equal(t, "coding", imported.Category)

    // verify directory was copied
    _, err = os.Stat(filepath.Join(root, "coding", "my-skill", "SKILLS.md"))
    assert.NoError(t, err)
}

func TestStorageConflictDetected(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "dup-skill")
    svc := skill.NewStorage(root)

    _, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
    require.NoError(t, err)

    _, err = svc.Import(skillDir, "coding", skill.SourceManual, "", "")
    assert.ErrorIs(t, err, skill.ErrSkillExists)
}

func TestStorageDeleteSkill(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "del-skill")
    svc := skill.NewStorage(root)

    s, err := svc.Import(skillDir, "", skill.SourceManual, "", "")
    require.NoError(t, err)
    require.NoError(t, svc.Delete(s.ID))

    skills, err := svc.ListAll()
    require.NoError(t, err)
    assert.Empty(t, skills)
}

func TestStorageMoveCategory(t *testing.T) {
    root := t.TempDir()
    src := t.TempDir()
    skillDir := makeTestSkillDir(t, src, "move-skill")
    svc := skill.NewStorage(root)
    require.NoError(t, svc.CreateCategory("cat-a"))
    require.NoError(t, svc.CreateCategory("cat-b"))

    s, err := svc.Import(skillDir, "cat-a", skill.SourceManual, "", "")
    require.NoError(t, err)

    err = svc.MoveCategory(s.ID, "cat-b")
    require.NoError(t, err)

    updated, err := svc.Get(s.ID)
    require.NoError(t, err)
    assert.Equal(t, "cat-b", updated.Category)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/skill/... -run TestStorage -v
```

**Step 3: Implement `core/skill/storage.go`**

```go
package skill

import (
    "encoding/json"
    "errors"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/google/uuid"
)

var ErrSkillExists = errors.New("skill already exists in target location")
var ErrSkillNotFound = errors.New("skill not found")

type Storage struct {
    root    string
    metaDir string
}

func NewStorage(root string) *Storage {
    return &Storage{
        root:    root,
        metaDir: filepath.Join(filepath.Dir(root), "meta"),
    }
}

func (s *Storage) CreateCategory(name string) error {
    return os.MkdirAll(filepath.Join(s.root, name), 0755)
}

func (s *Storage) ListCategories() ([]string, error) {
    entries, err := os.ReadDir(s.root)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }
    var cats []string
    for _, e := range entries {
        if e.IsDir() {
            cats = append(cats, e.Name())
        }
    }
    return cats, nil
}

func (s *Storage) Import(srcDir, category string, source SourceType, sourceURL, sourceSubPath string) (*Skill, error) {
    name := filepath.Base(srcDir)
    targetDir := filepath.Join(s.root, category, name)
    if _, err := os.Stat(targetDir); err == nil {
        return nil, ErrSkillExists
    }
    if err := copyDir(srcDir, targetDir); err != nil {
        return nil, err
    }
    sk := &Skill{
        ID:            uuid.New().String(),
        Name:          name,
        Path:          targetDir,
        Category:      category,
        Source:        source,
        SourceURL:     sourceURL,
        SourceSubPath: sourceSubPath,
        InstalledAt:   time.Now(),
        UpdatedAt:     time.Now(),
    }
    return sk, s.saveMeta(sk)
}

func (s *Storage) Get(id string) (*Skill, error) {
    skills, err := s.ListAll()
    if err != nil {
        return nil, err
    }
    for _, sk := range skills {
        if sk.ID == id {
            return sk, nil
        }
    }
    return nil, ErrSkillNotFound
}

func (s *Storage) ListAll() ([]*Skill, error) {
    if err := os.MkdirAll(s.metaDir, 0755); err != nil {
        return nil, err
    }
    entries, err := os.ReadDir(s.metaDir)
    if err != nil {
        return nil, err
    }
    var skills []*Skill
    for _, e := range entries {
        if filepath.Ext(e.Name()) != ".json" {
            continue
        }
        data, err := os.ReadFile(filepath.Join(s.metaDir, e.Name()))
        if err != nil {
            continue
        }
        var sk Skill
        if err := json.Unmarshal(data, &sk); err == nil {
            skills = append(skills, &sk)
        }
    }
    return skills, nil
}

func (s *Storage) Delete(id string) error {
    sk, err := s.Get(id)
    if err != nil {
        return err
    }
    if err := os.RemoveAll(sk.Path); err != nil {
        return err
    }
    return os.Remove(filepath.Join(s.metaDir, id+".json"))
}

func (s *Storage) MoveCategory(id, newCategory string) error {
    sk, err := s.Get(id)
    if err != nil {
        return err
    }
    newPath := filepath.Join(s.root, newCategory, sk.Name)
    if err := os.MkdirAll(filepath.Join(s.root, newCategory), 0755); err != nil {
        return err
    }
    if err := os.Rename(sk.Path, newPath); err != nil {
        return err
    }
    sk.Path = newPath
    sk.Category = newCategory
    sk.UpdatedAt = time.Now()
    return s.saveMeta(sk)
}

func (s *Storage) UpdateMeta(sk *Skill) error {
    sk.UpdatedAt = time.Now()
    return s.saveMeta(sk)
}

func (s *Storage) saveMeta(sk *Skill) error {
    if err := os.MkdirAll(s.metaDir, 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(sk, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(filepath.Join(s.metaDir, sk.ID+".json"), data, 0644)
}

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(src, path)
        target := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(target, info.Mode())
        }
        return copyFile(path, target)
    })
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}
```

**Step 4: Run tests**

```bash
go test ./core/skill/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/skill/
git commit -m "feat: add skill storage service with category and meta management"
```

---

## Phase 3: Install

### Task 8: Install Interfaces

**Files:**
- Create: `core/install/installer.go`

```go
package install

import "context"

type InstallSource struct {
    Type string // "github" | "local"
    URI  string
}

type SkillCandidate struct {
    Name      string
    Path      string // relative path within source
    Installed bool
}

type Installer interface {
    Type() string
    Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
    Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
}
```

**Step 1: Commit interface**

```bash
git add core/install/
git commit -m "feat: add installer interface"
```

---

### Task 9: GitHub Installer

**Files:**
- Create: `core/install/github.go`
- Create: `core/install/github_test.go`

**Step 1: Write failing tests (using httptest to mock GitHub API)**

```go
package install_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/shinerio/skillflow/core/install"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func mockGitHubServer(t *testing.T) *httptest.Server {
    t.Helper()
    mux := http.NewServeMux()
    // Mock: list skills directory contents
    mux.HandleFunc("/repos/user/repo/contents/skills", func(w http.ResponseWriter, r *http.Request) {
        items := []map[string]any{
            {"name": "skill-a", "type": "dir", "path": "skills/skill-a"},
            {"name": "skill-b", "type": "dir", "path": "skills/skill-b"},
            {"name": "readme.md", "type": "file", "path": "skills/readme.md"},
        }
        json.NewEncoder(w).Encode(items)
    })
    // Mock: check SKILLS.md existence for skill-a (returns file info)
    mux.HandleFunc("/repos/user/repo/contents/skills/skill-a/SKILLS.md", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]any{"name": "SKILLS.md", "type": "file"})
    })
    // Mock: skill-b has no SKILLS.md (404)
    mux.HandleFunc("/repos/user/repo/contents/skills/skill-b/SKILLS.md", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    })
    return httptest.NewServer(mux)
}

func TestGitHubInstallerScan(t *testing.T) {
    srv := mockGitHubServer(t)
    defer srv.Close()

    installer := install.NewGitHubInstaller(srv.URL)
    candidates, err := installer.Scan(context.Background(), install.InstallSource{
        Type: "github",
        URI:  srv.URL + "/repos/user/repo",
    })
    require.NoError(t, err)
    // Only skill-a has SKILLS.md, skill-b does not
    assert.Len(t, candidates, 1)
    assert.Equal(t, "skill-a", candidates[0].Name)
}
```

**Step 2: Run to verify failure**

```bash
go test ./core/install/... -v
```

**Step 3: Implement `core/install/github.go`**

```go
package install

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type githubContent struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Path        string `json:"path"`
    DownloadURL string `json:"download_url"`
}

type GitHubInstaller struct {
    baseURL string // overridable for tests
    client  *http.Client
}

func NewGitHubInstaller(baseURL string) *GitHubInstaller {
    if baseURL == "" {
        baseURL = "https://api.github.com"
    }
    return &GitHubInstaller{baseURL: baseURL, client: http.DefaultClient}
}

func (g *GitHubInstaller) Type() string { return "github" }

func (g *GitHubInstaller) Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error) {
    owner, repo, err := parseGitHubURI(source.URI)
    if err != nil {
        return nil, err
    }
    items, err := g.listContents(ctx, owner, repo, "skills")
    if err != nil {
        return nil, err
    }
    var candidates []SkillCandidate
    for _, item := range items {
        if item.Type != "dir" {
            continue
        }
        // Check SKILLS.md exists
        if g.fileExists(ctx, owner, repo, item.Path+"/SKILLS.md") {
            candidates = append(candidates, SkillCandidate{
                Name: item.Name,
                Path: item.Path,
            })
        }
    }
    return candidates, nil
}

func (g *GitHubInstaller) Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error {
    owner, repo, err := parseGitHubURI(source.URI)
    if err != nil {
        return err
    }
    for _, c := range selected {
        if err := g.downloadDir(ctx, owner, repo, c.Path, category, c.Name); err != nil {
            return fmt.Errorf("install %s: %w", c.Name, err)
        }
    }
    return nil
}

func (g *GitHubInstaller) listContents(ctx context.Context, owner, repo, path string) ([]githubContent, error) {
    url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, path)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var items []githubContent
    return items, json.NewDecoder(resp.Body).Decode(&items)
}

func (g *GitHubInstaller) fileExists(ctx context.Context, owner, repo, path string) bool {
    url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, path)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return false
    }
    resp.Body.Close()
    return resp.StatusCode == http.StatusOK
}

func (g *GitHubInstaller) downloadDir(ctx context.Context, owner, repo, remotePath, category, name string) error {
    items, err := g.listContents(ctx, owner, repo, remotePath)
    if err != nil {
        return err
    }
    for _, item := range items {
        if item.Type == "dir" {
            if err := g.downloadDir(ctx, owner, repo, item.Path, category, name); err != nil {
                return err
            }
        } else if item.DownloadURL != "" {
            if err := g.downloadFile(ctx, item.DownloadURL, category, name, item.Path, remotePath); err != nil {
                return err
            }
        }
    }
    return nil
}

func (g *GitHubInstaller) downloadFile(ctx context.Context, url, category, skillName, filePath, basePath string) error {
    rel := strings.TrimPrefix(filePath, basePath+"/")
    // caller (app layer) sets actual target; installer returns to tmp dir
    tmpDir := filepath.Join(os.TempDir(), "skillflow-install", skillName)
    target := filepath.Join(tmpDir, rel)
    if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
        return err
    }
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := g.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    f, err := os.Create(target)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, resp.Body)
    return err
}

func parseGitHubURI(uri string) (owner, repo string, err error) {
    // Accept: https://github.com/owner/repo or https://api.github.com/repos/owner/repo
    uri = strings.TrimSuffix(uri, "/")
    parts := strings.Split(uri, "/")
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid GitHub URI: %s", uri)
    }
    return parts[len(parts)-2], parts[len(parts)-1], nil
}
```

**Step 4: Run tests**

```bash
go test ./core/install/... -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add core/install/
git commit -m "feat: add GitHub installer with SKILLS.md validation"
```

---

### Task 10: Local Installer

**Files:**
- Create: `core/install/local.go`
- Create: `core/install/local_test.go`

**Step 1: Write failing tests**

```go
package install_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/install"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLocalInstallerScanValidSkill(t *testing.T) {
    dir := t.TempDir()
    require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# skill"), 0644))

    inst := install.NewLocalInstaller()
    candidates, err := inst.Scan(context.Background(), install.InstallSource{Type: "local", URI: dir})
    require.NoError(t, err)
    assert.Len(t, candidates, 1)
    assert.Equal(t, filepath.Base(dir), candidates[0].Name)
}

func TestLocalInstallerScanInvalidSkill(t *testing.T) {
    dir := t.TempDir() // no SKILLS.md
    inst := install.NewLocalInstaller()
    candidates, err := inst.Scan(context.Background(), install.InstallSource{Type: "local", URI: dir})
    require.NoError(t, err)
    assert.Empty(t, candidates)
}
```

**Step 2: Implement `core/install/local.go`**

```go
package install

import (
    "context"
    "os"
    "path/filepath"
    "github.com/shinerio/skillflow/core/skill"
)

type LocalInstaller struct {
    validator *skill.Validator
}

func NewLocalInstaller() *LocalInstaller {
    return &LocalInstaller{validator: skill.NewValidator()}
}

func (l *LocalInstaller) Type() string { return "local" }

func (l *LocalInstaller) Scan(_ context.Context, source InstallSource) ([]SkillCandidate, error) {
    dir := source.URI
    if err := l.validator.Validate(dir); err != nil {
        return nil, nil // not a valid skill dir — return empty, not error
    }
    return []SkillCandidate{{Name: filepath.Base(dir), Path: dir}}, nil
}

func (l *LocalInstaller) Install(_ context.Context, _ InstallSource, selected []SkillCandidate, _ string) error {
    // Local install: the app layer copies from candidate.Path directly via Storage.Import
    // This installer's Install is a no-op; the app layer calls Storage.Import
    _ = selected
    return nil
}

// Ensure os import used
var _ = os.Stat
```

**Step 3: Run tests**

```bash
go test ./core/install/... -v
```
Expected: PASS

**Step 4: Commit**

```bash
git add core/install/
git commit -m "feat: add local installer for manual skill import"
```

---

## Phase 4: Sync

### Task 11: Sync Interfaces and Tool Adapters

**Files:**
- Create: `core/sync/adapter.go`
- Create: `core/sync/filesystem_adapter.go`
- Create: `core/sync/filesystem_adapter_test.go`

**Step 1: Define sync interface**

```go
// core/sync/adapter.go
package sync

import (
    "context"
    "github.com/shinerio/skillflow/core/skill"
)

type ToolAdapter interface {
    Name() string
    DefaultSkillsDir() string
    // Push copies skills into targetDir, flattened (no category subdirs)
    Push(ctx context.Context, skills []*skill.Skill, targetDir string) error
    // Pull scans sourceDir and returns skill candidates (not yet imported)
    Pull(ctx context.Context, sourceDir string) ([]*skill.Skill, error)
}
```

**Step 2: Write failing tests**

```go
// core/sync/filesystem_adapter_test.go
package sync_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    toolsync "github.com/shinerio/skillflow/core/sync"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func makeSkillDir(t *testing.T, root, category, name string) *skill.Skill {
    t.Helper()
    dir := filepath.Join(root, category, name)
    require.NoError(t, os.MkdirAll(dir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
    return &skill.Skill{Name: name, Path: dir, Category: category}
}

func TestFilesystemAdapterPushFlattens(t *testing.T) {
    src := t.TempDir()
    dst := t.TempDir()
    sk := makeSkillDir(t, src, "coding", "my-skill")

    adapter := toolsync.NewFilesystemAdapter("test-tool", "")
    err := adapter.Push(context.Background(), []*skill.Skill{sk}, dst)
    require.NoError(t, err)

    // skill should be at dst/my-skill (no category subdir)
    _, err = os.Stat(filepath.Join(dst, "my-skill", "SKILLS.md"))
    assert.NoError(t, err)
}

func TestFilesystemAdapterPull(t *testing.T) {
    src := t.TempDir()
    // Create two valid skills directly in src (tool dir is flat)
    for _, name := range []string{"skill-x", "skill-y"} {
        dir := filepath.Join(src, name)
        require.NoError(t, os.MkdirAll(dir, 0755))
        require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
    }
    // Create a non-skill directory
    require.NoError(t, os.MkdirAll(filepath.Join(src, "not-a-skill"), 0755))

    adapter := toolsync.NewFilesystemAdapter("test-tool", "")
    skills, err := adapter.Pull(context.Background(), src)
    require.NoError(t, err)
    assert.Len(t, skills, 2)
}
```

**Step 3: Implement `core/sync/filesystem_adapter.go`**

```go
package sync

import (
    "context"
    "io"
    "os"
    "path/filepath"
    "github.com/shinerio/skillflow/core/skill"
)

// FilesystemAdapter works for all tools — they all share the same file-based skills directory model.
type FilesystemAdapter struct {
    name          string
    defaultSkillsDir string
}

func NewFilesystemAdapter(name, defaultSkillsDir string) *FilesystemAdapter {
    return &FilesystemAdapter{name: name, defaultSkillsDir: defaultSkillsDir}
}

func (f *FilesystemAdapter) Name() string             { return f.name }
func (f *FilesystemAdapter) DefaultSkillsDir() string { return f.defaultSkillsDir }

func (f *FilesystemAdapter) Push(_ context.Context, skills []*skill.Skill, targetDir string) error {
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        return err
    }
    for _, sk := range skills {
        dst := filepath.Join(targetDir, sk.Name)
        if err := copyDir(sk.Path, dst); err != nil {
            return err
        }
    }
    return nil
}

func (f *FilesystemAdapter) Pull(_ context.Context, sourceDir string) ([]*skill.Skill, error) {
    validator := skill.NewValidator()
    entries, err := os.ReadDir(sourceDir)
    if err != nil {
        return nil, err
    }
    var skills []*skill.Skill
    for _, e := range entries {
        if !e.IsDir() {
            continue
        }
        dir := filepath.Join(sourceDir, e.Name())
        if err := validator.Validate(dir); err == nil {
            skills = append(skills, &skill.Skill{
                Name:   e.Name(),
                Path:   dir,
                Source: skill.SourceManual, // pulled from external tool
            })
        }
    }
    return skills, nil
}

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(src, path)
        target := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(target, info.Mode())
        }
        return copyFile(path, target)
    })
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}
```

**Step 4: Run tests**

```bash
go test ./core/sync/... -v
```
Expected: PASS

**Step 5: Register all built-in adapters in `app/wails/adapters.go`**

```go
package main

import (
    "github.com/shinerio/skillflow/core/config"
    "github.com/shinerio/skillflow/core/registry"
    toolsync "github.com/shinerio/skillflow/core/sync"
    "runtime"
)

func registerAdapters() {
    tools := []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}
    for _, name := range tools {
        registry.RegisterAdapter(toolsync.NewFilesystemAdapter(name, config.DefaultToolsDir(name)))
    }
}
```

**Step 6: Commit**

```bash
git add core/sync/ app/wails/adapters.go
git commit -m "feat: add filesystem sync adapter shared by all tools"
```

---

## Phase 5: Cloud Backup

### Task 12: Cloud Backup Interface + Aliyun OSS

**Files:**
- Create: `core/backup/provider.go`
- Create: `core/backup/aliyun.go`

**Step 1: Define backup interface**

```go
// core/backup/provider.go
package backup

import "context"

type CredentialField struct {
    Key         string `json:"key"`
    Label       string `json:"label"`
    Placeholder string `json:"placeholder"`
    Secret      bool   `json:"secret"`
}

type RemoteFile struct {
    Path         string `json:"path"`
    Size         int64  `json:"size"`
    IsDir        bool   `json:"isDir"`
}

type CloudProvider interface {
    Name() string
    Init(credentials map[string]string) error
    // Sync mirrors localDir to cloud bucket at remotePath (incremental, no compression)
    Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error
    Restore(ctx context.Context, bucket, remotePath, localDir string) error
    List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error)
    RequiredCredentials() []CredentialField
}
```

**Step 2: Implement `core/backup/aliyun.go`**

```go
package backup

import (
    "context"
    "os"
    "path/filepath"
    "strings"
    "github.com/aliyun/aliyun-oss-go-sdk/v3/oss"
)

type AliyunProvider struct {
    client *oss.Client
}

func NewAliyunProvider() *AliyunProvider { return &AliyunProvider{} }

func (a *AliyunProvider) Name() string { return "aliyun" }

func (a *AliyunProvider) RequiredCredentials() []CredentialField {
    return []CredentialField{
        {Key: "access_key_id", Label: "Access Key ID", Secret: false},
        {Key: "access_key_secret", Label: "Access Key Secret", Secret: true},
        {Key: "endpoint", Label: "Endpoint", Placeholder: "oss-cn-hangzhou.aliyuncs.com"},
    }
}

func (a *AliyunProvider) Init(creds map[string]string) error {
    client, err := oss.New(creds["endpoint"], creds["access_key_id"], creds["access_key_secret"])
    if err != nil {
        return err
    }
    a.client = client
    return nil
}

func (a *AliyunProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
    b, err := a.client.Bucket(bucket)
    if err != nil {
        return err
    }
    return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }
        rel, _ := filepath.Rel(localDir, path)
        key := remotePath + strings.ReplaceAll(rel, string(filepath.Separator), "/")
        if onProgress != nil {
            onProgress(rel)
        }
        return b.PutObjectFromFile(key, path)
    })
}

func (a *AliyunProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
    b, err := a.client.Bucket(bucket)
    if err != nil {
        return err
    }
    marker := ""
    for {
        result, err := b.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
        if err != nil {
            return err
        }
        for _, obj := range result.Objects {
            rel := strings.TrimPrefix(obj.Key, remotePath)
            local := filepath.Join(localDir, filepath.FromSlash(rel))
            if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
                return err
            }
            if err := b.GetObjectToFile(obj.Key, local); err != nil {
                return err
            }
        }
        if !result.IsTruncated {
            break
        }
        marker = result.NextMarker
    }
    return nil
}

func (a *AliyunProvider) List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error) {
    b, err := a.client.Bucket(bucket)
    if err != nil {
        return nil, err
    }
    result, err := b.ListObjects(oss.Prefix(remotePath))
    if err != nil {
        return nil, err
    }
    var files []RemoteFile
    for _, obj := range result.Objects {
        files = append(files, RemoteFile{
            Path: strings.TrimPrefix(obj.Key, remotePath),
            Size: obj.Size,
        })
    }
    return files, nil
}
```

**Step 3: Implement Tencent COS and Huawei OBS similarly in `core/backup/tencent.go` and `core/backup/huawei.go`** (same interface, different SDK calls — follow same pattern as AliyunProvider)

**Step 4: Register providers in `app/wails/providers.go`**

```go
package main

import (
    "github.com/shinerio/skillflow/core/backup"
    "github.com/shinerio/skillflow/core/registry"
)

func registerProviders() {
    registry.RegisterCloudProvider(backup.NewAliyunProvider())
    registry.RegisterCloudProvider(backup.NewTencentProvider())
    registry.RegisterCloudProvider(backup.NewHuaweiProvider())
}
```

**Step 5: Commit**

```bash
git add core/backup/ app/wails/providers.go
git commit -m "feat: add cloud backup interface + Aliyun/Tencent/Huawei providers"
```

---

## Phase 6: Update Checker

### Task 13: GitHub Update Checker

**Files:**
- Create: `core/update/checker.go`
- Create: `core/update/checker_test.go`

**Step 1: Write failing tests (using httptest)**

```go
package update_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/shinerio/skillflow/core/update"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCheckerDetectsUpdate(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode([]map[string]any{{"sha": "newsha123"}})
    }))
    defer srv.Close()

    checker := update.NewChecker(srv.URL)
    sk := &skill.Skill{
        Source:        skill.SourceGitHub,
        SourceURL:     "https://github.com/user/repo",
        SourceSubPath: "skills/skill-a",
        SourceSHA:     "oldsha456",
    }
    result, err := checker.Check(context.Background(), sk)
    require.NoError(t, err)
    assert.True(t, result.HasUpdate)
    assert.Equal(t, "newsha123", result.LatestSHA)
}

func TestCheckerNoUpdateWhenSHAMatches(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode([]map[string]any{{"sha": "sameSHA"}})
    }))
    defer srv.Close()

    checker := update.NewChecker(srv.URL)
    sk := &skill.Skill{
        Source:    skill.SourceGitHub,
        SourceSHA: "sameSHA",
    }
    result, err := checker.Check(context.Background(), sk)
    require.NoError(t, err)
    assert.False(t, result.HasUpdate)
}
```

**Step 2: Implement `core/update/checker.go`**

```go
package update

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "github.com/shinerio/skillflow/core/skill"
)

type CheckResult struct {
    SkillID   string
    HasUpdate bool
    LatestSHA string
}

type Checker struct {
    baseURL string
    client  *http.Client
}

func NewChecker(baseURL string) *Checker {
    if baseURL == "" {
        baseURL = "https://api.github.com"
    }
    return &Checker{baseURL: baseURL, client: http.DefaultClient}
}

func (c *Checker) Check(ctx context.Context, sk *skill.Skill) (CheckResult, error) {
    if !sk.IsGitHub() {
        return CheckResult{}, nil
    }
    owner, repo, subPath := parseSourceURL(sk.SourceURL, sk.SourceSubPath)
    url := fmt.Sprintf("%s/repos/%s/%s/commits?path=%s&per_page=1", c.baseURL, owner, repo, subPath)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := c.client.Do(req)
    if err != nil {
        return CheckResult{}, err
    }
    defer resp.Body.Close()

    var commits []struct{ SHA string `json:"sha"` }
    if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
        return CheckResult{}, err
    }
    latestSHA := commits[0].SHA
    return CheckResult{
        SkillID:   sk.ID,
        LatestSHA: latestSHA,
        HasUpdate: latestSHA != sk.SourceSHA,
    }, nil
}

func parseSourceURL(sourceURL, subPath string) (owner, repo, path string) {
    sourceURL = strings.TrimSuffix(sourceURL, "/")
    parts := strings.Split(sourceURL, "/")
    owner = parts[len(parts)-2]
    repo = parts[len(parts)-1]
    return owner, repo, subPath
}
```

**Step 3: Run tests**

```bash
go test ./core/update/... -v
```
Expected: PASS

**Step 4: Commit**

```bash
git add core/update/
git commit -m "feat: add GitHub SHA-based update checker"
```

---

## Phase 7: Wails App Layer

### Task 14: Wails App Methods

**Files:**
- Modify: `app/wails/app.go`
- Create: `app/wails/events.go`

**Step 1: Implement `app/wails/app.go`** (all methods exposed to frontend)

```go
package main

import (
    "context"
    "github.com/shinerio/skillflow/core/config"
    "github.com/shinerio/skillflow/core/install"
    "github.com/shinerio/skillflow/core/notify"
    "github.com/shinerio/skillflow/core/registry"
    "github.com/shinerio/skillflow/core/skill"
    "github.com/shinerio/skillflow/core/update"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
    ctx     context.Context
    hub     *notify.Hub
    storage *skill.Storage
    config  *config.Service
    checker *update.Checker
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
    a.checker = update.NewChecker("")
    registerAdapters()
    registerProviders()
    go forwardEvents(ctx, a.hub)
    go a.checkUpdatesOnStartup()
}

// --- Skills ---

func (a *App) ListSkills() ([]*skill.Skill, error) {
    return a.storage.ListAll()
}

func (a *App) ListCategories() ([]string, error) {
    return a.storage.ListCategories()
}

func (a *App) CreateCategory(name string) error {
    return a.storage.CreateCategory(name)
}

func (a *App) MoveSkillCategory(skillID, category string) error {
    return a.storage.MoveCategory(skillID, category)
}

func (a *App) DeleteSkill(skillID string) error {
    return a.storage.Delete(skillID)
}

// --- Install ---

func (a *App) ScanGitHub(repoURL string) ([]install.SkillCandidate, error) {
    inst := install.NewGitHubInstaller("")
    return inst.Scan(a.ctx, install.InstallSource{Type: "github", URI: repoURL})
}

func (a *App) InstallFromGitHub(repoURL string, candidates []install.SkillCandidate, category string) error {
    // For each candidate, download to tmp then import via storage
    inst := install.NewGitHubInstaller("")
    return inst.Install(a.ctx, install.InstallSource{Type: "github", URI: repoURL}, candidates, category)
}

func (a *App) ImportLocal(dir, category string) (*skill.Skill, error) {
    return a.storage.Import(dir, category, skill.SourceManual, "", "")
}

// --- Sync ---

type ConflictResolution string

const (
    ConflictOverwrite ConflictResolution = "overwrite"
    ConflictSkip      ConflictResolution = "skip"
)

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

func (a *App) ScanToolSkills(toolName string) ([]*skill.Skill, error) {
    cfg, _ := a.config.Load()
    for _, t := range cfg.Tools {
        if t.Name == toolName {
            adapter, ok := registry.GetAdapter(toolName)
            if !ok {
                adapter = newFilesystemAdapterFromConfig(t)
            }
            return adapter.Pull(a.ctx, t.SkillsDir)
        }
    }
    return nil, nil
}

// --- Config ---

func (a *App) GetConfig() (config.AppConfig, error) {
    return a.config.Load()
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
    return a.config.Save(cfg)
}

// --- Updates ---

func (a *App) CheckUpdates() error {
    skills, err := a.storage.ListAll()
    if err != nil {
        return err
    }
    for _, sk := range skills {
        result, err := a.checker.Check(a.ctx, sk)
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

func (a *App) checkUpdatesOnStartup() {
    _ = a.CheckUpdates()
}
```

**Step 2: Implement `app/wails/events.go`**

```go
package main

import (
    "context"
    "encoding/json"
    "github.com/shinerio/skillflow/core/notify"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

func forwardEvents(ctx context.Context, hub *notify.Hub) {
    ch := hub.Subscribe()
    for {
        select {
        case evt, ok := <-ch:
            if !ok {
                return
            }
            data, _ := json.Marshal(evt.Payload)
            runtime.EventsEmit(ctx, string(evt.Type), string(data))
        case <-ctx.Done():
            return
        }
    }
}
```

**Step 3: Commit**

```bash
git add app/wails/
git commit -m "feat: add Wails app layer with all method bindings and event forwarding"
```

---

## Phase 8: Frontend

### Task 15: Frontend Setup

**Files:**
- Modify: `frontend/src/main.tsx`
- Create: `frontend/src/store/index.ts`
- Create: `frontend/src/App.tsx`

**Step 1: Install frontend dependencies**

```bash
cd frontend
npm install zustand @radix-ui/react-dialog react-router-dom lucide-react
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

**Step 2: Configure Tailwind in `frontend/src/index.css`**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

**Step 3: Create main layout with sidebar navigation in `frontend/src/App.tsx`**

```tsx
import { BrowserRouter, Route, Routes, NavLink } from 'react-router-dom'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'

export default function App() {
  return (
    <BrowserRouter>
      <div className="flex h-screen bg-gray-950 text-gray-100">
        <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col p-4 gap-1">
          <h1 className="text-lg font-bold mb-6 px-2">SkillFlow</h1>
          <NavItem to="/" icon={<Package size={16} />} label="我的 Skills" />
          <p className="text-xs text-gray-500 px-2 mt-3 mb-1">同步管理</p>
          <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label="推送到工具" />
          <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label="从工具拉取" />
          <div className="flex-1" />
          <NavItem to="/backup" icon={<Cloud size={16} />} label="云备份" />
          <NavItem to="/settings" icon={<Settings size={16} />} label="设置" />
        </aside>
        <main className="flex-1 overflow-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/sync/push" element={<SyncPush />} />
            <Route path="/sync/pull" element={<SyncPull />} />
            <Route path="/backup" element={<Backup />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

function NavItem({ to, icon, label }: { to: string; icon: React.ReactNode; label: string }) {
  return (
    <NavLink
      to={to}
      end
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
          isActive ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800 hover:text-white'
        }`
      }
    >
      {icon}
      {label}
    </NavLink>
  )
}
```

**Step 4: Commit**

```bash
git add frontend/
git commit -m "feat: add frontend layout with sidebar navigation"
```

---

### Task 16: Dashboard Page

**Files:**
- Create: `frontend/src/pages/Dashboard.tsx`
- Create: `frontend/src/components/SkillCard.tsx`
- Create: `frontend/src/components/CategoryPanel.tsx`

**Step 1: Implement SkillCard component**

```tsx
// frontend/src/components/SkillCard.tsx
import { Github, FolderOpen, RefreshCw } from 'lucide-react'

interface Skill {
  id: string; name: string; category: string
  source: 'github' | 'manual'; hasUpdate: boolean
}

interface Props { skill: Skill; onDelete: () => void; onUpdate?: () => void }

export default function SkillCard({ skill, onDelete, onUpdate }: Props) {
  return (
    <div
      draggable
      onDragStart={e => e.dataTransfer.setData('skillId', skill.id)}
      className="relative bg-gray-800 border border-gray-700 rounded-xl p-4 cursor-grab hover:border-indigo-500 transition-colors group"
    >
      {skill.hasUpdate && (
        <span className="absolute top-2 right-2 w-2.5 h-2.5 rounded-full bg-red-500" />
      )}
      <div className="flex items-center gap-2 mb-2">
        {skill.source === 'github'
          ? <Github size={14} className="text-gray-400" />
          : <FolderOpen size={14} className="text-gray-400" />
        }
        <span className="text-xs text-gray-400">{skill.source}</span>
      </div>
      <p className="font-medium text-sm truncate">{skill.name}</p>
      <div className="mt-3 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
        {skill.hasUpdate && (
          <button onClick={onUpdate} className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1">
            <RefreshCw size={12} /> 更新
          </button>
        )}
        <button onClick={onDelete} className="text-xs text-red-400 hover:text-red-300 ml-auto">删除</button>
      </div>
    </div>
  )
}
```

**Step 2: Implement CategoryPanel with drag-drop**

```tsx
// frontend/src/components/CategoryPanel.tsx
interface Props {
  categories: string[]
  selected: string | null
  onSelect: (cat: string | null) => void
  onDrop: (skillId: string, category: string) => void
}

export default function CategoryPanel({ categories, selected, onSelect, onDrop }: Props) {
  const handleDragOver = (e: React.DragEvent) => e.preventDefault()
  const handleDrop = (e: React.DragEvent, cat: string) => {
    const id = e.dataTransfer.getData('skillId')
    if (id) onDrop(id, cat)
  }

  return (
    <div className="w-48 flex-shrink-0 border-r border-gray-800 p-3">
      <CategoryItem
        label="全部" isSelected={selected === null}
        onSelect={() => onSelect(null)}
        onDragOver={handleDragOver} onDrop={() => {}}
      />
      {categories.map(cat => (
        <CategoryItem
          key={cat} label={cat} isSelected={selected === cat}
          onSelect={() => onSelect(cat)}
          onDragOver={handleDragOver}
          onDrop={e => handleDrop(e, cat)}
        />
      ))}
    </div>
  )
}

function CategoryItem({ label, isSelected, onSelect, onDragOver, onDrop }: any) {
  return (
    <div
      onClick={onSelect}
      onDragOver={onDragOver}
      onDrop={onDrop}
      className={`px-3 py-2 rounded-lg text-sm cursor-pointer mb-0.5 transition-colors ${
        isSelected ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'
      }`}
    >
      {label}
    </div>
  )
}
```

**Step 3: Implement Dashboard page**

```tsx
// frontend/src/pages/Dashboard.tsx
import { useEffect, useState } from 'react'
import { ListSkills, ListCategories, MoveSkillCategory, DeleteSkill } from '../../wailsjs/go/main/App'
import CategoryPanel from '../components/CategoryPanel'
import SkillCard from '../components/SkillCard'

export default function Dashboard() {
  const [skills, setSkills] = useState<any[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCat, setSelectedCat] = useState<string | null>(null)

  const load = async () => {
    const [s, c] = await Promise.all([ListSkills(), ListCategories()])
    setSkills(s ?? [])
    setCategories(c ?? [])
  }

  useEffect(() => { load() }, [])

  const filtered = selectedCat ? skills.filter(s => s.Category === selectedCat) : skills

  const handleDrop = async (skillId: string, category: string) => {
    await MoveSkillCategory(skillId, category)
    load()
  }

  return (
    <div className="flex h-full">
      <CategoryPanel
        categories={categories}
        selected={selectedCat}
        onSelect={setSelectedCat}
        onDrop={handleDrop}
      />
      <div className="flex-1 p-6">
        <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
          {filtered.map(sk => (
            <SkillCard
              key={sk.ID}
              skill={{ id: sk.ID, name: sk.Name, category: sk.Category, source: sk.Source, hasUpdate: sk.HasUpdate }}
              onDelete={async () => { await DeleteSkill(sk.ID); load() }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
```

**Step 4: Commit**

```bash
git add frontend/src/
git commit -m "feat: add Dashboard with category panel and skill cards with drag-drop"
```

---

### Task 17: Sync Push and Pull Pages

**Files:**
- Create: `frontend/src/pages/SyncPush.tsx`
- Create: `frontend/src/pages/SyncPull.tsx`

(Implement per the design doc UI spec — tool multi-select, scope selector for push; tool select + scan + skill checklist for pull)

**Step 1: Implement SyncPush.tsx** (follows design doc layout exactly)

**Step 2: Implement SyncPull.tsx** (scan button → skill checklist → category select → pull)

**Step 3: Commit**

```bash
git add frontend/src/pages/
git commit -m "feat: add sync push and pull pages"
```

---

### Task 18: Backup and Settings Pages

**Files:**
- Create: `frontend/src/pages/Backup.tsx`
- Create: `frontend/src/pages/Settings.tsx`

**Step 1: Implement Backup.tsx** (last backup time, manual trigger button, file list, restore button)

**Step 2: Implement Settings.tsx** (three tabs: Tools, Cloud, General)
- Tools tab: built-in tools with enabled toggle + path override; custom tools CRUD
- Cloud tab: provider selector → dynamic credential fields from `RequiredCredentials()`
- General tab: storage dir, default import category

**Step 3: Commit**

```bash
git add frontend/src/pages/
git commit -m "feat: add backup and settings pages with dynamic cloud credential form"
```

---

## Phase 9: Build & CI

### Task 19: Build Configuration and GitHub Actions

**Files:**
- Create: `.github/workflows/build.yml`
- Modify: `wails.json`

**Step 1: Configure wails.json**

```json
{
  "schemaVersion": "2",
  "name": "SkillFlow",
  "outputfilename": "SkillFlow",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto"
}
```

**Step 2: Create `.github/workflows/build.yml`**

```yaml
name: Build
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: macos-latest
            arch: amd64
            name: macos-intel
          - os: macos-latest
            arch: arm64
            name: macos-apple-silicon
          - os: windows-latest
            arch: amd64
            name: windows

    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: wails build -platform ${{ matrix.arch == 'arm64' && 'darwin/arm64' || matrix.os == 'windows-latest' && 'windows/amd64' || 'darwin/amd64' }}
      - uses: actions/upload-artifact@v4
        with:
          name: skillflow-${{ matrix.name }}
          path: build/bin/
```

**Step 3: Final local build test**

```bash
wails build
```
Expected: Binary produced in `build/bin/`

**Step 4: Commit**

```bash
git add .github/ wails.json
git commit -m "chore: add GitHub Actions matrix build for macOS (intel+arm) and Windows"
```

---

## Summary

| Phase | Tasks | Key Deliverables |
|-------|-------|-----------------|
| 1 Foundation | 1-5 | Go module, Wails init, models, notify hub, config, registry |
| 2 Skill Mgmt | 6-7 | Validator, storage with categories and meta |
| 3 Install | 8-10 | Install interface, GitHub installer, local installer |
| 4 Sync | 11 | Filesystem adapter (shared by all tools), registry wiring |
| 5 Backup | 12 | Cloud provider interface + Aliyun/Tencent/Huawei |
| 6 Update | 13 | GitHub SHA-based update checker |
| 7 App Layer | 14 | All Wails method bindings + event forwarding |
| 8 Frontend | 15-18 | All 5 pages + drag-drop + dynamic forms |
| 9 Build | 19 | GitHub Actions matrix build |

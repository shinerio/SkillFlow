# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation Sync Rule — MANDATORY

**Any time a feature is added, changed, or removed, you MUST update the following files in the same commit:**

| File | What to update |
|------|---------------|
| `feature.md` | Add / edit / remove the corresponding section(s) in English. Update the "Last updated" date at the bottom. |
| `feature_zh.md` | Same changes in Chinese. Update the "最后更新" date at the bottom. |
| `README.md` | Update the Features table row(s) if the high-level description changes. |
| `README_zh.md` | Same in Chinese. |

**Rules:**
- A "feature change" includes: any new UI element (button, dialog, toggle, input), behavior change, removal of a control, new backend method callable from the frontend, and new event type.
- Do **not** leave the docs stale. Never commit a feature change without the corresponding doc update in the same commit.
- The feature files are the source of truth for UX details. README files only carry high-level summaries with links to the feature files.

## Commands

### Make targets (recommended)

```bash
make dev              # Run in dev mode (hot-reload for Go + frontend)
make build            # Build production binary
make test             # Run all Go tests
make tidy             # Sync Go module dependencies
make generate         # Regenerate TypeScript bindings after App method changes
make install-frontend # Install frontend npm dependencies
make clean            # Remove build artifacts
make help             # List all targets
```

### Development (manual)

```bash
# Run the app in dev mode (hot-reload for both Go and frontend)
~/go/bin/wails dev

# Build production binary
~/go/bin/wails build

# Regenerate TypeScript bindings after changing App struct methods
~/go/bin/wails generate module
```

### Go (backend)

```bash
# Run all tests
go test ./core/...

# Run tests for a single package
go test ./core/skill/...
go test ./core/update/...
go test ./core/git/...

# Run a single test function
go test ./core/skill/... -run TestSkillHasUpdate

# Sync dependencies after modifying go.mod
go mod tidy
```

### Frontend

```bash
cd frontend
npm install        # install dependencies
npm run dev        # Vite dev server (used by wails dev)
npm run build      # production build (output: frontend/dist/)
```

## Architecture

SkillFlow is a Wails v2 desktop app (Go 1.23, Wails v2.11.0). The Go backend exposes methods directly to the React frontend via Wails method bindings. There is **no REST API** — frontend calls Go methods as async functions.

### Key Design Decisions

- **`core/sync` package name conflicts with Go stdlib `sync`** — always import it with alias: `toolsync "github.com/shinerio/skillflow/core/sync"`
- **Wails bindings are auto-generated** — after adding/removing exported methods on `App`, run `wails generate module` to update `frontend/wailsjs/go/main/App.{js,d.ts}`
- **`package main` files at root** — `app.go`, `adapters.go`, `providers.go`, `events.go` are all `package main` alongside `main.go` because Wails requires the app struct in the same package as `main`
- **No REST API** — direct Wails method bindings; faster and simpler
- **UUID-based skills** — skills are identified by UUID, metadata stored in JSON sidecars
- **Filesystem adapters** — all built-in tools share the same `FilesystemAdapter` pattern
- **GitHub as source of truth** — update checker polls GitHub API, not local timestamps

### Data Storage Layout

```
~/.skillflow/
  skills/              ← SkillsStorageDir (configured)
    <category>/
      <skill-name>/    ← copied skill directory
        skill.md       ← main file with YAML frontmatter
        ...other files
  meta/                ← JSON sidecars (sibling of skills/)
    <uuid>.json        ← one per skill, contains Skill struct
  config.json          ← AppConfig (tools, cloud, proxy)
  star_repos.json      ← StarredRepo[] array
  cache/               ← temporary cloned repos for starred repos
    <cached-repo-dirs>/
```

Skills are identified by UUID. The `meta/` directory is always `filepath.Join(filepath.Dir(root), "meta")`.

### Backend Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `core/skill` | `Skill` model, `Storage` (CRUD + categories), `Validator` (skill.md check) |
| `core/config` | `AppConfig` model, `Service` (load/save JSON), `DefaultToolsDir()` per tool |
| `core/notify` | `Hub` (buffered channel pub/sub), `EventType` constants |
| `core/install` | `Installer` interface, `GitHubInstaller` (scan/download/SHA), `LocalInstaller` |
| `core/sync` | `ToolAdapter` interface, `FilesystemAdapter` (shared by all built-in tools) |
| `core/backup` | `CloudProvider` interface, Aliyun/Tencent/Huawei implementations |
| `core/update` | `Checker` (GitHub Commits API SHA comparison) |
| `core/registry` | Global maps for Installer/ToolAdapter/CloudProvider — registered at startup |
| `core/git` | Git clone/update, repo scanning for skills, starred repo storage |

### Key Data Models

#### Skill (`core/skill/model.go`)

```go
type Skill struct {
    ID            string     // UUID
    Name          string     // skill name (dir name)
    Path          string     // absolute path to skill directory
    Category      string     // user-defined category
    Source        SourceType // "github" | "manual"
    SourceURL     string     // GitHub repo URL for GitHub sources
    SourceSubPath string     // relative path within repo (e.g. "skills/my-skill")
    SourceSHA     string     // installed commit SHA (from GitHub)
    LatestSHA     string     // detected newer SHA (for update checking)
    InstalledAt   time.Time
    UpdatedAt     time.Time
    LastCheckedAt time.Time
}

const (
    SourceGitHub SourceType = "github"
    SourceManual SourceType = "manual"
)
```

#### AppConfig (`core/config/model.go`)

```go
type ToolConfig struct {
    Name     string   // e.g. "claude-code", "opencode", "codex", "gemini-cli", "openclaw"
    ScanDirs []string // directories to scan for existing skills
    PushDir  string   // default directory to push skills to
    Enabled  bool
    Custom   bool     // true if user-added via Settings
}

type CloudConfig struct {
    Provider    string            // "aliyun", "tencent", "huawei"
    Enabled     bool
    BucketName  string
    RemotePath  string            // e.g. "skillflow/"
    Credentials map[string]string // provider-specific credentials
}

type ProxyConfig struct {
    Mode   ProxyMode // "none" | "system" | "manual"
    URL    string    // used when Mode == "manual"
}

type AppConfig struct {
    SkillsStorageDir string        // default: ~/.skillflow/skills
    DefaultCategory  string        // default: "Default"
    Tools            []ToolConfig
    Cloud            CloudConfig
    Proxy            ProxyConfig
}
```

#### StarredRepo (`core/git/model.go`)

```go
type StarredRepo struct {
    URL       string    // user-provided git repo URL
    Name      string    // parsed "owner/repo"
    Source    string    // canonical key "<host>/<path>"
    LocalDir  string    // cache directory on disk
    LastSync  time.Time
    SyncError string
}

type StarSkill struct {
    Name     string
    Path     string   // absolute local path to skill dir
    SubPath  string   // relative path in repo
    RepoURL  string
    RepoName string
    Source   string
    Imported bool     // already in My Skills?
}
```

### Startup Flow

`main.go` → `app.startup()`:
1. Load app data directory
2. Initialize `config.Service`, load config
3. Create `skill.Storage` with configured `SkillsStorageDir`
4. Call `registerAdapters()` (5 built-in tools → `FilesystemAdapter`)
5. Call `registerProviders()` (Aliyun, Tencent, Huawei)
6. Start `forwardEvents(ctx, hub)` goroutine — subscribes to Hub, emits each event via `runtime.EventsEmit`
7. Start `checkUpdatesOnStartup()` goroutine — scan skills for GitHub updates
8. Start `updateStarredReposOnStartup()` goroutine — sync starred repos

### Main App Struct

`app.go` (`package main`) contains the `App` struct and all exported methods:

```go
type App struct {
    ctx         context.Context
    hub         *notify.Hub           // event pub/sub
    storage     *skill.Storage        // skill CRUD
    config      *config.Service       // config persistence
    starStorage *coregit.StarStorage  // starred repos JSON persistence
    cacheDir    string                // ~/.skillflow/cache/
}
```

**Key exported methods (50+) — all callable from frontend:**

| Category | Methods |
|----------|---------|
| Skills | `ListSkills()`, `ListCategories()`, `DeleteSkill()`, `MoveSkillCategory()` |
| Import | `ScanGitHub()`, `InstallFromGitHub()`, `ImportLocal()` |
| Sync | `GetEnabledTools()`, `ScanToolSkills()`, `PushToTools()`, `PullFromTool()` |
| Config | `GetConfig()`, `SaveConfig()`, `AddCustomTool()`, `RemoveCustomTool()` |
| Backup | `BackupNow()`, `ListCloudFiles()`, `RestoreFromCloud()`, `ListCloudProviders()` |
| Updates | `CheckUpdates()`, `UpdateSkill()` |
| Starred repos | `AddStarredRepo()`, `ListAllStarSkills()`, `ImportStarSkills()`, `UpdateAllStarredRepos()` |
| UI helpers | `OpenFolderDialog()`, `OpenPath()` |

Auto-backup (`autoBackup()`) is triggered after mutations (delete, import, push, pull) when cloud backup is enabled.

### Event System

Backend → Frontend events flow through `core/notify.Hub`:
- Backend publishes via `hub.Publish(notify.Event{Type: ..., Payload: ...})`
- `forwardEvents()` goroutine subscribes to Hub, marshals `Payload` to JSON, and calls `runtime.EventsEmit(ctx, eventType, jsonData)`
- Frontend subscribes via `EventsOn('backup.progress', handler)` from `wailsjs/runtime/runtime`

Event types are defined in `core/notify/model.go`:

```go
const (
    EventBackupStarted    EventType = "backup.started"
    EventBackupProgress   EventType = "backup.progress"
    EventBackupCompleted  EventType = "backup.completed"
    EventBackupFailed     EventType = "backup.failed"
    EventSyncCompleted    EventType = "sync.completed"
    EventUpdateAvailable  EventType = "update.available"
    EventSkillConflict    EventType = "skill.conflict"
    EventStarSyncProgress EventType = "star.sync.progress"
    EventStarSyncDone     EventType = "star.sync.done"
)
```

The Hub uses a buffered channel (size 32) with drop-oldest behavior for slow subscribers.

### Tool Adapters

All 5 built-in tools use `FilesystemAdapter` from `core/sync`. Default push directories per tool:

| Tool | Default Push Directory |
|------|----------------------|
| `claude-code` | `~/.claude/skills` |
| `opencode` | `~/.config/opencode/skills` |
| `codex` | `~/.agents/skills` |
| `gemini-cli` | `~/.gemini/skills` |
| `openclaw` | `~/.openclaw/skills` |

**Adapter behavior:**
- `Pull()` — recursively scan directory tree for `skill.md` files, import each as a skill
- `Push()` — copy skill directories flat (no category subdir) into the target directory

Custom tools added via Settings also use `FilesystemAdapter` with user-provided directory.

### Installer Interface (`core/install`)

```go
type Installer interface {
    Type() string
    Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
    Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
}
```

- `GitHubInstaller` — scans GitHub repos via Contents API, downloads skill directories, records commit SHA
- `LocalInstaller` — imports from local filesystem path

### Cloud Provider Interface (`core/backup`)

```go
type CloudProvider interface {
    Name() string
    Init(credentials map[string]string) error
    Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error
    Restore(ctx context.Context, bucket, remotePath, localDir string) error
    List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error)
    RequiredCredentials() []CredentialField
}
```

The Settings page automatically renders credential input fields from `RequiredCredentials()`.

### Git Package (`core/git`)

Handles starred repo workflows:
- `CloneOrUpdate(ctx, repoURL, localDir, proxyURL)` — git clone or fetch+pull
- `ScanSkills(localDir, repoURL, repoName, source)` — find skill dirs in cloned repo
- `GetSubPathSHA(ctx, repoDir, subPath)` — get latest commit SHA for a path
- `ParseRepoRef()`, `ParseRepoName()`, `RepoSource()` — URL parsing utilities
- `StarStorage` — JSON persistence for `[]StarredRepo` at `~/.skillflow/star_repos.json`

### Adding a New Cloud Provider

1. Create `core/backup/<name>.go` implementing `backup.CloudProvider`
2. Register in `providers.go`: `registry.RegisterCloudProvider(NewXxxProvider())`
3. The Settings page automatically renders credential fields from `RequiredCredentials()`

### Adding a New Tool Adapter

If the tool uses a flat directory of skills (standard), just add it to `registerAdapters()` in `adapters.go`. For custom behavior, implement `toolsync.ToolAdapter` and register via `registry.RegisterAdapter()`.

### Adding a New App Method (Frontend-callable)

1. Add exported method to `App` struct in `app.go` (or a new `package main` file at root)
2. Run `make generate` (or `wails generate module`) to update `frontend/wailsjs/go/main/App.{js,d.ts}`
3. Import and call from frontend: `import { MyNewMethod } from '../../wailsjs/go/main/App'`

### Frontend Structure

```
frontend/src/
  App.tsx              ← BrowserRouter + sidebar layout + route definitions
  pages/               ← one file per route
    Dashboard.tsx      ← My Skills listing (categories, search, drag-drop)
    SyncPush.tsx       ← Push skills to external tools
    SyncPull.tsx       ← Pull skills from external tools
    StarredRepos.tsx   ← Browse and import from starred/watched repos
    Backup.tsx         ← Cloud backup management
    Settings.tsx       ← Tool config, cloud provider, proxy settings
  components/          ← shared UI components
    SkillCard.tsx      ← Individual skill display card
    SkillTooltip.tsx   ← Hover tooltips showing skill metadata
    CategoryPanel.tsx  ← Category sidebar/filter
    GitHubInstallDialog.tsx  ← GitHub repo scanner UI
    ConflictDialog.tsx ← Handle skill name conflicts on sync
    SyncSkillCard.tsx  ← Skill card for sync pages
    ContextMenu.tsx    ← Right-click context menus
  config/
    toolIcons.tsx      ← Tool name → icon mapping
  wailsjs/             ← auto-generated (do not edit manually)
    go/main/App.js     ← Go method bindings
    go/main/App.d.ts   ← TypeScript type declarations
    runtime/runtime.js ← Wails runtime (EventsOn, EventsEmit, etc.)
```

Frontend calls Go methods directly: `import { ListSkills } from '../../wailsjs/go/main/App'`. Go struct field names are PascalCase in JSON (e.g. `cfg.Tools`, `t.SkillsDir`, `cfg.Cloud.Enabled`).

**Frontend tech stack:** React 18, TypeScript, React Router v7, Tailwind CSS, Lucide React icons, Radix UI dialogs.

### Testing Approach

Tests use `httptest.NewServer` to mock GitHub API calls. Pass the mock server URL to `NewChecker(srv.URL)` or `NewGitHubInstaller(srv.URL)`. Filesystem tests use `t.TempDir()`.

**Test coverage by package:**

| Package | Test files | Notes |
|---------|-----------|-------|
| `core/skill` | `model_test.go`, `storage_test.go`, `validator_test.go` | Full coverage |
| `core/config` | `service_test.go` | Full coverage |
| `core/notify` | `hub_test.go` | Full coverage |
| `core/install` | `github_test.go`, `local_test.go` | Mocked GitHub API |
| `core/update` | `checker_test.go` | Mocked GitHub API |
| `core/sync` | `filesystem_adapter_test.go` | TempDir filesystem tests |
| `core/git` | `client_test.go`, `scanner_test.go`, `storage_test.go` | TempDir + mock |
| `core/backup` | none | Requires real cloud credentials |
| `core/registry` | none | Thin wrapper, tested via integration |

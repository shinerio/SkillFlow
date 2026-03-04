# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development

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

SkillFlow is a Wails v2 desktop app. The Go backend exposes methods directly to the React frontend via Wails method bindings. There is **no REST API** — frontend calls Go methods as async functions.

### Key Design Decisions

- **`core/sync` package name conflicts with Go stdlib `sync`** — always import it with alias: `toolsync "github.com/shinerio/skillflow/core/sync"`
- **Wails bindings are auto-generated** — after adding/removing exported methods on `App`, run `wails generate module` to update `frontend/wailsjs/go/main/App.{js,d.ts}`
- **`package main` files at root** — `app.go`, `adapters.go`, `providers.go`, `events.go` are all `package main` alongside `main.go` because Wails requires the app struct in the same package as `main`

### Data Storage Layout

```
~/.skillflow/
  skills/              ← SkillsStorageDir (configured)
    <category>/
      <skill-name>/    ← copied skill directory
  meta/                ← JSON sidecars (sibling of skills/)
    <uuid>.json        ← one per skill, contains Skill struct
  config.json
```

Skills are identified by UUID. The `meta/` directory is always `filepath.Join(filepath.Dir(root), "meta")`.

### Backend Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `core/skill` | `Skill` model, `Storage` (CRUD + categories), `Validator` (SKILLS.md check) |
| `core/config` | `AppConfig` model, `Service` (load/save JSON), `DefaultToolsDir()` per tool |
| `core/notify` | `Hub` (buffered channel pub/sub), `EventType` constants |
| `core/install` | `Installer` interface, `GitHubInstaller` (scan/download/SHA), `LocalInstaller` |
| `core/sync` | `ToolAdapter` interface, `FilesystemAdapter` (shared by all 5 built-in tools) |
| `core/backup` | `CloudProvider` interface, Aliyun/Tencent/Huawei implementations |
| `core/update` | `Checker` (GitHub Commits API SHA comparison) |
| `core/registry` | Global maps for Installer/ToolAdapter/CloudProvider — registered at startup |

### Startup Flow

`main.go` → `app.startup()`:
1. Loads config (`config.Service.Load()`)
2. Creates `skill.Storage` with configured `SkillsStorageDir`
3. Calls `registerAdapters()` (5 built-in tools → `FilesystemAdapter`)
4. Calls `registerProviders()` (Aliyun, Tencent, Huawei)
5. Starts `forwardEvents(ctx, hub)` goroutine — subscribes to Hub, emits each event via `runtime.EventsEmit`
6. Starts `checkUpdatesOnStartup()` goroutine

### Event System

Backend → Frontend events flow through `core/notify.Hub`:
- Backend publishes via `hub.Publish(notify.Event{Type: ..., Payload: ...})`
- `forwardEvents()` marshals `Payload` to JSON and calls `runtime.EventsEmit(ctx, eventType, jsonData)`
- Frontend subscribes via `EventsOn('backup.progress', handler)` from `wailsjs/runtime/runtime`

Event types are defined in `core/notify/model.go` as string constants (e.g. `"backup.started"`, `"update.available"`).

### Adding a New Cloud Provider

1. Create `core/backup/<name>.go` implementing `backup.CloudProvider`
2. Register in `providers.go`: `registry.RegisterCloudProvider(NewXxxProvider())`
3. The Settings page automatically renders credential fields from `RequiredCredentials()`

### Adding a New Tool Adapter

If the tool uses a flat directory of skills (standard), just add it to `registerAdapters()` in `adapters.go`. For custom behavior, implement `toolsync.ToolAdapter` and register via `registry.RegisterAdapter()`.

### Frontend Structure

```
frontend/src/
  App.tsx              ← BrowserRouter + sidebar layout + route definitions
  pages/               ← one file per route
  components/          ← shared UI components
  wailsjs/             ← auto-generated (do not edit manually)
    go/main/App.js     ← Go method bindings
    runtime/runtime.js ← Wails runtime (EventsOn, EventsEmit, etc.)
```

Frontend calls Go methods directly: `import { ListSkills } from '../../wailsjs/go/main/App'`. Go struct field names are PascalCase in JSON (e.g. `cfg.Tools`, `t.SkillsDir`, `cfg.Cloud.Enabled`).

### Testing Approach

Tests use `httptest.NewServer` to mock GitHub API calls. Pass the mock server URL to `NewChecker(srv.URL)` or `NewGitHubInstaller(srv.URL)`. Filesystem tests use `t.TempDir()`.

The `core/backup` and `core/registry` packages have no test files — they require real cloud credentials or are thin wrappers tested via integration.

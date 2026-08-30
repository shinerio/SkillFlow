# AGENTS.md

This file provides guidance to coding agents when working with this repository.

## Directory Organization Rule - MANDATORY

The root directory must contain **no Go source files**. All code lives in clearly scoped subdirectories:

```text
/ (project root - no .go files here)
  go.mod, go.sum
  Makefile
  AGENTS.md (→ CLAUDE.md symlink)
  README.md, README_zh.md
  contributing.md, contributing_zh.md
  LICENSE, .gitignore, .github/
  changelog/
  chats/
  docs/
    agents/
      skill_directory.md
      memory_directory.md
    architecture/
      README.md
      README_zh.md
      overview.md
      overview_zh.md
      contexts.md
      contexts_zh.md
      layers.md
      layers_zh.md
      use-cases.md
      use-cases_zh.md
      runtime-and-storage.md
      runtime-and-storage_zh.md
    features.md
    features_zh.md
    config.md
    config_zh.md
    plans/
    superpowers/
  core/
    config/
    platform/
    shared/
    orchestration/
    readmodel/
    skillcatalog/
      app/
      domain/
      infra/
    promptcatalog/
      app/
      domain/
      infra/
    agentintegration/
      app/
      domain/
      infra/
    skillsource/
      app/
      domain/
      infra/
    backup/
      app/
      domain/
      infra/
  native/
    macos/
    windows/
    scripts/
  build/
    native/
  cmd/
    skillflow/
      main.go
      app.go, app_*.go
      adapters.go, providers.go, events.go, version.go
      process_*.go
      tray_*.go
      window_*.go
      single_instance_*.go
      wails.json
      build/
      frontend/
```

**Rules:**
- Never add `.go` files to the project root.
- New backend business code must go under a bounded context in `core/<context>/app`, `core/<context>/domain`, or `core/<context>/infra`.
- Cross-context write coordination belongs in `core/orchestration/`.
- Cross-context read composition belongs in `core/readmodel/`.
- Pure technical capabilities with no business ownership belong in `core/platform/`.
- `core/config/` is a frontend-facing settings facade and split/merge persistence adapter. Do not treat it as a bounded context.
- `core/shared/` is only for highly stable shared kernel concepts. Do not move context-local IDs or business rules there unless they are genuinely cross-context.
- `cmd/skillflow/` remains the Go daemon composition root, process host, legacy Wails shell, and transport adapter layer.
- `native/` contains platform-native clients: macOS Swift uses SwiftPM, and Windows WinUI uses .NET; native build outputs go under `build/native/`.
- `wails.json` must stay co-located with `frontend/` inside `cmd/skillflow/`.
- The `//go:embed all:frontend/dist` directive in `main.go` works because both are in `cmd/skillflow/`.
- `go test ./core/...` is run from the module root.
- Import paths use the full module path: `github.com/shinerio/skillflow/core/...`.
- **`cmd/skillflow/*.go` files must remain flat.** Legacy Wails bindings require a single `package main` directory, so do not create subdirectories under `cmd/skillflow/`.
- Use file-name prefixes inside `cmd/skillflow/` as the organization convention:
  - `app.go`, `app_*.go` for legacy Wails-facing transport methods
  - `process_daemon_only.go` for the native client daemon-only process role
  - `events.go` for shell event types and emitters
  - `adapters.go`, `providers.go` for shell-side wiring
  - `process_*.go`, `tray_*.go`, `window_*.go`, `single_instance_*.go` for shell/runtime concerns
- When a concern grows large enough to warrant its own package, extract it to `core/` rather than creating a subdirectory inside `cmd/skillflow/`.

## Documentation Organization Rule - MANDATORY

**Root directory may contain only `README.md`, `README_zh.md`, `contributing.md`, `contributing_zh.md`, `REVIEW.md`, and `AGENTS.md` (with `CLAUDE.md` as symlink) as documentation files.**

All other documentation lives under `docs/`:

| File | Purpose |
|------|---------|
| `docs/agents/skill_directory.md` | Built-in agent scan/push directory reference |
| `docs/agents/memory_directory.md` | Common agent memory-file reference |
| `docs/features.md` | Complete UI/UX feature reference in English |
| `docs/features_zh.md` | Complete UI/UX feature reference in Chinese |
| `docs/architecture/README.md` | Architecture index and reading order (English) |
| `docs/architecture/README_zh.md` | Architecture index and reading order (Chinese) |
| `docs/architecture/overview.md` | High-level architecture overview (English) |
| `docs/architecture/overview_zh.md` | High-level architecture overview (Chinese) |
| `docs/architecture/` | Detailed DDD architecture set: overview, contexts, layers, use cases, and storage |
| `docs/config.md` | Persisted config and metadata file reference (English) |
| `docs/config_zh.md` | Same in Chinese |
| `docs/plans/` | Design and implementation plans |
| `docs/<module>/...` | Module-scoped design notes, plans, and reference docs such as `docs/superpowers/` |

**Rules:**
- `README.md` / `README_zh.md` are user-facing only: features overview, download/install links, skill format, cloud backup config, contributing/build instructions. No internal code snippets, no package tables, no architecture diagrams.
- `contributing.md` / `contributing_zh.md` are contributor-facing only: prerequisites, build/test/generate commands, and pointers to deeper architecture docs.
- Never add new standalone `.md` files to the root directory beyond `README.md`, `README_zh.md`, `contributing.md`, `contributing_zh.md`, `REVIEW.md`, and `AGENTS.md`/`CLAUDE.md`. If you need other documentation, put it under `docs/`.

## Documentation Sync Rule - MANDATORY

**Any time a meaningful user-facing feature is added, changed, or removed, you MUST update the following files in the same commit:**

| File | What to update |
|------|---------------|
| `docs/features.md` | Add, edit, or remove the corresponding section(s) in English. Update the "Last updated" date at the bottom. |
| `docs/features_zh.md` | Same changes in Chinese. Update the "最后更新" date at the bottom. |
| `README.md` | Update only when the high-level description changes or an existing Highlights row needs a coarse-grained capability update. |
| `README_zh.md` | Same in Chinese. |

**Rules:**
- A "feature change" means a clear user-facing capability or usage-flow change: a new UI control that enables a new action, a behavior change, removal of a control, a new backend method callable from the frontend, or a new event type.
- Pure presentation-only adjustments are **not** feature changes for this rule. Do not update `docs/features.md`, `docs/features_zh.md`, `README.md`, or `README_zh.md` for small copy tweaks, label renames, icon swaps, visual reordering, spacing/layout polish, or other changes that do not change how the feature is used.
- Do not leave the docs stale. Never commit a feature change without the corresponding doc update in the same commit.
- `docs/features.md` / `docs/features_zh.md` are the source of truth for UX details, but only for meaningful UX and feature behavior.
- `README.md` / `README_zh.md` must stay high-level and only describe coarse-grained product capabilities.
- Small user-facing helpers inside existing flows, narrow diagnostics, settings-page utilities, and similarly scoped UX additions should normally update `docs/features.md` / `docs/features_zh.md` only, and should not touch `README.md` / `README_zh.md` unless they materially change a top-level product capability.
- If a frontend or backend feature change also changes architecture, module boundaries, cross-module contracts, core data flow, persistence model, or extension points, update the relevant files under `docs/architecture/` in the same commit. Keep `docs/architecture/README.md` and `docs/architecture/README_zh.md` aligned as the architecture index and reading entry.
- If a change needs detailed module-level design or implementation documentation, create or update files under a coarse-grained module folder in `docs/` such as `docs/<module>/...` instead of adding more root-level markdown files.

## Configuration Documentation Sync Rule - MANDATORY

**Any time a repo-tracked persisted config or metadata file is added, removed, renamed, or its on-disk schema or semantics change, you MUST update the following files in the same commit:**

| File | What to update |
|------|---------------|
| `docs/config.md` | Update the English examples, key tables, file-name notes, path and sync-scope notes, and any added, removed, or renamed fields. |
| `docs/config_zh.md` | Same changes in Chinese. |

**Rules:**
- This rule applies to persisted config and metadata files such as `config.json`, `config_local.json`, `star_repos.json`, `meta/*.json`, and future repo-tracked files that document on-disk configuration or sync metadata.
- Update the docs when a field is added, removed, renamed, split, merged, moved between files, changes meaning, changes persistence scope, or when the canonical file name or path changes.
- Keep examples aligned with the actual persisted format on disk, not just the in-memory runtime model.
- Pure formatting-only changes that do not change persisted keys, file names, storage location, or semantics do not require doc churn.
- Do not merge config or schema changes while `docs/config.md` and `docs/config_zh.md` are stale.

## Upgrade Cutover Rule - MANDATORY

Any persisted schema, terminology, or config-structure upgrade that changes repo-tracked on-disk data must use an explicit startup cutover rather than long-term backward compatibility in business code.

**Rules:**
- Put upgrade and cutover code under `core/platform/upgrade/`.
- Run the cutover automatically during startup **before** loading config or other persisted business data.
- Rewrite persisted files in place to the latest schema and terminology so the rest of the codebase reads only the new format.
- Do **not** keep business-layer compatibility branches for the old on-disk schema after the cutover is in place.
- When an upgrade rewrites config or metadata files, update `docs/config.md` and `docs/config_zh.md` in the same commit.
- When the upgrade changes user-facing terminology or flows, also update `README.md`, `README_zh.md`, `docs/features.md`, `docs/features_zh.md`, and the relevant architecture docs in the same commit.

## Path Persistence Rule - MANDATORY

Any repo-tracked file that can be backed up or synced across devices must avoid machine-specific absolute paths.

- Synced files such as `config.json`, `meta/*.json`, `star_repos.json`, and future backup or sync data files must store local filesystem paths as **forward-slash relative paths** whenever the target is inside the synchronized root.
- The synchronized root is normally `config.AppDataDir()`. When `SkillsStorageDir` is moved outside that directory, treat the shared parent of `skills/` and `meta/` as the synchronized root for persisted skill metadata.
- Any path that points **outside** the synchronized root is platform-specific and must live only in `config_local.json`.
- `config_local.json` is local-only and must remain excluded from cloud backup and git sync.
- Runtime APIs may expand persisted relative paths back to absolute paths before returning them to frontend or backend callers, but the on-disk synced representation must stay relative.

## Logging Rule - MANDATORY

All backend code changes must follow consistent logging standards for troubleshooting.

### Log level policy

- `error`:
  - Required for any failed operation, exception, unexpected branch, or external dependency failure.
- `info`:
  - Required for important flow milestones (`start` / `completed`) of key operations.
- `debug`:
  - For detailed diagnostics and branch-level context, and must be suppressible by configured log level.

### Key operations that MUST log

The following operations must have reasonable logs, at minimum `info` on start and success, and `error` on failure:

- Git operations:
  - clone, fetch, pull, push, conflict detection and resolution, reset, force update
- API operations:
  - external API calls such as GitHub, cloud providers, and remote services, especially failures
- Sync operations:
  - scan, import, update, push, pull, backup, restore
- Resource mutations:
  - create, delete, rename, move, overwrite
- Config mutations:
  - settings save, log-level changes, provider or agent config updates

### Message quality requirements

- Log messages should include:
  - operation name
  - target or resource identifier such as skill id or name, repo url or name, agent or provider, or path
  - result status: `started`, `completed`, or `failed`
  - failure reason for `error` logs
- Keep wording stable and searchable across the same operation.
- Avoid noisy or duplicated logs and avoid logging every trivial getter.

### Security requirements

- Never log secrets or sensitive data:
  - access token, password, secret key, credential raw content, authorization header, cookie
- If needed for diagnosis, log only masked or non-sensitive metadata.

### Rotation and file-size rule

- Log file strategy must remain bounded:
  - keep only 2 files: `skillflow.log` and `skillflow.log.1`
  - max 1MB per file
  - rotate and overwrite the oldest file when the size limit is reached

## Python Tooling Rule - MANDATORY

Any Python-related work in this repository must use `uv` for interpreter management, dependency management, and script execution, while preserving functional correctness.

**Rules:**
- Never invoke system `python`, `python3`, `pip`, or `pip3` directly. If the interpreter itself is needed, run it via `uv run python ...`.
- Never install Python packages into the system environment.
- Prefer `uv run` for repo scripts, inline scripts, and module execution.
- When converting an existing Python command to `uv`, preserve the original entrypoint, arguments, working directory, environment variables, stdin/stdout behavior, and required Python version.
- Prefer `uv run python path/to/script.py` for script files, `uv run python - <<'PY'` for inline snippets, and `uv run -m <module>` for module-style commands.
- Use `uv add`, `uv remove`, and `uv sync` to manage project dependencies.
- Use `uv run --with <package>` or `uvx <tool>` for temporary or one-off tools instead of touching the system environment.
- If a documented or scripted Python workflow currently uses direct Python or pip invocation, convert it to the equivalent `uv` workflow before running it.
- Do not trade correctness for tooling purity: choose the `uv` invocation that faithfully reproduces the intended behavior.

## Commands

### Make targets (recommended)

```bash
make dev                     # Build and open the native macOS client
make dev-legacy              # Run the legacy Wails UI with hot reload
make build                   # Build the native client for the current machine
make build-native-macos      # Build SkillFlow.app with bundled skillflowd
make build-native-windows    # Build the WinUI client with bundled skillflowd.exe
make build-legacy            # Build the legacy Wails production binary
make test                    # Run Go tests and native build checks
make test-native             # Run the full current-platform native tests
make tidy                    # Sync Go module dependencies
make generate                # Regenerate legacy Wails TypeScript bindings
make install-frontend        # Install legacy Wails frontend npm dependencies
make clean                   # Remove build artifacts
make help                    # List all targets
```

### Development (manual)

```bash
# Build and run the native macOS app
make build-native-macos
open build/native/macos/SkillFlow.app

# Run the legacy Wails app in dev mode (hot-reload for both Go and frontend)
cd cmd/skillflow && ~/go/bin/wails dev

# Build the legacy Wails app
cd cmd/skillflow && ~/go/bin/wails build

# Regenerate legacy Wails TypeScript bindings after changing App struct methods
cd cmd/skillflow && ~/go/bin/wails generate module
```

### Go (backend)

```bash
# Run all core tests (from project root)
go test ./core/...

# Run shell tests when cmd/skillflow is touched
go test ./cmd/skillflow

# Run a single test name across core packages
go test ./core/... -run TestName

# Sync dependencies after modifying go.mod
go mod tidy
```

### Frontend

```bash
cd cmd/skillflow/frontend
npm install        # install dependencies
npm run dev        # Vite dev server (used by wails dev)
npm run build      # production build (output: cmd/skillflow/frontend/dist/)
```

## Architecture

SkillFlow is a native macOS (Swift/SwiftUI) and Windows (WinUI) desktop app backed by a shared Go daemon (Go 1.25). Native clients communicate through the local authenticated daemon API and are packaged beside `skillflowd` / `skillflowd.exe`. The Wails/React app remains the explicit legacy fallback during migration. There is **no public REST API**.

The target backend architecture is a DDD-oriented modular monolith:

- `cmd/skillflow/` is the Go daemon composition root, process host, legacy Wails shell, and transport adapter layer
- `native/macos/` and `native/windows/` contain the default native clients
- bounded contexts live under `core/`
- each bounded context is organized as `app`, `domain`, and `infra`
- cross-context write coordination goes through `core/orchestration/`
- cross-context read composition goes through `core/readmodel/`
- `core/config/` is a frontend-facing settings facade
- pure technical capabilities live in `core/platform/`
- only highly stable shared kernel concepts live in `core/shared/`

For comprehensive architecture docs, data models, and extension guides, see **[docs/architecture/README.md](docs/architecture/README.md)** and the detailed documents under `docs/architecture/`.

## Cross-Module Skill Identity Rule - MANDATORY

Any change touching skill identity, install/import/push/pull state, starred repo correlation, agent scan correlation, or skill update badges **must** follow the **"Unified Skill Identity and State Model"** section in `docs/architecture/contexts.md` and `docs/architecture/contexts_zh.md`.

- Distinguish **instance identity** (`SkillID`) from **logical identity** (`LogicalSkillKey`).
- Do **not** use `Name` or absolute `Path` as the primary cross-context identity.
- For repository-backed skills, `LogicalSkillKey` is derived from `SkillSourceRef` using the canonical form `git:<repo-source>#<subpath>` and should not be persisted as a separate field when `SkillSourceRef` already exists.
- For manually imported or non-repository-backed skills, generate `LogicalSkillKey` at import time as a stable content-based key such as `content:<hash>`, computed from a canonicalized content snapshot rather than absolute paths or local-only metadata.
- `imported` is the external-source wording alias for `installed`.
- `pushed` means the logical skill exists in an agent's configured `PushDir`.
- `seenInAgentScan` means the logical skill was detected in an agent's configured `ScanDirs`; it does **not** imply SkillFlow previously pushed it.
- Git-backed update detection must be keyed by normalized repo source plus subpath, and compare installed source state against the latest known remote state for that same logical source.
- One logical source identified by `repo + subpath` should correspond to exactly one installed skill in the library. Re-importing the same logical source must be treated as an already-installed conflict or update path, not as a second installed instance.

### Key Design Decisions

- Legacy Wails-bound transport adapters remain in `cmd/skillflow/` because bindings require a single `package main` directory.
- Native clients call the stable JSON daemon API; they must not directly read or write business data files.
- `cmd/skillflow/App` methods should stay thin and delegate to one bounded context, `core/orchestration/`, `core/readmodel/`, or `core/config`.
- Bounded contexts are `skillcatalog`, `promptcatalog`, `agentintegration`, `skillsource`, and `backup`.
- `Skill` and `Prompt` are parallel core business concepts.
- `Settings`, `Dashboard`, and `My Agents` are composed UI read surfaces, not bounded contexts.
- `core/config` is a settings facade for transport and shell coordination, not a source-of-truth bounded context.
- Shell concerns such as tray, window state, launch-at-login, single-instance behavior, and app update stay in `cmd/skillflow/` and `core/platform/`.
- In `skillsource`, `StarRepo` is the repository-level model and `SkillSource` is the skill-level source model identified by `repo + subpath`.
- The long-term settings model is context-owned configuration namespaces stored through a platform settings store, not one global domain config object.
- Wails bindings are auto-generated. After adding or removing exported methods on `App`, run `make generate` to update `cmd/skillflow/frontend/wailsjs/go/main/App.{js,d.ts}`. Native API contract changes must be covered by Go daemon tests.

### Adding a New Legacy Wails App Method (Frontend-callable)

1. Add an exported method to `App` in `cmd/skillflow/app.go` or another flat `package main` file under `cmd/skillflow/`.
2. Keep that method as a thin transport adapter: validate inputs, convert DTOs, and delegate to one bounded context application service, `core/orchestration/`, `core/readmodel/`, or `core/config`.
3. Run `make generate` or `cd cmd/skillflow && wails generate module` to update `cmd/skillflow/frontend/wailsjs/go/main/App.{js,d.ts}`.
4. Import and call it from the frontend via `../../wailsjs/go/main/App`.

### Adding a New Cloud Provider

1. Implement the provider adapter under `core/backup/infra/`.
2. Expose it through the backup context's ports and wire it in the shell composition root under `cmd/skillflow/`.
3. Keep provider-specific API integration inside backup infrastructure code, not in Wails transport methods.

### Adding a New Agent Adapter

1. Implement the adapter under `core/agentintegration/infra/`.
2. Wire it from the shell composition root in `cmd/skillflow/`.
3. Keep scan, push, pull, and conflict semantics inside `agentintegration`, not in `App` methods or UI DTO builders.

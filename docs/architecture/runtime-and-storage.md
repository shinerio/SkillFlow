# Runtime, Repository Layout, and Storage

## Wails Shell Constraints

SkillFlow remains a Wails desktop application.

The backend architecture preserves these constraints:

- repository root contains no Go source files
- `cmd/skillflow/*.go` stays flat because Wails bindings require a single `package main` directory
- Wails-bound transport adapters remain in `cmd/skillflow/`
- the desktop shell binds backend use cases directly to the frontend instead of exposing a REST API

## `cmd/skillflow/` Responsibilities

`cmd/skillflow/` is the shell-oriented composition layer.

It contains:

- Wails startup and binding registration
- Wails-facing transport adapters
- daemon/UI process bootstrapping
- tray and window integration
- single-instance coordination
- shell-level startup sequencing
- settings-save fan-out coordination where multiple contexts must be updated together
- dependency assembly for backend contexts, orchestration services, and read models

It is not the home for reusable business use cases such as skill import rules, prompt CRUD rules, source synchronization rules, or agent push/pull semantics.

## Daemon and UI Runtime Split

SkillFlow uses a daemon/UI split:

- the `daemon` process owns the long-lived backend runtime, tray or menu-bar presence, local control endpoints, the loopback service gateway, background timers, and UI relaunch/focus
- the `ui` process hosts Wails, React, and shell-only adapters such as dialogs or OS open actions
- frontend business calls are proxied from the `ui` process to the `daemon` over the local authenticated loopback gateway instead of directly binding the full backend runtime into the UI process
- closing or hiding the main window exits the `ui` process without killing the `daemon`; showing the app again cold-starts a fresh UI process

## Native Migration Runtime

During the native platform migration, SkillFlow has two client families:

- the existing Wails/React UI remains the behavioral baseline and continues to run from `cmd/skillflow/`
- the native macOS and Windows clients will run as platform UI processes and call the same Go `daemon`

The Go `daemon` is the only process allowed to read or mutate business data. Native clients must not read or write `config*.json`, `star_repos*.json`, `skills/`, `prompts/`, `meta/`, backup state, or runtime-derived business files directly.

The native daemon API is a stable JSON envelope defined under `core/platform/nativeapi`:

- request fields: `version`, `method`, `params`, `requestID`
- response fields: `ok`, `result`, `error`
- stable error codes such as `method_not_found` and `internal_error`

The migration starts with read-only contract methods such as `settings.get`, `skills.list`, `skills.categories.list`, `agents.listEnabled`, and `backup.providers.list`. The current daemon exposes the native envelope through the local authenticated daemon gateway while preserving the legacy Wails method proxy. Transport details can move from the current loopback implementation to Unix domain sockets on macOS or named pipes on Windows without changing the JSON contract.

## Repository Shape

```text
/
  go.mod
  go.sum
  Makefile
  README.md
  README_zh.md
  contributing.md
  contributing_zh.md
  docs/
    agents/
    architecture/
    config.md
    config_zh.md
    features.md
    features_zh.md
    plans/
    superpowers/
  core/
    platform/
    shared/
    orchestration/
    readmodel/
    skillcatalog/
    promptcatalog/
    agentintegration/
    skillsource/
    backup/
    config/
  cmd/
    skillflow/
      main.go
      app.go
      app_*.go
      app_startup.go
      adapters.go
      providers.go
      events.go
      process_*.go
      tray_*.go
      window_*.go
      single_instance_*.go
      frontend/
```

## Storage Layout

Logical ownership is split by bounded context, while physical storage remains operationally simple.

Current persisted layout is split between the fixed app-data root and the optional local repo-cache root:

```text
<AppDataDir>/
  config.json          # shared sync-safe settings payload
  config_local.json    # local-only settings, paths, secrets, and runtime state
  star_repos.json      # tracked repository state for skillsource
  star_repos_local.json
  skills/
    <category>/<skill>/
  meta/
  meta_local/
  prompts/
    <category>/<name>/
  cache/
    viewstate/
  runtime/
  logs/

<RepoCacheDir>/        # local-only repo clone cache root; defaults to <AppDataDir>/cache/repos
  <git-cache-hosts...>
```

`config.json` and `config_local.json` are flat shared/local payloads managed through `core/config`. Ownership is logical rather than literal top-level namespacing, so the physical files do not mirror bounded contexts 1:1.

## Configuration Ownership

`config.json` and `config_local.json` are shared storage files. Ownership of the fields inside them is still split by context and platform concern.

Logical ownership:

- `skillcatalog`
  - default skill category
- `skillsource`
  - local repo cache root for starred-repo clones
- `agentintegration`
  - agent profiles
  - auto-push policy
  - recursive repo/agent scan depth
- `backup`
  - active backup profile
  - provider selection and interval
  - cloud profile / credential split across shared and local config
- shell/platform
  - launch-at-login
  - window state
  - skipped update version
  - proxy and log-level preferences

Additional persisted ownership outside `config*.json`:

- `promptcatalog`
  - prompt content and metadata under `prompts/`
- `skillsource`
  - tracked repo state in `star_repos.json` / `star_repos_local.json`
  - repo cache directories under the current `repoCacheDir` (default `<AppDataDir>/cache/repos`)

Implementation ownership:

- app-data path ownership lives in `core/platform/appdata`
- shell proxy, window, log-level, and skipped-update preferences live in `core/platform/shellsettings` plus `core/platform/settingsstore`
- cross-context write flows for import, push/pull, update, and restore compensation live in `core/orchestration`
- installed-skill, starred-skill, and agent-presence composition lives in `core/readmodel/skills` plus `core/readmodel/viewstate`
- `core/config` is the frontend-facing settings facade and split/merge persistence adapter around these context- and platform-owned settings components

## Repository vs Gateway Examples

Repositories:

- installed skill metadata store
- prompt library store
- agent profile store
- source-tracking store
- settings facade persistence views

Gateways:

- agent workspace adapter
- Git client wrapper used to sync external sources
- cloud backup provider adapter
- GitHub release API client
- Wails runtime adapters such as file dialogs or shell open operations

## Events and Derived State

Event forwarding to the frontend remains a shell integration concern. Event publication belongs close to application services and orchestration services.

Derived snapshots such as installed-skill cards or aggregated agent presence should live under `readmodel/` or context-local `infra/projection/`, not inside domain models.

## Logging and Path Portability

Existing constraints remain in force:

- logs stay bounded to `skillflow.log` and `skillflow.log.1`
- synced paths should be stored as forward-slash relative paths when under the synchronized root
- local-only machine-specific paths stay in local settings namespaces
- secrets must never be written to logs

*Last updated: 2026-05-06*

# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

A cross-platform desktop app for managing LLM Skills (prompt libraries / slash commands) across multiple AI coding tools — with GitHub install, cloud backup, and cross-tool sync.

## Download & Install

Get the latest release from **[GitHub Releases →](https://github.com/shinerio/SkillFlow/releases/latest)**

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `SkillFlow-macos-apple-silicon` |
| macOS (Intel) | `SkillFlow-macos-intel` |
| Windows (x64) | `SkillFlow-windows-amd64.exe` |

## Features

| Feature | Description |
|---------|-------------|
| **Skill Library** | Central store with categories, real-time search, A-Z / Z-A sorting, drag-and-drop organization, batch delete, and safe empty-category deletion |
| **Prompt Library** | Save reusable prompts as synced `prompts/<category>/<name>/system.md` cards, with required unique names, optional descriptions, categories, import/export, drag-to-category management, one-click copy, and `and` / `or` keyword search |
| **GitHub Install** | Clone any repo, recursively discover nested skill candidates, and install selected ones with one click; candidate status badges follow the configurable per-page card-visibility policy, and same-name candidates stay distinct by normalized repo source + subpath |
| **Cross-tool Sync** | Push or pull skills to/from Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, or any custom tool; My Skills can persist auto-push target tools, immediately backfill the current library when a tool is enabled there, Push defaults to Manual Select, Pull starts with everything unchecked and adds a quick "Select Not Imported" checkbox, and pushed tools are rendered as compact icon lists with hover-to-reveal full tool sets |
| **Starred Repos** | Watch Git repos and recursively browse/import nested repo skills without adding them to your library first; repo skill cards show imported and pushed-tool state with imported correlation keyed by normalized repo source + subpath, and builtin starter repos (`anthropics/skills`, `ComposioHQ/awesome-claude-skills`, `affaan-m/everything-claude-code`) are seeded only on first initialization (won't be re-added after user deletion) |
| **Cloud Backup** | Mirror your library to Aliyun OSS, AWS S3, Azure Blob Storage, Google Cloud Storage, Tencent COS, Huawei OBS, or any Git repo, with a custom object-storage remote path preview, provider-specific saved profiles, local-only sensitive credentials, a manual Git-conflict folder shortcut, and a backup page that shows per-run changed files plus the latest sync-completed timestamp |
| **Update Checker** | Detects new commits for installed GitHub-sourced skills by normalized repo source + subpath, clears stale update markers when already current, and supports one-click instance updates |
| **App Auto-Update** | Modal dialog notifies when a new version is available; Windows supports one-click download and restart; macOS links to GitHub Releases; users can skip a version to suppress future startup prompts |
| **Background Tray** | Clicking the window close button keeps the app running in background; on macOS it hides the Dock icon and leaves a monochrome menu-bar status icon, on Windows it stays in the notification area |
| **Desktop Shell** | Fixed sidebar with the branded SkillFlow title, app icon, quick language/theme toggles, and feedback entry |
| **Startup Window** | Each device stores the most recently adjusted window size in local `config_local.json` and restores it on the next launch; the first launch still calculates the window size adaptively for the current display |
| **Settings** | A responsive settings page that expands with the window, with per-tool enable/disable, push & scan paths, custom tools, proxy configuration, launch-at-login toggle, configurable local/remote scan depth, per-page card-status visibility, local-only path/proxy/login settings kept out of sync, and a `Ctrl+S` / `Cmd+S` save shortcut; folder pickers reopen at the current location |
| **Bilingual UI** | Switch the frontend instantly between Chinese and English from the sidebar or Settings; language preference is stored locally |
| **Dark / Young / Light Themes** | Switch between a refined graphite Dark theme, a softened paper-blue Young theme evolved from the legacy Light palette, and a new Messor-inspired Light theme; persisted across restarts |

For a complete description of every button, dialog, and interaction, see **[docs/features.md](docs/features.md)**.

## SkillFlow vs cc-switch

### 1. Core Positioning

| | SkillFlow | cc-switch |
|---|---|---|
| **Goal** | A dedicated management tool for Skills (prompt libraries) | An all-in-one configuration assistant for AI CLI tools |
| **Core Value** | Discovery, installation, sync, and cloud backup for Skills. **Philosophy**: lightweight tooling focused on accumulating and managing skill assets | One-stop management for provider API switching + MCP + Skills + Prompts |

### 2. Supported Tools

| | SkillFlow | cc-switch |
|---|---|---|
| **Claude Code** | ✅ | ✅ |
| **OpenCode** | ✅ | ✅ |
| **Codex** | ✅ | ✅ |
| **Gemini CLI** | ✅ | ✅ |
| **OpenClaw** | ✅ | ❌ |
| **Custom Tools** | ✅ | ❌ |

### 3. Feature Comparison

| | SkillFlow | cc-switch |
|---|---|---|
| **Local Skill Library Management** | ✅ A local central library with categories, search, drag-and-drop organization, and batch delete | ❌ No local library concept; installs/uninstalls directly into `~/.claude/skills/` |
| **Install Sources** | GitHub repo cloning, **Starred Repos** browsing, manual import, and scanning skills built into tools | GitHub repo scanning (3 preconfigured repos + custom repos) |
| **GitHub Install** | ✅ Deep recursive scan with conflict handling | ✅ Basic support (preconfigured repos + custom repos) |
| **Cross-tool Sync** | ✅ Fine-grained control with per-skill conflict handling | ✅ Basic, mainly one-click install to `~/.claude/skills/` |
| **Update Detection** | ✅ Detects new commits per skill | ❌ |
| **Cloud Backup** | ✅ Multiple object storage providers / Git | ❌ (relies on external cloud-drive directory sync) |
| **Starred Repo Browsing** | ✅ | ❌ |

## Supported Tools

Built-in adapters for: **Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

Custom tools can be added in Settings with any local directory path.

## Skill Format

A valid skill directory must contain a `skill.md` file at its root. Any directory satisfying this requirement can be imported locally or via GitHub.

```
my-skill/
  skill.md     ← required
  ...other files
```

## Cloud Backup

Configure in **Settings → Cloud Storage**.

- Sync-safe settings and metadata live under the app data directory and use relative paths for cross-platform restore.
- Local-only filesystem paths and proxy settings (such as `SkillsStorageDir`, tool scan/push directories, and manual proxy URLs) live in `config_local.json` and are excluded from backup/sync.
- Sensitive cloud credentials (such as access key IDs, secret keys, account keys, service-account JSON, and access tokens) are stored only in per-provider entries inside `config_local.json`; synced `config.json` keeps only non-sensitive cloud settings such as bucket, endpoint, region, account name, service URL, repo URL, and branch.
- Reusable prompts are synced alongside skills under `prompts/<category>/<name>/system.md`, so Git backup and object storage keep the same prompt library on every device.
- Object storage providers support a custom parent `remotePath`; the final backup prefix is always rendered and stored as `<bucket>/<remotePath>/skillflow/` (or `<bucket>/skillflow/` when the parent path is empty).
- Each cloud provider keeps its own saved bucket/path/credential profile, so switching providers in Settings does not overwrite another provider's values.
- Git sync conflicts can be resolved by keeping local, keeping remote, or opening the backup folder for manual fixes.
- The Backup page shows only the files changed in the latest backup or restore operation for the current app session and displays the last successful sync-completed timestamp, instead of the full remote file listing.
- Aliyun OSS, Tencent COS, and Huawei OBS share the same bucket + endpoint configuration model. AWS S3 uses bucket + region. Azure Blob Storage uses container name (in the bucket field) plus account name, account key, and an optional service URL. Google Cloud Storage uses bucket plus a service-account JSON string or local key file path. For Tencent COS, the bucket always comes from the dedicated bucket field, while the endpoint field can store either a plain endpoint host or a full bucket host/URL and is preserved as entered.
- App data directory:
  - macOS: `~/Library/Application Support/SkillFlow/`
  - Windows: `%USERPROFILE%\.skillflow\`

Supported providers and required fields:

| Provider | Required Fields |
|----------|----------------|
| Aliyun OSS | Access Key ID (local-only), Access Key Secret (local-only), Endpoint (synced) |
| AWS S3 | Access Key ID (local-only), Secret Access Key (local-only), Region (synced) |
| Azure Blob Storage | Container name (bucket field, synced), Account Name (synced), Account Key (local-only), Service URL (synced, optional) |
| Google Cloud Storage | Service Account JSON or local key file path (local-only) |
| Tencent COS | SecretId (local-only), SecretKey (local-only), Endpoint (synced) |
| Huawei OBS | Access Key ID (local-only), Secret Access Key (local-only), Endpoint (synced) |
| Git Repo | Repo URL (synced), Branch (synced), Username (synced), Access Token (local-only) |

## Contributing & Building from Source

### Prerequisites

- macOS 11+ or Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Build Steps

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend   # install frontend dependencies
make dev                # hot-reload dev mode
make test               # run Go tests
make build              # production binary → build/bin/
```

Common `make` targets:

| Target | Description |
|--------|-------------|
| `make dev` | Hot-reload dev mode (Go + frontend) |
| `make build` | Build production binary |
| `make test` | Run all Go tests |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate TypeScript bindings after App method changes |
| `make clean` | Remove build artifacts |

For internal architecture details, see **[docs/architecture.md](docs/architecture.md)**.

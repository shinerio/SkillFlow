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
| **GitHub Install** | Clone any repo, recursively discover nested skill candidates, and install selected ones with one click; already-installed detection is keyed by normalized repo source + subpath so same-name candidates remain distinct |
| **Cross-tool Sync** | Push or pull skills to/from Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, or any custom tool; searchable, sortable skill pickers with backend-resolved imported/pushed status, pair-precise conflict handling, and no name-only heuristics |
| **Starred Repos** | Watch Git repos and recursively browse/import nested repo skills without adding them to your library first; imported badges follow normalized repo source + subpath, updatable installed instances are surfaced directly in repo skill cards, and builtin starter repos (`anthropics/skills`, `ComposioHQ/awesome-claude-skills`, `affaan-m/everything-claude-code`) are seeded only on first initialization (won't be re-added after user deletion) |
| **Cloud Backup** | Mirror your library to Aliyun OSS, Tencent COS, Huawei OBS, or any Git repo, with a custom object-storage remote path preview, provider-specific saved profiles, and local-only sensitive credentials |
| **Update Checker** | Detects new commits for installed GitHub-sourced skills by normalized repo source + subpath, clears stale update markers when already current, and supports one-click instance updates |
| **App Auto-Update** | Modal dialog notifies when a new version is available; Windows supports one-click download and restart; macOS links to GitHub Releases; users can skip a version to suppress future startup prompts |
| **Background Tray** | Clicking the window close button keeps the app running in background; on macOS it hides the Dock icon and leaves a monochrome menu-bar status icon, on Windows it stays in the notification area |
| **Desktop Shell** | Fixed sidebar with the branded SkillFlow title, app icon, quick language/theme toggles, and feedback entry |
| **Settings** | Per-tool enable/disable, push & scan paths, custom tools, proxy configuration, configurable local/remote scan depth, and local-only path settings kept out of sync; folder pickers reopen at the current location |
| **Bilingual UI** | Switch the frontend instantly between Chinese and English from the sidebar or Settings; language preference is stored locally |
| **Dark / Young / Light Themes** | Switch between a refined graphite Dark theme, a softened paper-blue Young theme evolved from the legacy Light palette, and a new Messor-inspired Light theme; tool icon tile contrast also adapts per theme and persists across restarts |

For a complete description of every button, dialog, and interaction, see **[docs/features.md](docs/features.md)**.

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
- Local-only filesystem paths (such as `SkillsStorageDir` and tool scan/push directories) live in `config_local.json` and are excluded from backup/sync.
- Sensitive cloud credentials (such as access key IDs, secret keys, and access tokens) are stored only in per-provider entries inside `config_local.json`; synced `config.json` keeps only non-sensitive cloud settings such as bucket, endpoint, repo URL, and branch.
- Reusable prompts are synced alongside skills under `prompts/<category>/<name>/system.md`, so Git backup and object storage keep the same prompt library on every device.
- Object storage providers support a custom parent `remotePath`; the final backup prefix is always rendered and stored as `<bucket>/<remotePath>/skillflow/` (or `<bucket>/skillflow/` when the parent path is empty).
- Each cloud provider keeps its own saved bucket/path/credential profile, so switching providers in Settings does not overwrite another provider's values.
- Aliyun OSS, Tencent COS, and Huawei OBS share the same bucket + endpoint configuration model. For Tencent COS, the bucket always comes from the dedicated bucket field, while the endpoint field can store either a plain endpoint host or a full bucket host/URL and is preserved as entered.
- App data directory:
  - macOS: `~/Library/Application Support/SkillFlow/`
  - Windows: `%USERPROFILE%\.skillflow\`

Supported providers and required fields:

| Provider | Required Fields |
|----------|----------------|
| Aliyun OSS | Access Key ID (local-only), Access Key Secret (local-only), Endpoint (synced) |
| Tencent COS | SecretId (local-only), SecretKey (local-only), Endpoint (synced) |
| Huawei OBS | Access Key (local-only), Secret Key (local-only), Endpoint (synced) |
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

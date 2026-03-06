# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

A cross-platform desktop app for managing LLM SKILLS (prompt libraries / slash commands) across multiple AI coding tools — with GitHub install, cloud backup, and cross-tool sync.

## Features

| Feature | Description |
|---------|-------------|
| **Skill Library** | Central store with categories, real-time search, drag-and-drop organization, and batch delete |
| **GitHub Install** | Clone any repo, browse skill candidates, select and install with one click; auto-pulls on subsequent scans |
| **Cross-tool Sync** | Push from an all/category sidebar scope or manual card selection, or pull skills to/from Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, or any custom tool; conflict handling per skill |
| **Starred Repos** | Watch Git repos and browse/import their skills without adding them to your library first; folder or flat view; bulk push directly to tools |
| **Cloud Backup** | Mirror your library to Aliyun OSS, Tencent COS, Huawei OBS, or any Git repo; all providers back up the same app-data scope (excluding `cache/` and `.git/`), the Backup page can browse the complete remote file list, and Git mode auto-migrates legacy nested backup metadata so real skill files stay trackable |
| **Update Checker** | Detects new commits for GitHub-sourced skills; one-click update |
| **App Auto-Update** | Startup banner notifies when a new app release is available; Windows supports one-click download and restart; macOS links to GitHub Releases |
| **Background Tray** | Clicking the window close button hides the window instead of quitting; macOS keeps a menu-bar status item with native click-to-open menu, Windows keeps a notification-area tray icon that uses the app icon and provides an exit menu |
| **Settings** | Per-tool enable/disable, push & scan paths (text input + folder picker), fixed fallback category `Default` for uncategorized pull/import, runtime log level (`debug`/`info`/`error`, default `error`) + one-click open log directory, custom tools, cloud credentials, proxy configuration |

For a complete description of every button, dialog, and interaction, see **[feature.md](feature.md)**.

The sidebar also includes a feedback entry that opens GitHub issue creation in your browser.

## Supported Tools

Built-in adapters for: **Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

Custom tools can be added in Settings with any local directory path.

## Requirements

- macOS 11+ or Windows 10+
- Go 1.23+
- Node.js 18+ (for frontend build)
- [Wails v2](https://wails.io) CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Development

```bash
# Clone and install frontend deps
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend

# Run in dev mode (hot-reload)
make dev

# Run Go tests
make test

# Build production binary
make build
```

Common `make` targets:

| Target | Description |
|--------|-------------|
| `make dev` | Hot-reload dev mode (Go + frontend) |
| `make build` | Build production binary |
| `make test` | Run all Go tests |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate TypeScript bindings |
| `make clean` | Remove build artifacts |

Binary output: `build/bin/SkillFlow` (macOS) / `build/bin/SkillFlow.exe` (Windows)

## Skill Format

A valid skill directory must contain a `skill.md` file at its root. Any directory satisfying this requirement can be imported locally or via GitHub.

```
my-skill/
  skill.md     ← required
  ...            ← other files
```

## Cloud Backup

Configure in **Settings → Cloud Storage**. Credentials are stored in the local config file at:

- macOS: `~/Library/Application Support/SkillFlow/config.json`
- Windows: `%APPDATA%\SkillFlow\config.json`

Supported providers and their required fields:

| Provider | Fields |
|----------|--------|
| Aliyun OSS | Access Key ID, Access Key Secret, Endpoint |
| Tencent COS | SecretId, SecretKey, Region |
| Huawei OBS | Access Key, Secret Key, Endpoint |

## CI / Releases

Builds are automated via GitHub Actions on `v*` tags, producing binaries for:
- macOS (Intel x86_64)
- macOS (Apple Silicon arm64)
- Windows (x86_64)

## Architecture

Go core library (`core/`) with interface-based plugin architecture. Wails v2 bridges Go backend to React 18 + TypeScript + Tailwind CSS frontend via direct method bindings (no HTTP API). See [CLAUDE.md](CLAUDE.md) for developer details.


## Supported Tools

Built-in adapters for: **Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

Custom tools can be added in Settings with any local directory path.

## Requirements

- macOS 11+ or Windows 10+
- Go 1.23+
- Node.js 18+ (for frontend build)
- [Wails v2](https://wails.io) CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Development

```bash
# Clone and install frontend deps
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend

# Run in dev mode (hot-reload)
make dev

# Run Go tests
make test

# Build production binary
make build
```

Common `make` targets:

| Target | Description |
|--------|-------------|
| `make dev` | Hot-reload dev mode (Go + frontend) |
| `make build` | Build production binary |
| `make test` | Run all Go tests |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate TypeScript bindings |
| `make clean` | Remove build artifacts |

Binary output: `build/bin/SkillFlow` (macOS) / `build/bin/SkillFlow.exe` (Windows)

## Skill Format

A valid skill directory must contain a `skill.md` file at its root. Any directory satisfying this requirement can be imported locally or via GitHub.

```
my-skill/
  skill.md     ← required
  ...            ← other files
```

## Cloud Backup

Configure in **Settings → Cloud Storage**. Credentials are stored in the local config file at:

- macOS: `~/Library/Application Support/SkillFlow/config.json`
- Windows: `%APPDATA%\SkillFlow\config.json`

Supported providers and their required fields:

| Provider | Fields |
|----------|--------|
| Aliyun OSS | Access Key ID, Access Key Secret, Endpoint |
| Tencent COS | SecretId, SecretKey, Region |
| Huawei OBS | Access Key, Secret Key, Endpoint |

## CI / Releases

Builds are automated via GitHub Actions on `v*` tags, producing binaries for:
- macOS (Intel x86_64)
- macOS (Apple Silicon arm64)
- Windows (x86_64)

## Architecture

Go core library (`core/`) with interface-based plugin architecture. Wails v2 bridges Go backend to React 18 + TypeScript + Tailwind CSS frontend via direct method bindings (no HTTP API). See [CLAUDE.md](CLAUDE.md) for developer details.

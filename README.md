# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

A cross-platform desktop app for managing LLM SKILLS (prompt libraries / slash commands) across multiple AI coding tools — with GitHub install, cloud backup, and cross-tool sync.

## What It Does

| Feature | Description |
|---------|-------------|
| **Skill Library** | Central store for all your skills with categories, search, and drag-drop organization |
| **GitHub Install** | Browse and install skills from any GitHub repo that follows the `skills/` convention |
| **Cross-tool Sync** | Push or pull skills to/from Claude Code, OpenCode, Codex, Gemini CLI, OpenClaw, or any custom tool |
| **Cloud Backup** | Mirror your skill library to Aliyun OSS, Tencent COS, or Huawei OBS |
| **Update Checker** | Detects when GitHub-sourced skills have new commits available |

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

A valid skill directory must contain a `SKILLS.md` file at its root. Any directory satisfying this requirement can be imported locally or via GitHub.

```
my-skill/
  SKILLS.md      ← required
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

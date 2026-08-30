[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shinerio/SkillFlow)

# SkillFlow

> 🌐 [中文](README_zh.md) | **English**

Slogan: Master once. Apply everywhere.

SkillFlow is a cross-platform desktop app for managing reusable skills, prompts, and memories across agent environments. It keeps local libraries for all three asset types, syncs selected skills and memories to multiple agents, tracks repo-backed sources, checks for updates, and backs up sync-safe data to object storage or Git.

![SkillFlow](docs/skillflow.gif)

## What SkillFlow Does

- **📦 Turn discoveries into assets**: Solve the "saved means forgotten" problem by turning scattered skills, prompts, and memory snippets from blogs, videos, and daily work into personal assets you can manage, iterate on, retire, and update deliberately.
- **🎛️ Stay in control**: Avoid black-box skill management by organizing large skill collections by scenario, so developers always understand and control an agent's capability boundaries.
- **🔄 Sync across agents**: Break model silos by reusing one setup across Claude Code, Codex, Gemini, and other agents with different strengths, limits, and use cases.
- **🧠 Build modular memory systems**: Maintain one main memory plus reusable module memories, then sync them automatically to multiple AI coding tools with merge or takeover modes.
- **🖥️ Keep environments consistent**: Build a cloud-backed workflow across macOS and Windows devices so your development environment and reusable assets stay aligned wherever you work.
- **⚡ Iterate automatically**: Replace tedious manual maintenance with automatic skill updates and memory sync, so distributed sources do not turn into stale versions.
- **📝 Version prompts intentionally**: Manage prompts as versioned assets instead of rewriting them ad hoc, improving consistency, optimization, and standardization over time.

## Download & Install

Get the latest release from **[GitHub Releases](https://github.com/shinerio/SkillFlow/releases/latest)**.

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `SkillFlow-macos-apple-silicon.dmg` |
| macOS (Intel) | `SkillFlow-macos-intel.dmg` |
| Windows (x64) | `SkillFlow-windows.exe` |

## Highlights

| Feature | Description |
|---------|-------------|
| **Skill Library** | Local library with categories, search, sorting, drag-and-drop organization, batch delete, update checks, and auto-push targets. |
| **My Prompts** | Synced prompt cards with category management, scoped export, conflict-aware import, copy, image previews, and web links. |
| **My Memory** | Unified memory workspace for main memory and module memories, with editing, preview, batch push, delete, and per-agent auto-sync modes. |
| **Cross-agent Sync** | Push skills and automatically sync memories across built-in and custom agents. |
| **Starred Repos** | Track Git repos, browse discovered skills, and import or push them without installing everything first. |
| **Cloud Backup** | Back up sync-safe skills, prompts, memories, and metadata to object storage providers or Git. |
| **Desktop Experience** | Native macOS and Windows clients with a bundled daemon-backed runtime, launch-at-login, and proxy testing; the legacy Wails build remains available for rollback. |

Detailed references:

- UI/UX behavior: [docs/features.md](docs/features.md)
- Persisted files and config schema: [docs/config.md](docs/config.md)

## Supported Agents

Built-in adapters:

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

You can also add **custom agents** in Settings by configuring local scan and push directories.

## Skill Format

A valid skill directory must contain a root `skill.md` file. Filename matching is case-insensitive.

```text
my-skill/
  skill.md
  ...other files
```

For contributing guidelines and building from source, see **[contributing.md](contributing.md)**.

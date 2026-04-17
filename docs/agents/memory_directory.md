# Built-in Agent Memory Directories

This file records the default memory directories used by SkillFlow's built-in agent profiles. These paths match the defaults in `core/agentintegration/domain/defaults.go`.

| Agent | Reference | Default Main directories          | Rules directories              |
|------|-----------|-----------------------------------|--------------------------------|
| Claude Code | [docs](https://code.claude.com/docs/en/memory) | `~/.claude/CLAUDE.md`             | `~/.claude/rules/`             |
| OpenCode | [docs](https://opencode.ai/docs/zh-cn/rules/) | `~/.config/opencode/AGENTS.md`    | `~/.config/opencode/rules/`    |
| Codex | [docs](https://developers.openai.com/codex/guides/agents-md) | `~/.codex/AGENTS.md`   | `~/.codex/rules/rules/`        |
| Gemini CLI | [docs](https://geminicli.com/docs/cli/gemini-md/) | `~/.gemini/GEMINI.md`             | `~/.gemini/rules/`             |
| OpenClaw | [docs](https://docs.openclaw.ai/concepts/system-prompt) | `~/.openclaw/workspace/MEMORY.md` | `~/.openclaw/workspace/rules/` |
| Copilot | [docs](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions) | `~/.copilot/copilot-instructions.md` | — |

Copilot does not define a first-party rules-directory equivalent in the GitHub Copilot CLI docs, so SkillFlow leaves its built-in `rulesDir` empty by default.

# Built-in Agent Skill Directories

This file records the built-in agent directory defaults used by SkillFlow. These paths match `core/agentintegration/domain/defaults.go`.

| Agent | Reference | Default push directory | Default scan directories |
|------|-----------|------------------------|--------------------------|
| Claude Code | [docs](https://code.claude.com/docs/en/skills) | `~/.claude/skills/` | `~/.claude/skills/`, `~/.claude/plugins/marketplaces/` |
| OpenCode | [docs](https://opencode.ai/docs/zh-cn/skills/) | `~/.config/opencode/skills/` | `~/.config/opencode/skills/`, `~/.agents/skills/` |
| Codex | [docs](https://developers.openai.com/codex/skills) | `~/.agents/skills/` | `~/.agents/skills/` |
| Gemini CLI | [docs](https://geminicli.com/docs/cli/skills) | `~/.gemini/skills/` | `~/.gemini/skills/`, `~/.agents/skills/` |
| OpenClaw | [docs](https://docs.openclaw.ai/tools/skills) | `~/.openclaw/skills/` | `~/.openclaw/skills/`, `~/.openclaw/workspace/skills/` |
| Copilot | [docs](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) | `~/.copilot/skills/` | `~/.claude/skills/`, `~/.agents/skills/`, runtime-detected `~/.copilot/pkg/universal/<version>/builtin-skills/` when present |

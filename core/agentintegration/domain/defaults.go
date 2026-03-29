package domain

import (
	"os"
	"path/filepath"
)

var builtinAgentNames = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func BuiltinAgentNames() []string {
	return append([]string(nil), builtinAgentNames...)
}

func DefaultProfile(agentName string) AgentProfile {
	profile := AgentProfile{
		Name:    agentName,
		Enabled: true,
		Custom:  false,
	}
	profile.ScanDirs = defaultScanDirs(agentName)
	if len(profile.ScanDirs) > 0 {
		profile.PushDir = profile.ScanDirs[0]
	}
	profile.MemoryPath, profile.RulesDir = defaultMemoryPaths(agentName)
	return profile
}

func defaultMemoryPaths(agentName string) (memoryPath, rulesDir string) {
	home, _ := os.UserHomeDir()

	paths := map[string][2]string{
		"claude-code": {
			filepath.Join(home, ".claude", "CLAUDE.md"),
			filepath.Join(home, ".claude", "rules"),
		},
		"codex": {
			filepath.Join(home, ".codex", "AGENTS.md"),
			filepath.Join(home, ".codex", "rules"),
		},
		"gemini-cli": {
			filepath.Join(home, ".gemini", "GEMINI.md"),
			filepath.Join(home, ".gemini", "rules"),
		},
		"opencode": {
			filepath.Join(home, ".config", "opencode", "AGENTS.md"),
			filepath.Join(home, ".config", "opencode", "rules"),
		},
		"openclaw": {
			filepath.Join(home, ".openclaw", "workspace", "AGENTS.md"),
			filepath.Join(home, ".openclaw", "workspace", "rules"),
		},
	}
	if p, ok := paths[agentName]; ok {
		return p[0], p[1]
	}
	return "", ""
}

func defaultScanDirs(agentName string) []string {
	home, _ := os.UserHomeDir()
	agentsDir := filepath.Join(home, ".agents", "skills")

	dirs := map[string][]string{
		"claude-code": {
			filepath.Join(home, ".claude", "skills"),
			filepath.Join(home, ".claude", "plugins", "marketplaces"),
		},
		"opencode": {
			filepath.Join(home, ".config", "opencode", "skills"),
			agentsDir,
		},
		"codex": {
			agentsDir,
		},
		"gemini-cli": {
			filepath.Join(home, ".gemini", "skills"),
			agentsDir,
		},
		"openclaw": {
			filepath.Join(home, ".openclaw", "skills"),
			filepath.Join(home, ".openclaw", "workspace", "skills"),
		},
	}
	return append([]string(nil), dirs[agentName]...)
}

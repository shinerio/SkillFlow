package domain

import (
	"os"
	"path/filepath"
	"sort"
)

var builtinAgentNames = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw", "copilot"}

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
	profile.PushDir = defaultPushDir(agentName, profile.ScanDirs)
	profile.MemoryPath, profile.RulesDir = defaultMemoryPaths(agentName)
	return profile
}

func defaultPushDir(agentName string, scanDirs []string) string {
	home, _ := os.UserHomeDir()

	dirs := map[string]string{
		"claude-code": filepath.Join(home, ".claude", "skills"),
		"opencode":    filepath.Join(home, ".config", "opencode", "skills"),
		"codex":       filepath.Join(home, ".agents", "skills"),
		"gemini-cli":  filepath.Join(home, ".gemini", "skills"),
		"openclaw":    filepath.Join(home, ".openclaw", "skills"),
		"copilot":     filepath.Join(home, ".copilot", "skills"),
	}
	if dir, ok := dirs[agentName]; ok {
		return dir
	}
	if len(scanDirs) > 0 {
		return scanDirs[0]
	}
	return ""
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
		"copilot": {
			filepath.Join(home, ".copilot", "copilot-instructions.md"),
			"",
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
		"copilot": {
			filepath.Join(home, ".claude", "skills"),
			agentsDir,
		},
	}
	scanDirs := append([]string(nil), dirs[agentName]...)
	if agentName == "copilot" {
		scanDirs = append(scanDirs, discoverCopilotBuiltinSkillDirs(home)...)
	}
	return scanDirs
}

func discoverCopilotBuiltinSkillDirs(home string) []string {
	baseDir := filepath.Join(home, ".copilot", "pkg", "universal")
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	type builtinDir struct {
		path    string
		modTime int64
	}
	dirs := make([]builtinDir, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(baseDir, entry.Name(), "builtin-skills")
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}
		dirs = append(dirs, builtinDir{path: candidate, modTime: info.ModTime().UnixNano()})
	}
	sort.Slice(dirs, func(i, j int) bool {
		if dirs[i].modTime == dirs[j].modTime {
			return dirs[i].path < dirs[j].path
		}
		return dirs[i].modTime > dirs[j].modTime
	})
	result := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		result = append(result, dir.path)
	}
	return result
}

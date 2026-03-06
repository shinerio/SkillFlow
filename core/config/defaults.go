package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func AppDataDir() string {
	switch runtime.GOOS {
	case "windows":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".skillflow")
	default: // darwin / linux
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "SkillFlow")
	}
}

func defaultAgentsSkillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agents", "skills")
}

func DefaultToolScanDirs(toolName string) []string {
	home, _ := os.UserHomeDir()
	agentsDir := defaultAgentsSkillsDir()

	// All platforms use home-relative paths.
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
	return dirs[toolName]
}

// DefaultToolsDir returns the default push path for a tool.
func DefaultToolsDir(toolName string) string {
	scanDirs := DefaultToolScanDirs(toolName)
	if len(scanDirs) == 0 {
		return ""
	}
	return scanDirs[0]
}

var builtinTools = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
	tools := make([]ToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		scanDirs := DefaultToolScanDirs(name)
		pushDir := DefaultToolsDir(name)
		tools = append(tools, ToolConfig{
			Name:     name,
			ScanDirs: scanDirs,
			PushDir:  pushDir,
			Enabled:  true,
			Custom:   false,
		})
	}
	return AppConfig{
		SkillsStorageDir: filepath.Join(dataDir, "skills"),
		DefaultCategory:  "Default",
		LogLevel:         DefaultLogLevel,
		Tools:            tools,
		Cloud:            CloudConfig{RemotePath: "skillflow/"},
	}
}

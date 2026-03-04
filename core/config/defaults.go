package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func AppDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "SkillFlow")
	default: // darwin
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "SkillFlow")
	}
}

func DefaultToolsDir(toolName string) string {
	home, _ := os.UserHomeDir()
	dirs := map[string]map[string]string{
		"darwin": {
			"claude-code": filepath.Join(home, ".claude"),
			"opencode":    filepath.Join(home, ".opencode"),
			"codex":       filepath.Join(home, ".codex"),
			"gemini-cli":  filepath.Join(home, ".gemini"),
			"openclaw":    filepath.Join(home, ".openclaw"),
		},
		"windows": {
			"claude-code": filepath.Join(os.Getenv("APPDATA"), "claude"),
			"opencode":    filepath.Join(os.Getenv("APPDATA"), "opencode"),
			"codex":       filepath.Join(os.Getenv("APPDATA"), "codex"),
			"gemini-cli":  filepath.Join(os.Getenv("APPDATA"), "gemini"),
			"openclaw":    filepath.Join(os.Getenv("APPDATA"), "openclaw"),
		},
	}
	goos := runtime.GOOS
	if goos != "windows" {
		goos = "darwin"
	}
	return dirs[goos][toolName]
}

var builtinTools = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
	tools := make([]ToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		dir := DefaultToolsDir(name)
		_, err := os.Stat(dir)
		tools = append(tools, ToolConfig{
			Name:      name,
			SkillsDir: dir,
			Enabled:   err == nil,
			Custom:    false,
		})
	}
	return AppConfig{
		SkillsStorageDir: filepath.Join(dataDir, "skills"),
		DefaultCategory:  "Imported",
		Tools:            tools,
		Cloud:            CloudConfig{RemotePath: "skillflow/"},
	}
}

package main

import (
	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/registry"
	toolsync "github.com/shinerio/skillflow/core/sync"
)

func registerAdapters() {
	tools := []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}
	for _, name := range tools {
		registry.RegisterAdapter(toolsync.NewFilesystemAdapter(name, config.DefaultToolsDir(name)))
	}
}

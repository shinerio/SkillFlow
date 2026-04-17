package app

import (
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSettingsIncludeBuiltinAgents(t *testing.T) {
	settings := DefaultSettings()
	assert.Equal(t, DefaultRepoScanMaxDepth, settings.Shared.RepoScanMaxDepth)
	assert.NotEmpty(t, settings.Shared.Agents)
	assert.NotEmpty(t, settings.Local.Agents)
}

func TestDefaultSettingsIncludeBuiltinAgentMemoryPaths(t *testing.T) {
	settings := DefaultSettings()
	assert.NotEmpty(t, settings.Local.Agents)

	for _, agent := range settings.Local.Agents {
		assert.NotEmpty(t, agent.MemoryPath)
		if agent.Name == "copilot" {
			assert.Empty(t, agent.RulesDir)
			continue
		}
		assert.NotEmpty(t, agent.RulesDir)
	}
}

func TestDefaultSettingsIncludeCopilotProfile(t *testing.T) {
	settings := DefaultSettings()
	require.NotEmpty(t, settings.Local.Agents)

	var found bool
	for _, agent := range settings.Local.Agents {
		if agent.Name != "copilot" {
			continue
		}
		found = true
		expected := domain.DefaultProfile("copilot")
		assert.Equal(t, expected.ScanDirs, agent.ScanDirs)
		assert.Equal(t, expected.PushDir, agent.PushDir)
		assert.Equal(t, expected.MemoryPath, agent.MemoryPath)
		assert.Equal(t, expected.RulesDir, agent.RulesDir)
		break
	}

	assert.True(t, found)
}

func TestNormalizeAutoPushAgentNames(t *testing.T) {
	normalized := NormalizeAutoPushAgentNames([]string{" codex ", "gemini-cli", "codex", ""})
	assert.Equal(t, []string{"codex", "gemini-cli"}, normalized)
}

func TestNormalizeRepoScanMaxDepth(t *testing.T) {
	assert.Equal(t, DefaultRepoScanMaxDepth, NormalizeRepoScanMaxDepth(0))
	assert.Equal(t, MaxRepoScanMaxDepth, NormalizeRepoScanMaxDepth(999))
	assert.Equal(t, 7, NormalizeRepoScanMaxDepth(7))
}

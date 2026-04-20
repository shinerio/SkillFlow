package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinAgentNamesAreStable(t *testing.T) {
	assert.Equal(t, []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw", "copilot"}, domain.BuiltinAgentNames())
}

func TestDefaultProfileUsesExpectedDirectories(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	claude := domain.DefaultProfile("claude-code")
	require.True(t, claude.Enabled)
	assert.False(t, claude.Custom)
	assert.Equal(t, filepath.Join(home, ".claude", "skills"), claude.PushDir)
	assert.Equal(t, []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".claude", "plugins", "marketplaces"),
	}, claude.ScanDirs)

	codex := domain.DefaultProfile("codex")
	assert.Equal(t, filepath.Join(home, ".agents", "skills"), codex.PushDir)
	assert.Equal(t, []string{filepath.Join(home, ".agents", "skills")}, codex.ScanDirs)
}

func TestDefaultProfileUsesExpectedDirectoriesForCopilot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	builtinDir := filepath.Join(home, ".copilot", "pkg", "universal", "1.0.31", "builtin-skills")
	require.NoError(t, os.MkdirAll(builtinDir, 0o755))

	copilot := domain.DefaultProfile("copilot")
	require.True(t, copilot.Enabled)
	assert.False(t, copilot.Custom)
	assert.Equal(t, filepath.Join(home, ".copilot", "skills"), copilot.PushDir)
	assert.Equal(t, []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".agents", "skills"),
		builtinDir,
	}, copilot.ScanDirs)
	assert.Equal(t, filepath.Join(home, ".copilot", "copilot-instructions.md"), copilot.MemoryPath)
	assert.Empty(t, copilot.RulesDir)
}

func TestDefaultProfileReturnsEmptyForUnknownAgent(t *testing.T) {
	profile := domain.DefaultProfile("unknown")

	assert.Equal(t, "unknown", profile.Name)
	assert.Empty(t, profile.ScanDirs)
	assert.Empty(t, profile.PushDir)
	assert.True(t, profile.Enabled)
	assert.False(t, profile.Custom)
}

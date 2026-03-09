package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportLocalAutoPushesToSelectedTools(t *testing.T) {
	app, pushDir, skillsDir := newAutoPushTestApp(t, []string{"codex"})
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nImported\n")

	_, err := app.ImportLocal(sourceDir, "")
	require.NoError(t, err)

	importedPath := filepath.Join(skillsDir, defaultCategoryName, "demo-skill", "skill.md")
	pushedPath := filepath.Join(pushDir, "demo-skill", "skill.md")
	assert.FileExists(t, importedPath)
	assert.FileExists(t, pushedPath)

	pushedContent, err := os.ReadFile(pushedPath)
	require.NoError(t, err)
	assert.Equal(t, "# Demo\nImported\n", string(pushedContent))
}

func TestImportLocalAutoPushSkipsExistingToolSkill(t *testing.T) {
	app, pushDir, skillsDir := newAutoPushTestApp(t, []string{"codex"})
	existingDir := writeTestSkillDir(t, pushDir, "demo-skill", "# Demo\nExisting\n")
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nImported\n")

	_, err := app.ImportLocal(sourceDir, "")
	require.NoError(t, err)

	importedPath := filepath.Join(skillsDir, defaultCategoryName, "demo-skill", "skill.md")
	existingPath := filepath.Join(existingDir, "skill.md")
	assert.FileExists(t, importedPath)

	existingContent, err := os.ReadFile(existingPath)
	require.NoError(t, err)
	assert.Equal(t, "# Demo\nExisting\n", string(existingContent))
}

func TestSaveConfigAutoPushesExistingSkillsToNewTool(t *testing.T) {
	app, pushDir, _ := newAutoPushTestApp(t, nil)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nImported\n")

	_, err := app.ImportLocal(sourceDir, "")
	require.NoError(t, err)

	cfg, err := app.GetConfig()
	require.NoError(t, err)
	cfg.AutoPushTools = []string{"codex"}

	require.NoError(t, app.SaveConfig(cfg))

	pushedPath := filepath.Join(pushDir, "demo-skill", "skill.md")
	assert.FileExists(t, pushedPath)

	pushedContent, err := os.ReadFile(pushedPath)
	require.NoError(t, err)
	assert.Equal(t, "# Demo\nImported\n", string(pushedContent))
}

func newAutoPushTestApp(t *testing.T, autoPushTools []string) (*App, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	pushDir := filepath.Join(dataDir, "tool-skills")
	skillsDir := filepath.Join(dataDir, "library", "skills")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.SkillsStorageDir = skillsDir
	cfg.AutoPushTools = autoPushTools
	cfg.Tools = []config.ToolConfig{
		{
			Name:     "codex",
			ScanDirs: []string{pushDir},
			PushDir:  pushDir,
			Enabled:  true,
		},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.storage = skill.NewStorage(skillsDir)
	return app, pushDir, skillsDir
}

func writeTestSkillDir(t *testing.T, parentDir, name, content string) string {
	t.Helper()

	skillDir := filepath.Join(parentDir, name)
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "skill.md"), []byte(content), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "notes.txt"), []byte("notes"), 0644))
	return skillDir
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleRestoredCloudStateAutoPushesNewlyRestoredSkills(t *testing.T) {
	app, pushDir, _, _ := newRestoreTestApp(t, []string{"codex"})
	before := app.captureCloudRestoreState()
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nRestored\n")

	_, err := app.storage.Import(sourceDir, defaultCategoryName, skill.SourceManual, "", "")
	require.NoError(t, err)

	require.NoError(t, app.handleRestoredCloudState(before, "test.restore"))

	pushedPath := filepath.Join(pushDir, "demo-skill", "skill.md")
	assert.FileExists(t, pushedPath)

	pushedContent, err := os.ReadFile(pushedPath)
	require.NoError(t, err)
	assert.Equal(t, "# Demo\nRestored\n", string(pushedContent))
}

func TestHandleRestoredCloudStateClonesNewlyRestoredStarredRepos(t *testing.T) {
	if err := coregit.CheckGitInstalled(); err != nil {
		t.Skip("git not installed")
	}

	app, _, _, starsPath := newRestoreTestApp(t, nil)
	before := app.captureCloudRestoreState()
	sourceRepo := newLocalGitRepo(t)
	cloneDir := filepath.Join(filepath.Dir(starsPath), "cache", "restored-repo")

	require.NoError(t, app.starStorage.Save([]coregit.StarredRepo{{
		URL:      sourceRepo,
		Name:     "local/test-repo",
		Source:   "local/test-repo",
		LocalDir: cloneDir,
	}}))

	require.NoError(t, app.handleRestoredCloudState(before, "test.restore"))

	assert.DirExists(t, filepath.Join(cloneDir, ".git"))
	assert.FileExists(t, filepath.Join(cloneDir, "README.md"))

	repos, err := app.starStorage.Load()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Empty(t, repos[0].SyncError)
	assert.False(t, repos[0].LastSync.IsZero())
}

func newRestoreTestApp(t *testing.T, autoPushTools []string) (*App, string, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	pushDir := filepath.Join(dataDir, "tool-skills")
	skillsDir := filepath.Join(dataDir, "library", "skills")
	starsPath := filepath.Join(dataDir, "star_repos.json")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.SkillsStorageDir = skillsDir
	cfg.AutoPushTools = autoPushTools
	cfg.Tools = []config.ToolConfig{{
		Name:     "codex",
		ScanDirs: []string{pushDir},
		PushDir:  pushDir,
		Enabled:  true,
	}}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.storage = skill.NewStorage(skillsDir)
	app.cacheDir = filepath.Join(dataDir, "cache")
	app.starStorage = coregit.NewStarStorage(starsPath)
	return app, pushDir, skillsDir, starsPath
}

func newLocalGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644))
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(output))
}

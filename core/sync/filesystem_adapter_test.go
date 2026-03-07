package sync_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skill"
	toolsync "github.com/shinerio/skillflow/core/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSkill(t *testing.T, dir, mdName string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, mdName), []byte("# skill"), 0644))
}

func TestFilesystemAdapterPushFlattens(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	skillDir := filepath.Join(src, "coding", "my-skill")
	writeSkill(t, skillDir, "skill.md")
	sk := &skill.Skill{Name: "my-skill", Path: skillDir}

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	require.NoError(t, adapter.Push(context.Background(), []*skill.Skill{sk}, dst))

	_, err := os.Stat(filepath.Join(dst, "my-skill", "skill.md"))
	assert.NoError(t, err)
}

func TestFilesystemAdapterPullFlat(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "skill-x"), "skill.md")
	writeSkill(t, filepath.Join(src, "skill-y"), "SKILL.MD")
	require.NoError(t, os.MkdirAll(filepath.Join(src, "not-a-skill"), 0755)) // no skill.md

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, skills, 2)
}

func TestFilesystemAdapterPullNested(t *testing.T) {
	src := t.TempDir()
	// skills nested under category dirs
	writeSkill(t, filepath.Join(src, "coding", "skill-a"), "skill.md")
	writeSkill(t, filepath.Join(src, "coding", "skill-b"), "skill.md")
	writeSkill(t, filepath.Join(src, "writing", "skill-c"), "Skill.md")
	// deeply nested
	writeSkill(t, filepath.Join(src, "a", "b", "c", "skill-d"), "skill.md")
	// category dir itself has no skill.md — should not be returned
	require.NoError(t, os.MkdirAll(filepath.Join(src, "empty-category"), 0755))

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	assert.ElementsMatch(t, []string{"skill-a", "skill-b", "skill-c", "skill-d"}, names)
}

func TestFilesystemAdapterPullWithMaxDepth(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "skills", "skill-a"), "skill.md")
	writeSkill(t, filepath.Join(src, "a", "b", "c", "skill-d"), "skill.md")

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	skills, err := adapter.PullWithMaxDepth(context.Background(), src, 2)
	require.NoError(t, err)

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	assert.ElementsMatch(t, []string{"skill-a"}, names)
}

func TestFilesystemAdapterPullDefaultRespectsDepthLimit(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "a", "b", "c", "d", "e", "f", "skill-g"), "skill.md")

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestFilesystemAdapterPullSkillNotRecursed(t *testing.T) {
	// A skill dir that itself has subdirs should NOT have those subdirs pulled as skills.
	src := t.TempDir()
	skillDir := filepath.Join(src, "parent-skill")
	writeSkill(t, skillDir, "skill.md")
	// sub-dir inside the skill that also has a skill.md
	writeSkill(t, filepath.Join(skillDir, "nested"), "skill.md")

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	// only parent-skill, not nested
	assert.Len(t, skills, 1)
	assert.Equal(t, "parent-skill", skills[0].Name)
}

func TestFilesystemAdapterPullDirNotExist(t *testing.T) {
	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	_, err := adapter.Pull(context.Background(), "/nonexistent/path")
	assert.Error(t, err)
}

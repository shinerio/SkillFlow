package skill_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestSkillDir(t *testing.T, baseDir, name string) string {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
	return dir
}

func TestStorageListCategories(t *testing.T) {
	root := t.TempDir()
	svc := skill.NewStorage(root)
	require.NoError(t, svc.CreateCategory("coding"))
	require.NoError(t, svc.CreateCategory("writing"))
	cats, err := svc.ListCategories()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"coding", "writing"}, cats)
}

func TestStorageImportSkill(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "my-skill")
	svc := skill.NewStorage(root)

	imported, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
	require.NoError(t, err)
	assert.Equal(t, "my-skill", imported.Name)
	assert.Equal(t, "coding", imported.Category)

	// verify directory was copied
	_, err = os.Stat(filepath.Join(root, "coding", "my-skill", "SKILLS.md"))
	assert.NoError(t, err)
}

func TestStorageConflictDetected(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "dup-skill")
	svc := skill.NewStorage(root)

	_, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
	require.NoError(t, err)

	_, err = svc.Import(skillDir, "coding", skill.SourceManual, "", "")
	assert.ErrorIs(t, err, skill.ErrSkillExists)
}

func TestStorageDeleteSkill(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "del-skill")
	svc := skill.NewStorage(root)

	s, err := svc.Import(skillDir, "", skill.SourceManual, "", "")
	require.NoError(t, err)
	require.NoError(t, svc.Delete(s.ID))

	skills, err := svc.ListAll()
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestStorageMoveCategory(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "move-skill")
	svc := skill.NewStorage(root)
	require.NoError(t, svc.CreateCategory("cat-a"))
	require.NoError(t, svc.CreateCategory("cat-b"))

	s, err := svc.Import(skillDir, "cat-a", skill.SourceManual, "", "")
	require.NoError(t, err)

	err = svc.MoveCategory(s.ID, "cat-b")
	require.NoError(t, err)

	updated, err := svc.Get(s.ID)
	require.NoError(t, err)
	assert.Equal(t, "cat-b", updated.Category)
}

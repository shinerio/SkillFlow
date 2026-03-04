package skill_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorAcceptsDirectoryWithSKILLSmd(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILLS.md"), []byte("# skill"), 0644))

	v := skill.NewValidator()
	err := v.Validate(skillDir)
	assert.NoError(t, err)
}

func TestValidatorRejectsDirectoryWithoutSKILLSmd(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "not-a-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))

	v := skill.NewValidator()
	err := v.Validate(skillDir)
	assert.ErrorIs(t, err, skill.ErrNoSKILLSmd)
}

func TestValidatorRejectsNonDirectory(t *testing.T) {
	v := skill.NewValidator()
	err := v.Validate("/nonexistent/path")
	assert.Error(t, err)
}

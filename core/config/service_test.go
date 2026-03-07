package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg, err := svc.Load()
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.SkillsStorageDir)
	assert.Equal(t, "Default", cfg.DefaultCategory)
	assert.Equal(t, config.DefaultLogLevel, cfg.LogLevel)
	assert.NotEmpty(t, cfg.Tools)
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.DefaultCategory = "MyCategory"
	cfg.SkippedUpdateVersion = "v1.2.3"
	err := svc.Save(cfg)
	require.NoError(t, err)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "MyCategory", loaded.DefaultCategory)
	assert.Equal(t, "v1.2.3", loaded.SkippedUpdateVersion)
}

func TestSkippedUpdateVersionPersistsInSharedConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.SkippedUpdateVersion = "v9.9.9"

	require.NoError(t, svc.Save(cfg))

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"skippedUpdateVersion": "v9.9.9"`)

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(localData), "skippedUpdateVersion")

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "v9.9.9", loaded.SkippedUpdateVersion)
}

func TestSaveAndLoadConfigNormalizesLogLevel(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.LogLevel = "BAD_LEVEL"
	err := svc.Save(cfg)
	require.NoError(t, err)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, config.DefaultLogLevel, loaded.LogLevel)
}

func TestConfigFileCreatedOnFirstLoad(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	_, err := svc.Load()
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "config.json"))
	assert.NoError(t, err, "config.json should be created on first load")
	_, err = os.Stat(filepath.Join(dir, "config_local.json"))
	assert.NoError(t, err, "config_local.json should be created on first load")
}

func TestSaveCreatesLocalConfigWithPaths(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.SkillsStorageDir = filepath.Join(dir, "custom-skills")
	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.SkillsStorageDir, loaded.SkillsStorageDir)

	// config.json must NOT contain skillsStorageDir (it belongs in config_local.json)
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "skillsStorageDir")

	// config_local.json must contain the path
	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), "skillsStorageDir")
}

func TestMigrationFromLegacyConfig(t *testing.T) {
	dir := t.TempDir()
	// Write a legacy config.json that includes skillsStorageDir inline
	legacy := `{"skillsStorageDir":"` + filepath.ToSlash(filepath.Join(dir, "skills")) + `","defaultCategory":"Legacy"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(legacy), 0644))

	svc := config.NewService(dir)
	cfg, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "Legacy", cfg.DefaultCategory)

	// After migration config_local.json must exist
	_, err = os.Stat(filepath.Join(dir, "config_local.json"))
	assert.NoError(t, err, "migration should create config_local.json")

	// config.json must no longer contain skillsStorageDir
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "skillsStorageDir")
}

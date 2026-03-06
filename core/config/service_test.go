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
	err := svc.Save(cfg)
	require.NoError(t, err)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "MyCategory", loaded.DefaultCategory)
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
	assert.NoError(t, err)
}

package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLaunchAtLoginController struct {
	enabled      bool
	enableCalls  int
	disableCalls int
	enableErr    error
	disableErr   error
}

func (f *fakeLaunchAtLoginController) IsEnabled() bool {
	return f.enabled
}

func (f *fakeLaunchAtLoginController) Enable() error {
	f.enableCalls++
	if f.enableErr != nil {
		return f.enableErr
	}
	f.enabled = true
	return nil
}

func (f *fakeLaunchAtLoginController) Disable() error {
	f.disableCalls++
	if f.disableErr != nil {
		return f.disableErr
	}
	f.enabled = false
	return nil
}

func TestSyncLaunchAtLoginRewritesEnabledEntry(t *testing.T) {
	controller := &fakeLaunchAtLoginController{enabled: true}
	app := &App{
		autostartFactory: func() (launchAtLoginController, error) {
			return controller, nil
		},
	}

	err := app.syncLaunchAtLogin(true)
	require.NoError(t, err)
	assert.Equal(t, 1, controller.enableCalls)
	assert.Zero(t, controller.disableCalls)
	assert.True(t, controller.enabled)
}

func TestSyncLaunchAtLoginSkipsDisableWhenAlreadyDisabled(t *testing.T) {
	controller := &fakeLaunchAtLoginController{enabled: false}
	app := &App{
		autostartFactory: func() (launchAtLoginController, error) {
			return controller, nil
		},
	}

	err := app.syncLaunchAtLogin(false)
	require.NoError(t, err)
	assert.Zero(t, controller.enableCalls)
	assert.Zero(t, controller.disableCalls)
	assert.False(t, controller.enabled)
}

func TestSyncLaunchAtLoginIgnoresMissingEntryDuringDisable(t *testing.T) {
	controller := &fakeLaunchAtLoginController{
		enabled: true,
		disableErr: &os.PathError{
			Op:   "remove",
			Path: "entry",
			Err:  os.ErrNotExist,
		},
	}
	app := &App{
		autostartFactory: func() (launchAtLoginController, error) {
			return controller, nil
		},
	}

	err := app.syncLaunchAtLogin(false)
	require.NoError(t, err)
	assert.Equal(t, 1, controller.disableCalls)
}

func TestSaveConfigRollsBackLaunchAtLoginWhenSyncFails(t *testing.T) {
	dataDir := t.TempDir()
	svc := config.NewService(dataDir)
	controller := &fakeLaunchAtLoginController{
		enableErr: errors.New("enable failed"),
	}
	app := &App{
		config: svc,
		autostartFactory: func() (launchAtLoginController, error) {
			return controller, nil
		},
	}

	cfg := config.DefaultConfig(dataDir)
	cfg.RepoScanMaxDepth = 7
	cfg.LaunchAtLogin = true

	err := app.SaveConfig(cfg)
	require.ErrorContains(t, err, "enable failed")

	loaded, loadErr := svc.Load()
	require.NoError(t, loadErr)
	assert.Equal(t, 7, loaded.RepoScanMaxDepth)
	assert.False(t, loaded.LaunchAtLogin)
	assert.Equal(t, 1, controller.enableCalls)
}

func TestLaunchAtLoginExecutablePathUsesSiblingNativeClient(t *testing.T) {
	daemonName, clientName := "skillflowd", "SkillFlow"
	if runtime.GOOS == "windows" {
		daemonName, clientName = "skillflowd.exe", "SkillFlow.exe"
	}
	dir := t.TempDir()
	clientPath := filepath.Join(dir, clientName)
	require.NoError(t, os.WriteFile(clientPath, []byte("client"), 0o755))

	assert.Equal(
		t,
		clientPath,
		launchAtLoginExecutablePath(filepath.Join(dir, daemonName)),
	)
	assert.Equal(
		t,
		filepath.Join(dir, "other"),
		launchAtLoginExecutablePath(filepath.Join(dir, "other")),
	)
}

func TestLaunchAtLoginExecutablePathFallsBackWhenSiblingMissing(t *testing.T) {
	daemonName := "skillflowd"
	if runtime.GOOS == "windows" {
		daemonName = "skillflowd.exe"
	}
	daemonPath := filepath.Join(t.TempDir(), daemonName)

	assert.Equal(t, daemonPath, launchAtLoginExecutablePath(daemonPath))
}

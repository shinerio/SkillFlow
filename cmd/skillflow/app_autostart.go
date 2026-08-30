package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type launchAtLoginController interface {
	IsEnabled() bool
	Enable() error
	Disable() error
}

func (a *App) autostartController() (launchAtLoginController, error) {
	if a.autostartFactory != nil {
		return a.autostartFactory()
	}
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path failed: %w", err)
	}
	return newLaunchAtLoginController(launchAtLoginExecutablePath(filepath.Clean(exePath)))
}

const launchAtLoginAppName = "SkillFlow"

// launchAtLoginExecutablePath points login launchers at the visible native client when
// settings are reconciled by the bundled daemon process. The daemon is intentionally
// packaged beside the UI executable, so this preserves the user-facing behavior without
// adding a daemon-specific launch-at-login mode.
func launchAtLoginExecutablePath(exePath string) string {
	daemonName, clientName := "skillflowd", launchAtLoginAppName
	if runtime.GOOS == "windows" {
		daemonName, clientName = daemonName+".exe", clientName+".exe"
	}
	if filepath.Base(exePath) != daemonName {
		return exePath
	}
	clientPath := filepath.Join(filepath.Dir(exePath), clientName)
	if info, err := os.Stat(clientPath); err == nil && !info.IsDir() {
		return clientPath
	}
	return exePath
}

func (a *App) syncLaunchAtLogin(enabled bool) error {
	a.logInfof("launch-at-login update started: desired=%t", enabled)
	app, err := a.autostartController()
	if err != nil {
		a.logErrorf("launch-at-login update failed: desired=%t err=%v", enabled, err)
		return err
	}
	if enabled {
		if err := app.Enable(); err != nil {
			a.logErrorf("launch-at-login update failed: desired=%t action=enable err=%v", enabled, err)
			return err
		}
		a.logInfof("launch-at-login update completed: desired=%t action=enable", enabled)
		return nil
	}
	if !app.IsEnabled() {
		a.logInfof("launch-at-login update completed: desired=%t action=noop", enabled)
		return nil
	}
	if err := app.Disable(); err != nil && !errors.Is(err, os.ErrNotExist) {
		a.logErrorf("launch-at-login update failed: desired=%t action=disable err=%v", enabled, err)
		return err
	} else {
		a.logInfof("launch-at-login update completed: desired=%t action=disable", enabled)
	}
	return nil
}

//go:build !darwin && !windows

package main

func setupTray(_ *App) error {
	return nil
}

func teardownTray() {}

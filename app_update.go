package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/shinerio/skillflow/core/notify"
)

const (
	githubOwner = "shinerio"
	githubRepo  = "SkillFlow"
)

// AppUpdateInfo holds information about an available application update.
type AppUpdateInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	DownloadURL    string `json:"downloadUrl"`
	ReleaseNotes   string `json:"releaseNotes"`
	CanAutoUpdate  bool   `json:"canAutoUpdate"`
}

// GetAppVersion returns the current application version.
func (a *App) GetAppVersion() string {
	return Version
}

// CheckAppUpdate queries GitHub Releases API and returns update information.
func (a *App) CheckAppUpdate() (*AppUpdateInfo, error) {
	client := a.proxyHTTPClient()
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubOwner, githubRepo)
	req, err := http.NewRequestWithContext(a.ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	current := Version
	latest := release.TagName
	hasUpdate := latest != "" && latest != current && latest != "v"+strings.TrimPrefix(current, "v")

	// Match asset for current platform.
	downloadURL := ""
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if goruntime.GOOS == "windows" && goruntime.GOARCH == "amd64" && strings.Contains(name, "windows") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
		if goruntime.GOOS == "darwin" && goruntime.GOARCH == "amd64" && strings.Contains(name, "macos-intel") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
		if goruntime.GOOS == "darwin" && goruntime.GOARCH == "arm64" && strings.Contains(name, "macos-apple-silicon") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	return &AppUpdateInfo{
		HasUpdate:      hasUpdate,
		CurrentVersion: current,
		LatestVersion:  latest,
		ReleaseURL:     release.HTMLURL,
		DownloadURL:    downloadURL,
		ReleaseNotes:   release.Body,
		CanAutoUpdate:  goruntime.GOOS == "windows",
	}, nil
}

// DownloadAppUpdate downloads the new version to a temp file and emits progress events.
// Windows only: emits EventAppUpdateDownloadDone on success or EventAppUpdateDownloadFail on error.
func (a *App) DownloadAppUpdate(downloadURL string) error {
	go func() {
		tmpDir := os.TempDir()
		tmpPath := filepath.Join(tmpDir, "skillflow_update.exe")

		client := a.proxyHTTPClient()
		resp, err := client.Get(downloadURL)
		if err != nil {
			a.hub.Publish(notify.Event{Type: notify.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		defer resp.Body.Close()

		f, err := os.Create(tmpPath)
		if err != nil {
			a.hub.Publish(notify.Event{Type: notify.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			a.hub.Publish(notify.Event{Type: notify.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		a.hub.Publish(notify.Event{Type: notify.EventAppUpdateDownloadDone, Payload: tmpPath})
	}()
	return nil
}

// ApplyAppUpdate writes a batch script to replace the running exe then exits.
// Windows only.
func (a *App) ApplyAppUpdate() error {
	if goruntime.GOOS != "windows" {
		return fmt.Errorf("auto-update is only supported on Windows")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	tmpNew := filepath.Join(os.TempDir(), "skillflow_update.exe")
	batPath := filepath.Join(os.TempDir(), "skillflow_update.bat")
	batContent := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak > nul
move /y "%s" "%s"
start "" "%s"
del "%%~f0"
`, tmpNew, exe, exe)
	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return err
	}
	cmd := exec.Command("cmd", "/C", batPath)
	cmd.SysProcAttr = nil
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

// CheckAppUpdateAndNotify checks for updates and, if a new version is found, publishes
// EventAppUpdateAvailable so the top banner activates. Returns the update info.
func (a *App) CheckAppUpdateAndNotify() (*AppUpdateInfo, error) {
	info, err := a.CheckAppUpdate()
	if err != nil {
		return nil, err
	}
	if info.HasUpdate {
		a.hub.Publish(notify.Event{
			Type:    notify.EventAppUpdateAvailable,
			Payload: info,
		})
	}
	return info, nil
}

// checkAppUpdateOnStartup checks for app updates and emits EventAppUpdateAvailable if found.
// Skipped in dev builds to avoid noise during development.
func (a *App) checkAppUpdateOnStartup() {
	if Version == "dev" {
		return
	}
	info, err := a.CheckAppUpdate()
	if err != nil || !info.HasUpdate {
		return
	}
	a.hub.Publish(notify.Event{
		Type:    notify.EventAppUpdateAvailable,
		Payload: info,
	})
}

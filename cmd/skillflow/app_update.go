package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/shinerio/skillflow/core/platform/eventbus"
)

var newRequestWithContextFn = http.NewRequestWithContext

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

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetAppVersion returns the current application version.
func (a *App) GetAppVersion() string {
	return Version
}

// CheckAppUpdate queries GitHub Releases API and returns update information.
func (a *App) CheckAppUpdate() (*AppUpdateInfo, error) {
	a.logInfof("check app update started")
	client := a.proxyHTTPClient()
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubOwner, githubRepo)
	releasePageURL := fmt.Sprintf("https://github.com/%s/%s/releases/latest", githubOwner, githubRepo)
	req, err := newRequestWithContextFn(a.cloneContext(), "GET", apiURL, nil)
	if err != nil {
		a.logErrorf("check app update failed: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		a.logErrorf("check app update failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string         `json:"tag_name"`
		HTMLURL string         `json:"html_url"`
		Body    string         `json:"body"`
		Assets  []releaseAsset `json:"assets"`
	}
	if err := decodeGitHubJSONResponse(resp, &release); err != nil {
		a.logErrorf("check app update failed: %v", err)
		return nil, err
	}

	current := Version
	latest := release.TagName
	releaseURL := strings.TrimSpace(release.HTMLURL)
	if releaseURL == "" {
		releaseURL = releasePageURL
	}
	hasUpdate := latest != "" && latest != current && latest != "v"+strings.TrimPrefix(current, "v")

	downloadURL := selectReleaseAssetDownloadURL(goruntime.GOOS, goruntime.GOARCH, release.Assets)

	info := &AppUpdateInfo{
		HasUpdate:      hasUpdate,
		CurrentVersion: current,
		LatestVersion:  latest,
		ReleaseURL:     releaseURL,
		DownloadURL:    downloadURL,
		ReleaseNotes:   release.Body,
		CanAutoUpdate:  goruntime.GOOS == "windows",
	}
	a.logInfof("check app update completed: hasUpdate=%v latest=%s", info.HasUpdate, info.LatestVersion)
	return info, nil
}

func decodeGitHubJSONResponse(resp *http.Response, target any) error {
	if resp.StatusCode != http.StatusOK {
		return githubStatusError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func githubStatusError(resp *http.Response) error {
	var payload struct {
		Message string `json:"message"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	msg := strings.TrimSpace(string(body))
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.Message) != "" {
			msg = strings.TrimSpace(payload.Message)
		}
	}
	if msg == "" {
		return fmt.Errorf("github status %d", resp.StatusCode)
	}
	return fmt.Errorf("github status %d: %s", resp.StatusCode, msg)
}

// DownloadAppUpdate downloads the new version to a temp file and emits progress events.
// Windows only: emits EventAppUpdateDownloadDone on success or EventAppUpdateDownloadFail on error.
func (a *App) DownloadAppUpdate(downloadURL string) error {
	a.logInfof("download app update requested")
	go func() {
		tmpDir := os.TempDir()
		tmpPath := filepath.Join(tmpDir, "skillflow_update.exe")
		partPath := tmpPath + ".part"

		client := a.proxyHTTPClient()
		resp, err := client.Get(downloadURL)
		if err != nil {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("download status %d", resp.StatusCode)
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}

		f, err := os.Create(partPath)
		if err != nil {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}

		written, err := io.Copy(f, resp.Body)
		if err != nil {
			_ = f.Close()
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		if err := f.Close(); err != nil {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		if resp.ContentLength > 0 && written != resp.ContentLength {
			err := fmt.Errorf("download size mismatch: expected=%d actual=%d", resp.ContentLength, written)
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		checksum, err := fetchReleaseSHA256(client, releaseChecksumURL(downloadURL))
		if err != nil {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		if err := validateDownloadedWindowsUpdate(partPath, checksum); err != nil {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		if err := os.Rename(partPath, tmpPath); err != nil {
			a.logErrorf("download app update failed: %v", err)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadFail, Payload: err.Error()})
			return
		}
		a.logInfof("download app update completed: %s", tmpPath)
		a.hub.Publish(eventbus.Event{Type: eventbus.EventAppUpdateDownloadDone, Payload: tmpPath})
	}()
	return nil
}

func selectReleaseAssetDownloadURL(goos, goarch string, assets []releaseAsset) string {
	preferredNames := preferredReleaseAssetNames(goos, goarch)
	for _, preferred := range preferredNames {
		for _, asset := range assets {
			if strings.EqualFold(strings.TrimSpace(asset.Name), preferred) {
				return strings.TrimSpace(asset.BrowserDownloadURL)
			}
		}
	}

	for _, asset := range assets {
		name := strings.ToLower(strings.TrimSpace(asset.Name))
		if goos == "windows" && goarch == "amd64" &&
			strings.HasSuffix(name, ".exe") &&
			strings.Contains(name, "windows") &&
			!strings.Contains(name, "installer") &&
			!strings.HasSuffix(name, ".sha256") {
			return strings.TrimSpace(asset.BrowserDownloadURL)
		}
	}
	return ""
}

func preferredReleaseAssetNames(goos, goarch string) []string {
	switch {
	case goos == "windows" && goarch == "amd64":
		return []string{"SkillFlow-windows.exe"}
	case goos == "darwin" && goarch == "amd64":
		return []string{"SkillFlow-macos-intel.dmg"}
	case goos == "darwin" && goarch == "arm64":
		return []string{"SkillFlow-macos-apple-silicon.dmg"}
	default:
		return nil
	}
}

func releaseChecksumURL(downloadURL string) string {
	trimmed := strings.TrimSpace(downloadURL)
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(strings.ToLower(trimmed), ".sha256") {
		return trimmed
	}
	return trimmed + ".sha256"
}

func fetchReleaseSHA256(client *http.Client, checksumURL string) (string, error) {
	if strings.TrimSpace(checksumURL) == "" {
		return "", nil
	}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(body))
	if len(fields) == 0 {
		return "", fmt.Errorf("checksum payload is empty")
	}
	return strings.ToLower(strings.TrimSpace(fields[0])), nil
}

func validateDownloadedWindowsUpdate(path string, expectedSHA256 string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := validatePEExecutable(data); err != nil {
		return err
	}
	if strings.TrimSpace(expectedSHA256) == "" {
		return nil
	}
	actual := sha256.Sum256(data)
	actualHex := strings.ToLower(hex.EncodeToString(actual[:]))
	if actualHex != strings.ToLower(strings.TrimSpace(expectedSHA256)) {
		return fmt.Errorf("sha256 mismatch: expected=%s actual=%s", expectedSHA256, actualHex)
	}
	return nil
}

func validatePEExecutable(data []byte) error {
	if len(data) < 0x44 {
		return fmt.Errorf("downloaded file is too small to be a Windows executable")
	}
	if data[0] != 'M' || data[1] != 'Z' {
		return fmt.Errorf("downloaded file is not a Windows executable")
	}
	peOffset := int(data[0x3c]) | int(data[0x3d])<<8 | int(data[0x3e])<<16 | int(data[0x3f])<<24
	if peOffset < 0 || peOffset+4 > len(data) {
		return fmt.Errorf("downloaded file has an invalid PE header")
	}
	if data[peOffset] != 'P' || data[peOffset+1] != 'E' || data[peOffset+2] != 0 || data[peOffset+3] != 0 {
		return fmt.Errorf("downloaded file has an invalid PE signature")
	}
	return nil
}

// ApplyAppUpdate writes a batch script to replace the running exe then exits.
// Windows only.
func (a *App) ApplyAppUpdate() error {
	if goruntime.GOOS != "windows" {
		a.logErrorf("apply app update failed: unsupported os")
		return fmt.Errorf("auto-update is only supported on Windows")
	}
	exe, err := os.Executable()
	if err != nil {
		a.logErrorf("apply app update failed: %v", err)
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		a.logErrorf("apply app update failed: %v", err)
		return err
	}
	tmpNew := filepath.Join(os.TempDir(), "skillflow_update.exe")
	batPath := filepath.Join(os.TempDir(), "skillflow_update.bat")
	helperPID := lookupHelperPID(os.Getpid())
	a.logInfof("apply app update started: target=%s helperPID=%d", exe, helperPID)
	batContent := buildWindowsUpdateScript(tmpNew, exe, helperPID)
	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		a.logErrorf("apply app update failed: %v", err)
		return err
	}
	cmd := exec.Command("cmd", "/C", batPath)
	cmd.SysProcAttr = nil
	if err := cmd.Start(); err != nil {
		a.logErrorf("apply app update failed: %v", err)
		return err
	}
	a.logInfof("apply app update handoff completed: script=%s", batPath)
	os.Exit(0)
	return nil
}

func lookupHelperPID(currentPID int) int {
	if err := pruneStaleLoopbackControlState(helperControlPath()); err != nil {
		return 0
	}
	endpoint, err := readControlEndpoint(helperControlPath())
	if err != nil {
		return 0
	}
	if endpoint.PID <= 0 || endpoint.PID == currentPID {
		return 0
	}
	return endpoint.PID
}

func buildWindowsUpdateScript(tmpNew, exe string, helperPID int) string {
	killHelper := ""
	if helperPID > 0 {
		killHelper = fmt.Sprintf("taskkill /PID %d /T /F > nul 2>&1\n", helperPID)
	}
	return fmt.Sprintf(`@echo off
setlocal
timeout /t 1 /nobreak > nul
%sset /a attempts=0
:retry
set /a attempts+=1
move /y "%s" "%s" > nul 2>&1
if errorlevel 1 (
  if %%attempts%% GEQ 15 exit /b 1
  timeout /t 1 /nobreak > nul
  goto retry
)
start "" "%s"
del "%%~f0"
`, killHelper, tmpNew, exe, exe)
}

// GetSkippedUpdateVersion returns the version tag that the user chose to skip on startup prompts.
func (a *App) GetSkippedUpdateVersion() string {
	cfg, err := a.config.Load()
	if err != nil {
		return ""
	}
	return cfg.SkippedUpdateVersion
}

// SetSkippedUpdateVersion persists a version tag so that the startup update prompt is
// suppressed for that specific version. Pass an empty string to clear the skip.
func (a *App) SetSkippedUpdateVersion(version string) error {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("set skipped update version failed: %v", err)
		return err
	}
	cfg.SkippedUpdateVersion = version
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("set skipped update version failed: %v", err)
		return err
	}
	a.logInfof("set skipped update version completed: %s", version)
	return nil
}

// CheckAppUpdateAndNotify checks for updates and, if a new version is found, publishes
// EventAppUpdateAvailable so the update dialog opens. Always notifies regardless of the
// skipped version (used by the manual check in Settings).
func (a *App) CheckAppUpdateAndNotify() (*AppUpdateInfo, error) {
	info, err := a.CheckAppUpdate()
	if err != nil {
		return nil, err
	}
	if info.HasUpdate {
		a.hub.Publish(eventbus.Event{
			Type:    eventbus.EventAppUpdateAvailable,
			Payload: info,
		})
	}
	return info, nil
}

// checkAppUpdateOnStartup checks for app updates and emits EventAppUpdateAvailable if found.
// Skipped in dev builds to avoid noise during development.
// If the user previously chose "skip this version" for the detected version, the event is not emitted.
func (a *App) checkAppUpdateOnStartup() {
	if Version == "dev" {
		return
	}
	info, err := a.CheckAppUpdate()
	if err != nil || !info.HasUpdate {
		return
	}
	// Suppress the startup prompt when the user explicitly skipped this version.
	skipped := a.GetSkippedUpdateVersion()
	if skipped != "" && skipped == info.LatestVersion {
		a.logDebugf("check app update: version %s is skipped by user, suppressing startup prompt", info.LatestVersion)
		return
	}
	a.hub.Publish(eventbus.Event{
		Type:    eventbus.EventAppUpdateAvailable,
		Payload: info,
	})
}

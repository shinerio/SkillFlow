package config

import "strings"

type ToolConfig struct {
	Name     string   `json:"name"`
	ScanDirs []string `json:"scanDirs"`
	PushDir  string   `json:"pushDir"`
	Enabled  bool     `json:"enabled"`
	Custom   bool     `json:"custom"`
}

type CloudConfig struct {
	Provider            string            `json:"provider"`
	Enabled             bool              `json:"enabled"`
	BucketName          string            `json:"bucketName"`
	RemotePath          string            `json:"remotePath"`
	Credentials         map[string]string `json:"credentials"`
	SyncIntervalMinutes int               `json:"syncIntervalMinutes"` // 0 = on mutation only
}

type CloudProviderConfig struct {
	BucketName  string            `json:"bucketName"`
	RemotePath  string            `json:"remotePath"`
	Credentials map[string]string `json:"credentials"`
}

const DefaultCloudRemotePath = "skillflow/"

// ProxyMode controls how outbound HTTP requests are routed.
// "none" = direct, "system" = read HTTP_PROXY/HTTPS_PROXY env vars, "manual" = use URL field.
type ProxyMode string

const (
	ProxyModeNone   ProxyMode = "none"
	ProxyModeSystem ProxyMode = "system"
	ProxyModeManual ProxyMode = "manual"
)

type ProxyConfig struct {
	Mode ProxyMode `json:"mode"` // "none" | "system" | "manual"
	URL  string    `json:"url"`  // used when Mode == "manual", e.g. "http://127.0.0.1:7890"
}

type WindowState struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type AppConfig struct {
	SkillsStorageDir      string                         `json:"skillsStorageDir"`
	DefaultCategory       string                         `json:"defaultCategory"`
	LogLevel              string                         `json:"logLevel"` // "debug" | "info" | "error"
	RepoScanMaxDepth      int                            `json:"repoScanMaxDepth"`
	SkillStatusVisibility SkillStatusVisibilityConfig    `json:"skillStatusVisibility"`
	Tools                 []ToolConfig                   `json:"tools"`
	Cloud                 CloudConfig                    `json:"cloud"`
	CloudProfiles         map[string]CloudProviderConfig `json:"cloudProfiles,omitempty"`
	Proxy                 ProxyConfig                    `json:"proxy"`
	GitHub                GitHubConfig                   `json:"github"`
	SkippedUpdateVersion  string                         `json:"skippedUpdateVersion,omitempty"` // version tag to suppress startup update prompt
}

type GitHubConfig struct {
	PAT string `json:"pat"`
}

const (
	LogLevelDebug           = "debug"
	LogLevelInfo            = "info"
	LogLevelError           = "error"
	DefaultLogLevel         = LogLevelError
	DefaultRepoScanMaxDepth = 5
	MinRepoScanMaxDepth     = 1
	MaxRepoScanMaxDepth     = 20
	MinWindowWidth          = 640
	MinWindowHeight         = 480
)

func NormalizeLogLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case LogLevelDebug:
		return LogLevelDebug
	case LogLevelError:
		return LogLevelError
	default:
		return DefaultLogLevel
	}
}

func NormalizeRepoScanMaxDepth(depth int) int {
	if depth < MinRepoScanMaxDepth {
		return DefaultRepoScanMaxDepth
	}
	if depth > MaxRepoScanMaxDepth {
		return MaxRepoScanMaxDepth
	}
	return depth
}

func NormalizeCloudRemotePath(path string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if trimmed == "" {
		return DefaultCloudRemotePath
	}

	parts := make([]string, 0)
	for _, part := range strings.Split(trimmed, "/") {
		part = strings.TrimSpace(part)
		if part == "" || part == "." {
			continue
		}
		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return DefaultCloudRemotePath
	}
	if parts[len(parts)-1] != "skillflow" {
		parts = append(parts, "skillflow")
	}
	return strings.Join(parts, "/") + "/"
}

func NormalizeProxyConfig(proxy ProxyConfig) ProxyConfig {
	mode := ProxyMode(strings.ToLower(strings.TrimSpace(string(proxy.Mode))))
	switch mode {
	case ProxyModeSystem, ProxyModeManual:
		proxy.Mode = mode
	default:
		proxy.Mode = ProxyModeNone
	}
	proxy.URL = strings.TrimSpace(proxy.URL)
	return proxy
}

func NormalizeWindowState(state WindowState) WindowState {
	if state.Width < MinWindowWidth {
		state.Width = 0
	}
	if state.Height < MinWindowHeight {
		state.Height = 0
	}
	if state.Width == 0 || state.Height == 0 {
		return WindowState{}
	}
	return state
}

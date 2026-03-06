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

type AppConfig struct {
	SkillsStorageDir string       `json:"skillsStorageDir"`
	DefaultCategory  string       `json:"defaultCategory"`
	LogLevel         string       `json:"logLevel"` // "debug" | "info" | "error"
	Tools            []ToolConfig `json:"tools"`
	Cloud            CloudConfig  `json:"cloud"`
	Proxy            ProxyConfig  `json:"proxy"`
}

const (
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelError   = "error"
	DefaultLogLevel = LogLevelError
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

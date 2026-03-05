package config

import "time"

type FavoriteRepo struct {
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AddedAt     time.Time `json:"addedAt"`
}

type ToolConfig struct {
	Name      string `json:"name"`
	SkillsDir string `json:"skillsDir"`
	Enabled   bool   `json:"enabled"`
	Custom    bool   `json:"custom"`
}

type CloudConfig struct {
	Provider    string            `json:"provider"`
	Enabled     bool              `json:"enabled"`
	BucketName  string            `json:"bucketName"`
	RemotePath  string            `json:"remotePath"`
	Credentials map[string]string `json:"credentials"`
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
	SkillsStorageDir string         `json:"skillsStorageDir"`
	DefaultCategory  string         `json:"defaultCategory"`
	Tools            []ToolConfig   `json:"tools"`
	Cloud            CloudConfig    `json:"cloud"`
	Proxy            ProxyConfig    `json:"proxy"`
	FavoriteRepos    []FavoriteRepo `json:"favoriteRepos"`
}


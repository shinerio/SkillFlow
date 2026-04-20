package config

import (
	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	backupapp "github.com/shinerio/skillflow/core/backup/app"
	"github.com/shinerio/skillflow/core/platform/logging"
	"github.com/shinerio/skillflow/core/platform/shellsettings"
)

type AgentConfig = agentdomain.AgentProfile
type CloudConfig = backupapp.CloudConfig
type CloudProviderConfig = backupapp.CloudProviderConfig
type ProxyMode = shellsettings.ProxyMode
type ProxyConfig = shellsettings.ProxyConfig
type WindowState = shellsettings.WindowState

type AgentSkillManagementConfig = agentdomain.SkillManagementConfig
type AgentSkillGroupAssignment = agentdomain.SkillGroupAssignment
type AgentSkillAgentState = agentdomain.AgentSkillState

const (
	DefaultCloudRemotePath  = backupapp.DefaultCloudRemotePath
	ProxyModeNone           = shellsettings.ProxyModeNone
	ProxyModeSystem         = shellsettings.ProxyModeSystem
	ProxyModeManual         = shellsettings.ProxyModeManual
	LogLevelDebug           = "debug"
	LogLevelInfo            = "info"
	LogLevelError           = "error"
	DefaultLogLevel         = logging.DefaultLevelString
	DefaultRepoScanMaxDepth = agentapp.DefaultRepoScanMaxDepth
	MinRepoScanMaxDepth     = agentapp.MinRepoScanMaxDepth
	MaxRepoScanMaxDepth     = agentapp.MaxRepoScanMaxDepth
)

type AppConfig struct {
	RepoCacheDir          string                         `json:"repoCacheDir"`
	AutoUpdateSkills      bool                           `json:"autoUpdateSkills"`
	AutoPushAgents        []string                       `json:"autoPushAgents"`
	LaunchAtLogin         bool                           `json:"launchAtLogin"`
	DefaultCategory       string                         `json:"defaultCategory"`
	LogLevel              string                         `json:"logLevel"` // "debug" | "info" | "error"
	RepoScanMaxDepth      int                            `json:"repoScanMaxDepth"`
	Agents                []AgentConfig                  `json:"agents"`
	Cloud                 CloudConfig                    `json:"cloud"`
	CloudProfiles         map[string]CloudProviderConfig `json:"cloudProfiles,omitempty"`
	Proxy                 ProxyConfig                    `json:"proxy"`
	AgentSkillManagement  AgentSkillManagementConfig     `json:"agentSkillManagement,omitempty"`
	SkippedUpdateVersion  string                         `json:"skippedUpdateVersion,omitempty"` // version tag to suppress startup update prompt
}

func NormalizeLogLevel(level string) string {
	return logging.NormalizeLevelString(level)
}

func NormalizeAgentNameList(names []string) []string {
	return agentapp.NormalizeAutoPushAgentNames(names)
}

func NormalizeRepoScanMaxDepth(depth int) int {
	return agentapp.NormalizeRepoScanMaxDepth(depth)
}

func NormalizeCloudRemotePath(path string) string {
	return backupapp.NormalizeCloudRemotePath(path)
}

func NormalizeProxyConfig(proxy ProxyConfig) ProxyConfig {
	return shellsettings.NormalizeProxyConfig(proxy)
}

func NormalizeWindowState(state WindowState) WindowState {
	return shellsettings.NormalizeWindowState(state)
}

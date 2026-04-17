package config

import (
	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	backupapp "github.com/shinerio/skillflow/core/backup/app"
	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/shinerio/skillflow/core/platform/shellsettings"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
)

func AppDataDir() string {
	return appdata.Dir()
}

func DefaultConfig(dataDir string) AppConfig {
	skillSettings := skillcatalogapp.DefaultSettings(dataDir)
	agentSettings := agentapp.DefaultSettings()
	backupSettings := backupapp.DefaultSettings()
	shellSharedSettings := shellsettings.DefaultSharedSettings()
	shellLocalSettings := shellsettings.DefaultLocalSettings()

	agents := make([]AgentConfig, 0, len(agentSettings.Local.Agents))
	for _, agent := range agentSettings.Local.Agents {
		agents = append(agents, AgentConfig{
			Name:       agent.Name,
			ScanDirs:   append([]string(nil), agent.ScanDirs...),
			PushDir:    agent.PushDir,
			MemoryPath: agent.MemoryPath,
			RulesDir:   agent.RulesDir,
			Enabled:    true,
			Custom:     agent.Custom,
		})
	}
	return AppConfig{
		RepoCacheDir:     appdata.RepoCacheDir(dataDir),
		DefaultCategory:  skillSettings.Shared.DefaultCategory,
		LogLevel:         shellSharedSettings.LogLevel,
		RepoScanMaxDepth: agentSettings.Shared.RepoScanMaxDepth,
		Agents:           agents,
		Cloud:            backupSettings.Cloud,
		Proxy:            shellLocalSettings.Proxy,
	}
}

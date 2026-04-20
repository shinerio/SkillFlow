package main

import (
	"os"
	"path/filepath"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnabledAgentsReturnsEnabledAgents(t *testing.T) {
	app, _, _ := newAutoPushTestApp(t, []string{"codex"})

	agents, err := app.GetEnabledAgents()
	require.NoError(t, err)
	require.NotEmpty(t, agents)

	enabledNames := make([]string, 0, len(agents))
	for _, agent := range agents {
		assert.True(t, agent.Enabled)
		enabledNames = append(enabledNames, agent.Name)
	}
	assert.Contains(t, enabledNames, "codex")
}

func TestPushConflictUsesAgentNameField(t *testing.T) {
	conflict := agentdomain.PushConflict{
		SkillName:  "demo-skill",
		AgentName:  "codex",
		TargetPath: "/tmp/codex/demo-skill",
	}

	assert.Equal(t, "codex", conflict.AgentName)
}

func TestGetAgentMemoryPreviewReturnsAgentPreview(t *testing.T) {
	dataDir := t.TempDir()
	memoryPath := filepath.Join(dataDir, "codex", "AGENTS.md")
	rulesDir := filepath.Join(dataDir, "codex", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	require.NoError(t, os.WriteFile(memoryPath, []byte("main memory"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "sf-style.md"), []byte("style"), 0o644))

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{{
		Name:       "codex",
		ScanDirs:   []string{filepath.Join(dataDir, "skills")},
		PushDir:    filepath.Join(dataDir, "skills"),
		MemoryPath: memoryPath,
		RulesDir:   rulesDir,
		Enabled:    true,
	}}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc

	preview, err := app.GetAgentMemoryPreview("codex")
	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.Equal(t, "codex", preview.AgentName)
	assert.Equal(t, memoryPath, preview.MemoryPath)
	assert.Equal(t, rulesDir, preview.RulesDir)
	assert.True(t, preview.MainExists)
	assert.Equal(t, "main memory", preview.MainContent)
	require.Len(t, preview.Rules, 1)
	assert.Equal(t, "sf-style.md", preview.Rules[0].Name)
	assert.True(t, preview.Rules[0].Managed)
}

func TestListManagedAgentSkillsAndAllAgentSkills(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("HOME", dataDir)
	codexPushDir := filepath.Join(dataDir, "codex-push")
	codexScanDir := filepath.Join(dataDir, "codex-scan")
	claudeScanDir := filepath.Join(dataDir, "claude-scan")
	require.NoError(t, os.MkdirAll(codexPushDir, 0o755))
	require.NoError(t, os.MkdirAll(codexScanDir, 0o755))
	require.NoError(t, os.MkdirAll(claudeScanDir, 0o755))
	writeTestSkillDir(t, codexPushDir, "react-expert", "# react\n")
	writeTestSkillDir(t, codexScanDir, "go-reviewer", "# go\n")
	writeTestSkillDir(t, claudeScanDir, "react-expert", "# react\n")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{Name: "codex", PushDir: codexPushDir, ScanDirs: []string{codexPushDir, codexScanDir}, Enabled: true},
		{Name: "claude-code", PushDir: claudeScanDir, ScanDirs: []string{claudeScanDir}, Enabled: true},
	}
	cfg.AgentSkillManagement = config.AgentSkillManagementConfig{
		Groups: []string{"frontend", "backend"},
		Assignments: []config.AgentSkillGroupAssignment{
			{SkillName: "react-expert", GroupName: "frontend"},
			{SkillName: "go-reviewer", GroupName: "backend"},
		},
		AgentStates: []config.AgentSkillAgentState{
			{AgentName: "codex", DisabledGroupNames: []string{"backend"}},
		},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc

	managed, err := app.ListManagedAgentSkills("codex")
	require.NoError(t, err)
	require.Len(t, managed, 2)
	assert.Equal(t, "go-reviewer", managed[0].Name)
	assert.False(t, managed[0].Enabled)

	allEntries, err := app.ListAllAgentSkills()
	require.NoError(t, err)
	require.Len(t, allEntries, 2)
	assert.Equal(t, "react-expert", allEntries[1].Name)
	assert.ElementsMatch(t, []string{"claude-code", "codex"}, allEntries[1].Agents)
}

func TestAgentSkillManagementMutationMethods(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("HOME", dataDir)
	codexPushDir := filepath.Join(dataDir, "codex-push")
	require.NoError(t, os.MkdirAll(codexPushDir, 0o755))
	writeTestSkillDir(t, codexPushDir, "react-expert", "# react\n")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{Name: "codex", PushDir: codexPushDir, ScanDirs: []string{codexPushDir}, Enabled: true},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc

	require.NoError(t, app.CreateAgentSkillGroup("frontend"))
	require.NoError(t, app.AssignAgentSkillGroup("react-expert", "frontend"))
	require.NoError(t, app.SetAgentSkillEnabled("codex", "react-expert", false))
	require.NoError(t, app.SetAgentSkillGroupEnabled("codex", "frontend", false))
	require.NoError(t, app.RenameAgentSkillGroup("frontend", "web"))
	require.NoError(t, app.ClearAgentSkillGroup("react-expert"))
	require.NoError(t, app.DeleteAgentSkillGroup("web"))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded.AgentSkillManagement.Groups)
	assert.Empty(t, loaded.AgentSkillManagement.Assignments)
	require.Len(t, loaded.AgentSkillManagement.AgentStates, 1)
	assert.Equal(t, []string{"react-expert"}, loaded.AgentSkillManagement.AgentStates[0].DisabledSkillNames)
	assert.Empty(t, loaded.AgentSkillManagement.AgentStates[0].DisabledGroupNames)
}

func TestSetAgentSkillEnabledIgnoresUnsupportedEnabledAgents(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("HOME", dataDir)
	codexPushDir := filepath.Join(dataDir, "codex-push")
	claudePushDir := filepath.Join(dataDir, "claude-push")
	require.NoError(t, os.MkdirAll(codexPushDir, 0o755))
	require.NoError(t, os.MkdirAll(claudePushDir, 0o755))
	writeTestSkillDir(t, codexPushDir, "react-expert", "# react\n")
	writeTestSkillDir(t, claudePushDir, "react-expert", "# react\n")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{Name: "codex", PushDir: codexPushDir, ScanDirs: []string{codexPushDir}, Enabled: true},
		{Name: "claude-code", PushDir: claudePushDir, ScanDirs: []string{claudePushDir}, Enabled: true},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc

	require.NoError(t, app.SetAgentSkillEnabled("codex", "react-expert", false))

	loaded, err := svc.Load()
	require.NoError(t, err)
	require.Len(t, loaded.AgentSkillManagement.AgentStates, 1)
	assert.Equal(t, "codex", loaded.AgentSkillManagement.AgentStates[0].AgentName)
	assert.Equal(t, []string{"react-expert"}, loaded.AgentSkillManagement.AgentStates[0].DisabledSkillNames)
}

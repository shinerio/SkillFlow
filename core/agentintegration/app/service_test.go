package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	gatewayport "github.com/shinerio/skillflow/core/agentintegration/app/port/gateway"
	"github.com/shinerio/skillflow/core/agentintegration/domain"
	gatewayinfra "github.com/shinerio/skillflow/core/agentintegration/infra/gateway"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newService() *agentapp.Service {
	return agentapp.NewService(func(profile domain.AgentProfile) gatewayport.AgentGateway {
		return gatewayinfra.NewFilesystemAdapter(profile.Name, profile.PushDir)
	})
}

func writeSkillDir(t *testing.T, root, name, content string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte(content), 0644))
	return dir
}

func findCandidateByName(candidates []domain.AgentSkillCandidate, name string) (domain.AgentSkillCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.Name == name {
			return candidate, true
		}
	}
	return domain.AgentSkillCandidate{}, false
}

func findEntryByName(entries []domain.AgentSkillEntry, name string) (domain.AgentSkillEntry, bool) {
	for _, entry := range entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return domain.AgentSkillEntry{}, false
}

func TestServiceEnabledProfilesAndMissingPushDirs(t *testing.T) {
	service := newService()
	existingDir := t.TempDir()
	profiles := []domain.AgentProfile{
		{Name: "codex", PushDir: existingDir, Enabled: true},
		{Name: "claude-code", PushDir: filepath.Join(existingDir, "missing"), Enabled: false},
	}

	enabled := service.EnabledProfiles(profiles)
	require.Len(t, enabled, 1)
	assert.Equal(t, "codex", enabled[0].Name)

	missing, err := service.CheckMissingPushDirs(profiles, []string{"codex", "claude-code"})
	require.NoError(t, err)
	require.Len(t, missing, 1)
	assert.Equal(t, "claude-code", missing[0].Name)
}

func TestServicePushSkillsHandlesConflictsAndForce(t *testing.T) {
	service := newService()
	ctx := context.Background()
	srcRoot := t.TempDir()
	pushDir := t.TempDir()

	alphaDir := writeSkillDir(t, srcRoot, "alpha", "# alpha\nNew\n")
	betaDir := writeSkillDir(t, srcRoot, "beta", "# beta\nNew\n")
	existingAlpha := writeSkillDir(t, pushDir, "alpha", "# alpha\nOld\n")
	require.DirExists(t, existingAlpha)

	profiles := []domain.AgentProfile{{
		Name:    "codex",
		PushDir: pushDir,
		Enabled: true,
	}}
	skills := []*skilldomain.InstalledSkill{
		{ID: "alpha-id", Name: "alpha", Path: alphaDir},
		{ID: "beta-id", Name: "beta", Path: betaDir},
	}

	conflicts, err := service.PushSkills(ctx, profiles, []string{"codex"}, skills, false)
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	assert.Equal(t, "alpha", conflicts[0].SkillName)
	assert.FileExists(t, filepath.Join(pushDir, "beta", "skill.md"))

	_, err = service.PushSkills(ctx, profiles, []string{"codex"}, skills, true)
	require.NoError(t, err)

	alphaData, err := os.ReadFile(filepath.Join(pushDir, "alpha", "skill.md"))
	require.NoError(t, err)
	assert.Equal(t, "# alpha\nNew\n", string(alphaData))
}

func TestServiceScansAndListsAgentSkillsWithPresence(t *testing.T) {
	service := newService()
	ctx := context.Background()
	installedRoot := t.TempDir()
	pushDir := t.TempDir()
	scanDir := t.TempDir()

	installedAlphaDir := writeSkillDir(t, installedRoot, "alpha", "# alpha\nShared\n")
	writeSkillDir(t, pushDir, "alpha", "# alpha\nShared\n")
	writeSkillDir(t, scanDir, "alpha", "# alpha\nShared\n")
	writeSkillDir(t, scanDir, "beta", "# beta\nScan only\n")

	profile := domain.AgentProfile{
		Name:     "codex",
		PushDir:  pushDir,
		ScanDirs: []string{scanDir},
		Enabled:  true,
	}
	installed := []*skilldomain.InstalledSkill{{
		ID:     "alpha-id",
		Name:   "alpha",
		Path:   installedAlphaDir,
		Source: skilldomain.SourceManual,
	}}
	idx := skillquery.BuildInstalledIndex(installed)

	presence, err := service.BuildPresenceIndex(ctx, []domain.AgentProfile{profile}, idx, 5)
	require.NoError(t, err)

	candidates, err := service.ScanAgentSkills(ctx, profile, idx, presence, 5)
	require.NoError(t, err)
	require.Len(t, candidates, 2)

	alphaCandidate, ok := findCandidateByName(candidates, "alpha")
	require.True(t, ok)
	assert.True(t, alphaCandidate.Installed)
	assert.True(t, alphaCandidate.Imported)
	assert.True(t, alphaCandidate.Pushed)
	assert.Equal(t, []string{"codex"}, alphaCandidate.PushedAgents)

	betaCandidate, ok := findCandidateByName(candidates, "beta")
	require.True(t, ok)
	assert.False(t, betaCandidate.Imported)
	assert.False(t, betaCandidate.Pushed)

	entries, err := service.ListAgentSkills(ctx, profile, idx, presence, 5)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	alphaEntry, ok := findEntryByName(entries, "alpha")
	require.True(t, ok)
	assert.True(t, alphaEntry.Pushed)
	assert.True(t, alphaEntry.SeenInAgentScan)

	betaEntry, ok := findEntryByName(entries, "beta")
	require.True(t, ok)
	assert.False(t, betaEntry.Pushed)
	assert.True(t, betaEntry.SeenInAgentScan)
}

func TestServiceListManagedSkillsGroupsSameNameAndResolvesEnablement(t *testing.T) {
	service := newService()
	ctx := context.Background()
	pushDir := t.TempDir()
	scanDir := t.TempDir()

	writeSkillDir(t, pushDir, "react-expert", "# react\nPush\n")
	writeSkillDir(t, scanDir, "react-expert", "# react\nScan\n")
	writeSkillDir(t, scanDir, "go-reviewer", "# go\nScan\n")

	profile := domain.AgentProfile{
		Name:     "codex",
		PushDir:  pushDir,
		ScanDirs: []string{scanDir},
		Enabled:  true,
	}
	management := domain.SkillManagementConfig{
		Groups: []string{"frontend", "backend"},
		Assignments: []domain.SkillGroupAssignment{
			{SkillName: "react-expert", GroupName: "frontend"},
			{SkillName: "go-reviewer", GroupName: "backend"},
		},
		AgentStates: []domain.AgentSkillState{
			{
				AgentName:          "codex",
				DisabledGroupNames: []string{"backend"},
			},
		},
	}

	managed, err := service.ListManagedSkills(ctx, profile, management, 5)
	require.NoError(t, err)
	require.Len(t, managed, 2)

	assert.Equal(t, "go-reviewer", managed[0].Name)
	assert.Equal(t, "backend", managed[0].GroupName)
	assert.False(t, managed[0].Enabled)
	assert.Len(t, managed[0].Paths, 1)

	assert.Equal(t, "react-expert", managed[1].Name)
	assert.Equal(t, "frontend", managed[1].GroupName)
	assert.True(t, managed[1].Enabled)
	assert.Len(t, managed[1].Paths, 2)
}

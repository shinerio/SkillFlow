package gateway_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	agentgateway "github.com/shinerio/skillflow/core/agentintegration/infra/gateway"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSkill(t *testing.T, dir, mdName string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, mdName), []byte("# skill"), 0644))
}

func TestFilesystemAdapterPushFlattens(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	skillDir := filepath.Join(src, "coding", "my-skill")
	writeSkill(t, skillDir, "skill.md")
	sk := &skilldomain.InstalledSkill{Name: "my-skill", Path: skillDir}

	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	require.NoError(t, adapter.Push(context.Background(), []*skilldomain.InstalledSkill{sk}, dst))

	_, err := os.Stat(filepath.Join(dst, "my-skill", "skill.md"))
	assert.NoError(t, err)
}

func TestFilesystemAdapterPullFlat(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "skill-x"), "skill.md")
	writeSkill(t, filepath.Join(src, "skill-y"), "SKILL.MD")
	require.NoError(t, os.MkdirAll(filepath.Join(src, "not-a-skill"), 0755))

	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, skills, 2)
}

func TestFilesystemAdapterPullNested(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "coding", "skill-a"), "skill.md")
	writeSkill(t, filepath.Join(src, "coding", "skill-b"), "skill.md")
	writeSkill(t, filepath.Join(src, "writing", "skill-c"), "Skill.md")
	writeSkill(t, filepath.Join(src, "a", "b", "c", "skill-d"), "skill.md")
	require.NoError(t, os.MkdirAll(filepath.Join(src, "empty-category"), 0755))

	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	assert.ElementsMatch(t, []string{"skill-a", "skill-b", "skill-c", "skill-d"}, names)
}

func TestFilesystemAdapterPullWithMaxDepth(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "skills", "skill-a"), "skill.md")
	writeSkill(t, filepath.Join(src, "a", "b", "c", "skill-d"), "skill.md")

	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	skills, err := adapter.PullWithMaxDepth(context.Background(), src, 2)
	require.NoError(t, err)

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	assert.ElementsMatch(t, []string{"skill-a"}, names)
}

func TestFilesystemAdapterPullDefaultRespectsDepthLimit(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, filepath.Join(src, "a", "b", "c", "d", "e", "f", "skill-g"), "skill.md")

	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestFilesystemAdapterPullSkillNotRecursed(t *testing.T) {
	src := t.TempDir()
	skillDir := filepath.Join(src, "parent-skill")
	writeSkill(t, skillDir, "skill.md")
	writeSkill(t, filepath.Join(skillDir, "nested"), "skill.md")

	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "parent-skill", skills[0].Name)
}

func TestFilesystemAdapterPullDirNotExist(t *testing.T) {
	adapter := agentgateway.NewFilesystemAdapter("test-agent", "")
	_, err := adapter.Pull(context.Background(), "/nonexistent/path")
	assert.Error(t, err)
}

func TestCodexApplySkillEnablementDisablesEveryMatchingPath(t *testing.T) {
	configRoot := t.TempDir()
	configPath := filepath.Join(configRoot, ".codex", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, os.WriteFile(configPath, []byte("[profiles.default]\nmodel = \"gpt-5\"\n"), 0o644))
	skillRoot := t.TempDir()
	firstSkillDir := filepath.Join(skillRoot, "a", "react-expert")
	secondSkillDir := filepath.Join(skillRoot, "b", "react-expert")
	writeSkill(t, firstSkillDir, "SKILL.md")
	writeSkill(t, secondSkillDir, "skill.md")

	adapter := agentgateway.NewFilesystemAdapter("codex", "")
	skills := []agentdomain.ManagedAgentSkill{
		{
			Name:    "react-expert",
			Enabled: false,
			Paths: []string{
				firstSkillDir,
				secondSkillDir,
			},
		},
	}

	configured := adapter
	configuredValue := *configured
	configuredValueName := &configuredValue
	configuredValueNameTestHackSetCodexConfigPath(configuredValueName, configPath)

	require.NoError(t, configuredValueName.ApplySkillEnablement(agentdomain.AgentProfile{Name: "codex"}, skills))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "[profiles.default]")
	assert.NotContains(t, content, "skills.config = [")
	assert.Equal(t, 2, strings.Count(content, "[[skills.config]]"))
	assert.Contains(t, content, `path = "`+filepath.Join(firstSkillDir, "SKILL.md")+`"`)
	assert.Contains(t, content, `path = "`+filepath.Join(secondSkillDir, "skill.md")+`"`)
	assert.Contains(t, content, "enabled = false")
}

func TestCodexApplySkillEnablementRemovesManagedEntriesWhenEnabled(t *testing.T) {
	configRoot := t.TempDir()
	configPath := filepath.Join(configRoot, ".codex", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, os.WriteFile(configPath, []byte(strings.Join([]string{
		"[profiles.default]",
		`model = "gpt-5"`,
		"",
		"[[skills.config]]",
		`path = "/tmp/a/react-expert/SKILL.md"`,
		"enabled = false",
		"",
		"[[skills.config]]",
		`path = "/tmp/other/unmanaged/SKILL.md"`,
		"enabled = false",
		"",
	}, "\n")), 0o644))

	adapter := agentgateway.NewFilesystemAdapter("codex", "")
	configured := *adapter
	configuredValueName := &configured
	configuredValueNameTestHackSetCodexConfigPath(configuredValueName, configPath)

	require.NoError(t, configuredValueName.ApplySkillEnablement(agentdomain.AgentProfile{Name: "codex"}, []agentdomain.ManagedAgentSkill{
		{
			Name:    "react-expert",
			Enabled: true,
			Paths:   []string{"/tmp/a/react-expert/SKILL.md"},
		},
	}))

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.NotContains(t, content, "skills.config = [")
	assert.Equal(t, 1, strings.Count(content, "[[skills.config]]"))
	assert.NotContains(t, content, `path = "/tmp/a/react-expert/SKILL.md"`)
	assert.NotContains(t, content, "enabled = true")
	assert.Contains(t, content, `path = "/tmp/other/unmanaged/SKILL.md"`)
}

func configuredValueNameTestHackSetCodexConfigPath(adapter interface{}, configPath string) {
	switch typed := adapter.(type) {
	case interface{ TestSetCodexConfigPath(string) }:
		typed.TestSetCodexConfigPath(configPath)
	case *agentgateway.FilesystemAdapter:
		typed.TestSetCodexConfigPath(configPath)
	}
}

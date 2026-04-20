package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
)

func TestResolveSkillGroupByName(t *testing.T) {
	group := domain.ResolveSkillGroup("react-expert", []domain.SkillGroupAssignment{
		{SkillName: "react-expert", GroupName: "frontend"},
		{SkillName: "go-reviewer", GroupName: "backend"},
	})
	assert.Equal(t, "frontend", group)
}

func TestDisabledWhenSkillNameIsDirectlyDisabled(t *testing.T) {
	state := domain.ResolveManagedSkillState("codex", "react-expert", domain.SkillManagementConfig{
		Assignments: []domain.SkillGroupAssignment{
			{SkillName: "react-expert", GroupName: "frontend"},
		},
		AgentStates: []domain.AgentSkillState{
			{AgentName: "codex", DisabledSkillNames: []string{"react-expert"}},
		},
	})

	assert.True(t, state.Disabled)
	assert.Equal(t, "skill", state.DisabledBy)
	assert.Equal(t, "frontend", state.GroupName)
}

func TestDisabledWhenGroupIsDisabled(t *testing.T) {
	state := domain.ResolveManagedSkillState("codex", "react-expert", domain.SkillManagementConfig{
		Assignments: []domain.SkillGroupAssignment{
			{SkillName: "react-expert", GroupName: "frontend"},
		},
		AgentStates: []domain.AgentSkillState{
			{AgentName: "codex", DisabledGroupNames: []string{"frontend"}},
		},
	})

	assert.True(t, state.Disabled)
	assert.Equal(t, "group", state.DisabledBy)
}

func TestUngroupedSkillIgnoresGroupDisable(t *testing.T) {
	state := domain.ResolveManagedSkillState("codex", "react-expert", domain.SkillManagementConfig{
		AgentStates: []domain.AgentSkillState{
			{AgentName: "codex", DisabledGroupNames: []string{"frontend"}},
		},
	})

	assert.False(t, state.Disabled)
	assert.Equal(t, domain.UngroupedSkillGroup, state.GroupName)
}

func TestCollapseSameNameInstancesForManagement(t *testing.T) {
	collapsed := domain.CollapseManagedSkillInstances("codex", []domain.ManagedSkillInstance{
		{SkillName: "react-expert", Path: "/a/one"},
		{SkillName: "react-expert", Path: "/a/two"},
		{SkillName: "go-reviewer", Path: "/b/one"},
	}, domain.SkillManagementConfig{
		Assignments: []domain.SkillGroupAssignment{
			{SkillName: "react-expert", GroupName: "frontend"},
		},
		AgentStates: []domain.AgentSkillState{
			{AgentName: "codex", DisabledGroupNames: []string{"frontend"}},
		},
	})

	assert.Len(t, collapsed, 2)
	for _, entry := range collapsed {
		if entry.SkillName == "react-expert" {
			assert.ElementsMatch(t, []string{"/a/one", "/a/two"}, entry.Paths)
			assert.True(t, entry.Disabled)
			assert.Equal(t, "frontend", entry.GroupName)
		}
	}
}

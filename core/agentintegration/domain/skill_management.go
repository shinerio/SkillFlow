package domain

import "strings"

const UngroupedSkillGroup = "Ungrouped"

type SkillGroupAssignment struct {
	SkillName string
	GroupName string
}

type AgentSkillState struct {
	AgentName          string
	DisabledSkillNames []string
	DisabledGroupNames []string
}

type SkillManagementConfig struct {
	Groups      []string
	Assignments []SkillGroupAssignment
	AgentStates []AgentSkillState
}

type AgentSkillManagement = SkillManagementConfig

type ManagedSkillInstance struct {
	SkillName string
	Path      string
}

type ManagedSkillName struct {
	SkillName   string
	GroupName   string
	Paths       []string
	Disabled    bool
	DisabledBy  string
}

func ResolveSkillGroup(skillName string, assignments []SkillGroupAssignment) string {
	skillName = strings.TrimSpace(skillName)
	if skillName == "" {
		return ""
	}
	for _, assignment := range assignments {
		if strings.TrimSpace(assignment.SkillName) != skillName {
			continue
		}
		groupName := strings.TrimSpace(assignment.GroupName)
		if groupName == "" {
			return ""
		}
		return groupName
	}
	return ""
}

func FindAgentSkillState(agentName string, states []AgentSkillState) (AgentSkillState, bool) {
	agentName = strings.TrimSpace(agentName)
	for _, state := range states {
		if strings.TrimSpace(state.AgentName) == agentName {
			return state, true
		}
	}
	return AgentSkillState{}, false
}

func FindAssignedGroup(assignments []SkillGroupAssignment, skillName string) string {
	return normalizedGroupName(ResolveSkillGroup(skillName, assignments))
}

func IsSkillDisabled(state AgentSkillState, groupName, skillName string) bool {
	if containsNormalized(state.DisabledSkillNames, skillName) {
		return true
	}
	groupName = strings.TrimSpace(groupName)
	return groupName != "" && groupName != UngroupedSkillGroup && containsNormalized(state.DisabledGroupNames, groupName)
}

func ResolveManagedSkillState(agentName, skillName string, cfg SkillManagementConfig) ManagedSkillName {
	groupName := ResolveSkillGroup(skillName, cfg.Assignments)
	state, _ := FindAgentSkillState(agentName, cfg.AgentStates)
	if containsNormalized(state.DisabledSkillNames, skillName) {
		return ManagedSkillName{
			SkillName:  strings.TrimSpace(skillName),
			GroupName:  normalizedGroupName(groupName),
			Disabled:   true,
			DisabledBy: "skill",
		}
	}
	if groupName != "" && containsNormalized(state.DisabledGroupNames, groupName) {
		return ManagedSkillName{
			SkillName:  strings.TrimSpace(skillName),
			GroupName:  groupName,
			Disabled:   true,
			DisabledBy: "group",
		}
	}
	return ManagedSkillName{
		SkillName: strings.TrimSpace(skillName),
		GroupName: normalizedGroupName(groupName),
	}
}

func CollapseManagedSkillInstances(agentName string, instances []ManagedSkillInstance, cfg SkillManagementConfig) []ManagedSkillName {
	byName := map[string]ManagedSkillName{}
	for _, instance := range instances {
		skillName := strings.TrimSpace(instance.SkillName)
		path := strings.TrimSpace(instance.Path)
		if skillName == "" || path == "" {
			continue
		}
		entry, ok := byName[skillName]
		if !ok {
			entry = ResolveManagedSkillState(agentName, skillName, cfg)
		}
		entry.Paths = appendIfMissing(entry.Paths, path)
		byName[skillName] = entry
	}
	result := make([]ManagedSkillName, 0, len(byName))
	for _, entry := range byName {
		result = append(result, entry)
	}
	return result
}

func normalizedGroupName(groupName string) string {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		return UngroupedSkillGroup
	}
	return groupName
}

func containsNormalized(values []string, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == want {
			return true
		}
	}
	return false
}

func appendIfMissing(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}

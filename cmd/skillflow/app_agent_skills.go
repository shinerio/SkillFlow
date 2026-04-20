package main

import (
	"fmt"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	readmodelagentskills "github.com/shinerio/skillflow/core/readmodel/agentskills"
	"github.com/shinerio/skillflow/core/config"
)

func (a *App) ListManagedAgentSkills(agentName string) ([]agentdomain.ManagedAgentSkill, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	profile, ok := findEnabledAgentProfile(cfg.Agents, agentName)
	if !ok {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	return newAgentIntegrationService().ListManagedSkills(a.ctx, profile, cfg.AgentSkillManagement, a.repoScanMaxDepth())
}

func (a *App) ListAllAgentSkills() ([]readmodelagentskills.AllEntry, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	entriesByAgent := map[string][]agentdomain.ManagedAgentSkill{}
	for _, profile := range cfg.Agents {
		if !profile.Enabled {
			continue
		}
		entries, err := newAgentIntegrationService().ListManagedSkills(a.ctx, profile, cfg.AgentSkillManagement, a.repoScanMaxDepth())
		if err != nil {
			return nil, err
		}
		entriesByAgent[profile.Name] = entries
	}
	return readmodelagentskills.ComposeAllEntries(entriesByAgent), nil
}

func (a *App) CreateAgentSkillGroup(name string) error {
	name = strings.TrimSpace(name)
	a.logInfof("create agent skill group started: group=%s", name)
	if name == "" {
		err := fmt.Errorf("group name is required")
		a.logErrorf("create agent skill group failed: group=%s err=%v", name, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		cfg.AgentSkillManagement.Groups = append(cfg.AgentSkillManagement.Groups, name)
		return nil
	})
}

func (a *App) RenameAgentSkillGroup(oldName, newName string) error {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	a.logInfof("rename agent skill group started: from=%s to=%s", oldName, newName)
	if oldName == "" || newName == "" {
		err := fmt.Errorf("group name is required")
		a.logErrorf("rename agent skill group failed: from=%s to=%s err=%v", oldName, newName, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		for index, groupName := range cfg.AgentSkillManagement.Groups {
			if groupName == oldName {
				cfg.AgentSkillManagement.Groups[index] = newName
			}
		}
		for index, assignment := range cfg.AgentSkillManagement.Assignments {
			if assignment.GroupName == oldName {
				cfg.AgentSkillManagement.Assignments[index].GroupName = newName
			}
		}
		for index, state := range cfg.AgentSkillManagement.AgentStates {
			cfg.AgentSkillManagement.AgentStates[index].DisabledGroupNames = replaceString(state.DisabledGroupNames, oldName, newName)
		}
		return nil
	})
}

func (a *App) DeleteAgentSkillGroup(name string) error {
	name = strings.TrimSpace(name)
	a.logInfof("delete agent skill group started: group=%s", name)
	if name == "" {
		err := fmt.Errorf("group name is required")
		a.logErrorf("delete agent skill group failed: group=%s err=%v", name, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		cfg.AgentSkillManagement.Groups = removeString(cfg.AgentSkillManagement.Groups, name)
		assignments := make([]config.AgentSkillGroupAssignment, 0, len(cfg.AgentSkillManagement.Assignments))
		for _, assignment := range cfg.AgentSkillManagement.Assignments {
			if assignment.GroupName == name {
				continue
			}
			assignments = append(assignments, assignment)
		}
		cfg.AgentSkillManagement.Assignments = assignments
		for index, state := range cfg.AgentSkillManagement.AgentStates {
			cfg.AgentSkillManagement.AgentStates[index].DisabledGroupNames = removeString(state.DisabledGroupNames, name)
		}
		return nil
	})
}

func (a *App) AssignAgentSkillGroup(skillName, groupName string) error {
	skillName = strings.TrimSpace(skillName)
	groupName = strings.TrimSpace(groupName)
	a.logInfof("assign agent skill group started: skill=%s group=%s", skillName, groupName)
	if skillName == "" || groupName == "" {
		err := fmt.Errorf("skill and group are required")
		a.logErrorf("assign agent skill group failed: skill=%s group=%s err=%v", skillName, groupName, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		for index, assignment := range cfg.AgentSkillManagement.Assignments {
			if assignment.SkillName == skillName {
				cfg.AgentSkillManagement.Assignments[index].GroupName = groupName
				return nil
			}
		}
		cfg.AgentSkillManagement.Assignments = append(cfg.AgentSkillManagement.Assignments, config.AgentSkillGroupAssignment{
			SkillName: skillName,
			GroupName: groupName,
		})
		return nil
	})
}

func (a *App) ClearAgentSkillGroup(skillName string) error {
	skillName = strings.TrimSpace(skillName)
	a.logInfof("clear agent skill group started: skill=%s", skillName)
	if skillName == "" {
		err := fmt.Errorf("skill is required")
		a.logErrorf("clear agent skill group failed: skill=%s err=%v", skillName, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		assignments := make([]config.AgentSkillGroupAssignment, 0, len(cfg.AgentSkillManagement.Assignments))
		for _, assignment := range cfg.AgentSkillManagement.Assignments {
			if assignment.SkillName == skillName {
				continue
			}
			assignments = append(assignments, assignment)
		}
		cfg.AgentSkillManagement.Assignments = assignments
		return nil
	})
}

func (a *App) SetAgentSkillEnabled(agentName, skillName string, enabled bool) error {
	agentName = strings.TrimSpace(agentName)
	skillName = strings.TrimSpace(skillName)
	a.logInfof("set agent skill enabled started: agent=%s skill=%s enabled=%t", agentName, skillName, enabled)
	if agentName == "" || skillName == "" {
		err := fmt.Errorf("agent and skill are required")
		a.logErrorf("set agent skill enabled failed: agent=%s skill=%s enabled=%t err=%v", agentName, skillName, enabled, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		state := ensureAgentSkillState(&cfg.AgentSkillManagement, agentName)
		if enabled {
			state.DisabledSkillNames = removeString(state.DisabledSkillNames, skillName)
		} else {
			state.DisabledSkillNames = appendUniqueString(state.DisabledSkillNames, skillName)
		}
		return nil
	})
}

func (a *App) SetAgentSkillGroupEnabled(agentName, groupName string, enabled bool) error {
	agentName = strings.TrimSpace(agentName)
	groupName = strings.TrimSpace(groupName)
	a.logInfof("set agent skill group enabled started: agent=%s group=%s enabled=%t", agentName, groupName, enabled)
	if agentName == "" || groupName == "" {
		err := fmt.Errorf("agent and group are required")
		a.logErrorf("set agent skill group enabled failed: agent=%s group=%s enabled=%t err=%v", agentName, groupName, enabled, err)
		return err
	}
	return a.mutateAgentSkillManagement(func(cfg *config.AppConfig) error {
		state := ensureAgentSkillState(&cfg.AgentSkillManagement, agentName)
		if enabled {
			state.DisabledGroupNames = removeString(state.DisabledGroupNames, groupName)
		} else {
			state.DisabledGroupNames = appendUniqueString(state.DisabledGroupNames, groupName)
		}
		return nil
	})
}

func (a *App) mutateAgentSkillManagement(mutate func(cfg *config.AppConfig) error) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	if err := mutate(&cfg); err != nil {
		return err
	}
	if err := a.config.Save(cfg); err != nil {
		return err
	}
	for _, profile := range cfg.Agents {
		if !profile.Enabled {
			continue
		}
		if _, err := newAgentIntegrationService().ApplyManagedSkillEnablement(a.ctx, profile, cfg.AgentSkillManagement, a.repoScanMaxDepth()); err != nil {
			a.logErrorf("apply agent skill enablement failed: agent=%s err=%v", profile.Name, err)
			return err
		}
	}
	return nil
}

func ensureAgentSkillState(cfg *config.AgentSkillManagementConfig, agentName string) *config.AgentSkillAgentState {
	for index := range cfg.AgentStates {
		if cfg.AgentStates[index].AgentName == agentName {
			return &cfg.AgentStates[index]
		}
	}
	cfg.AgentStates = append(cfg.AgentStates, config.AgentSkillAgentState{AgentName: agentName})
	return &cfg.AgentStates[len(cfg.AgentStates)-1]
}

func findEnabledAgentProfile(profiles []config.AgentConfig, agentName string) (config.AgentConfig, bool) {
	for _, profile := range profiles {
		if profile.Name == agentName && profile.Enabled {
			return profile, true
		}
	}
	return config.AgentConfig{}, false
}

func removeString(values []string, target string) []string {
	filtered := values[:0]
	for _, value := range values {
		if value == target {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func replaceString(values []string, oldValue, newValue string) []string {
	for index, value := range values {
		if value == oldValue {
			values[index] = newValue
		}
	}
	return values
}

func appendUniqueString(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}

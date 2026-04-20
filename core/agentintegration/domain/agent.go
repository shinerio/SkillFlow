package domain

import (
	"sort"
	"strings"
)

type AgentProfile struct {
	Name       string   `json:"name"`
	ScanDirs   []string `json:"scanDirs"`
	PushDir    string   `json:"pushDir"`
	MemoryPath string   `json:"memoryPath"` // path to agent's main memory file
	RulesDir   string   `json:"rulesDir"`   // path to agent's rules directory
	Enabled    bool     `json:"enabled"`
	Custom     bool     `json:"custom"`
}

type MissingPushDir struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
}

type PushConflict struct {
	SkillID    string `json:"skillId,omitempty"`
	SkillName  string `json:"skillName"`
	SkillPath  string `json:"skillPath,omitempty"`
	AgentName  string `json:"agentName"`
	TargetPath string `json:"targetPath"`
}

type AgentSkillCandidate struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Source       string   `json:"source"`
	LogicalKey   string   `json:"logicalKey"`
	GroupName    string   `json:"groupName,omitempty"`
	Enabled      bool     `json:"enabled"`
	Installed    bool     `json:"installed"`
	Imported     bool     `json:"imported"`
	Updatable    bool     `json:"updatable"`
	Pushed       bool     `json:"pushed"`
	PushedAgents []string `json:"pushedAgents"`
}

type AgentSkillEntry struct {
	Name            string   `json:"name"`
	Path            string   `json:"path"`
	Source          string   `json:"source"`
	LogicalKey      string   `json:"logicalKey"`
	GroupName       string   `json:"groupName,omitempty"`
	Enabled         bool     `json:"enabled"`
	Installed       bool     `json:"installed"`
	Imported        bool     `json:"imported"`
	Updatable       bool     `json:"updatable"`
	Pushed          bool     `json:"pushed"`
	PushedAgents    []string `json:"pushedAgents"`
	SeenInAgentScan bool     `json:"seenInAgentScan"`
}

type ManagedAgentSkill struct {
	Name      string   `json:"name"`
	GroupName string   `json:"groupName,omitempty"`
	Paths     []string `json:"paths"`
	Agents    []string `json:"agents,omitempty"`
	Enabled   bool     `json:"enabled"`
}

type AgentPresenceIndex struct {
	agentsByKey map[string][]string
}

func NewAgentPresenceIndex() *AgentPresenceIndex {
	return &AgentPresenceIndex{agentsByKey: map[string][]string{}}
}

func (p *AgentPresenceIndex) Add(agentName string, keys ...string) {
	if p == nil || strings.TrimSpace(agentName) == "" {
		return
	}
	if p.agentsByKey == nil {
		p.agentsByKey = map[string][]string{}
	}
	for _, key := range compactKeys(keys...) {
		p.agentsByKey[key] = appendUniqueAgent(p.agentsByKey[key], agentName)
	}
}

func (p *AgentPresenceIndex) Agents(keys ...string) []string {
	if p == nil {
		return nil
	}
	var merged []string
	for _, key := range compactKeys(keys...) {
		merged = mergeAgentLists(merged, p.agentsByKey[key])
	}
	sort.Strings(merged)
	return merged
}

func (p *AgentPresenceIndex) KeysForAgent(agentName string) []string {
	if p == nil || strings.TrimSpace(agentName) == "" {
		return nil
	}
	keys := make([]string, 0, len(p.agentsByKey))
	for key, agentNames := range p.agentsByKey {
		for _, existing := range agentNames {
			if existing == agentName {
				keys = append(keys, key)
				break
			}
		}
	}
	sort.Strings(keys)
	return keys
}

func compactKeys(keys ...string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

func mergeAgentLists(primary, secondary []string) []string {
	merged := append([]string{}, primary...)
	for _, agentName := range secondary {
		merged = appendUniqueAgent(merged, agentName)
	}
	return merged
}

func appendUniqueAgent(agents []string, agentName string) []string {
	for _, existing := range agents {
		if existing == agentName {
			return agents
		}
	}
	return append(agents, agentName)
}

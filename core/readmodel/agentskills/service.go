package agentskills

import (
	"sort"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
)

func ComposeAllEntries(entriesByAgent map[string][]agentdomain.ManagedAgentSkill) []AllEntry {
	byName := map[string]*AllEntry{}
	for agentName, skills := range entriesByAgent {
		for _, skill := range skills {
			entry := byName[skill.Name]
			if entry == nil {
				entry = &AllEntry{
					Name:      skill.Name,
					GroupName: skill.GroupName,
				}
				byName[skill.Name] = entry
			}
			if entry.GroupName == "" {
				entry.GroupName = skill.GroupName
			}
			entry.Agents = appendUnique(entry.Agents, agentName)
			entry.InstanceCount += len(skill.Paths)
		}
	}
	result := make([]AllEntry, 0, len(byName))
	for _, entry := range byName {
		sort.Strings(entry.Agents)
		result = append(result, *entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].GroupName == result[j].GroupName {
			return result[i].Name < result[j].Name
		}
		return result[i].GroupName < result[j].GroupName
	})
	return result
}

func appendUnique(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}


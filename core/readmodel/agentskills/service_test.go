package agentskills

import (
	"encoding/json"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeAllEntriesAggregatesByNameAcrossAgents(t *testing.T) {
	entries := ComposeAllEntries(map[string][]agentdomain.ManagedAgentSkill{
		"codex": {
			{
				Name:      "react-expert",
				GroupName: "frontend",
				Paths:     []string{"/a/one", "/a/two"},
				Enabled:   true,
			},
		},
		"claude-code": {
			{
				Name:      "react-expert",
				GroupName: "frontend",
				Paths:     []string{"/b/one"},
				Enabled:   true,
			},
			{
				Name:      "go-reviewer",
				GroupName: "backend",
				Paths:     []string{"/b/two"},
				Enabled:   false,
			},
		},
	})

	assert.Len(t, entries, 2)
	assert.Equal(t, "go-reviewer", entries[0].Name)
	assert.Equal(t, "backend", entries[0].GroupName)
	assert.Equal(t, []string{"claude-code"}, entries[0].Agents)
	assert.Equal(t, 1, entries[0].InstanceCount)

	assert.Equal(t, "react-expert", entries[1].Name)
	assert.Equal(t, "frontend", entries[1].GroupName)
	assert.Equal(t, []string{"claude-code", "codex"}, entries[1].Agents)
	assert.Equal(t, 3, entries[1].InstanceCount)
}

func TestAllEntryJSONUsesFrontendFieldNames(t *testing.T) {
	data, err := json.Marshal(AllEntry{
		Name:          "react-expert",
		GroupName:     "frontend",
		Agents:        []string{"codex"},
		InstanceCount: 2,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"name": "react-expert",
		"groupName": "frontend",
		"agents": ["codex"],
		"instanceCount": 2
	}`, string(data))
}

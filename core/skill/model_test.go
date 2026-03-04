package skill_test

import (
	"testing"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
)

func TestSkillSourceTypes(t *testing.T) {
	s := skill.Skill{
		ID:       "test-id",
		Name:     "my-skill",
		Source:   skill.SourceGitHub,
		Category: "coding",
	}
	assert.Equal(t, skill.SourceType("github"), s.Source)
	assert.True(t, s.IsGitHub())
	assert.False(t, s.IsManual())
}

func TestSkillIsManual(t *testing.T) {
	s := skill.Skill{Source: skill.SourceManual}
	assert.True(t, s.IsManual())
	assert.False(t, s.IsGitHub())
}

func TestSkillHasUpdate(t *testing.T) {
	s := skill.Skill{
		Source:    skill.SourceGitHub,
		SourceSHA: "abc123",
		LatestSHA: "def456",
	}
	assert.True(t, s.HasUpdate())

	s.LatestSHA = "abc123"
	assert.False(t, s.HasUpdate())
}

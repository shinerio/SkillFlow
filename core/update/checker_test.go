package update_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shinerio/skillflow/core/skill"
	"github.com/shinerio/skillflow/core/update"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckerDetectsUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"sha": "newsha123"}})
	}))
	defer srv.Close()

	checker := update.NewChecker(srv.URL)
	sk := &skill.Skill{
		Source:        skill.SourceGitHub,
		SourceURL:     "https://github.com/user/repo",
		SourceSubPath: "skills/skill-a",
		SourceSHA:     "oldsha456",
	}
	result, err := checker.Check(context.Background(), sk)
	require.NoError(t, err)
	assert.True(t, result.HasUpdate)
	assert.Equal(t, "newsha123", result.LatestSHA)
}

func TestCheckerNoUpdateWhenSHAMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"sha": "sameSHA"}})
	}))
	defer srv.Close()

	checker := update.NewChecker(srv.URL)
	sk := &skill.Skill{
		Source:    skill.SourceGitHub,
		SourceSHA: "sameSHA",
	}
	result, err := checker.Check(context.Background(), sk)
	require.NoError(t, err)
	assert.False(t, result.HasUpdate)
}

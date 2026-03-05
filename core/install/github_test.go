package install_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shinerio/skillflow/core/install"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockGitHubServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	// Mock: list skills directory contents
	mux.HandleFunc("/repos/user/repo/contents/skills", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "skill-a", "type": "dir", "path": "skills/skill-a"},
			{"name": "skill-b", "type": "dir", "path": "skills/skill-b"},
			{"name": "readme.md", "type": "file", "path": "skills/readme.md"},
		}
		json.NewEncoder(w).Encode(items)
	})
	// Mock: check SKILLS.md existence for skill-a (returns file info)
	mux.HandleFunc("/repos/user/repo/contents/skills/skill-a/SKILLS.md", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"name": "SKILLS.md", "type": "file"})
	})
	// Mock: skill-b has no SKILLS.md (404)
	mux.HandleFunc("/repos/user/repo/contents/skills/skill-b/SKILLS.md", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

func TestGitHubInstallerScan(t *testing.T) {
	srv := mockGitHubServer(t)
	defer srv.Close()

	installer := install.NewGitHubInstaller(srv.URL, nil)
	candidates, err := installer.Scan(context.Background(), install.InstallSource{
		Type: "github",
		URI:  srv.URL + "/repos/user/repo",
	})
	require.NoError(t, err)
	// Only skill-a has SKILLS.md, skill-b does not
	assert.Len(t, candidates, 1)
	assert.Equal(t, "skill-a", candidates[0].Name)
}

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

	// Root of skills subdir: two skill dirs
	mux.HandleFunc("/repos/user/repo/contents/skills", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "skill-a", "type": "dir", "path": "skills/skill-a"},
			{"name": "skill-b", "type": "dir", "path": "skills/skill-b"},
			{"name": "readme.md", "type": "file", "path": "skills/readme.md"},
		}
		json.NewEncoder(w).Encode(items)
	})

	// skill-a contains a skill.md file → valid skill
	mux.HandleFunc("/repos/user/repo/contents/skills/skill-a", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "skill.md", "type": "file", "path": "skills/skill-a/skill.md"},
			{"name": "prompt.md", "type": "file", "path": "skills/skill-a/prompt.md"},
		}
		json.NewEncoder(w).Encode(items)
	})

	// skill-b has no skill.md → not a valid skill
	mux.HandleFunc("/repos/user/repo/contents/skills/skill-b", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "README.md", "type": "file", "path": "skills/skill-b/README.md"},
		}
		json.NewEncoder(w).Encode(items)
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
	// Only skill-a has skill.md
	assert.Len(t, candidates, 1)
	assert.Equal(t, "skill-a", candidates[0].Name)
}

func TestGitHubInstallerScanFallbackToRoot(t *testing.T) {
	// Repo with no "skills/" subdir — skills live at root
	mux := http.NewServeMux()

	// skills/ path returns 404
	mux.HandleFunc("/repos/user/repo/contents/skills", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
	})
	// Root listing has one skill dir
	mux.HandleFunc("/repos/user/repo/contents/", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "my-skill", "type": "dir", "path": "my-skill"},
		}
		json.NewEncoder(w).Encode(items)
	})
	// my-skill has SKILLS.md (uppercase variant)
	mux.HandleFunc("/repos/user/repo/contents/my-skill", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{
			{"name": "SKILLS.md", "type": "file", "path": "my-skill/SKILLS.md"},
		}
		json.NewEncoder(w).Encode(items)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	installer := install.NewGitHubInstaller(srv.URL, nil)
	candidates, err := installer.Scan(context.Background(), install.InstallSource{
		Type: "github",
		URI:  srv.URL + "/repos/user/repo",
	})
	require.NoError(t, err)
	assert.Len(t, candidates, 1)
	assert.Equal(t, "my-skill", candidates[0].Name)
}

func TestParseGitHubURIStripsGitSuffix(t *testing.T) {
	// Verify .git suffix is stripped by scanning a .git URL
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/anthropics/skills/contents/skills", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
	})
	mux.HandleFunc("/repos/anthropics/skills/contents/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	installer := install.NewGitHubInstaller(srv.URL, nil)
	// Should not error (would error with 404 on /repos/anthropics/skills.git/... if .git not stripped)
	_, err := installer.Scan(context.Background(), install.InstallSource{
		Type: "github",
		URI:  srv.URL + "/anthropics/skills.git",
	})
	require.NoError(t, err)
}

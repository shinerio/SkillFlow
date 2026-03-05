package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type githubContent struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

type GitHubInstaller struct {
	baseURL string // overridable for tests
	client  *http.Client
}

// NewGitHubInstaller creates a GitHubInstaller. Pass nil for client to use http.DefaultClient.
func NewGitHubInstaller(baseURL string, client *http.Client) *GitHubInstaller {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &GitHubInstaller{baseURL: baseURL, client: client}
}

func (g *GitHubInstaller) Type() string { return "github" }

// Scan finds skill candidates in a GitHub repo.
// It first tries the "skills/" subdirectory, then falls back to the repo root,
// so repos like anthropics/skills (skills at root) are handled correctly.
func (g *GitHubInstaller) Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error) {
	owner, repo, err := parseGitHubURI(source.URI)
	if err != nil {
		return nil, err
	}

	// Try "skills/" subdir first; if not found, scan repo root.
	items, err := g.listContents(ctx, owner, repo, "skills")
	if err != nil {
		// Fall back to root — repos like anthropics/skills keep skills at root level.
		items, err = g.listContents(ctx, owner, repo, "")
		if err != nil {
			return nil, err
		}
	}

	var candidates []SkillCandidate
	for _, item := range items {
		if item.Type != "dir" {
			continue
		}
		if g.hasSkillFile(ctx, owner, repo, item.Path) {
			candidates = append(candidates, SkillCandidate{
				Name: item.Name,
				Path: item.Path,
			})
		}
	}
	return candidates, nil
}

// hasSkillFile reports whether a directory in the repo contains a skill.md / skills.md
// file (case-insensitive). Uses one API call per directory instead of per-filename probing.
func (g *GitHubInstaller) hasSkillFile(ctx context.Context, owner, repo, path string) bool {
	files, err := g.listContents(ctx, owner, repo, path)
	if err != nil {
		return false
	}
	for _, f := range files {
		if f.Type == "file" {
			lower := strings.ToLower(f.Name)
			if lower == "skill.md" || lower == "skills.md" {
				return true
			}
		}
	}
	return false
}

func (g *GitHubInstaller) Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error {
	owner, repo, err := parseGitHubURI(source.URI)
	if err != nil {
		return err
	}
	for _, c := range selected {
		if err := g.downloadDir(ctx, owner, repo, c.Path, category, c.Name); err != nil {
			return fmt.Errorf("install %s: %w", c.Name, err)
		}
	}
	return nil
}

// listContents calls the GitHub Contents API and returns the items in a directory.
// Returns an error (with the API message) if the HTTP response is not 200.
func (g *GitHubInstaller) listContents(ctx context.Context, owner, repo, path string) ([]githubContent, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, path)
	req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// GitHub returns {"message":"Not Found",...} for 404 and similar errors.
		var apiErr struct{ Message string `json:"message"` }
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("GitHub API %d: %s", resp.StatusCode, apiErr.Message)
		}
		return nil, fmt.Errorf("GitHub API %d", resp.StatusCode)
	}
	var items []githubContent
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("unexpected GitHub response: %w", err)
	}
	return items, nil
}

func (g *GitHubInstaller) downloadDir(ctx context.Context, owner, repo, remotePath, category, name string) error {
	items, err := g.listContents(ctx, owner, repo, remotePath)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Type == "dir" {
			if err := g.downloadDir(ctx, owner, repo, item.Path, category, name); err != nil {
				return err
			}
		} else if item.DownloadURL != "" {
			if err := g.downloadFile(ctx, item.DownloadURL, name, item.Path, remotePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *GitHubInstaller) downloadFile(ctx context.Context, url, skillName, filePath, basePath string) error {
	rel := strings.TrimPrefix(filePath, basePath+"/")
	tmpDir := filepath.Join(os.TempDir(), "skillflow-install", skillName)
	target := filepath.Join(tmpDir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// DownloadTo downloads a skill candidate from GitHub into targetDir.
func (g *GitHubInstaller) DownloadTo(ctx context.Context, source InstallSource, c SkillCandidate, targetDir string) error {
	owner, repo, err := parseGitHubURI(source.URI)
	if err != nil {
		return err
	}
	return g.downloadDirTo(ctx, owner, repo, c.Path, targetDir)
}

func (g *GitHubInstaller) downloadDirTo(ctx context.Context, owner, repo, remotePath, targetDir string) error {
	items, err := g.listContents(ctx, owner, repo, remotePath)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Type == "dir" {
			subTarget := filepath.Join(targetDir, item.Name)
			if err := g.downloadDirTo(ctx, owner, repo, item.Path, subTarget); err != nil {
				return err
			}
		} else if item.DownloadURL != "" {
			rel := strings.TrimPrefix(item.Path, remotePath+"/")
			target := filepath.Join(targetDir, rel)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			req, _ := http.NewRequestWithContext(ctx, "GET", item.DownloadURL, nil)
			resp, err := g.client.Do(req)
			if err != nil {
				return err
			}
			f, err := os.Create(target)
			if err != nil {
				resp.Body.Close()
				return err
			}
			_, err = io.Copy(f, resp.Body)
			resp.Body.Close()
			f.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetLatestSHA fetches the latest commit SHA for a skill's subdirectory path.
func (g *GitHubInstaller) GetLatestSHA(ctx context.Context, repoURL, subPath string) (string, error) {
	owner, repo, err := parseGitHubURI(repoURL)
	if err != nil {
		return "", err
	}
	apiURL := fmt.Sprintf("%s/repos/%s/%s/commits?path=%s&per_page=1", g.baseURL, owner, repo, subPath)
	req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var commits []struct{ SHA string `json:"sha"` }
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
		return "", err
	}
	return commits[0].SHA, nil
}

// parseGitHubURI extracts owner and repo from a GitHub URL.
// Handles trailing slashes and optional .git suffix.
func parseGitHubURI(uri string) (owner, repo string, err error) {
	uri = strings.TrimSuffix(uri, "/")
	uri = strings.TrimSuffix(uri, ".git")
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URI: %s", uri)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}

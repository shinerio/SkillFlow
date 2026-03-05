package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/shinerio/skillflow/core/skill"
)

type CheckResult struct {
	SkillID   string
	HasUpdate bool
	LatestSHA string
}

type Checker struct {
	baseURL string
	client  *http.Client
}

// NewChecker creates a Checker. Pass nil for client to use http.DefaultClient.
func NewChecker(baseURL string, client *http.Client) *Checker {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Checker{baseURL: baseURL, client: client}
}

func (c *Checker) Check(ctx context.Context, sk *skill.Skill) (CheckResult, error) {
	if !sk.IsGitHub() {
		return CheckResult{}, nil
	}
	owner, repo, subPath := parseSourceURL(sk.SourceURL, sk.SourceSubPath)
	url := fmt.Sprintf("%s/repos/%s/%s/commits?path=%s&per_page=1", c.baseURL, owner, repo, subPath)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return CheckResult{}, err
	}
	defer resp.Body.Close()

	var commits []struct{ SHA string `json:"sha"` }
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil || len(commits) == 0 {
		return CheckResult{}, err
	}
	latestSHA := commits[0].SHA
	return CheckResult{
		SkillID:   sk.ID,
		LatestSHA: latestSHA,
		HasUpdate: latestSHA != sk.SourceSHA,
	}, nil
}

func parseSourceURL(sourceURL, subPath string) (owner, repo, path string) {
	sourceURL = strings.TrimSuffix(sourceURL, "/")
	parts := strings.Split(sourceURL, "/")
	if len(parts) < 2 {
		return "", "", subPath
	}
	owner = parts[len(parts)-2]
	repo = parts[len(parts)-1]
	return owner, repo, subPath
}

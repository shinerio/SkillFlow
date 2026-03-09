package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/notify"
)

const githubStarWatchInterval = 5 * time.Minute
const githubStarScanCooldown = 24 * time.Hour

type githubStarRepoItem struct {
	ID            int64  `json:"id"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
	Name string `json:"name"`
}

type githubTreeResp struct {
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"tree"`
}

func (a *App) startGitHubStarWatcher() {
	ticker := time.NewTicker(githubStarWatchInterval)
	go func() {
		defer ticker.Stop()
		a.runGitHubStarWatcherCycle()
		for {
			select {
			case <-ticker.C:
				a.runGitHubStarWatcherCycle()
			case <-a.ctx.Done():
				return
			}
		}
	}()
}

func (a *App) runGitHubStarWatcherCycle() {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("github star watcher failed: load config failed: %v", err)
		return
	}
	pat := strings.TrimSpace(cfg.GitHub.PAT)
	if pat == "" {
		a.logDebugf("github star watcher skipped: missing pat")
		return
	}
	state, err := a.githubStarStorage.Load()
	if err != nil {
		a.logErrorf("github star watcher failed: load state failed: %v", err)
		return
	}
	etag, repos, status, err := a.fetchGitHubStarredRepos(pat, state.Metadata.LastETag)
	if err != nil {
		a.logErrorf("github star watcher failed: fetch starred failed: %v", err)
		return
	}
	if status == http.StatusNotModified {
		a.logDebugf("github star watcher completed: not modified")
		return
	}
	for _, repo := range repos {
		repoID := repo.ID
		if repoID == 0 {
			continue
		}
		if _, ok := state.Repos[strconv.FormatInt(repoID, 10)]; ok {
			continue
		}
		now := time.Now()
		shouldScan, scanErr := a.githubStarStorage.ShouldScan(repoID, now, githubStarScanCooldown)
		if scanErr != nil {
			a.logErrorf("github star watcher failed: repo=%s shouldScan err=%v", repo.FullName, scanErr)
			continue
		}
		if !shouldScan {
			continue
		}
		hasSkill, scanErr := a.scanGitHubRepoForSkill(pat, repo)
		if scanErr != nil {
			a.logErrorf("github star watcher failed: scan repo=%s err=%v", repo.FullName, scanErr)
			_ = a.githubStarStorage.Upsert(repoID, repo.FullName, false, now)
			continue
		}
		_ = a.githubStarStorage.Upsert(repoID, repo.FullName, hasSkill, now)
		if !hasSkill {
			continue
		}
		if _, addErr := a.addStarredRepoFromGitHub(repo); addErr != nil {
			a.logErrorf("github star watcher failed: add repo=%s err=%v", repo.FullName, addErr)
			continue
		}
		a.hub.Publish(notify.Event{Type: notify.EventGitHubStarCaptured, Payload: repo.FullName})
	}
	if etag != "" {
		if err := a.githubStarStorage.UpdateETag(etag); err != nil {
			a.logErrorf("github star watcher failed: save etag failed: %v", err)
		}
	}
	a.logInfof("github star watcher completed: candidates=%d", len(repos))
}

func (a *App) fetchGitHubStarredRepos(pat, etag string) (string, []githubStarRepoItem, int, error) {
	apiURL := "https://api.github.com/user/starred?per_page=30"
	req, _ := http.NewRequestWithContext(a.ctx, http.MethodGet, apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+pat)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	resp, err := a.proxyHTTPClient().Do(req)
	if err != nil {
		return "", nil, 0, err
	}
	defer resp.Body.Close()
	a.handleGitHubRateLimit(resp)
	if resp.StatusCode == http.StatusNotModified {
		return "", nil, resp.StatusCode, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", nil, resp.StatusCode, fmt.Errorf("github status %d", resp.StatusCode)
	}
	var repos []githubStarRepoItem
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return "", nil, resp.StatusCode, err
	}
	return strings.TrimSpace(resp.Header.Get("ETag")), repos, resp.StatusCode, nil
}

func (a *App) scanGitHubRepoForSkill(pat string, repo githubStarRepoItem) (bool, error) {
	branch := strings.TrimSpace(repo.DefaultBranch)
	if branch == "" {
		branch = "HEAD"
	}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", repo.Owner.Login, repo.Name, branch)
	req, _ := http.NewRequestWithContext(a.ctx, http.MethodGet, apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+pat)
	resp, err := a.proxyHTTPClient().Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	a.handleGitHubRateLimit(resp)
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("github tree status %d", resp.StatusCode)
	}
	var tree githubTreeResp
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return false, err
	}
	for _, node := range tree.Tree {
		if node.Type != "blob" {
			continue
		}
		if node.Path == "skill.md" || node.Path == "SKILL.md" || strings.HasSuffix(node.Path, "/skill.md") || strings.HasSuffix(node.Path, "/SKILL.md") {
			return true, nil
		}
	}
	return false, nil
}

func (a *App) handleGitHubRateLimit(resp *http.Response) {
	if resp == nil {
		return
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		if retryAfter := strings.TrimSpace(resp.Header.Get("Retry-After")); retryAfter != "" {
			if sec, err := strconv.Atoi(retryAfter); err == nil && sec > 0 {
				time.Sleep(time.Duration(sec) * time.Second)
			}
		}
	}
	remaining, _ := strconv.Atoi(strings.TrimSpace(resp.Header.Get("X-RateLimit-Remaining")))
	resetUnix, _ := strconv.ParseInt(strings.TrimSpace(resp.Header.Get("X-RateLimit-Reset")), 10, 64)
	if remaining <= 1 && resetUnix > 0 {
		wait := time.Until(time.Unix(resetUnix, 0))
		if wait > 0 {
			time.Sleep(wait)
		}
	}
}

func (a *App) addStarredRepoFromGitHub(repo githubStarRepoItem) (*coregit.StarredRepo, error) {
	repoURL := fmt.Sprintf("https://github.com/%s.git", repo.FullName)
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	for i := range repos {
		if coregit.SameRepo(repos[i].URL, repoURL) {
			repos[i].Starred = true
			if !repos[i].Manual && repos[i].Starred {
				// keep existing manual flag as-is; do not force true for auto-only
			}
			if saveErr := a.starStorage.Save(repos); saveErr != nil {
				return nil, saveErr
			}
			return &repos[i], nil
		}
	}
	name, err := coregit.ParseRepoName(repoURL)
	if err != nil {
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		return nil, err
	}
	localDir, err := coregit.CacheDir(filepath.Dir(a.cacheDir), repoURL)
	if err != nil {
		return nil, err
	}
	newRepo := coregit.StarredRepo{URL: repoURL, Name: name, Source: source, LocalDir: localDir, Starred: true}
	if syncErr := coregit.CloneOrUpdate(a.ctx, repoURL, localDir, a.gitProxyURL()); syncErr != nil {
		newRepo.SyncError = syncErr.Error()
	} else {
		newRepo.LastSync = time.Now()
	}
	repos = append(repos, newRepo)
	if err := a.starStorage.Save(repos); err != nil {
		return nil, err
	}
	return &repos[len(repos)-1], nil
}

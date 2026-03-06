package git

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type RepoRef struct {
	Host string // e.g. github.com, gitee.com, git.example.com:2222
	Path string // e.g. owner/repo or group/subgroup/repo
}

// CheckGitInstalled returns nil if git is in PATH, or a user-friendly error.
func CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git 未安装，请先安装 git（https://git-scm.com）再使用此功能")
	}
	return nil
}

// ParseRepoRef extracts remote host + repository path from a git remote URL.
func ParseRepoRef(repoURL string) (RepoRef, error) {
	u := strings.TrimSpace(repoURL)
	if u == "" {
		return RepoRef{}, fmt.Errorf("无效的远程仓库地址: %s", repoURL)
	}

	host, rawPath, ok := splitRemoteHostPath(u)
	if !ok {
		return RepoRef{}, fmt.Errorf("无效的远程仓库地址: %s", repoURL)
	}

	repoPath, ok := normalizeRepoPath(rawPath)
	if !ok {
		return RepoRef{}, fmt.Errorf("无效的远程仓库地址: %s", repoURL)
	}

	return RepoRef{
		Host: strings.ToLower(host),
		Path: repoPath,
	}, nil
}

// ParseRepoName extracts repository path from a remote URL, e.g. "owner/repo".
func ParseRepoName(repoURL string) (string, error) {
	ref, err := ParseRepoRef(repoURL)
	if err != nil {
		return "", err
	}
	return ref.Path, nil
}

func RepoSource(repoURL string) (string, error) {
	ref, err := ParseRepoRef(repoURL)
	if err != nil {
		return "", err
	}
	return strings.ToLower(ref.Host + "/" + ref.Path), nil
}

func CanonicalRepoURL(repoURL string) (string, error) {
	ref, err := ParseRepoRef(repoURL)
	if err != nil {
		return "", err
	}
	// Strip port — SSH ports (e.g. 22) are not valid for HTTPS; use default 443.
	host := ref.Host
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	return "https://" + host + "/" + ref.Path, nil
}

func SameRepo(repoA, repoB string) bool {
	sourceA, errA := RepoSource(repoA)
	sourceB, errB := RepoSource(repoB)
	if errA != nil || errB != nil {
		return strings.EqualFold(strings.TrimSpace(repoA), strings.TrimSpace(repoB))
	}
	return strings.EqualFold(sourceA, sourceB)
}

func splitRemoteHostPath(remote string) (host, path string, ok bool) {
	// URL form: https://host/owner/repo.git, ssh://git@host/owner/repo.git
	if strings.Contains(remote, "://") {
		parsed, err := url.Parse(remote)
		if err != nil {
			return "", "", false
		}
		if parsed.Host == "" {
			return "", "", false
		}
		host = parsed.Hostname()
		if parsed.Port() != "" {
			host = host + ":" + parsed.Port()
		}
		return host, parsed.Path, host != "" && parsed.Path != ""
	}

	// SCP-like SSH form: git@host:owner/repo.git
	if strings.Contains(remote, "@") && strings.Contains(remote, ":") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) != 2 {
			return "", "", false
		}
		hostPart := parts[0]
		if at := strings.LastIndex(hostPart, "@"); at >= 0 {
			hostPart = hostPart[at+1:]
		}
		return hostPart, parts[1], hostPart != "" && parts[1] != ""
	}

	// Host/path form without scheme: gitee.com/owner/repo
	parts := strings.SplitN(strings.Trim(remote, "/"), "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	if !strings.Contains(parts[0], ".") {
		return "", "", false
	}
	return parts[0], parts[1], parts[0] != "" && parts[1] != ""
}

func normalizeRepoPath(path string) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", false
	}

	// GitHub/Gitea-like API path: /repos/{owner}/{repo}
	if strings.EqualFold(parts[0], "repos") {
		if len(parts) < 3 {
			return "", false
		}
		parts = parts[1:3]
	}

	// Web URL variants like /owner/repo/tree/main or /owner/repo/blob/main/...
	if len(parts) >= 4 {
		switch strings.ToLower(parts[2]) {
		case "tree", "blob", "raw", "commit":
			parts = parts[:2]
		}
	}

	parts[len(parts)-1] = strings.TrimSuffix(parts[len(parts)-1], ".git")
	if parts[len(parts)-1] == "" {
		return "", false
	}
	return strings.Join(parts, "/"), true
}

// CacheDir returns the local clone directory for a repo URL under dataDir/cache/.
func CacheDir(dataDir, repoURL string) (string, error) {
	source, err := RepoSource(repoURL)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(source, "/", 2)
	hostPart := strings.ReplaceAll(parts[0], ":", "_")
	repoPath := ""
	if len(parts) == 2 {
		repoPath = parts[1]
	}
	return filepath.Join(dataDir, "cache", hostPart, filepath.FromSlash(repoPath)), nil
}

// CloneOrUpdate clones repoURL into dir, or force-updates it if already present.
// proxyURL is optional (empty = inherit environment).
func CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error {
	if err := CheckGitInstalled(); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		// already cloned — force-pull (handles force-push on remote)
		if err := runGit(ctx, dir, proxyURL, "fetch", "origin"); err != nil {
			return fmt.Errorf("git fetch: %w", err)
		}
		return runGit(ctx, dir, proxyURL, "reset", "--hard", "FETCH_HEAD")
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return err
	}
	return runGit(ctx, "", proxyURL, "clone", repoURL, dir)
}

// GetSubPathSHA returns the latest commit SHA for a path within a local git repo.
func GetSubPathSHA(ctx context.Context, repoDir, subPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-n", "1", "--format=%H", "--", subPath)
	cmd.Dir = repoDir
	hideConsole(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// IsAuthError reports whether err looks like an HTTP authentication failure from git.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "authentication failed") ||
		strings.Contains(msg, "invalid username or password") ||
		strings.Contains(msg, "could not read username") ||
		strings.Contains(msg, "terminal prompts disabled") ||
		strings.Contains(msg, "repository not found") ||
		strings.Contains(msg, "http basic: access denied") ||
		strings.Contains(msg, "the requested url returned error: 403") ||
		strings.Contains(msg, "the requested url returned error: 401")
}

// IsSSHAuthError reports whether err looks like an SSH key authentication failure.
func IsSSHAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "permission denied (publickey") ||
		strings.Contains(msg, "no supported authentication methods") ||
		(strings.Contains(msg, "could not read from remote repository") &&
			!strings.Contains(msg, "http"))
}

// CloneOrUpdateWithCreds clones or updates a repo using embedded username/password for HTTP(S) URLs.
// It always removes any partial clone directory first to ensure a clean state.
func CloneOrUpdateWithCreds(ctx context.Context, repoURL, dir, proxyURL, username, password string) error {
	cloneURL := repoURL
	if (username != "" || password != "") && strings.Contains(repoURL, "://") {
		if parsed, err := url.Parse(repoURL); err == nil &&
			(parsed.Scheme == "https" || parsed.Scheme == "http") {
			parsed.User = url.UserPassword(username, password)
			cloneURL = parsed.String()
		}
	}
	// Remove any partial clone so CloneOrUpdate always does a fresh clone.
	_ = os.RemoveAll(dir)
	return CloneOrUpdate(ctx, cloneURL, dir, proxyURL)
}

func runGit(ctx context.Context, dir, proxyURL string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	hideConsole(cmd)
	env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if proxyURL != "" {
		env = append(env,
			"HTTP_PROXY="+proxyURL,
			"HTTPS_PROXY="+proxyURL,
			"http_proxy="+proxyURL,
			"https_proxy="+proxyURL,
		)
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

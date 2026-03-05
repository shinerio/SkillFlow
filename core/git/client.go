package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckGitInstalled returns nil if git is in PATH, or a user-friendly error.
func CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git 未安装，请先安装 git（https://git-scm.com）再使用此功能")
	}
	return nil
}

// ParseRepoName extracts "owner/repo" from a GitHub URL.
func ParseRepoName(repoURL string) (string, error) {
	u := strings.TrimSuffix(strings.TrimSuffix(repoURL, "/"), ".git")
	parts := strings.Split(u, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("无效的 GitHub URL: %s", repoURL)
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1], nil
}

// CacheDir returns the local clone directory for a repo URL under dataDir/cache/.
func CacheDir(dataDir, repoURL string) (string, error) {
	name, err := ParseRepoName(repoURL)
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "cache", filepath.FromSlash(name)), nil
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
		return runGit(ctx, dir, proxyURL, "reset", "--hard", "origin/HEAD")
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
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func runGit(ctx context.Context, dir, proxyURL string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	if proxyURL != "" {
		cmd.Env = append(os.Environ(),
			"HTTP_PROXY="+proxyURL,
			"HTTPS_PROXY="+proxyURL,
			"http_proxy="+proxyURL,
			"https_proxy="+proxyURL,
		)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

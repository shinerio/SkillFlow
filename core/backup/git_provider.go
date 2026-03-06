package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const GitProviderName = "git"

// GitConflictError is returned by Restore (git pull) when merge conflicts are detected,
// and by Sync (git push) when the remote has diverged.
type GitConflictError struct {
	Output string
}

func (e *GitConflictError) Error() string {
	return fmt.Sprintf("git conflict: %s", e.Output)
}

// GitProvider implements CloudProvider using a remote Git repository as the backup target.
// The skills storage directory itself becomes the git working tree.
// bucket and remotePath parameters are ignored; the repo URL comes from credentials.
type GitProvider struct {
	repoURL  string
	branch   string
	username string
	token    string
	localDir string // cached from the last Sync/Restore call, used by List
}

func NewGitProvider() *GitProvider { return &GitProvider{} }

func (p *GitProvider) Name() string { return GitProviderName }

func (p *GitProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{
			Key:         "repo_url",
			Label:       "Git 仓库地址",
			Placeholder: "https://github.com/user/my-backup.git",
		},
		{
			Key:         "branch",
			Label:       "分支（留空默认 main）",
			Placeholder: "main",
		},
		{
			Key:         "username",
			Label:       "用户名（HTTPS 认证，可选）",
			Placeholder: "your-username",
		},
		{
			Key:         "token",
			Label:       "访问令牌（HTTPS 认证，可选）",
			Placeholder: "ghp_xxxx",
			Secret:      true,
		},
	}
}

func (p *GitProvider) Init(credentials map[string]string) error {
	p.repoURL = strings.TrimSpace(credentials["repo_url"])
	if p.repoURL == "" {
		return fmt.Errorf("git 仓库地址不能为空")
	}
	p.branch = strings.TrimSpace(credentials["branch"])
	if p.branch == "" {
		p.branch = "main"
	}
	p.username = strings.TrimSpace(credentials["username"])
	p.token = strings.TrimSpace(credentials["token"])
	return nil
}

// authenticatedURL injects credentials into an HTTPS URL when username+token are provided.
func (p *GitProvider) authenticatedURL() string {
	if p.username != "" && p.token != "" && strings.HasPrefix(p.repoURL, "https://") {
		return "https://" + p.username + ":" + p.token + "@" + p.repoURL[8:]
	}
	return p.repoURL
}

func (p *GitProvider) run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ensureRepo initialises a git repo in localDir (if not yet a repo) and sets the remote origin.
func (p *GitProvider) ensureRepo(localDir string) error {
	gitDir := filepath.Join(localDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if _, err := p.run(localDir, "init"); err != nil {
			return fmt.Errorf("git init 失败")
		}
		if _, err := p.run(localDir, "remote", "add", "origin", p.authenticatedURL()); err != nil {
			return fmt.Errorf("git remote add 失败")
		}
	} else {
		// Update remote URL (credentials may have changed)
		p.run(localDir, "remote", "set-url", "origin", p.authenticatedURL()) //nolint
	}
	// Ensure user identity is set locally so commits don't fail
	p.run(localDir, "config", "user.email", "skillflow@local") //nolint
	p.run(localDir, "config", "user.name", "SkillFlow")        //nolint
	return nil
}

// Sync commits all local changes and pushes to the remote repository.
func (p *GitProvider) Sync(_ context.Context, localDir, _, _ string, onProgress func(file string)) error {
	p.localDir = localDir
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}

	onProgress("git add")
	if out, err := p.run(localDir, "add", "-A"); err != nil {
		return fmt.Errorf("git add 失败: %s", out)
	}

	// Nothing to commit?
	statusOut, _ := p.run(localDir, "status", "--porcelain")
	if strings.TrimSpace(statusOut) == "" {
		onProgress("up-to-date")
	} else {
		onProgress("git commit")
		if out, err := p.run(localDir, "commit", "-m", "SkillFlow auto-backup"); err != nil {
			return fmt.Errorf("git commit 失败: %s", out)
		}
	}

	onProgress("git push")
	out, err := p.run(localDir, "push", "origin", p.branch)
	if err != nil {
		// First push to a new empty remote: set upstream
		if out2, err2 := p.run(localDir, "push", "--set-upstream", "origin", p.branch); err2 != nil {
			// Remote has diverged → conflict
			if strings.Contains(out, "rejected") || strings.Contains(out, "non-fast-forward") ||
				strings.Contains(out2, "rejected") || strings.Contains(out2, "non-fast-forward") {
				return &GitConflictError{Output: out + out2}
			}
			return fmt.Errorf("git push 失败: %s %s", out, out2)
		}
	}
	return nil
}

// Restore runs git pull to bring the local directory up to date with the remote.
// Returns *GitConflictError when merge conflicts are detected.
func (p *GitProvider) Restore(_ context.Context, _, _, localDir string) error {
	p.localDir = localDir
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}
	out, err := p.run(localDir, "pull", "origin", p.branch, "--allow-unrelated-histories")
	if err != nil {
		if strings.Contains(out, "CONFLICT") || strings.Contains(strings.ToLower(out), "conflict") ||
			strings.Contains(out, "Automatic merge failed") {
			return &GitConflictError{Output: out}
		}
		return fmt.Errorf("git pull 失败: %s", out)
	}
	return nil
}

// List returns the files tracked by git in the local working tree.
func (p *GitProvider) List(_ context.Context, _, _ string) ([]RemoteFile, error) {
	if p.localDir == "" {
		return nil, nil
	}
	out, err := p.run(p.localDir, "ls-files")
	if err != nil {
		return nil, nil
	}
	var files []RemoteFile
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		full := filepath.Join(p.localDir, line)
		info, statErr := os.Stat(full)
		var size int64
		if statErr == nil {
			size = info.Size()
		}
		files = append(files, RemoteFile{Path: line, Size: size})
	}
	return files, nil
}

// ResolveConflictUseLocal aborts the in-progress merge and force-pushes local state to remote.
func (p *GitProvider) ResolveConflictUseLocal(localDir string) error {
	p.run(localDir, "merge", "--abort") //nolint — may not be in merge state
	// Ensure all local changes are committed before force-push
	p.run(localDir, "add", "-A")                                              //nolint
	p.run(localDir, "commit", "-m", "SkillFlow: resolve conflict (use local)") //nolint — may have nothing to commit
	out, err := p.run(localDir, "push", "origin", p.branch, "--force-with-lease")
	if err != nil {
		// Fallback to force push when --force-with-lease fails
		if out2, err2 := p.run(localDir, "push", "origin", p.branch, "--force"); err2 != nil {
			return fmt.Errorf("git push --force 失败: %s %s", out, out2)
		}
	}
	return nil
}

// ResolveConflictUseRemote aborts the in-progress merge and resets local state to the remote branch.
func (p *GitProvider) ResolveConflictUseRemote(localDir string) error {
	p.run(localDir, "merge", "--abort") //nolint
	if out, err := p.run(localDir, "fetch", "origin"); err != nil {
		return fmt.Errorf("git fetch 失败: %s", out)
	}
	if out, err := p.run(localDir, "reset", "--hard", "origin/"+p.branch); err != nil {
		return fmt.Errorf("git reset 失败: %s", out)
	}
	return nil
}

// GetBranch returns the configured branch.
func (p *GitProvider) GetBranch() string { return p.branch }

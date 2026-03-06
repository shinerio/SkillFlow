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
	Files  []string
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
	hideConsole(cmd)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ensureIgnoredPath(localDir, ignoredPath string) error {
	gitignorePath := filepath.Join(localDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		if strings.TrimSpace(line) == ignoredPath {
			return nil
		}
	}

	if len(normalized) > 0 && !strings.HasSuffix(normalized, "\n") {
		normalized += "\n"
	}
	normalized += ignoredPath + "\n"
	return os.WriteFile(gitignorePath, []byte(normalized), 0644)
}

func (p *GitProvider) isGitRepo(localDir string) bool {
	out, err := p.run(localDir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "true"
}

func isMissingRemoteRef(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "couldn't find remote ref") ||
		strings.Contains(lower, "remote ref does not exist") ||
		strings.Contains(lower, "no such ref was fetched")
}

func isGitConflictOutput(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "conflict") ||
		strings.Contains(lower, "non-fast-forward") ||
		strings.Contains(lower, "rejected") ||
		strings.Contains(lower, "divergent") ||
		strings.Contains(lower, "not possible to fast-forward") ||
		strings.Contains(lower, "need to specify how to reconcile divergent branches")
}

func parseConflictFilesFromOutput(out string) []string {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "CONFLICT") {
			continue
		}
		if idx := strings.LastIndex(strings.ToLower(line), " in "); idx >= 0 && idx+4 < len(line) {
			path := strings.TrimSpace(line[idx+4:])
			if path != "" {
				files = append(files, path)
			}
		}
	}
	return uniqueStrings(files)
}

func parseNameOnlyOutput(out string) []string {
	var files []string
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, line)
	}
	return files
}

func parsePorcelainConflictFiles(out string) []string {
	var files []string
	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		if status == "UU" || status == "AA" || status == "DD" ||
			status == "AU" || status == "UA" || status == "DU" || status == "UD" {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	return files
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func (p *GitProvider) collectConflictFiles(localDir, branch, output string) []string {
	files := parseConflictFilesFromOutput(output)

	if out, err := p.run(localDir, "diff", "--name-only", "--diff-filter=U"); err == nil {
		files = append(files, parseNameOnlyOutput(out)...)
	}

	// Best-effort fetch for divergence cases like non-fast-forward push.
	p.run(localDir, "fetch", "origin", branch) //nolint

	if _, err := p.run(localDir, "rev-parse", "--verify", "HEAD"); err == nil {
		if out, err := p.run(localDir, "diff", "--name-only", "HEAD..origin/"+branch); err == nil {
			files = append(files, parseNameOnlyOutput(out)...)
		}
		if out, err := p.run(localDir, "diff", "--name-only", "origin/"+branch+"..HEAD"); err == nil {
			files = append(files, parseNameOnlyOutput(out)...)
		}
	}

	if out, err := p.run(localDir, "status", "--porcelain"); err == nil {
		files = append(files, parsePorcelainConflictFiles(out)...)
	}

	return uniqueStrings(files)
}

// ensureRepo initializes a git repo in localDir (if needed) and ensures remote origin is configured.
func (p *GitProvider) ensureRepo(localDir string) error {
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	gitDir := filepath.Join(localDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) || !p.isGitRepo(localDir) {
		if out, err := p.run(localDir, "init"); err != nil {
			return fmt.Errorf("git init 失败: %s", out)
		}
	}

	authURL := p.authenticatedURL()
	remoteOut, remoteErr := p.run(localDir, "remote", "get-url", "origin")
	if remoteErr != nil {
		if out, err := p.run(localDir, "remote", "add", "origin", authURL); err != nil {
			return fmt.Errorf("git remote add 失败: %s", out)
		}
	} else {
		current := strings.TrimSpace(remoteOut)
		if current != authURL {
			if out, err := p.run(localDir, "remote", "set-url", "origin", authURL); err != nil {
				return fmt.Errorf("git remote set-url 失败: %s", out)
			}
		}
	}

	// Ensure user identity is set locally so commits don't fail
	if out, err := p.run(localDir, "config", "user.email", "skillflow@local"); err != nil {
		return fmt.Errorf("git config user.email 失败: %s", out)
	}
	if out, err := p.run(localDir, "config", "user.name", "SkillFlow"); err != nil {
		return fmt.Errorf("git config user.name 失败: %s", out)
	}
	for _, dir := range excludedDirs {
		if err := ensureIgnoredPath(localDir, dir+"/"); err != nil {
			return fmt.Errorf("写入 .gitignore 失败: %w", err)
		}
	}
	for _, file := range excludedFiles {
		if err := ensureIgnoredPath(localDir, file); err != nil {
			return fmt.Errorf("写入 .gitignore 失败: %w", err)
		}
	}
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
	// Ensure excluded paths are never tracked in git backup.
	for _, dir := range excludedDirs {
		if out, err := p.run(localDir, "rm", "-r", "--cached", "--ignore-unmatch", dir); err != nil {
			return fmt.Errorf("git rm --cached %s 失败: %s", dir, out)
		}
	}

	// Nothing to commit?
	statusOut, _ := p.run(localDir, "status", "--porcelain")
	_, headErr := p.run(localDir, "rev-parse", "--verify", "HEAD")
	// Fresh repo + no files: treat as no-op.
	if strings.TrimSpace(statusOut) == "" && headErr != nil {
		onProgress("up-to-date")
		return nil
	}
	if strings.TrimSpace(statusOut) == "" {
		onProgress("up-to-date")
	} else {
		onProgress("git commit")
		if out, err := p.run(localDir, "commit", "-m", "SkillFlow auto-backup"); err != nil {
			return fmt.Errorf("git commit 失败: %s", out)
		}
	}

	onProgress("git push")
	out, err := p.run(localDir, "push", "origin", "HEAD:"+p.branch)
	if err != nil {
		// First push to a new remote: set upstream
		if out2, err2 := p.run(localDir, "push", "--set-upstream", "origin", "HEAD:"+p.branch); err2 != nil {
			// Remote has diverged → conflict
			if isGitConflictOutput(out) || isGitConflictOutput(out2) {
				return &GitConflictError{
					Output: out + out2,
					Files:  p.collectConflictFiles(localDir, p.branch, out+out2),
				}
			}
			return fmt.Errorf("git push 失败: %s %s", out, out2)
		}
	}
	return nil
}

// autoCommitLocal stages and commits any local changes before a pull so that
// untracked or modified files do not block the merge.
func (p *GitProvider) autoCommitLocal(localDir string) {
	p.run(localDir, "add", "-A") //nolint
	// Remove excluded paths from index if accidentally staged.
	for _, dir := range excludedDirs {
		p.run(localDir, "rm", "-r", "--cached", "--ignore-unmatch", dir) //nolint
	}
	statusOut, _ := p.run(localDir, "status", "--porcelain")
	if strings.TrimSpace(statusOut) != "" {
		p.run(localDir, "commit", "-m", "SkillFlow: pre-pull auto-commit") //nolint
	}
}

// Restore runs git pull to bring the local directory up to date with the remote.
// Returns *GitConflictError when merge conflicts are detected.
func (p *GitProvider) Restore(_ context.Context, _, _, localDir string) error {
	p.localDir = localDir
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}
	// Commit any local changes before pulling to prevent
	// "untracked working tree files would be overwritten by merge" errors.
	p.autoCommitLocal(localDir)
	out, err := p.run(localDir, "pull", "origin", p.branch, "--allow-unrelated-histories")
	if err != nil {
		if isMissingRemoteRef(out) {
			// Remote branch does not exist yet; nothing to restore.
			return nil
		}
		if isGitConflictOutput(out) || strings.Contains(out, "Automatic merge failed") {
			return &GitConflictError{
				Output: out,
				Files:  p.collectConflictFiles(localDir, p.branch, out),
			}
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
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}
	p.run(localDir, "merge", "--abort") //nolint – may not be in merge state
	// Ensure all local changes are committed before force-push
	p.run(localDir, "add", "-A")                                               //nolint
	p.run(localDir, "commit", "-m", "SkillFlow: resolve conflict (use local)") //nolint – may have nothing to commit
	out, err := p.run(localDir, "push", "origin", "HEAD:"+p.branch, "--force-with-lease")
	if err != nil {
		// Fallback to force push when --force-with-lease fails
		if out2, err2 := p.run(localDir, "push", "origin", "HEAD:"+p.branch, "--force"); err2 != nil {
			return fmt.Errorf("git push --force 失败: %s %s", out, out2)
		}
	}
	return nil
}

// ResolveConflictUseRemote aborts the in-progress merge and resets local state to the remote branch.
func (p *GitProvider) ResolveConflictUseRemote(localDir string) error {
	if err := p.ensureRepo(localDir); err != nil {
		return err
	}
	p.run(localDir, "merge", "--abort") //nolint
	if out, err := p.run(localDir, "fetch", "origin", p.branch); err != nil {
		return fmt.Errorf("git fetch 失败: %s", out)
	}
	if out, err := p.run(localDir, "reset", "--hard", "origin/"+p.branch); err != nil {
		return fmt.Errorf("git reset 失败: %s", out)
	}
	return nil
}

// GetBranch returns the configured branch.
func (p *GitProvider) GetBranch() string { return p.branch }

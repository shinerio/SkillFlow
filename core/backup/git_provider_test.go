package backup

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output: %s", args, err, string(out))
	}
	return string(out)
}

func TestGitProviderSyncInitializesRepoAndPushes(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("mkdir localDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "skill.md"), []byte("# test"), 0644); err != nil {
		t.Fatalf("write skill.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(localDir, "cache"), 0755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "cache", "tmp.bin"), []byte("tmp"), 0644); err != nil {
		t.Fatalf("write cache file: %v", err)
	}

	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}
	if err := p.Sync(context.Background(), localDir, "", "", func(string) {}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(localDir, ".git")); err != nil {
		t.Fatalf("expected .git directory: %v", err)
	}

	origin := strings.TrimSpace(runGit(t, localDir, "remote", "get-url", "origin"))
	if origin != remoteDir {
		t.Fatalf("unexpected origin: got %q want %q", origin, remoteDir)
	}
	gitignore, err := os.ReadFile(filepath.Join(localDir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), "cache/") {
		t.Fatalf("expected .gitignore to contain cache/, got: %q", string(gitignore))
	}

	_ = runGit(t, "", "--git-dir", remoteDir, "rev-parse", "--verify", "refs/heads/main")
	remoteFiles := runGit(t, "", "--git-dir", remoteDir, "ls-tree", "-r", "--name-only", "main")
	if strings.Contains(remoteFiles, "cache/") {
		t.Fatalf("cache should not be tracked, remote files: %s", remoteFiles)
	}
}

func TestGitProviderSyncAddsOriginForExistingRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatalf("mkdir localDir: %v", err)
	}
	runGit(t, localDir, "init")
	if err := os.WriteFile(filepath.Join(localDir, "skill.md"), []byte("# test"), 0644); err != nil {
		t.Fatalf("write skill.md: %v", err)
	}

	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}
	if err := p.Sync(context.Background(), localDir, "", "", func(string) {}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	origin := strings.TrimSpace(runGit(t, localDir, "remote", "get-url", "origin"))
	if origin != remoteDir {
		t.Fatalf("unexpected origin: got %q want %q", origin, remoteDir)
	}
}

func TestGitProviderRestoreAllowsMissingRemoteBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	base := t.TempDir()
	remoteDir := filepath.Join(base, "remote.git")
	runGit(t, "", "init", "--bare", remoteDir)

	localDir := filepath.Join(base, "skills")
	p := NewGitProvider()
	if err := p.Init(map[string]string{
		"repo_url": remoteDir,
		"branch":   "main",
	}); err != nil {
		t.Fatalf("init provider: %v", err)
	}

	if err := p.Restore(context.Background(), "", "", localDir); err != nil {
		t.Fatalf("restore should allow missing remote branch, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(localDir, ".git")); err != nil {
		t.Fatalf("expected .git directory: %v", err)
	}
	origin := strings.TrimSpace(runGit(t, localDir, "remote", "get-url", "origin"))
	if origin != remoteDir {
		t.Fatalf("unexpected origin: got %q want %q", origin, remoteDir)
	}
}

func TestParseConflictFilesFromOutput(t *testing.T) {
	out := `
Auto-merging skills/a/skill.md
CONFLICT (content): Merge conflict in skills/a/skill.md
Auto-merging meta/123.json
CONFLICT (content): Merge conflict in meta/123.json
`
	files := parseConflictFilesFromOutput(out)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %#v", len(files), files)
	}
	if files[0] != "skills/a/skill.md" || files[1] != "meta/123.json" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

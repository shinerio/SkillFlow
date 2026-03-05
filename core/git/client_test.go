package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// makeLocalRepo creates a bare-minimum local git repo in dir with one commit.
func makeLocalRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		if err := runGit(context.Background(), dir, "", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		if err := runGit(context.Background(), dir, "", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
}

func TestParseRepoName(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://github.com/owner/repo", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo/", "owner/repo"},
	}
	for _, c := range cases {
		got, err := ParseRepoName(c.url)
		if err != nil || got != c.want {
			t.Errorf("ParseRepoName(%q) = %q, %v; want %q", c.url, got, err, c.want)
		}
	}
}

func TestCloneOrUpdate(t *testing.T) {
	if err := CheckGitInstalled(); err != nil {
		t.Skip("git not installed")
	}
	src := t.TempDir()
	makeLocalRepo(t, src)

	dst := filepath.Join(t.TempDir(), "clone")

	// First call: clone
	if err := CloneOrUpdate(context.Background(), src, dst, ""); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Fatal("README.md missing after clone")
	}

	// Add a new file to source
	if err := os.WriteFile(filepath.Join(src, "NEW.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "add file"}} {
		if err := runGit(context.Background(), src, "", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}

	// Second call: update
	if err := CloneOrUpdate(context.Background(), src, dst, ""); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "NEW.md")); err != nil {
		t.Fatal("NEW.md missing after update")
	}
}

func TestCacheDir(t *testing.T) {
	dir, err := CacheDir("/data", "https://github.com/owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/data", "cache", "owner", "repo")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestParseRepoNameError(t *testing.T) {
	_, err := ParseRepoName("notavalidurl")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestCloneOrUpdateForceOverwrite(t *testing.T) {
	if err := CheckGitInstalled(); err != nil {
		t.Skip("git not installed")
	}
	src := t.TempDir()
	makeLocalRepo(t, src)
	dst := filepath.Join(t.TempDir(), "clone")

	// Clone
	if err := CloneOrUpdate(context.Background(), src, dst, ""); err != nil {
		t.Fatalf("clone: %v", err)
	}

	// Locally modify a file in dst
	if err := os.WriteFile(filepath.Join(dst, "README.md"), []byte("local modification"), 0644); err != nil {
		t.Fatal(err)
	}

	// Update should overwrite local change (force-push safety)
	if err := CloneOrUpdate(context.Background(), src, dst, ""); err != nil {
		t.Fatalf("update: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dst, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("expected README.md = %q after reset, got %q", "hello", string(data))
	}
}

func TestGetSubPathSHA(t *testing.T) {
	if err := CheckGitInstalled(); err != nil {
		t.Skip("git not installed")
	}
	src := t.TempDir()
	makeLocalRepo(t, src)

	sha, err := GetSubPathSHA(context.Background(), src, "README.md")
	if err != nil {
		t.Fatalf("GetSubPathSHA: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("expected 40-char SHA, got %q (len %d)", sha, len(sha))
	}
}

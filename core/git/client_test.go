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

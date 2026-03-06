package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyNestedGitDirMovesNestedRepo(t *testing.T) {
	base := t.TempDir()
	skillsDir := filepath.Join(base, "skills")
	backupRoot := base
	legacyGitDir := filepath.Join(skillsDir, ".git")

	if err := os.MkdirAll(filepath.Join(legacyGitDir, "objects"), 0755); err != nil {
		t.Fatalf("mkdir legacy git dir: %v", err)
	}

	target, moved, err := MigrateLegacyNestedGitDir(skillsDir, backupRoot)
	if err != nil {
		t.Fatalf("migrate legacy git dir: %v", err)
	}
	if !moved {
		t.Fatalf("expected nested git dir to be moved")
	}
	if filepath.Base(target) != ".git.skillflow-legacy-backup" {
		t.Fatalf("unexpected target path: %s", target)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected nested .git to be moved away, stat err: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected migrated backup dir to exist: %v", err)
	}
}

func TestMigrateLegacyNestedGitDirNoopWhenBackupRootEqualsSkillsDir(t *testing.T) {
	base := t.TempDir()
	skillsDir := filepath.Join(base, "skills")
	if err := os.MkdirAll(filepath.Join(skillsDir, ".git"), 0755); err != nil {
		t.Fatalf("mkdir nested git dir: %v", err)
	}

	target, moved, err := MigrateLegacyNestedGitDir(skillsDir, skillsDir)
	if err != nil {
		t.Fatalf("migrate legacy git dir: %v", err)
	}
	if moved || target != "" {
		t.Fatalf("expected noop, got target=%q moved=%v", target, moved)
	}
	if _, err := os.Stat(filepath.Join(skillsDir, ".git")); err != nil {
		t.Fatalf("expected .git to remain in place: %v", err)
	}
}

package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSkillsEmpty(t *testing.T) {
	dir := t.TempDir()
	skills, err := ScanSkills(dir, "https://github.com/a/b", "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(skills))
	}
}

func TestScanSkills(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	// create two valid skills and one invalid (no SKILLS.md)
	for _, name := range []string{"alpha", "beta"} {
		d := filepath.Join(skillsDir, name)
		os.MkdirAll(d, 0755)
		os.WriteFile(filepath.Join(d, "SKILLS.md"), []byte("# "+name), 0644)
	}
	os.MkdirAll(filepath.Join(skillsDir, "no-skills-md"), 0755)

	skills, err := ScanSkills(dir, "https://github.com/a/b", "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(skills), skills)
	}
	for _, sk := range skills {
		if sk.RepoURL != "https://github.com/a/b" {
			t.Errorf("RepoURL wrong: %s", sk.RepoURL)
		}
		if sk.SubPath != "skills/"+sk.Name {
			t.Errorf("SubPath wrong: %s", sk.SubPath)
		}
		if sk.Path == "" {
			t.Errorf("Path empty for skill %s", sk.Name)
		}
	}
}

// TestScanSkillsRootFallback covers repos where skill dirs live at the repo
// root (no skills/ subdirectory), e.g. github.com/anthropics/skills.
func TestScanSkillsRootFallback(t *testing.T) {
	dir := t.TempDir()
	// Skills placed directly under repo root.
	for _, name := range []string{"gamma", "delta"} {
		d := filepath.Join(dir, name)
		os.MkdirAll(d, 0755)
		os.WriteFile(filepath.Join(d, "SKILLS.md"), []byte("# "+name), 0644)
	}
	// A dir without SKILLS.md should be ignored.
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	skills, err := ScanSkills(dir, "https://github.com/a/skills", "a/skills")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(skills), skills)
	}
	for _, sk := range skills {
		if sk.SubPath != sk.Name {
			t.Errorf("SubPath should equal Name for root-level skill, got SubPath=%s Name=%s", sk.SubPath, sk.Name)
		}
	}
}

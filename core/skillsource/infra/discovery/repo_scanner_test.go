package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSkillsEmpty(t *testing.T) {
	dir := t.TempDir()
	skills, err := NewRepoScanner().ScanSkills(dir, "https://github.com/a/b", "a/b", "github.com/a/b")
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
	for _, name := range []string{"alpha", "beta"} {
		d := filepath.Join(skillsDir, name)
		os.MkdirAll(d, 0755)
		os.WriteFile(filepath.Join(d, "skill.md"), []byte("# "+name), 0644)
	}
	os.MkdirAll(filepath.Join(skillsDir, "no-skills-md"), 0755)

	skills, err := NewRepoScanner().ScanSkills(dir, "https://github.com/a/b", "a/b", "github.com/a/b")
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
		if sk.Source != "github.com/a/b" {
			t.Errorf("Source wrong: %s", sk.Source)
		}
		if sk.Path == "" {
			t.Errorf("Path empty for skill %s", sk.Name)
		}
	}
}

func TestScanSkillsRootFallback(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"gamma", "delta"} {
		d := filepath.Join(dir, name)
		os.MkdirAll(d, 0755)
		os.WriteFile(filepath.Join(d, "skill.md"), []byte("# "+name), 0644)
	}
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	skills, err := NewRepoScanner().ScanSkills(dir, "https://github.com/a/skills", "a/skills", "github.com/a/skills")
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

func TestScanSkillsRepoRootSkill(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# repo-root"), 0644); err != nil {
		t.Fatalf("write repo root skill: %v", err)
	}

	skills, err := NewRepoScanner().ScanSkills(dir, "https://github.com/a/root-skill", "a/root-skill", "github.com/a/root-skill")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1, got %d: %+v", len(skills), skills)
	}
	if skills[0].Name != filepath.Base(dir) {
		t.Fatalf("expected repo root name %q, got %+v", filepath.Base(dir), skills[0])
	}
	if skills[0].SubPath != "." {
		t.Fatalf("expected repo root subpath '.', got %+v", skills[0])
	}
}

func TestScanSkillsRepoRootSkillDoesNotBlockNestedSkills(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# repo-root"), 0644); err != nil {
		t.Fatalf("write repo root skill: %v", err)
	}

	nestedSkillDir := filepath.Join(dir, "skills", "nested")
	if err := os.MkdirAll(nestedSkillDir, 0755); err != nil {
		t.Fatalf("mkdir nested skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedSkillDir, "SKILL.md"), []byte("# nested"), 0644); err != nil {
		t.Fatalf("write nested skill: %v", err)
	}

	skills, err := NewRepoScanner().ScanSkills(dir, "https://github.com/a/root-plus-nested", "a/root-plus-nested", "github.com/a/root-plus-nested")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d: %+v", len(skills), skills)
	}

	got := map[string]string{}
	for _, sk := range skills {
		got[sk.Name] = sk.SubPath
	}
	if got[filepath.Base(dir)] != "." {
		t.Fatalf("missing repo root skill: %+v", got)
	}
	if got["nested"] != "skills/nested" {
		t.Fatalf("missing nested skill: %+v", got)
	}
}

func TestScanSkillsNestedPluginSkills(t *testing.T) {
	dir := t.TempDir()
	pluginSkillDir := filepath.Join(dir, "plugins", "shinerio-note-plugin", "skills", "embed-mindmap")
	if err := os.MkdirAll(pluginSkillDir, 0755); err != nil {
		t.Fatalf("mkdir plugin skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginSkillDir, "SKILL.md"), []byte("# embed-mindmap"), 0644); err != nil {
		t.Fatalf("write plugin skill: %v", err)
	}

	topSkillDir := filepath.Join(dir, "skills", "top-level")
	if err := os.MkdirAll(topSkillDir, 0755); err != nil {
		t.Fatalf("mkdir top-level skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(topSkillDir, "skill.md"), []byte("# top-level"), 0644); err != nil {
		t.Fatalf("write top-level skill: %v", err)
	}

	skills, err := NewRepoScanner().ScanSkills(dir, "https://github.com/shinerio/shinerio-marketplace", "shinerio/shinerio-marketplace", "github.com/shinerio/shinerio-marketplace")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(skills), skills)
	}

	got := map[string]string{}
	for _, sk := range skills {
		got[sk.Name] = sk.SubPath
	}
	if got["embed-mindmap"] != "plugins/shinerio-note-plugin/skills/embed-mindmap" {
		t.Fatalf("nested plugin skill missing or wrong subpath: %+v", got)
	}
	if got["top-level"] != "skills/top-level" {
		t.Fatalf("top-level skill missing or wrong subpath: %+v", got)
	}
}

func TestScanSkillsRespectsMaxDepth(t *testing.T) {
	dir := t.TempDir()
	deepSkillDir := filepath.Join(dir, "plugins", "shinerio-note-plugin", "skills", "embed-mindmap")
	if err := os.MkdirAll(deepSkillDir, 0755); err != nil {
		t.Fatalf("mkdir deep skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deepSkillDir, "skill.md"), []byte("# embed-mindmap"), 0644); err != nil {
		t.Fatalf("write deep skill: %v", err)
	}

	shallowSkillDir := filepath.Join(dir, "skills", "top-level")
	if err := os.MkdirAll(shallowSkillDir, 0755); err != nil {
		t.Fatalf("mkdir shallow skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(shallowSkillDir, "skill.md"), []byte("# top-level"), 0644); err != nil {
		t.Fatalf("write shallow skill: %v", err)
	}

	skills, err := NewRepoScanner().ScanSkillsWithMaxDepth(dir, "https://github.com/shinerio/shinerio-marketplace", "shinerio/shinerio-marketplace", "github.com/shinerio/shinerio-marketplace", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill with depth limit 2, got %d: %+v", len(skills), skills)
	}
	if skills[0].Name != "top-level" || skills[0].SubPath != "skills/top-level" {
		t.Fatalf("unexpected shallow result: %+v", skills[0])
	}
}

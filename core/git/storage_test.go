package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStarStorageLoadEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorage(path)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("expected nil, got %v", repos)
	}
}

func TestStarStorageSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	s := NewStarStorage(path)
	localDir, err := CacheDir(dir, "https://github.com/a/b")
	if err != nil {
		t.Fatal(err)
	}
	want := []StarredRepo{
		{URL: "https://github.com/a/b", Name: "a/b", LocalDir: localDir, Manual: true, LastSync: time.Time{}},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"localDir": "cache/github.com/a/b"`) {
		t.Fatalf("expected relative localDir in persisted file, got %s", string(raw))
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestStarStorageLoadCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewStarStorage(path)
	_, err := s.Load()
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestStarStorageLoadMigratesAbsoluteLocalDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	localDir, err := CacheDir(dir, "https://github.com/a/b")
	if err != nil {
		t.Fatal(err)
	}
	repos := []StarredRepo{{URL: "https://github.com/a/b", Name: "a/b", LocalDir: localDir}}
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s := NewStarStorage(path)
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].LocalDir != localDir {
		t.Fatalf("unexpected load result: %+v", got)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"localDir": "cache/github.com/a/b"`) {
		t.Fatalf("expected migrated relative localDir, got %s", string(raw))
	}
}

func TestStarStorageLoadSeedsBuiltinsOnlyOnFirstInit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	builtins := []string{
		"https://github.com/anthropics/skills.git",
		"https://github.com/ComposioHQ/awesome-claude-skills.git",
		"https://github.com/affaan-m/everything-claude-code.git",
	}
	s := NewStarStorageWithBuiltins(path, builtins)

	first, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(builtins) {
		t.Fatalf("expected %d builtin repos, got %d", len(builtins), len(first))
	}
	for _, wantURL := range builtins {
		found := false
		for _, gotRepo := range first {
			if SameRepo(gotRepo.URL, wantURL) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected seeded repo list to include %s, got %+v", wantURL, first)
		}
	}

	if err := s.Save([]StarredRepo{}); err != nil {
		t.Fatal(err)
	}

	second, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("expected empty repos after user deletion, got %d", len(second))
	}
}

func TestStarStorageLoadDoesNotSeedWhenBuiltinsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorageWithBuiltins(path, nil)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("expected nil repos without builtins, got %+v", repos)
	}
}

func TestStarStorageLoadMigratesLegacySourceFlags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	legacy := []StarredRepo{{URL: "https://github.com/a/b", Name: "a/b"}}
	data, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	s := NewStarStorage(path)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || !repos[0].Manual {
		t.Fatalf("expected manual=true migration, got %+v", repos)
	}
}

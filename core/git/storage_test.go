package git

import (
	"path/filepath"
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
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorage(path)
	want := []StarredRepo{
		{URL: "https://github.com/a/b", Name: "a/b", LocalDir: "/tmp/a/b", LastSync: time.Time{}},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].URL != want[0].URL || got[0].Name != want[0].Name {
		t.Fatalf("mismatch: %+v", got)
	}
}

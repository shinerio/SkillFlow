package git

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGitHubStarStorageUpsertAndShouldScan(t *testing.T) {
	s := NewGitHubStarStorage(filepath.Join(t.TempDir(), "github_star_repo.json"))
	now := time.Now()
	if err := s.Upsert(123, "owner/repo", true, now); err != nil {
		t.Fatal(err)
	}
	state, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	rec, ok := state.Repos["123"]
	if !ok || !rec.HasSkill || rec.FullName != "owner/repo" {
		t.Fatalf("unexpected record: %+v", state.Repos)
	}
	should, err := s.ShouldScan(123, now.Add(2*time.Hour), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if should {
		t.Fatal("expected cooldown to block scan")
	}
	should, err = s.ShouldScan(123, now.Add(25*time.Hour), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !should {
		t.Fatal("expected scan after cooldown")
	}
}

func TestGitHubStarStorageUpdateETag(t *testing.T) {
	s := NewGitHubStarStorage(filepath.Join(t.TempDir(), "github_star_repo.json"))
	if err := s.UpdateETag("W/\"abc\""); err != nil {
		t.Fatal(err)
	}
	state, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if state.Metadata.LastETag != "W/\"abc\"" {
		t.Fatalf("unexpected etag: %s", state.Metadata.LastETag)
	}
}

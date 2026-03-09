package git

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type GitHubStarRepoRecord struct {
	FullName      string `json:"full_name"`
	HasSkill      bool   `json:"has_skill"`
	LastScannedAt int64  `json:"last_scanned_at"`
}

type GitHubStarRepoState struct {
	Metadata struct {
		LastETag string `json:"last_etag"`
	} `json:"metadata"`
	Repos map[string]GitHubStarRepoRecord `json:"repos"`
}

type GitHubStarStorage struct {
	path string
	mu   sync.Mutex
}

func NewGitHubStarStorage(path string) *GitHubStarStorage {
	return &GitHubStarStorage{path: filepath.Clean(path)}
}

func (s *GitHubStarStorage) Load() (GitHubStarRepoState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *GitHubStarStorage) Save(state GitHubStarRepoState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(state)
}

func (s *GitHubStarStorage) Upsert(repoID int64, fullName string, hasSkill bool, scannedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return err
	}
	key := strconv.FormatInt(repoID, 10)
	record := state.Repos[key]
	record.FullName = fullName
	record.HasSkill = hasSkill
	record.LastScannedAt = scannedAt.Unix()
	state.Repos[key] = record
	return s.saveLocked(state)
}

func (s *GitHubStarStorage) UpdateETag(etag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return err
	}
	state.Metadata.LastETag = etag
	return s.saveLocked(state)
}

func (s *GitHubStarStorage) ShouldScan(repoID int64, now time.Time, cooldown time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return false, err
	}
	record, ok := state.Repos[strconv.FormatInt(repoID, 10)]
	if !ok || record.LastScannedAt <= 0 {
		return true, nil
	}
	last := time.Unix(record.LastScannedAt, 0)
	return now.Sub(last) >= cooldown, nil
}

func (s *GitHubStarStorage) loadLocked() (GitHubStarRepoState, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return GitHubStarRepoState{Repos: map[string]GitHubStarRepoRecord{}}, nil
	}
	if err != nil {
		return GitHubStarRepoState{}, err
	}
	var state GitHubStarRepoState
	if err := json.Unmarshal(data, &state); err != nil {
		return GitHubStarRepoState{}, err
	}
	if state.Repos == nil {
		state.Repos = map[string]GitHubStarRepoRecord{}
	}
	return state, nil
}

func (s *GitHubStarStorage) saveLocked(state GitHubStarRepoState) error {
	if state.Repos == nil {
		state.Repos = map[string]GitHubStarRepoRecord{}
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".github_star_repo_*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

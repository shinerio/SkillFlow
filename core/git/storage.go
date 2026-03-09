package git

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/shinerio/skillflow/core/pathutil"
)

type StarStorage struct {
	path            string
	dataDir         string
	builtinRepoURLs []string
	mu              sync.Mutex
}

func NewStarStorage(path string) *StarStorage {
	return NewStarStorageWithBuiltins(path, nil)
}

func NewStarStorageWithBuiltins(path string, builtinRepoURLs []string) *StarStorage {
	cleanPath := filepath.Clean(path)
	builtins := append([]string(nil), builtinRepoURLs...)
	return &StarStorage{path: cleanPath, dataDir: filepath.Dir(cleanPath), builtinRepoURLs: builtins}
}

func (s *StarStorage) Load() ([]StarredRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		if len(s.builtinRepoURLs) == 0 {
			return nil, nil
		}
		repos, buildErr := s.buildBuiltinReposLocked()
		if buildErr != nil {
			return nil, buildErr
		}
		if err := s.saveLocked(repos); err != nil {
			return nil, err
		}
		return repos, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []StarredRepo
	if err := json.Unmarshal(data, &repos); err != nil {
		return repos, err
	}
	changed := false
	for i := range repos {
		if s.resolveLocalDir(&repos[i]) {
			changed = true
		}
		if !repos[i].Manual && !repos[i].Starred {
			repos[i].Manual = true
			changed = true
		}
	}
	if changed {
		if err := s.saveLocked(repos); err != nil {
			return nil, err
		}
	}
	return repos, nil
}

func (s *StarStorage) Save(repos []StarredRepo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(repos)
}

func (s *StarStorage) saveLocked(repos []StarredRepo) error {
	if repos == nil {
		repos = []StarredRepo{}
	}
	snapshot := make([]StarredRepo, len(repos))
	for i := range repos {
		snapshot[i] = s.serializedRepo(repos[i])
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".star_repos_*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

func (s *StarStorage) serializedRepo(repo StarredRepo) StarredRepo {
	snapshot := repo
	snapshot.LocalDir = pathutil.StorePath(s.dataDir, repo.LocalDir, s.derivedLocalDir(repo.URL))
	return snapshot
}

func (s *StarStorage) resolveLocalDir(repo *StarredRepo) bool {
	resolved, needsMigration := pathutil.ResolveStoredPath(s.dataDir, repo.LocalDir, s.derivedLocalDir(repo.URL))
	repo.LocalDir = resolved
	return needsMigration
}

func (s *StarStorage) derivedLocalDir(repoURL string) string {
	dir, err := CacheDir(s.dataDir, repoURL)
	if err != nil {
		return ""
	}
	return dir
}

func (s *StarStorage) buildBuiltinReposLocked() ([]StarredRepo, error) {
	repos := make([]StarredRepo, 0, len(s.builtinRepoURLs))
	for _, repoURL := range s.builtinRepoURLs {
		name, err := ParseRepoName(repoURL)
		if err != nil {
			return nil, err
		}
		source, err := RepoSource(repoURL)
		if err != nil {
			return nil, err
		}
		repos = append(repos, StarredRepo{
			URL:      repoURL,
			Name:     name,
			Source:   source,
			LocalDir: s.derivedLocalDir(repoURL),
			Manual:   true,
		})
	}
	return repos, nil
}

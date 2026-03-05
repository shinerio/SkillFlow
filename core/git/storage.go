package git

import (
	"encoding/json"
	"os"
	"sync"
)

type StarStorage struct {
	path string
	mu   sync.Mutex
}

func NewStarStorage(path string) *StarStorage {
	return &StarStorage{path: path}
}

func (s *StarStorage) Load() ([]StarredRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []StarredRepo
	return repos, json.Unmarshal(data, &repos)
}

func (s *StarStorage) Save(repos []StarredRepo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Service struct {
	dataDir    string
	configPath string
}

func NewService(dataDir string) *Service {
	return &Service{
		dataDir:    dataDir,
		configPath: filepath.Join(dataDir, "config.json"),
	}
}

func (s *Service) Load() (AppConfig, error) {
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		cfg := DefaultConfig(s.dataDir)
		if err := s.Save(cfg); err != nil {
			return AppConfig{}, err
		}
		return cfg, nil
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return AppConfig{}, err
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, err
	}
	cfg.LogLevel = NormalizeLogLevel(cfg.LogLevel)
	return cfg, nil
}

func (s *Service) Save(cfg AppConfig) error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}
	cfg.LogLevel = NormalizeLogLevel(cfg.LogLevel)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

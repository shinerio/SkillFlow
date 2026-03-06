package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// sharedConfig is stored in config.json and safe to sync across platforms.
// It contains no file system paths.
type sharedConfig struct {
	DefaultCategory string             `json:"defaultCategory"`
	LogLevel        string             `json:"logLevel"`
	Tools           []sharedToolConfig `json:"tools"`
	Cloud           CloudConfig        `json:"cloud"`
	Proxy           ProxyConfig        `json:"proxy"`
}

// sharedToolConfig stores only the platform-agnostic settings for a built-in tool.
type sharedToolConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// localConfig is stored in config_local.json and never synced to cloud/git.
// It holds all file system paths that differ between platforms.
type localConfig struct {
	SkillsStorageDir string            `json:"skillsStorageDir"`
	Tools            []localToolConfig `json:"tools"`
}

// localToolConfig holds path settings for one tool.
// Custom tools are stored only here (name + paths + enabled).
type localToolConfig struct {
	Name     string   `json:"name"`
	ScanDirs []string `json:"scanDirs"`
	PushDir  string   `json:"pushDir"`
	Custom   bool     `json:"custom"`
	Enabled  bool     `json:"enabled"` // only meaningful for custom tools
}

// legacyAppConfig is used to detect the old single-file format that included
// skillsStorageDir directly in config.json.
type legacyAppConfig struct {
	SkillsStorageDir string `json:"skillsStorageDir"`
}

type Service struct {
	dataDir         string
	configPath      string
	localConfigPath string
}

func NewService(dataDir string) *Service {
	return &Service{
		dataDir:         dataDir,
		configPath:      filepath.Join(dataDir, "config.json"),
		localConfigPath: filepath.Join(dataDir, "config_local.json"),
	}
}

// LocalConfigPath returns the path to the local (non-synced) config file.
func (s *Service) LocalConfigPath() string { return s.localConfigPath }

func (s *Service) Load() (AppConfig, error) {
	s.maybeMigrate()

	shared, err := s.loadShared()
	if err != nil {
		return AppConfig{}, err
	}
	local := s.loadLocal()
	cfg := s.merge(shared, local)

	// Persist defaults for any file that does not exist yet (fresh install).
	_ = os.MkdirAll(s.dataDir, 0755)
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		_ = s.saveShared(s.splitShared(cfg))
	}
	if _, err := os.Stat(s.localConfigPath); os.IsNotExist(err) {
		_ = s.saveLocal(s.splitLocal(cfg))
	}

	return cfg, nil
}

func (s *Service) Save(cfg AppConfig) error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}
	cfg.LogLevel = NormalizeLogLevel(cfg.LogLevel)
	if err := s.saveShared(s.splitShared(cfg)); err != nil {
		return err
	}
	return s.saveLocal(s.splitLocal(cfg))
}

// maybeMigrate converts the old single-file config.json (which contained paths)
// into the new split format. It is a no-op when config_local.json already exists
// or when config.json does not exist yet.
func (s *Service) maybeMigrate() {
	if _, err := os.Stat(s.localConfigPath); err == nil {
		return // already migrated
	}
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return // config.json doesn't exist yet — fresh install
	}
	var legacy legacyAppConfig
	if err := json.Unmarshal(data, &legacy); err != nil || legacy.SkillsStorageDir == "" {
		return // not the old format
	}
	// Old format detected: unmarshal full AppConfig and re-save in split format.
	var old AppConfig
	if err := json.Unmarshal(data, &old); err != nil {
		return
	}
	old.LogLevel = NormalizeLogLevel(old.LogLevel)
	_ = s.saveShared(s.splitShared(old))
	_ = s.saveLocal(s.splitLocal(old))
}

func (s *Service) loadShared() (sharedConfig, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.defaultShared(), nil
		}
		return sharedConfig{}, err
	}
	var sc sharedConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return sharedConfig{}, err
	}
	return sc, nil
}

func (s *Service) loadLocal() localConfig {
	data, err := os.ReadFile(s.localConfigPath)
	if err != nil {
		return s.defaultLocal()
	}
	var lc localConfig
	if err := json.Unmarshal(data, &lc); err != nil {
		return s.defaultLocal()
	}
	if lc.SkillsStorageDir == "" {
		lc.SkillsStorageDir = filepath.Join(s.dataDir, "skills")
	}
	return lc
}

func (s *Service) saveShared(sc sharedConfig) error {
	sc.LogLevel = NormalizeLogLevel(sc.LogLevel)
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

func (s *Service) saveLocal(lc localConfig) error {
	data, err := json.MarshalIndent(lc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.localConfigPath, data, 0644)
}

// defaultShared returns the default shared config (no paths).
func (s *Service) defaultShared() sharedConfig {
	tools := make([]sharedToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		tools = append(tools, sharedToolConfig{Name: name, Enabled: true})
	}
	return sharedConfig{
		DefaultCategory: "Default",
		LogLevel:        DefaultLogLevel,
		Tools:           tools,
		Cloud:           CloudConfig{RemotePath: "skillflow/"},
	}
}

// defaultLocal returns the default local config using platform-specific paths.
func (s *Service) defaultLocal() localConfig {
	tools := make([]localToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		tools = append(tools, localToolConfig{
			Name:     name,
			ScanDirs: DefaultToolScanDirs(name),
			PushDir:  DefaultToolsDir(name),
		})
	}
	return localConfig{
		SkillsStorageDir: filepath.Join(s.dataDir, "skills"),
		Tools:            tools,
	}
}

// merge combines shared and local configs into the single AppConfig used by the app.
func (s *Service) merge(shared sharedConfig, local localConfig) AppConfig {
	localMap := make(map[string]localToolConfig, len(local.Tools))
	for _, lt := range local.Tools {
		localMap[lt.Name] = lt
	}

	var tools []ToolConfig
	// Built-in tools: enabled/name from shared, paths from local (fall back to platform defaults).
	for _, st := range shared.Tools {
		lt := localMap[st.Name]
		scanDirs := lt.ScanDirs
		pushDir := lt.PushDir
		if len(scanDirs) == 0 {
			scanDirs = DefaultToolScanDirs(st.Name)
		}
		if pushDir == "" {
			pushDir = DefaultToolsDir(st.Name)
		}
		tools = append(tools, ToolConfig{
			Name:     st.Name,
			ScanDirs: scanDirs,
			PushDir:  pushDir,
			Enabled:  st.Enabled,
			Custom:   false,
		})
	}
	// Custom tools are stored entirely in local config.
	for _, lt := range local.Tools {
		if lt.Custom {
			tools = append(tools, ToolConfig{
				Name:     lt.Name,
				ScanDirs: lt.ScanDirs,
				PushDir:  lt.PushDir,
				Enabled:  lt.Enabled,
				Custom:   true,
			})
		}
	}

	return AppConfig{
		SkillsStorageDir: local.SkillsStorageDir,
		DefaultCategory:  shared.DefaultCategory,
		LogLevel:         NormalizeLogLevel(shared.LogLevel),
		Tools:            tools,
		Cloud:            shared.Cloud,
		Proxy:            shared.Proxy,
	}
}

// splitShared extracts the platform-agnostic fields from AppConfig.
func (s *Service) splitShared(cfg AppConfig) sharedConfig {
	var tools []sharedToolConfig
	for _, t := range cfg.Tools {
		if !t.Custom {
			tools = append(tools, sharedToolConfig{Name: t.Name, Enabled: t.Enabled})
		}
	}
	return sharedConfig{
		DefaultCategory: cfg.DefaultCategory,
		LogLevel:        NormalizeLogLevel(cfg.LogLevel),
		Tools:           tools,
		Cloud:           cfg.Cloud,
		Proxy:           cfg.Proxy,
	}
}

// splitLocal extracts the path-sensitive fields from AppConfig.
func (s *Service) splitLocal(cfg AppConfig) localConfig {
	tools := make([]localToolConfig, 0, len(cfg.Tools))
	for _, t := range cfg.Tools {
		tools = append(tools, localToolConfig{
			Name:     t.Name,
			ScanDirs: t.ScanDirs,
			PushDir:  t.PushDir,
			Custom:   t.Custom,
			Enabled:  t.Enabled,
		})
	}
	return localConfig{
		SkillsStorageDir: cfg.SkillsStorageDir,
		Tools:            tools,
	}
}

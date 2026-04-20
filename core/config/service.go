package config

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strings"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/shinerio/skillflow/core/platform/settingsstore"
	"github.com/shinerio/skillflow/core/platform/shellsettings"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
)

// sharedConfig is stored in config.json and safe to sync across platforms.
// It contains no file system paths or sensitive cloud credentials.
type sharedConfig struct {
	DefaultCategory      string                         `json:"defaultCategory"`
	LogLevel             string                         `json:"logLevel"`
	RepoScanMaxDepth     int                            `json:"repoScanMaxDepth"`
	Agents               []sharedAgentConfig            `json:"agents"`
	Cloud                sharedCloudState               `json:"cloud"`
	CloudProfiles        map[string]CloudProviderConfig `json:"cloudProfiles,omitempty"`
	SkippedUpdateVersion string                         `json:"skippedUpdateVersion,omitempty"`
	legacyCloudMigrated  bool                           `json:"-"`
	legacyProxyMigrated  bool                           `json:"-"`
	legacyProxy          ProxyConfig                    `json:"-"`
}

type sharedCloudState struct {
	Provider            string `json:"provider"`
	Enabled             bool   `json:"enabled"`
	SyncIntervalMinutes int    `json:"syncIntervalMinutes"`
}

// sharedAgentConfig stores only the platform-agnostic settings for a built-in agent.
type sharedAgentConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// localConfig is stored in config_local.json and never synced to cloud/git.
// It holds all file system paths and sensitive cloud credentials.
type localConfig struct {
	RepoCacheDir               string                       `json:"repoCacheDir"`
	AutoUpdateSkills           bool                         `json:"autoUpdateSkills"`
	AutoPushAgents             []string                     `json:"autoPushAgents"`
	LaunchAtLogin              bool                         `json:"launchAtLogin"`
	Agents                     []localAgentConfig           `json:"agents"`
	AgentSkillManagement       localAgentSkillManagement    `json:"agentSkillManagement,omitempty"`
	CloudCredentialsByProvider map[string]map[string]string `json:"cloudCredentialsByProvider,omitempty"`
	CloudCredentials           map[string]string            `json:"cloudCredentials,omitempty"`
	Proxy                      ProxyConfig                  `json:"proxy"`
	Window                     *WindowState                 `json:"window,omitempty"`
}

type localAgentSkillManagement struct {
	Groups      []string                          `json:"groups,omitempty"`
	Assignments []AgentSkillGroupAssignment       `json:"assignments,omitempty"`
	AgentStates []localAgentSkillManagementState  `json:"agentStates,omitempty"`
}

type localAgentSkillManagementState struct {
	AgentName          string   `json:"agentName"`
	DisabledSkillNames []string `json:"disabledSkillNames,omitempty"`
	DisabledGroupNames []string `json:"disabledGroupNames,omitempty"`
}

// localAgentConfig holds path settings for one agent.
// Custom agents are stored only here (name + paths + enabled).
type localAgentConfig struct {
	Name       string   `json:"name"`
	ScanDirs   []string `json:"scanDirs"`
	PushDir    string   `json:"pushDir"`
	MemoryPath string   `json:"memoryPath"`
	RulesDir   string   `json:"rulesDir"`
	Custom     bool     `json:"custom"`
	Enabled    bool     `json:"enabled"` // only meaningful for custom agents
}

// legacyAppConfig is used to detect the old single-file format that included
// skillsStorageDir directly in config.json.
type legacyAppConfig struct {
	SkillsStorageDir string `json:"skillsStorageDir"`
}

type Service struct {
	store *settingsstore.Store
}

type LocalRuntimeConfig struct {
	RepoCacheDir string
}

func NewService(dataDir string) *Service {
	return &Service{
		store: settingsstore.New(dataDir),
	}
}

// DataDir returns the app data root used by this config service.
func (s *Service) DataDir() string { return s.store.DataDir() }

// LocalConfigPath returns the path to the local (non-synced) config file.
func (s *Service) LocalConfigPath() string { return s.store.LocalPath() }

// LoadLocalRuntimeConfig returns the local-only settings that remain readable
// even when synced files like config.json are temporarily conflicted.
func (s *Service) LoadLocalRuntimeConfig() LocalRuntimeConfig {
	local := s.loadLocal()
	return LocalRuntimeConfig{
		RepoCacheDir: local.RepoCacheDir,
	}
}

func (s *Service) Load() (AppConfig, error) {
	s.maybeMigrate()

	shared, err := s.loadShared()
	if err != nil {
		return AppConfig{}, err
	}
	local := s.loadLocal()
	if s.migrateCloudStorage(&shared, &local) {
		_ = os.MkdirAll(s.DataDir(), 0755)
		_ = s.saveShared(shared)
		_ = s.saveLocal(local)
	}
	cfg := s.merge(shared, local)

	_ = os.MkdirAll(s.DataDir(), 0755)
	if _, err := os.Stat(s.store.SharedPath()); os.IsNotExist(err) {
		_ = s.saveShared(s.splitShared(cfg))
	}
	if _, err := os.Stat(s.store.LocalPath()); os.IsNotExist(err) {
		_ = s.saveLocal(s.splitLocal(cfg))
	}

	return cfg, nil
}

func (s *Service) Save(cfg AppConfig) error {
	if err := os.MkdirAll(s.DataDir(), 0755); err != nil {
		return err
	}
	cfg.LogLevel = NormalizeLogLevel(cfg.LogLevel)
	cfg.AutoPushAgents = NormalizeAgentNameList(cfg.AutoPushAgents)
	cfg.RepoScanMaxDepth = NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	cfg.Proxy = NormalizeProxyConfig(cfg.Proxy)
	cfg.Agents = normalizeAgentConfigs(cfg.Agents)
	cfg.AgentSkillManagement = normalizeAgentSkillManagement(cfg.AgentSkillManagement)

	shared, err := s.loadShared()
	if err != nil {
		return err
	}
	local := s.loadLocal()
	_ = s.migrateCloudStorage(&shared, &local)

	existingProfiles := mergeCloudProfiles(shared.CloudProfiles, local.CloudCredentialsByProvider)
	cfg.CloudProfiles = mergeRuntimeCloudProfiles(existingProfiles, cfg.CloudProfiles, cfg.Cloud)
	cfg.Cloud = buildRuntimeCloudConfig(sharedCloudState{
		Provider:            strings.TrimSpace(cfg.Cloud.Provider),
		Enabled:             cfg.Cloud.Enabled,
		SyncIntervalMinutes: cfg.Cloud.SyncIntervalMinutes,
	}, cfg.CloudProfiles)

	if err := s.saveShared(s.splitShared(cfg)); err != nil {
		return err
	}
	nextLocal := s.splitLocal(cfg)
	nextLocal.Window = local.Window
	return s.saveLocal(nextLocal)
}

// maybeMigrate converts the old single-file config.json (which contained paths)
// into the new split format. It is a no-op when config_local.json already exists
// or when config.json does not exist yet.
func (s *Service) maybeMigrate() {
	if _, err := os.Stat(s.store.LocalPath()); err == nil {
		return
	}
	data, err := os.ReadFile(s.store.SharedPath())
	if err != nil {
		return
	}
	var legacy legacyAppConfig
	if err := json.Unmarshal(data, &legacy); err != nil || legacy.SkillsStorageDir == "" {
		return
	}
	var old AppConfig
	if err := json.Unmarshal(data, &old); err != nil {
		return
	}
	old.LogLevel = NormalizeLogLevel(old.LogLevel)
	_ = s.saveShared(s.splitShared(old))
	_ = s.saveLocal(s.splitLocal(old))
}

func (s *Service) loadShared() (sharedConfig, error) {
	var sc sharedConfig
	exists, err := s.store.ReadShared(&sc)
	if err != nil {
		return sharedConfig{}, err
	}
	if !exists {
		return s.defaultShared(), nil
	}
	sc.LogLevel = NormalizeLogLevel(sc.LogLevel)
	sc.RepoScanMaxDepth = NormalizeRepoScanMaxDepth(sc.RepoScanMaxDepth)
	sc.CloudProfiles = normalizeCloudProfiles(sc.CloudProfiles)

	var legacy struct {
		Cloud CloudConfig  `json:"cloud"`
		Proxy *ProxyConfig `json:"proxy"`
	}
	data, err := os.ReadFile(s.store.SharedPath())
	if err == nil && json.Unmarshal(data, &legacy) == nil {
		provider := strings.TrimSpace(legacy.Cloud.Provider)
		if provider != "" {
			if sc.Cloud.Provider == "" {
				sc.Cloud.Provider = provider
			}
			if len(sc.CloudProfiles) == 0 {
				sc.CloudProfiles = map[string]CloudProviderConfig{
					provider: cloudProviderConfigFromCloud(legacy.Cloud),
				}
				sc.legacyCloudMigrated = true
			}
		}
		if legacy.Proxy != nil {
			sc.legacyProxy = NormalizeProxyConfig(*legacy.Proxy)
			sc.legacyProxyMigrated = true
		}
	}

	return sc, nil
}

func (s *Service) loadLocal() localConfig {
	var lc localConfig
	exists, err := s.store.ReadLocal(&lc)
	if err != nil || !exists {
		return s.defaultLocal()
	}
	lc.RepoCacheDir = normalizeRepoCacheDir(lc.RepoCacheDir, s.DataDir())
	lc.AutoPushAgents = agentapp.NormalizeAutoPushAgentNames(lc.AutoPushAgents)
	lc.CloudCredentialsByProvider = normalizeCredentialProfiles(lc.CloudCredentialsByProvider)
	lc.AgentSkillManagement = toLocalAgentSkillManagement(fromLocalAgentSkillManagement(lc.AgentSkillManagement))
	normalizedShell := shellsettings.NormalizeLocalSettings(shellsettings.LocalSettings{
		LaunchAtLogin: lc.LaunchAtLogin,
		Proxy:         lc.Proxy,
		Window:        lc.Window,
	})
	lc.LaunchAtLogin = normalizedShell.LaunchAtLogin
	lc.Proxy = normalizedShell.Proxy
	lc.Window = normalizedShell.Window
	return lc
}

func (s *Service) saveShared(sc sharedConfig) error {
	sc.LogLevel = NormalizeLogLevel(sc.LogLevel)
	sc.RepoScanMaxDepth = NormalizeRepoScanMaxDepth(sc.RepoScanMaxDepth)
	sc.CloudProfiles = splitSharedCloudProfiles(sc.CloudProfiles)
	return s.store.WriteShared(sc)
}

func (s *Service) saveLocal(lc localConfig) error {
	lc.CloudCredentials = nil
	lc.RepoCacheDir = normalizeRepoCacheDir(lc.RepoCacheDir, s.DataDir())
	lc.AutoPushAgents = agentapp.NormalizeAutoPushAgentNames(lc.AutoPushAgents)
	lc.CloudCredentialsByProvider = normalizeCredentialProfiles(lc.CloudCredentialsByProvider)
	normalizedShell := shellsettings.NormalizeLocalSettings(shellsettings.LocalSettings{
		LaunchAtLogin: lc.LaunchAtLogin,
		Proxy:         lc.Proxy,
		Window:        lc.Window,
	})
	lc.LaunchAtLogin = normalizedShell.LaunchAtLogin
	lc.Proxy = normalizedShell.Proxy
	lc.Window = normalizedShell.Window
	return s.store.WriteLocal(lc)
}

func (s *Service) LoadWindowState() (WindowState, bool) {
	state, ok, err := s.store.LoadWindowState()
	if err != nil {
		return WindowState{}, false
	}
	return state, ok
}

func (s *Service) SaveWindowState(state WindowState) error {
	return s.store.SaveWindowState(state)
}

func (s *Service) defaultShared() sharedConfig {
	defaultAgentSettings := agentapp.DefaultSharedSettings()
	defaultSkillSettings := skillcatalogapp.DefaultSharedSettings()
	defaultShellSettings := shellsettings.DefaultSharedSettings()

	agents := make([]sharedAgentConfig, 0, len(defaultAgentSettings.Agents))
	for _, agent := range defaultAgentSettings.Agents {
		agents = append(agents, sharedAgentConfig{Name: agent.Name, Enabled: agent.Enabled})
	}
	return sharedConfig{
		DefaultCategory:  defaultSkillSettings.DefaultCategory,
		LogLevel:         defaultShellSettings.LogLevel,
		RepoScanMaxDepth: defaultAgentSettings.RepoScanMaxDepth,
		Agents:           agents,
	}
}

func (s *Service) defaultLocal() localConfig {
	defaultAgentSettings := agentapp.DefaultLocalSettings()
	defaultShellSettings := shellsettings.DefaultLocalSettings()

	agents := make([]localAgentConfig, 0, len(defaultAgentSettings.Agents))
	for _, agent := range defaultAgentSettings.Agents {
		agents = append(agents, localAgentConfig{
			Name:       agent.Name,
			ScanDirs:   append([]string(nil), agent.ScanDirs...),
			PushDir:    agent.PushDir,
			MemoryPath: agent.MemoryPath,
			RulesDir:   agent.RulesDir,
			Custom:     agent.Custom,
			Enabled:    agent.Enabled,
		})
	}
	return localConfig{
		RepoCacheDir:         appdata.RepoCacheDir(s.DataDir()),
		Agents:               agents,
		AgentSkillManagement: localAgentSkillManagement{},
		Proxy:                defaultShellSettings.Proxy,
	}
}

// merge combines shared and local configs into the single AppConfig used by the app.
func (s *Service) merge(shared sharedConfig, local localConfig) AppConfig {
	localMap := make(map[string]localAgentConfig, len(local.Agents))
	for _, la := range local.Agents {
		localMap[la.Name] = la
	}

	sharedMap := make(map[string]sharedAgentConfig, len(shared.Agents))
	for _, sa := range shared.Agents {
		if strings.TrimSpace(sa.Name) == "" {
			continue
		}
		sharedMap[sa.Name] = sa
	}

	builtinNames := agentdomain.BuiltinAgentNames()
	agents := make([]AgentConfig, 0, len(builtinNames)+len(local.Agents))
	for _, name := range builtinNames {
		sa, ok := sharedMap[name]
		enabled := true
		if ok {
			enabled = sa.Enabled
		}
		la := localMap[name]
		scanDirs := la.ScanDirs
		pushDir := la.PushDir
		memoryPath := la.MemoryPath
		rulesDir := la.RulesDir
		defaultProfile := agentdomain.DefaultProfile(name)
		if len(scanDirs) == 0 {
			scanDirs = defaultProfile.ScanDirs
		}
		if pushDir == "" {
			pushDir = defaultProfile.PushDir
		}
		if memoryPath == "" {
			memoryPath = defaultProfile.MemoryPath
		}
		if rulesDir == "" {
			rulesDir = defaultProfile.RulesDir
		}
		agents = append(agents, AgentConfig{
			Name:       name,
			ScanDirs:   scanDirs,
			PushDir:    pushDir,
			MemoryPath: memoryPath,
			RulesDir:   rulesDir,
			Enabled:    enabled,
			Custom:     false,
		})
	}
	for _, la := range local.Agents {
		if la.Custom {
			agents = append(agents, AgentConfig{
				Name:       la.Name,
				ScanDirs:   la.ScanDirs,
				PushDir:    la.PushDir,
				MemoryPath: la.MemoryPath,
				RulesDir:   la.RulesDir,
				Enabled:    la.Enabled,
				Custom:     true,
			})
		}
	}
	agents = normalizeAgentConfigs(agents)

	cloudProfiles := mergeCloudProfiles(shared.CloudProfiles, local.CloudCredentialsByProvider)
	return AppConfig{
		RepoCacheDir:         local.RepoCacheDir,
		AutoUpdateSkills:     local.AutoUpdateSkills,
		AutoPushAgents:       NormalizeAgentNameList(local.AutoPushAgents),
		LaunchAtLogin:        local.LaunchAtLogin,
		DefaultCategory:      shared.DefaultCategory,
		LogLevel:             NormalizeLogLevel(shared.LogLevel),
		RepoScanMaxDepth:     NormalizeRepoScanMaxDepth(shared.RepoScanMaxDepth),
		Agents:               agents,
		Cloud:                buildRuntimeCloudConfig(shared.Cloud, cloudProfiles),
		CloudProfiles:        cloudProfiles,
		Proxy:                NormalizeProxyConfig(local.Proxy),
		AgentSkillManagement: fromLocalAgentSkillManagement(local.AgentSkillManagement),
		SkippedUpdateVersion: shared.SkippedUpdateVersion,
	}
}

func normalizeAgentConfigs(agents []AgentConfig) []AgentConfig {
	if len(agents) == 0 {
		return nil
	}
	normalized := make([]AgentConfig, 0, len(agents))
	for _, agent := range agents {
		next := agent
		next.Name = strings.TrimSpace(next.Name)
		next.PushDir = strings.TrimSpace(next.PushDir)
		next.MemoryPath = strings.TrimSpace(next.MemoryPath)
		next.RulesDir = strings.TrimSpace(next.RulesDir)
		next.ScanDirs = normalizeAgentPathList(next.ScanDirs)
		if next.Custom && len(next.ScanDirs) == 0 && next.PushDir != "" {
			next.ScanDirs = []string{next.PushDir}
		}
		normalized = append(normalized, next)
	}
	return normalized
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeAgentSkillManagement(cfg AgentSkillManagementConfig) AgentSkillManagementConfig {
	cfg.Groups = normalizeStringList(cfg.Groups)

	assignments := make([]AgentSkillGroupAssignment, 0, len(cfg.Assignments))
	seenAssignments := make(map[string]struct{}, len(cfg.Assignments))
	for _, assignment := range cfg.Assignments {
		skillName := strings.TrimSpace(assignment.SkillName)
		groupName := strings.TrimSpace(assignment.GroupName)
		if skillName == "" || groupName == "" {
			continue
		}
		if _, ok := seenAssignments[skillName]; ok {
			continue
		}
		seenAssignments[skillName] = struct{}{}
		assignments = append(assignments, AgentSkillGroupAssignment{
			SkillName: skillName,
			GroupName: groupName,
		})
	}
	if len(assignments) == 0 {
		cfg.Assignments = nil
	} else {
		cfg.Assignments = assignments
	}

	agentStates := make([]AgentSkillAgentState, 0, len(cfg.AgentStates))
	seenAgentStates := make(map[string]struct{}, len(cfg.AgentStates))
	for _, state := range cfg.AgentStates {
		agentName := strings.TrimSpace(state.AgentName)
		if agentName == "" {
			continue
		}
		if _, ok := seenAgentStates[agentName]; ok {
			continue
		}
		seenAgentStates[agentName] = struct{}{}
		agentStates = append(agentStates, AgentSkillAgentState{
			AgentName:          agentName,
			DisabledSkillNames: normalizeStringList(state.DisabledSkillNames),
			DisabledGroupNames: normalizeStringList(state.DisabledGroupNames),
		})
	}
	if len(agentStates) == 0 {
		cfg.AgentStates = nil
	} else {
		cfg.AgentStates = agentStates
	}

	return cfg
}

func fromLocalAgentSkillManagement(local localAgentSkillManagement) AgentSkillManagementConfig {
	cfg := AgentSkillManagementConfig{
		Groups:      append([]string(nil), local.Groups...),
		Assignments: append([]AgentSkillGroupAssignment(nil), local.Assignments...),
	}
	if len(local.AgentStates) > 0 {
		cfg.AgentStates = make([]AgentSkillAgentState, 0, len(local.AgentStates))
		for _, state := range local.AgentStates {
			cfg.AgentStates = append(cfg.AgentStates, AgentSkillAgentState{
				AgentName:          state.AgentName,
				DisabledSkillNames: append([]string(nil), state.DisabledSkillNames...),
				DisabledGroupNames: append([]string(nil), state.DisabledGroupNames...),
			})
		}
	}
	return normalizeAgentSkillManagement(cfg)
}

func toLocalAgentSkillManagement(cfg AgentSkillManagementConfig) localAgentSkillManagement {
	cfg = normalizeAgentSkillManagement(cfg)
	local := localAgentSkillManagement{
		Groups:      append([]string(nil), cfg.Groups...),
		Assignments: append([]AgentSkillGroupAssignment(nil), cfg.Assignments...),
	}
	if len(cfg.AgentStates) > 0 {
		local.AgentStates = make([]localAgentSkillManagementState, 0, len(cfg.AgentStates))
		for _, state := range cfg.AgentStates {
			local.AgentStates = append(local.AgentStates, localAgentSkillManagementState{
				AgentName:          state.AgentName,
				DisabledSkillNames: append([]string(nil), state.DisabledSkillNames...),
				DisabledGroupNames: append([]string(nil), state.DisabledGroupNames...),
			})
		}
	}
	return local
}

func normalizeAgentPathList(paths []string) []string {
	return normalizeStringList(paths)
}

// splitShared extracts the platform-agnostic fields from AppConfig.
func (s *Service) splitShared(cfg AppConfig) sharedConfig {
	var agents []sharedAgentConfig
	for _, agent := range cfg.Agents {
		if !agent.Custom {
			agents = append(agents, sharedAgentConfig{Name: agent.Name, Enabled: agent.Enabled})
		}
	}
	profiles := mergeRuntimeCloudProfiles(nil, cfg.CloudProfiles, cfg.Cloud)
	return sharedConfig{
		DefaultCategory:      cfg.DefaultCategory,
		LogLevel:             NormalizeLogLevel(cfg.LogLevel),
		RepoScanMaxDepth:     NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth),
		Agents:               agents,
		Cloud:                sharedCloudState{Provider: strings.TrimSpace(cfg.Cloud.Provider), Enabled: cfg.Cloud.Enabled, SyncIntervalMinutes: cfg.Cloud.SyncIntervalMinutes},
		CloudProfiles:        splitSharedCloudProfiles(profiles),
		SkippedUpdateVersion: cfg.SkippedUpdateVersion,
	}
}

// splitLocal extracts the path-sensitive fields from AppConfig.
func (s *Service) splitLocal(cfg AppConfig) localConfig {
	agents := make([]localAgentConfig, 0, len(cfg.Agents))
	for _, agent := range cfg.Agents {
		agents = append(agents, localAgentConfig{
			Name:       agent.Name,
			ScanDirs:   agent.ScanDirs,
			PushDir:    agent.PushDir,
			MemoryPath: agent.MemoryPath,
			RulesDir:   agent.RulesDir,
			Custom:     agent.Custom,
			Enabled:    agent.Enabled,
		})
	}
	profiles := mergeRuntimeCloudProfiles(nil, cfg.CloudProfiles, cfg.Cloud)
	return localConfig{
		RepoCacheDir:               cfg.RepoCacheDir,
		AutoUpdateSkills:           cfg.AutoUpdateSkills,
		AutoPushAgents:             NormalizeAgentNameList(cfg.AutoPushAgents),
		LaunchAtLogin:              cfg.LaunchAtLogin,
		Agents:                     agents,
		AgentSkillManagement:       toLocalAgentSkillManagement(cfg.AgentSkillManagement),
		CloudCredentialsByProvider: splitLocalCloudCredentialsByProvider(profiles),
		Proxy:                      NormalizeProxyConfig(cfg.Proxy),
	}
}

func normalizeRepoCacheDir(path string, dataDir string) string {
	if strings.TrimSpace(path) == "" {
		return appdata.RepoCacheDir(dataDir)
	}
	return strings.TrimSpace(path)
}

func (s *Service) migrateCloudStorage(shared *sharedConfig, local *localConfig) bool {
	changed := shared.legacyCloudMigrated

	if shared.legacyProxyMigrated {
		if shellsettings.IsZeroProxyConfig(local.Proxy) {
			local.Proxy = NormalizeProxyConfig(shared.legacyProxy)
		}
		changed = true
	}

	normalizedSharedProfiles := normalizeCloudProfiles(shared.CloudProfiles)
	if !cloudProfilesEqual(shared.CloudProfiles, normalizedSharedProfiles) {
		shared.CloudProfiles = normalizedSharedProfiles
		changed = true
	}

	normalizedLocalCredentials := normalizeCredentialProfiles(local.CloudCredentialsByProvider)
	if !credentialProfileMapsEqual(local.CloudCredentialsByProvider, normalizedLocalCredentials) {
		local.CloudCredentialsByProvider = normalizedLocalCredentials
		changed = true
	}

	if len(local.CloudCredentials) > 0 {
		provider := strings.TrimSpace(shared.Cloud.Provider)
		if provider != "" {
			nextCredentials := cloneCredentialProfiles(local.CloudCredentialsByProvider)
			if nextCredentials == nil {
				nextCredentials = make(map[string]map[string]string)
			}
			nextCredentials[provider] = mergeCredentialMaps(nextCredentials[provider], local.CloudCredentials)
			local.CloudCredentialsByProvider = normalizeCredentialProfiles(nextCredentials)
		}
		local.CloudCredentials = nil
		changed = true
	}

	if len(shared.CloudProfiles) > 0 {
		nextSharedProfiles := cloneCloudProfiles(shared.CloudProfiles)
		nextLocalCredentials := cloneCredentialProfiles(local.CloudCredentialsByProvider)
		if nextSharedProfiles == nil {
			nextSharedProfiles = make(map[string]CloudProviderConfig)
		}
		if nextLocalCredentials == nil {
			nextLocalCredentials = make(map[string]map[string]string)
		}
		for provider, profile := range shared.CloudProfiles {
			sharedCredsProfile := normalizeCloudProviderConfig(provider, profile)
			sharedCreds := splitSharedCloudCredentials(sharedCredsProfile.Credentials)
			localCreds := splitLocalCloudCredentials(sharedCredsProfile.Credentials)
			nextProfile := sharedCredsProfile
			nextProfile.Credentials = sharedCreds
			nextSharedProfiles[provider] = nextProfile
			if len(localCreds) > 0 {
				nextLocalCredentials[provider] = mergeCredentialMaps(nextLocalCredentials[provider], localCreds)
			}
		}
		nextSharedProfiles = normalizeCloudProfiles(nextSharedProfiles)
		nextLocalCredentials = normalizeCredentialProfiles(nextLocalCredentials)
		if !cloudProfilesEqual(shared.CloudProfiles, nextSharedProfiles) {
			shared.CloudProfiles = nextSharedProfiles
			changed = true
		}
		if !credentialProfileMapsEqual(local.CloudCredentialsByProvider, nextLocalCredentials) {
			local.CloudCredentialsByProvider = nextLocalCredentials
			changed = true
		}
	}

	return changed
}

func buildRuntimeCloudConfig(state sharedCloudState, profiles map[string]CloudProviderConfig) CloudConfig {
	provider := strings.TrimSpace(state.Provider)
	profile := normalizeCloudProviderConfig(provider, profiles[provider])
	return CloudConfig{
		Provider:            provider,
		Enabled:             state.Enabled,
		BucketName:          profile.BucketName,
		RemotePath:          profile.RemotePath,
		Credentials:         cloneStringMap(profile.Credentials),
		SyncIntervalMinutes: state.SyncIntervalMinutes,
	}
}

func cloudProviderConfigFromCloud(cloud CloudConfig) CloudProviderConfig {
	provider := strings.TrimSpace(cloud.Provider)
	return normalizeCloudProviderConfig(provider, CloudProviderConfig{
		BucketName:  strings.TrimSpace(cloud.BucketName),
		RemotePath:  cloud.RemotePath,
		Credentials: cloneStringMap(cloud.Credentials),
	})
}

func normalizeCloudProviderConfig(provider string, profile CloudProviderConfig) CloudProviderConfig {
	normalized := CloudProviderConfig{
		BucketName:  strings.TrimSpace(profile.BucketName),
		RemotePath:  NormalizeCloudRemotePath(profile.RemotePath),
		Credentials: cloneStringMap(profile.Credentials),
	}
	switch strings.TrimSpace(provider) {
	case "aws":
		normalized = normalizeAWSCloudProviderConfig(normalized)
	case "azure":
		normalized = normalizeAzureCloudProviderConfig(normalized)
	case "tencent":
		normalized = normalizeTencentCloudProviderConfig(normalized)
	}
	if len(normalized.Credentials) == 0 {
		normalized.Credentials = nil
	}
	return normalized
}

func normalizeTencentCloudProviderConfig(profile CloudProviderConfig) CloudProviderConfig {
	profile.BucketName = normalizeTencentBucketNameValue(profile.BucketName)
	credentials := cloneStringMap(profile.Credentials)

	if endpoint := normalizeTencentEndpointValue(credentials["endpoint"]); endpoint != "" {
		credentials["endpoint"] = endpoint
	} else {
		delete(credentials, "endpoint")
	}
	delete(credentials, "bucket_url")

	if len(credentials) == 0 {
		profile.Credentials = nil
		return profile
	}
	profile.Credentials = credentials
	return profile
}

func normalizeAWSCloudProviderConfig(profile CloudProviderConfig) CloudProviderConfig {
	credentials := cloneStringMap(profile.Credentials)

	if region := strings.TrimSpace(credentials["region"]); region != "" {
		credentials["region"] = region
	} else {
		delete(credentials, "region")
	}

	if len(credentials) == 0 {
		profile.Credentials = nil
		return profile
	}
	profile.Credentials = credentials
	return profile
}

func normalizeAzureCloudProviderConfig(profile CloudProviderConfig) CloudProviderConfig {
	credentials := cloneStringMap(profile.Credentials)

	if accountName := strings.TrimSpace(credentials["account_name"]); accountName != "" {
		credentials["account_name"] = accountName
	} else {
		delete(credentials, "account_name")
	}

	if serviceURL := normalizeConfigHostLikeValue(credentials["service_url"]); serviceURL != "" {
		credentials["service_url"] = serviceURL
	} else {
		delete(credentials, "service_url")
	}

	if len(credentials) == 0 {
		profile.Credentials = nil
		return profile
	}
	profile.Credentials = credentials
	return profile
}

func normalizeTencentBucketNameValue(raw string) string {
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return ""
	}
	if isValidTencentBucketNameForConfig(bucket) {
		return bucket
	}
	if bucketName, _ := splitTencentBucketURLParts(bucket); bucketName != "" {
		return bucketName
	}
	return bucket
}

func normalizeTencentEndpointValue(raw string) string {
	return normalizeConfigHostLikeValue(raw)
}

func splitTencentBucketURLParts(raw string) (string, string) {
	host := normalizeConfigHostLikeValue(raw)
	if host == "" {
		return "", ""
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 5 && parts[1] == "cos" && isValidTencentBucketNameForConfig(parts[0]) {
		return parts[0], strings.Join(parts[1:], ".")
	}
	return "", ""
}

func normalizeConfigHostLikeValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
			value = parsed.Host
		}
	}
	value = strings.TrimPrefix(value, "//")
	if slash := strings.Index(value, "/"); slash >= 0 {
		value = value[:slash]
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	return strings.TrimSuffix(strings.TrimSpace(value), ".")
}

func isValidTencentBucketNameForConfig(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	for i, r := range name {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isHyphen := r == '-'
		if !isLower && !isDigit && !isHyphen {
			return false
		}
		if (i == 0 || i == len(name)-1) && !isLower && !isDigit {
			return false
		}
	}
	return true
}

func mergeRuntimeCloudProfiles(existing map[string]CloudProviderConfig, incoming map[string]CloudProviderConfig, active CloudConfig) map[string]CloudProviderConfig {
	profiles := cloneCloudProfiles(existing)
	if profiles == nil {
		profiles = make(map[string]CloudProviderConfig)
	}
	for provider, profile := range incoming {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		profiles[provider] = normalizeCloudProviderConfig(provider, profile)
	}
	if provider := strings.TrimSpace(active.Provider); provider != "" {
		profiles[provider] = cloudProviderConfigFromCloud(active)
	}
	return normalizeCloudProfiles(profiles)
}

func mergeCloudProfiles(sharedProfiles map[string]CloudProviderConfig, localCredentialsByProvider map[string]map[string]string) map[string]CloudProviderConfig {
	profiles := cloneCloudProfiles(sharedProfiles)
	if profiles == nil {
		profiles = make(map[string]CloudProviderConfig)
	}
	for provider, credentials := range localCredentialsByProvider {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		profile := profiles[provider]
		profile.Credentials = mergeCredentialMaps(profile.Credentials, credentials)
		profiles[provider] = normalizeCloudProviderConfig(provider, profile)
	}
	return normalizeCloudProfiles(profiles)
}

func splitSharedCloudProfiles(profiles map[string]CloudProviderConfig) map[string]CloudProviderConfig {
	sharedProfiles := make(map[string]CloudProviderConfig)
	for provider, profile := range profiles {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		sharedProfile := normalizeCloudProviderConfig(provider, profile)
		sharedProfile.Credentials = splitSharedCloudCredentials(sharedProfile.Credentials)
		if sharedProfile.RemotePath == "" {
			sharedProfile.RemotePath = DefaultCloudRemotePath
		}
		sharedProfiles[provider] = sharedProfile
	}
	if len(sharedProfiles) == 0 {
		return nil
	}
	return sharedProfiles
}

func splitLocalCloudCredentialsByProvider(profiles map[string]CloudProviderConfig) map[string]map[string]string {
	localProfiles := make(map[string]map[string]string)
	for provider, profile := range profiles {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		normalizedProfile := normalizeCloudProviderConfig(provider, profile)
		credentials := splitLocalCloudCredentials(normalizedProfile.Credentials)
		if len(credentials) == 0 {
			continue
		}
		localProfiles[provider] = credentials
	}
	if len(localProfiles) == 0 {
		return nil
	}
	return localProfiles
}

func splitSharedCloudCredentials(credentials map[string]string) map[string]string {
	shared := make(map[string]string)
	for key, value := range credentials {
		if isSharedCloudCredentialKey(key) {
			shared[key] = value
		}
	}
	if len(shared) == 0 {
		return nil
	}
	return shared
}

func splitLocalCloudCredentials(credentials map[string]string) map[string]string {
	local := make(map[string]string)
	for key, value := range credentials {
		if !isSharedCloudCredentialKey(key) {
			local[key] = value
		}
	}
	if len(local) == 0 {
		return nil
	}
	return local
}

func normalizeCloudProfiles(profiles map[string]CloudProviderConfig) map[string]CloudProviderConfig {
	normalized := make(map[string]CloudProviderConfig)
	for provider, profile := range profiles {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		normalized[provider] = normalizeCloudProviderConfig(provider, profile)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeCredentialProfiles(profiles map[string]map[string]string) map[string]map[string]string {
	normalized := make(map[string]map[string]string)
	for provider, credentials := range profiles {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		copied := cloneStringMap(credentials)
		if len(copied) == 0 {
			continue
		}
		normalized[provider] = copied
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func cloneCloudProfiles(profiles map[string]CloudProviderConfig) map[string]CloudProviderConfig {
	cloned := make(map[string]CloudProviderConfig)
	for provider, profile := range profiles {
		cloned[provider] = normalizeCloudProviderConfig(provider, profile)
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func cloneCredentialProfiles(profiles map[string]map[string]string) map[string]map[string]string {
	cloned := make(map[string]map[string]string)
	for provider, credentials := range profiles {
		copied := cloneStringMap(credentials)
		if len(copied) == 0 {
			continue
		}
		cloned[provider] = copied
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func mergeCredentialMaps(mapsToMerge ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, current := range mapsToMerge {
		for key, value := range current {
			merged[key] = value
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func cloudProfilesEqual(left, right map[string]CloudProviderConfig) bool {
	if len(left) != len(right) {
		return false
	}
	for provider, leftProfile := range left {
		rightProfile, ok := right[provider]
		if !ok {
			return false
		}
		if leftProfile.BucketName != rightProfile.BucketName || leftProfile.RemotePath != rightProfile.RemotePath || !credentialMapsEqual(leftProfile.Credentials, rightProfile.Credentials) {
			return false
		}
	}
	return true
}

func credentialProfileMapsEqual(left, right map[string]map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for provider, leftCredentials := range left {
		if !credentialMapsEqual(leftCredentials, right[provider]) {
			return false
		}
	}
	return true
}

func credentialMapsEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func isSharedCloudCredentialKey(key string) bool {
	switch key {
	case "endpoint", "repo_url", "branch", "username", "region", "account_name", "service_url":
		return true
	default:
		return false
	}
}

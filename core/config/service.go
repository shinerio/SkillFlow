package config

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// sharedConfig is stored in config.json and safe to sync across platforms.
// It contains no file system paths or sensitive cloud credentials.
type sharedConfig struct {
	DefaultCategory       string                         `json:"defaultCategory"`
	LogLevel              string                         `json:"logLevel"`
	RepoScanMaxDepth      int                            `json:"repoScanMaxDepth"`
	SkillStatusVisibility SkillStatusVisibilityConfig    `json:"skillStatusVisibility"`
	Tools                 []sharedToolConfig             `json:"tools"`
	Cloud                 sharedCloudState               `json:"cloud"`
	CloudProfiles         map[string]CloudProviderConfig `json:"cloudProfiles,omitempty"`
	SkippedUpdateVersion  string                         `json:"skippedUpdateVersion,omitempty"`
	legacyCloudMigrated   bool                           `json:"-"`
	legacyProxyMigrated   bool                           `json:"-"`
	legacyProxy           ProxyConfig                    `json:"-"`
}

type sharedCloudState struct {
	Provider            string `json:"provider"`
	Enabled             bool   `json:"enabled"`
	SyncIntervalMinutes int    `json:"syncIntervalMinutes"`
}

// sharedToolConfig stores only the platform-agnostic settings for a built-in tool.
type sharedToolConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// localConfig is stored in config_local.json and never synced to cloud/git.
// It holds all file system paths and sensitive cloud credentials.
type localConfig struct {
	SkillsStorageDir           string                       `json:"skillsStorageDir"`
	Tools                      []localToolConfig            `json:"tools"`
	CloudCredentialsByProvider map[string]map[string]string `json:"cloudCredentialsByProvider,omitempty"`
	CloudCredentials           map[string]string            `json:"cloudCredentials,omitempty"`
	Proxy                      ProxyConfig                  `json:"proxy"`
	GitHubPAT                  string                       `json:"githubPat,omitempty"`
	Window                     *WindowState                 `json:"window,omitempty"`
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

type LocalRuntimeConfig struct {
	SkillsStorageDir string
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

// LoadLocalRuntimeConfig returns the local-only settings that remain readable
// even when synced files like config.json are temporarily conflicted.
func (s *Service) LoadLocalRuntimeConfig() LocalRuntimeConfig {
	local := s.loadLocal()
	return LocalRuntimeConfig{
		SkillsStorageDir: local.SkillsStorageDir,
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
		_ = os.MkdirAll(s.dataDir, 0755)
		_ = s.saveShared(shared)
		_ = s.saveLocal(local)
	}
	cfg := s.merge(shared, local)

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
	cfg.RepoScanMaxDepth = NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	cfg.SkillStatusVisibility = NormalizeSkillStatusVisibility(cfg.SkillStatusVisibility)
	cfg.Proxy = NormalizeProxyConfig(cfg.Proxy)

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
	if _, err := os.Stat(s.localConfigPath); err == nil {
		return
	}
	data, err := os.ReadFile(s.configPath)
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
	sc.LogLevel = NormalizeLogLevel(sc.LogLevel)
	sc.RepoScanMaxDepth = NormalizeRepoScanMaxDepth(sc.RepoScanMaxDepth)
	sc.SkillStatusVisibility = NormalizeSkillStatusVisibility(sc.SkillStatusVisibility)
	sc.CloudProfiles = normalizeCloudProfiles(sc.CloudProfiles)

	var legacy struct {
		Cloud CloudConfig  `json:"cloud"`
		Proxy *ProxyConfig `json:"proxy"`
	}
	if err := json.Unmarshal(data, &legacy); err == nil {
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
	lc.CloudCredentialsByProvider = normalizeCredentialProfiles(lc.CloudCredentialsByProvider)
	lc.Proxy = NormalizeProxyConfig(lc.Proxy)
	if lc.Window != nil {
		normalized := NormalizeWindowState(*lc.Window)
		if normalized.Width == 0 || normalized.Height == 0 {
			lc.Window = nil
		} else {
			lc.Window = &normalized
		}
	}
	return lc
}

func (s *Service) saveShared(sc sharedConfig) error {
	sc.LogLevel = NormalizeLogLevel(sc.LogLevel)
	sc.RepoScanMaxDepth = NormalizeRepoScanMaxDepth(sc.RepoScanMaxDepth)
	sc.SkillStatusVisibility = NormalizeSkillStatusVisibility(sc.SkillStatusVisibility)
	sc.CloudProfiles = splitSharedCloudProfiles(sc.CloudProfiles)
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

func (s *Service) saveLocal(lc localConfig) error {
	lc.CloudCredentials = nil
	lc.CloudCredentialsByProvider = normalizeCredentialProfiles(lc.CloudCredentialsByProvider)
	lc.Proxy = NormalizeProxyConfig(lc.Proxy)
	if lc.Window != nil {
		normalized := NormalizeWindowState(*lc.Window)
		if normalized.Width == 0 || normalized.Height == 0 {
			lc.Window = nil
		} else {
			lc.Window = &normalized
		}
	}
	data, err := json.MarshalIndent(lc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.localConfigPath, data, 0644)
}

func (s *Service) LoadWindowState() (WindowState, bool) {
	local := s.loadLocal()
	if local.Window == nil {
		return WindowState{}, false
	}
	return *local.Window, true
}

func (s *Service) SaveWindowState(state WindowState) error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}
	normalized := NormalizeWindowState(state)
	if normalized.Width == 0 || normalized.Height == 0 {
		return nil
	}
	local := s.loadLocal()
	local.Window = &normalized
	return s.saveLocal(local)
}

func (s *Service) defaultShared() sharedConfig {
	tools := make([]sharedToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		tools = append(tools, sharedToolConfig{Name: name, Enabled: true})
	}
	return sharedConfig{
		DefaultCategory:       "Default",
		LogLevel:              DefaultLogLevel,
		RepoScanMaxDepth:      DefaultRepoScanMaxDepth,
		SkillStatusVisibility: DefaultSkillStatusVisibility(),
		Tools:                 tools,
	}
}

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
		Proxy:            ProxyConfig{Mode: ProxyModeNone},
	}
}

// merge combines shared and local configs into the single AppConfig used by the app.
func (s *Service) merge(shared sharedConfig, local localConfig) AppConfig {
	localMap := make(map[string]localToolConfig, len(local.Tools))
	for _, lt := range local.Tools {
		localMap[lt.Name] = lt
	}

	var tools []ToolConfig
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

	cloudProfiles := mergeCloudProfiles(shared.CloudProfiles, local.CloudCredentialsByProvider)
	return AppConfig{
		SkillsStorageDir:      local.SkillsStorageDir,
		DefaultCategory:       shared.DefaultCategory,
		LogLevel:              NormalizeLogLevel(shared.LogLevel),
		RepoScanMaxDepth:      NormalizeRepoScanMaxDepth(shared.RepoScanMaxDepth),
		SkillStatusVisibility: NormalizeSkillStatusVisibility(shared.SkillStatusVisibility),
		Tools:                 tools,
		Cloud:                 buildRuntimeCloudConfig(shared.Cloud, cloudProfiles),
		CloudProfiles:         cloudProfiles,
		Proxy:                 NormalizeProxyConfig(local.Proxy),
		GitHub:                GitHubConfig{PAT: strings.TrimSpace(local.GitHubPAT)},
		SkippedUpdateVersion:  shared.SkippedUpdateVersion,
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
	profiles := mergeRuntimeCloudProfiles(nil, cfg.CloudProfiles, cfg.Cloud)
	return sharedConfig{
		DefaultCategory:       cfg.DefaultCategory,
		LogLevel:              NormalizeLogLevel(cfg.LogLevel),
		RepoScanMaxDepth:      NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth),
		SkillStatusVisibility: NormalizeSkillStatusVisibility(cfg.SkillStatusVisibility),
		Tools:                 tools,
		Cloud:                 sharedCloudState{Provider: strings.TrimSpace(cfg.Cloud.Provider), Enabled: cfg.Cloud.Enabled, SyncIntervalMinutes: cfg.Cloud.SyncIntervalMinutes},
		CloudProfiles:         splitSharedCloudProfiles(profiles),
		SkippedUpdateVersion:  cfg.SkippedUpdateVersion,
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
	profiles := mergeRuntimeCloudProfiles(nil, cfg.CloudProfiles, cfg.Cloud)
	return localConfig{
		SkillsStorageDir:           cfg.SkillsStorageDir,
		Tools:                      tools,
		CloudCredentialsByProvider: splitLocalCloudCredentialsByProvider(profiles),
		Proxy:                      NormalizeProxyConfig(cfg.Proxy),
		GitHubPAT:                  strings.TrimSpace(cfg.GitHub.PAT),
	}
}

func (s *Service) migrateCloudStorage(shared *sharedConfig, local *localConfig) bool {
	changed := shared.legacyCloudMigrated

	if shared.legacyProxyMigrated {
		if isZeroProxyConfig(local.Proxy) {
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

func isZeroProxyConfig(proxy ProxyConfig) bool {
	mode := ProxyMode(strings.ToLower(strings.TrimSpace(string(proxy.Mode))))
	url := strings.TrimSpace(proxy.URL)
	return url == "" && (mode == "" || mode == ProxyModeNone)
}

func isSharedCloudCredentialKey(key string) bool {
	switch key {
	case "endpoint", "repo_url", "branch", "username", "region", "account_name", "service_url":
		return true
	default:
		return false
	}
}

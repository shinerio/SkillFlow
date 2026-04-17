package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg, err := svc.Load()
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "cache", "repos"), cfg.RepoCacheDir)
	assert.Equal(t, "Default", cfg.DefaultCategory)
	assert.False(t, cfg.AutoUpdateSkills)
	assert.Equal(t, config.DefaultLogLevel, cfg.LogLevel)
	assert.Equal(t, config.DefaultRepoScanMaxDepth, cfg.RepoScanMaxDepth)
	assert.NotEmpty(t, cfg.Agents)

	var copilot config.AgentConfig
	var found bool
	for _, agent := range cfg.Agents {
		if agent.Name != "copilot" {
			continue
		}
		copilot = agent
		found = true
		break
	}
	require.True(t, found)
	assert.NotEmpty(t, copilot.ScanDirs)
	assert.Equal(t, filepath.Join(home, ".copilot", "skills"), copilot.PushDir)
	assert.Contains(t, copilot.ScanDirs, filepath.Join(home, ".claude", "skills"))
	assert.Contains(t, copilot.ScanDirs, filepath.Join(home, ".agents", "skills"))
	assert.NotContains(t, copilot.ScanDirs, copilot.PushDir)
	assert.NotEmpty(t, copilot.MemoryPath)
	assert.Empty(t, copilot.RulesDir)
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.DefaultCategory = "MyCategory"
	cfg.AutoUpdateSkills = true
	cfg.AutoPushAgents = []string{"codex", "gemini-cli"}
	cfg.RepoScanMaxDepth = 7
	cfg.Proxy = config.ProxyConfig{
		Mode: config.ProxyModeManual,
		URL:  "http://127.0.0.1:7890",
	}
	cfg.SkippedUpdateVersion = "v1.2.3"
	err := svc.Save(cfg)
	require.NoError(t, err)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "MyCategory", loaded.DefaultCategory)
	assert.True(t, loaded.AutoUpdateSkills)
	assert.Equal(t, []string{"codex", "gemini-cli"}, loaded.AutoPushAgents)
	assert.Equal(t, 7, loaded.RepoScanMaxDepth)
	assert.Equal(t, cfg.Proxy, loaded.Proxy)
	assert.Equal(t, "v1.2.3", loaded.SkippedUpdateVersion)
}

func TestSaveAndLoadPreservesAgentMemoryPaths(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	require.NotEmpty(t, cfg.Agents)

	cfg.Agents[0].MemoryPath = filepath.Join(dir, "custom-agent-memory.md")
	cfg.Agents[0].RulesDir = filepath.Join(dir, "custom-agent-rules")

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	require.NotEmpty(t, loaded.Agents)
	assert.Equal(t, filepath.Join(dir, "custom-agent-memory.md"), loaded.Agents[0].MemoryPath)
	assert.Equal(t, filepath.Join(dir, "custom-agent-rules"), loaded.Agents[0].RulesDir)

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	var local struct {
		Agents []config.AgentConfig `json:"agents"`
	}
	require.NoError(t, json.Unmarshal(localData, &local))
	require.NotEmpty(t, local.Agents)
	assert.Equal(t, filepath.Join(dir, "custom-agent-memory.md"), local.Agents[0].MemoryPath)
	assert.Equal(t, filepath.Join(dir, "custom-agent-rules"), local.Agents[0].RulesDir)
}

func TestSaveAndLoadCustomAgentPreservesExplicitScanDirs(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Agents = append(cfg.Agents, config.AgentConfig{
		Name:       "custom-agent",
		PushDir:    filepath.Join(dir, "push"),
		ScanDirs:   []string{filepath.Join(dir, "scan-a"), filepath.Join(dir, "scan-b")},
		MemoryPath: filepath.Join(dir, "custom-agent-memory.md"),
		RulesDir:   filepath.Join(dir, "custom-agent-rules"),
		Enabled:    true,
		Custom:     true,
	})

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Agents, len(cfg.Agents))

	customAgent := loaded.Agents[len(loaded.Agents)-1]
	assert.Equal(t, "custom-agent", customAgent.Name)
	assert.Equal(t, []string{filepath.Join(dir, "scan-a"), filepath.Join(dir, "scan-b")}, customAgent.ScanDirs)
}

func TestSaveAndLoadCustomAgentFallsBackScanDirsToPushDir(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Agents = append(cfg.Agents, config.AgentConfig{
		Name:       "custom-agent",
		PushDir:    filepath.Join(dir, "push"),
		ScanDirs:   nil,
		MemoryPath: filepath.Join(dir, "custom-agent-memory.md"),
		RulesDir:   filepath.Join(dir, "custom-agent-rules"),
		Enabled:    true,
		Custom:     true,
	})

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Agents, len(cfg.Agents))

	customAgent := loaded.Agents[len(loaded.Agents)-1]
	assert.Equal(t, []string{filepath.Join(dir, "push")}, customAgent.ScanDirs)
}

func TestAutoUpdateSkillsStoredOnlyInLocalConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.AutoUpdateSkills = true

	require.NoError(t, svc.Save(cfg))

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(sharedData), "autoUpdateSkills")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"autoUpdateSkills": true`)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.True(t, loaded.AutoUpdateSkills)
}

func TestAutoPushAgentsStoredOnlyInLocalConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.AutoPushAgents = []string{" codex ", "gemini-cli", "codex", ""}

	require.NoError(t, svc.Save(cfg))

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(sharedData), "autoPushAgents")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"autoPushAgents"`)
	assert.Contains(t, string(localData), `"codex"`)
	assert.Contains(t, string(localData), `"gemini-cli"`)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"codex", "gemini-cli"}, loaded.AutoPushAgents)
}

func TestSkippedUpdateVersionPersistsInSharedConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.SkippedUpdateVersion = "v9.9.9"

	require.NoError(t, svc.Save(cfg))

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"skippedUpdateVersion": "v9.9.9"`)

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(localData), "skippedUpdateVersion")

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "v9.9.9", loaded.SkippedUpdateVersion)
}

func TestConfigDoesNotPersistSkillStatusVisibility(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)

	require.NoError(t, svc.Save(cfg))

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(sharedData), "skillStatusVisibility")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(localData), "skillStatusVisibility")
}

func TestSaveAndLoadConfigNormalizesLogLevel(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.LogLevel = "BAD_LEVEL"
	err := svc.Save(cfg)
	require.NoError(t, err)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, config.DefaultLogLevel, loaded.LogLevel)
}

func TestSaveAndLoadConfigNormalizesRepoScanMaxDepth(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.RepoScanMaxDepth = 999
	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, config.MaxRepoScanMaxDepth, loaded.RepoScanMaxDepth)

	cfg.RepoScanMaxDepth = 0
	require.NoError(t, svc.Save(cfg))

	loaded, err = svc.Load()
	require.NoError(t, err)
	assert.Equal(t, config.DefaultRepoScanMaxDepth, loaded.RepoScanMaxDepth)
}

func TestSaveAndLoadConfigNormalizesCloudRemotePath(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Cloud.Provider = "aliyun"
	cfg.Cloud.RemotePath = "team-a/nightly"

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "team-a/nightly/skillflow/", loaded.Cloud.RemotePath)
	assert.Equal(t, "team-a/nightly/skillflow/", loaded.CloudProfiles["aliyun"].RemotePath)

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), `"cloudProfiles"`)
	assert.Contains(t, string(sharedData), `"remotePath": "team-a/nightly/skillflow/"`)

	cfg.Cloud.RemotePath = "skillflow/"
	require.NoError(t, svc.Save(cfg))

	loaded, err = svc.Load()
	require.NoError(t, err)
	assert.Equal(t, config.DefaultCloudRemotePath, loaded.Cloud.RemotePath)
}

func TestConfigFileCreatedOnFirstLoad(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	_, err := svc.Load()
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "config.json"))
	assert.NoError(t, err, "config.json should be created on first load")
	_, err = os.Stat(filepath.Join(dir, "config_local.json"))
	assert.NoError(t, err, "config_local.json should be created on first load")
}

func TestSaveCreatesLocalConfigWithRepoCacheDir(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.RepoCacheDir = filepath.Join(dir, "volumes", "repo-cache")
	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.RepoCacheDir, loaded.RepoCacheDir)

	// config.json must NOT contain repoCacheDir (it belongs in config_local.json)
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "repoCacheDir")
	assert.NotContains(t, string(data), "skillsStorageDir")

	// config_local.json must contain the repo cache path
	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), "repoCacheDir")
	assert.NotContains(t, string(localData), "skillsStorageDir")
}

func TestProxyStoredOnlyInLocalConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Proxy = config.ProxyConfig{
		Mode: config.ProxyModeManual,
		URL:  "http://127.0.0.1:7890",
	}

	require.NoError(t, svc.Save(cfg))

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(sharedData), `"proxy"`)
	assert.NotContains(t, string(sharedData), "127.0.0.1:7890")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"proxy"`)
	assert.Contains(t, string(localData), `"mode": "manual"`)
	assert.Contains(t, string(localData), `"url": "http://127.0.0.1:7890"`)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.Proxy, loaded.Proxy)
}

func TestLaunchAtLoginStoredOnlyInLocalConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.LaunchAtLogin = true

	require.NoError(t, svc.Save(cfg))

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(sharedData), "launchAtLogin")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"launchAtLogin": true`)

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.True(t, loaded.LaunchAtLogin)
}

func TestWindowStateStoredOnlyInLocalConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)

	require.NoError(t, svc.SaveWindowState(config.WindowState{Width: 1440, Height: 920}))

	state, ok := svc.LoadWindowState()
	require.True(t, ok)
	assert.Equal(t, config.WindowState{Width: 1440, Height: 920}, state)

	sharedPath := filepath.Join(dir, "config.json")
	sharedData, err := os.ReadFile(sharedPath)
	if err == nil {
		assert.NotContains(t, string(sharedData), `"window"`)
	} else {
		assert.ErrorIs(t, err, os.ErrNotExist)
	}

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"window"`)
	assert.Contains(t, string(localData), `"width": 1440`)
	assert.Contains(t, string(localData), `"height": 920`)
}

func TestSaveConfigPreservesWindowState(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)

	require.NoError(t, svc.SaveWindowState(config.WindowState{Width: 1320, Height: 860}))

	cfg := config.DefaultConfig(dir)
	cfg.DefaultCategory = "Saved"
	require.NoError(t, svc.Save(cfg))

	state, ok := svc.LoadWindowState()
	require.True(t, ok)
	assert.Equal(t, config.WindowState{Width: 1320, Height: 860}, state)
}

func TestCloudSensitiveCredentialsStoredOnlyInLocalConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Cloud.Provider = "aliyun"
	cfg.Cloud.Enabled = true
	cfg.Cloud.BucketName = "skillflow-bucket"
	cfg.Cloud.RemotePath = "skillflow/"
	cfg.Cloud.Credentials = map[string]string{
		"access_key_id":     "test-ak",
		"access_key_secret": "test-sk",
		"endpoint":          "oss-cn-hangzhou.aliyuncs.com",
	}

	require.NoError(t, svc.Save(cfg))

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), "cloudProfiles")
	assert.Contains(t, string(sharedData), "bucketName")
	assert.Contains(t, string(sharedData), "endpoint")
	assert.NotContains(t, string(sharedData), "access_key_id")
	assert.NotContains(t, string(sharedData), "access_key_secret")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), "cloudCredentialsByProvider")
	assert.Contains(t, string(localData), "aliyun")
	assert.Contains(t, string(localData), "access_key_id")
	assert.Contains(t, string(localData), "access_key_secret")
	assert.NotContains(t, string(localData), "endpoint")

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.Cloud.Credentials, loaded.Cloud.Credentials)
	assert.Equal(t, cfg.Cloud.Credentials, loaded.CloudProfiles["aliyun"].Credentials)
}

func TestCloudProviderProfilesPersistSeparately(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Cloud.Provider = "aliyun"
	cfg.Cloud.Enabled = true
	cfg.Cloud.BucketName = "aliyun-bucket"
	cfg.Cloud.RemotePath = "nightly"
	cfg.Cloud.Credentials = map[string]string{
		"access_key_id":     "aliyun-ak",
		"access_key_secret": "aliyun-sk",
		"endpoint":          "oss-cn-hangzhou.aliyuncs.com",
	}
	cfg.CloudProfiles = map[string]config.CloudProviderConfig{
		"git": {
			RemotePath: "skillflow/",
			Credentials: map[string]string{
				"repo_url": "https://example.com/org/repo.git",
				"branch":   "main",
				"username": "alice",
				"token":    "git-token",
			},
		},
		"tencent": {
			BucketName: "bucket-125000",
			RemotePath: "team-b/backup",
			Credentials: map[string]string{
				"endpoint":   "bucket-125000.cos.ap-shanghai.myqcloud.com",
				"secret_id":  "tx-id",
				"secret_key": "tx-key",
			},
		},
	}

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "aliyun", loaded.Cloud.Provider)
	assert.Equal(t, "aliyun-bucket", loaded.Cloud.BucketName)
	assert.Equal(t, "nightly/skillflow/", loaded.Cloud.RemotePath)
	assert.Equal(t, "aliyun-sk", loaded.CloudProfiles["aliyun"].Credentials["access_key_secret"])
	assert.Equal(t, "git-token", loaded.CloudProfiles["git"].Credentials["token"])
	assert.Equal(t, "https://example.com/org/repo.git", loaded.CloudProfiles["git"].Credentials["repo_url"])
	assert.Equal(t, "bucket-125000", loaded.CloudProfiles["tencent"].BucketName)
	assert.Equal(t, "tx-key", loaded.CloudProfiles["tencent"].Credentials["secret_key"])
	assert.Equal(t, "bucket-125000.cos.ap-shanghai.myqcloud.com", loaded.CloudProfiles["tencent"].Credentials["endpoint"])
	assert.Equal(t, "team-b/backup/skillflow/", loaded.CloudProfiles["tencent"].RemotePath)

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), "cloudProfiles")
	assert.Contains(t, string(sharedData), "aliyun-bucket")
	assert.Contains(t, string(sharedData), "repo_url")
	assert.Contains(t, string(sharedData), `"endpoint": "bucket-125000.cos.ap-shanghai.myqcloud.com"`)
	assert.NotContains(t, string(sharedData), "bucket_url")
	assert.NotContains(t, string(sharedData), "aliyun-sk")
	assert.NotContains(t, string(sharedData), "git-token")
	assert.NotContains(t, string(sharedData), "tx-key")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), "cloudCredentialsByProvider")
	assert.Contains(t, string(localData), `"aliyun"`)
	assert.Contains(t, string(localData), `"git"`)
	assert.Contains(t, string(localData), `"tencent"`)
	assert.Contains(t, string(localData), "aliyun-sk")
	assert.Contains(t, string(localData), "git-token")
	assert.Contains(t, string(localData), "tx-key")
	assert.NotContains(t, string(localData), "repo_url")
	assert.NotContains(t, string(localData), "bucket_url")
	assert.NotContains(t, string(localData), "endpoint")
}

func TestLoadMigratesCloudSecretsOutOfSharedConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	shared := `{
	  "defaultCategory": "Default",
	  "logLevel": "info",
	  "repoScanMaxDepth": 5,
	  "agents": [],
	  "cloud": {
	    "provider": "git",
	    "enabled": true,
	    "remotePath": "skillflow/",
	    "credentials": {
	      "repo_url": "https://example.com/org/repo.git",
	      "branch": "main",
	      "username": "alice",
	      "token": "secret-token"
	    }
	  },
	  "proxy": {
	    "mode": "none",
	    "url": ""
	  }
	}`
	local := `{
	  "skillsStorageDir": "` + filepath.ToSlash(filepath.Join(dir, "skills")) + `",
	  "agents": []
	}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(shared), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(local), 0644))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/org/repo.git", loaded.Cloud.Credentials["repo_url"])
	assert.Equal(t, "main", loaded.Cloud.Credentials["branch"])
	assert.Equal(t, "alice", loaded.Cloud.Credentials["username"])
	assert.Equal(t, "secret-token", loaded.Cloud.Credentials["token"])

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), "cloudProfiles")
	assert.Contains(t, string(sharedData), "repo_url")
	assert.Contains(t, string(sharedData), "branch")
	assert.Contains(t, string(sharedData), "username")
	assert.NotContains(t, string(sharedData), `"proxy"`)
	assert.NotContains(t, string(sharedData), "secret-token")
	assert.NotContains(t, string(sharedData), `"token"`)

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), "cloudCredentialsByProvider")
	assert.Contains(t, string(localData), `"git"`)
	assert.Contains(t, string(localData), `"proxy"`)
	assert.Contains(t, string(localData), "repoCacheDir")
	assert.Contains(t, string(localData), "secret-token")
	assert.NotContains(t, string(localData), "repo_url")
	assert.NotContains(t, string(localData), "branch")
	assert.NotContains(t, string(localData), "username")
	assert.NotContains(t, string(localData), "skillsStorageDir")
}

func TestLoadMigratesProxyOutOfSharedConfig(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	shared := `{
	  "defaultCategory": "Default",
	  "logLevel": "info",
	  "repoScanMaxDepth": 5,
	  "agents": [],
	  "proxy": {
	    "mode": "manual",
	    "url": "http://127.0.0.1:7890"
	  }
	}`
	local := `{
	  "skillsStorageDir": "` + filepath.ToSlash(filepath.Join(dir, "skills")) + `",
	  "agents": []
	}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(shared), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(local), 0644))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, config.ProxyModeManual, loaded.Proxy.Mode)
	assert.Equal(t, "http://127.0.0.1:7890", loaded.Proxy.URL)

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(sharedData), `"proxy"`)
	assert.NotContains(t, string(sharedData), "127.0.0.1:7890")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"proxy"`)
	assert.Contains(t, string(localData), `"mode": "manual"`)
	assert.Contains(t, string(localData), `"url": "http://127.0.0.1:7890"`)
}

func TestLoadFallsBackToBuiltinAgentsWhenSharedAgentsMissing(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
	  "defaultCategory": "Default",
	  "logLevel": "info",
	  "repoScanMaxDepth": 5,
	  "agents": []
	}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(`{
	  "skillsStorageDir": "`+filepath.ToSlash(filepath.Join(dir, "skills"))+`",
	  "agents": []
	}`), 0o644))

	loaded, err := svc.Load()
	require.NoError(t, err)

	defaultCfg := config.DefaultConfig(dir)
	require.Len(t, loaded.Agents, len(defaultCfg.Agents))
	for i, agent := range defaultCfg.Agents {
		assert.Equal(t, agent.Name, loaded.Agents[i].Name)
		assert.Equal(t, agent.Enabled, loaded.Agents[i].Enabled)
		assert.Equal(t, agent.Custom, loaded.Agents[i].Custom)
		assert.Equal(t, agent.ScanDirs, loaded.Agents[i].ScanDirs)
		assert.Equal(t, agent.PushDir, loaded.Agents[i].PushDir)
	}
}

func TestMigrationFromLegacyConfig(t *testing.T) {
	dir := t.TempDir()
	// Write a legacy config.json that includes skillsStorageDir inline
	legacy := `{"skillsStorageDir":"` + filepath.ToSlash(filepath.Join(dir, "skills")) + `","defaultCategory":"Legacy"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(legacy), 0644))

	svc := config.NewService(dir)
	cfg, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "Legacy", cfg.DefaultCategory)

	// After migration config_local.json must exist
	_, err = os.Stat(filepath.Join(dir, "config_local.json"))
	assert.NoError(t, err, "migration should create config_local.json")

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), "repoCacheDir")
	assert.NotContains(t, string(localData), "skillsStorageDir")

	// config.json must no longer contain skillsStorageDir
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "skillsStorageDir")
}

func TestSaveAndLoadTencentEndpointWithBucketHostPreservesBucketField(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.Cloud.Provider = "tencent"
	cfg.Cloud.Enabled = true
	cfg.Cloud.BucketName = "shinerio-1258556983"
	cfg.Cloud.RemotePath = "nightly"
	cfg.Cloud.Credentials = map[string]string{
		"secret_id":  "tx-id",
		"secret_key": "tx-key",
		"endpoint":   "shinerio-1258556983.cos.ap-guangzhou.myqcloud.com",
	}

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, "shinerio-1258556983", loaded.Cloud.BucketName)
	assert.Equal(t, "shinerio-1258556983.cos.ap-guangzhou.myqcloud.com", loaded.Cloud.Credentials["endpoint"])
	assert.Equal(t, "nightly/skillflow/", loaded.Cloud.RemotePath)

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), `"bucketName": "shinerio-1258556983"`)
	assert.Contains(t, string(sharedData), `"endpoint": "shinerio-1258556983.cos.ap-guangzhou.myqcloud.com"`)
}

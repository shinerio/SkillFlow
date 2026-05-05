package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/shinerio/skillflow/core/platform/nativeapi"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeAPIReadOnlyMethods(t *testing.T) {
	app := newNativeAPITestApp(t)

	handler, ok := daemonServiceHandlers(app)["native.api"]
	require.True(t, ok, "native api daemon handler should be registered")

	resp := invokeNativeAPITestMethod(t, handler, "settings.get")
	require.True(t, resp.OK)
	var cfg config.AppConfig
	require.NoError(t, json.Unmarshal(resp.Result, &cfg))
	assert.Equal(t, defaultCategoryName, cfg.DefaultCategory)
	assert.Contains(t, nativeAPITestAgentNames(cfg.Agents), "codex")

	resp = invokeNativeAPITestMethod(t, handler, "skills.list")
	require.True(t, resp.OK)
	var skills []InstalledSkillEntry
	require.NoError(t, json.Unmarshal(resp.Result, &skills))
	assert.Empty(t, skills)

	resp = invokeNativeAPITestMethod(t, handler, "skills.categories.list")
	require.True(t, resp.OK)
	var categories []string
	require.NoError(t, json.Unmarshal(resp.Result, &categories))
	assert.Contains(t, categories, defaultCategoryName)

	resp = invokeNativeAPITestMethod(t, handler, "agents.listEnabled")
	require.True(t, resp.OK)
	var agents []config.AgentConfig
	require.NoError(t, json.Unmarshal(resp.Result, &agents))
	enabledNames := nativeAPITestAgentNames(agents)
	assert.Contains(t, enabledNames, "codex")
	assert.NotContains(t, enabledNames, "disabled")

	resp = invokeNativeAPITestMethod(t, handler, "backup.providers.list")
	require.True(t, resp.OK)
	var providers []map[string]any
	require.NoError(t, json.Unmarshal(resp.Result, &providers))
	names := make([]string, 0, len(providers))
	for _, provider := range providers {
		if name, ok := provider["name"].(string); ok {
			names = append(names, name)
		}
	}
	assert.Contains(t, names, "git")
}

func nativeAPITestAgentNames(agents []config.AgentConfig) []string {
	names := make([]string, 0, len(agents))
	for _, agent := range agents {
		names = append(names, agent.Name)
	}
	return names
}

func invokeNativeAPITestMethod(t *testing.T, handler func(context.Context, json.RawMessage) (any, error), method string) nativeapi.Response {
	t.Helper()

	payload, err := json.Marshal(nativeapi.Request{
		Version:   "2026-04-25",
		Method:    method,
		Params:    json.RawMessage(`null`),
		RequestID: "test-" + method,
	})
	require.NoError(t, err)

	result, err := handler(context.Background(), payload)
	require.NoError(t, err)
	resp, ok := result.(nativeapi.Response)
	require.True(t, ok)
	require.Nil(t, resp.Error)
	return resp
}

func newNativeAPITestApp(t *testing.T) *App {
	t.Helper()

	dataDir := t.TempDir()
	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{
			Name:     "codex",
			ScanDirs: []string{filepath.Join(dataDir, "codex-scan")},
			PushDir:  filepath.Join(dataDir, "codex-push"),
			Enabled:  true,
		},
		{
			Name:     "disabled",
			ScanDirs: []string{filepath.Join(dataDir, "disabled-scan")},
			PushDir:  filepath.Join(dataDir, "disabled-push"),
			Enabled:  false,
		},
	}
	require.NoError(t, svc.Save(cfg))

	prevProviders := registeredCloudProviderFactories
	t.Cleanup(func() {
		registeredCloudProviderFactories = prevProviders
	})
	registerProviders()

	app := NewApp()
	app.config = svc
	app.cacheDir = filepath.Join(dataDir, "cache")
	app.storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(appdata.SkillsDir(dataDir)))
	require.NoError(t, app.storage.CreateCategory(defaultCategoryName))
	return app
}

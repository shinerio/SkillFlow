package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestNativeAPISettingsLifecycleMethods(t *testing.T) {
	app := newNativeAPITestApp(t)
	handler, ok := daemonServiceHandlers(app)["native.api"]
	require.True(t, ok, "native api daemon handler should be registered")

	cfg, err := app.GetConfig()
	require.NoError(t, err)
	cfg.LogLevel = config.LogLevelDebug
	cfg.Proxy.Mode = config.ProxyModeNone
	resp := invokeNativeAPITestMethodWithParams(t, handler, "settings.save", cfg)
	require.True(t, resp.OK)

	saved, err := app.config.Load()
	require.NoError(t, err)
	assert.Equal(t, config.LogLevelDebug, saved.LogLevel)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()
	resp = invokeNativeAPITestMethodWithParams(t, handler, "proxy.test", nativeProxyTestParams{
		TargetURL: target.URL,
		Proxy:     config.ProxyConfig{Mode: config.ProxyModeNone},
	})
	require.True(t, resp.OK)
	var proxyResult ProxyConnectionTestResult
	require.NoError(t, json.Unmarshal(resp.Result, &proxyResult))
	assert.True(t, proxyResult.Success)
	assert.Equal(t, http.StatusNoContent, proxyResult.StatusCode)

	release := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v999.0.0","html_url":"https://example.com/release","body":"Release notes","assets":[]}`))
	}))
	defer release.Close()
	prevNewRequestWithContextFn := newRequestWithContextFn
	newRequestWithContextFn = func(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, method, release.URL, body)
	}
	t.Cleanup(func() {
		newRequestWithContextFn = prevNewRequestWithContextFn
	})
	resp = invokeNativeAPITestMethod(t, handler, "app.update.check")
	require.True(t, resp.OK)
	var update AppUpdateInfo
	require.NoError(t, json.Unmarshal(resp.Result, &update))
	assert.True(t, update.HasUpdate)
	assert.Equal(t, "v999.0.0", update.LatestVersion)
	assert.Equal(t, "https://example.com/release", update.ReleaseURL)

	openedPath := ""
	app.openExternalPathFn = func(target string) error {
		openedPath = target
		return nil
	}
	resp = invokeNativeAPITestMethod(t, handler, "logs.openDir")
	require.True(t, resp.OK)
	assert.Equal(t, app.logDir(), openedPath)

	controller := &fakeLaunchAtLoginController{}
	app.autostartFactory = func() (launchAtLoginController, error) {
		return controller, nil
	}
	resp = invokeNativeAPITestMethodWithParams(t, handler, "autostart.set", nativeAutostartSetParams{Enabled: true})
	require.True(t, resp.OK)
	assert.True(t, controller.enabled)
}

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

func TestNativeAPISkillsMutationMethods(t *testing.T) {
	app := newNativeAPITestApp(t)
	handler, ok := daemonServiceHandlers(app)["native.api"]
	require.True(t, ok, "native api daemon handler should be registered")

	// Ensure agent push directories exist for push tests.
	for _, agent := range []string{"codex-push", "disabled-push"} {
		require.NoError(t, os.MkdirAll(filepath.Join(app.config.DataDir(), agent), 0755))
	}

	// Import a local skill.
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nImported\n")
	resp := invokeNativeAPITestMethodWithParams(t, handler, "skills.importLocal", nativeSkillsImportParams{
		Dir:      sourceDir,
		Category: defaultCategoryName,
	})
	require.True(t, resp.OK)
	var imported map[string]any
	require.NoError(t, json.Unmarshal(resp.Result, &imported))
	assert.Equal(t, "demo-skill", imported["name"])

	// List skills — imported skill should appear.
	resp = invokeNativeAPITestMethod(t, handler, "skills.list")
	require.True(t, resp.OK)
	var skills []InstalledSkillEntry
	require.NoError(t, json.Unmarshal(resp.Result, &skills))
	require.Len(t, skills, 1)
	assert.Equal(t, "demo-skill", skills[0].Name)
	skillID := skills[0].ID

	// Move category — create a new category first, then move.
	resp = invokeNativeAPITestMethodWithParams(t, handler, "skills.moveCategory", nativeSkillsMoveCategoryParams{
		SkillID:  skillID,
		Category: "moved",
	})
	require.True(t, resp.OK)

	// Verify the category changed.
	resp = invokeNativeAPITestMethod(t, handler, "skills.list")
	require.True(t, resp.OK)
	require.NoError(t, json.Unmarshal(resp.Result, &skills))
	require.Len(t, skills, 1)
	assert.Equal(t, "moved", skills[0].Category)

	// Push to agents — codex is enabled, disabled is not.
	resp = invokeNativeAPITestMethodWithParams(t, handler, "skills.push", nativeSkillsPushParams{
		SkillIDs:   []string{skillID},
		AgentNames: []string{"codex"},
	})
	require.True(t, resp.OK)

	// Force push to agents.
	resp = invokeNativeAPITestMethodWithParams(t, handler, "skills.pushForce", nativeSkillsPushParams{
		SkillIDs:   []string{skillID},
		AgentNames: []string{"codex"},
	})
	require.True(t, resp.OK)

	// Check updates — non-GitHub skills are skipped, should succeed.
	resp = invokeNativeAPITestMethod(t, handler, "skills.updateCheck")
	require.True(t, resp.OK)

	// Delete the skill.
	resp = invokeNativeAPITestMethodWithParams(t, handler, "skills.delete", nativeSkillsDeleteParams{
		SkillID: skillID,
	})
	require.True(t, resp.OK)

	// Verify the skill is gone.
	resp = invokeNativeAPITestMethod(t, handler, "skills.list")
	require.True(t, resp.OK)
	require.NoError(t, json.Unmarshal(resp.Result, &skills))
	assert.Empty(t, skills)
}

func TestNativeAPISkillsBatchDelete(t *testing.T) {
	app := newNativeAPITestApp(t)
	handler, ok := daemonServiceHandlers(app)["native.api"]
	require.True(t, ok, "native api daemon handler should be registered")

	// Import two skills.
	var skillIDs []string
	for _, name := range []string{"skill-a", "skill-b"} {
		sourceDir := writeTestSkillDir(t, t.TempDir(), name, "# "+name+"\n")
		resp := invokeNativeAPITestMethodWithParams(t, handler, "skills.importLocal", nativeSkillsImportParams{
			Dir:      sourceDir,
			Category: defaultCategoryName,
		})
		require.True(t, resp.OK)
	}

	// List to get IDs.
	resp := invokeNativeAPITestMethod(t, handler, "skills.list")
	require.True(t, resp.OK)
	var skills []InstalledSkillEntry
	require.NoError(t, json.Unmarshal(resp.Result, &skills))
	require.Len(t, skills, 2)
	for _, s := range skills {
		skillIDs = append(skillIDs, s.ID)
	}

	// Batch delete.
	resp = invokeNativeAPITestMethodWithParams(t, handler, "skills.deleteBatch", nativeSkillsDeleteBatchParams{
		SkillIDs: skillIDs,
	})
	require.True(t, resp.OK)

	// Verify all gone.
	resp = invokeNativeAPITestMethod(t, handler, "skills.list")
	require.True(t, resp.OK)
	require.NoError(t, json.Unmarshal(resp.Result, &skills))
	assert.Empty(t, skills)
}

func TestNativeAPIAgentsMethods(t *testing.T) {
	app := newNativeAPITestApp(t)
	handler, ok := daemonServiceHandlers(app)["native.api"]
	require.True(t, ok, "native api daemon handler should be registered")

	// agents.list — should return enabled agents.
	resp := invokeNativeAPITestMethod(t, handler, "agents.list")
	require.True(t, resp.OK)
	var agents []config.AgentConfig
	require.NoError(t, json.Unmarshal(resp.Result, &agents))
	enabledNames := nativeAPITestAgentNames(agents)
	assert.Contains(t, enabledNames, "codex")
	assert.NotContains(t, enabledNames, "disabled")

	// agents.scanSkills — scan codex's scan dir (empty, should return empty list).
	resp = invokeNativeAPITestMethodWithParams(t, handler, "agents.scanSkills", nativeAgentNameParams{
		AgentName: "codex",
	})
	require.True(t, resp.OK)

	// agents.listSkills — list codex's pushed skills (empty initially).
	resp = invokeNativeAPITestMethodWithParams(t, handler, "agents.listSkills", nativeAgentNameParams{
		AgentName: "codex",
	})
	require.True(t, resp.OK)

	// agents.memoryPreview — preview codex memory.
	resp = invokeNativeAPITestMethodWithParams(t, handler, "agents.memoryPreview", nativeAgentNameParams{
		AgentName: "codex",
	})
	require.True(t, resp.OK)
	var preview AgentMemoryPreviewDTO
	require.NoError(t, json.Unmarshal(resp.Result, &preview))
	assert.Equal(t, "codex", preview.AgentName)
}

func nativeAPITestAgentNames(agents []config.AgentConfig) []string {
	names := make([]string, 0, len(agents))
	for _, agent := range agents {
		names = append(names, agent.Name)
	}
	return names
}

func invokeNativeAPITestMethodWithParams(t *testing.T, handler func(context.Context, json.RawMessage) (any, error), method string, params any) nativeapi.Response {
	t.Helper()

	if params == nil {
		return invokeNativeAPITestMethod(t, handler, method)
	}
	payload, err := json.Marshal(params)
	require.NoError(t, err)
	return invokeNativeAPITestMethodWithRawParams(t, handler, method, payload)
}

func invokeNativeAPITestMethodWithRawParams(t *testing.T, handler func(context.Context, json.RawMessage) (any, error), method string, params json.RawMessage) nativeapi.Response {
	t.Helper()

	payload, err := json.Marshal(nativeapi.Request{
		Version:   "2026-04-25",
		Method:    method,
		Params:    params,
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

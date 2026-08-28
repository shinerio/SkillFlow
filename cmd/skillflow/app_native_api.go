package main

import (
	"context"
	"encoding/json"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/nativeapi"
)

type nativeProxyTestParams struct {
	TargetURL string             `json:"targetURL"`
	Proxy     config.ProxyConfig `json:"proxy"`
}

type nativeAutostartSetParams struct {
	Enabled bool `json:"enabled"`
}

func nativeAPIRouter(app *App) *nativeapi.Router {
	router := nativeapi.NewRouter()
	registerNativeReadOnlyAPI(router, app)
	return router
}

func registerNativeReadOnlyAPI(router *nativeapi.Router, app *App) {
	router.Register("settings.get", func(context.Context, json.RawMessage) (any, error) {
		return app.GetConfig()
	})
	router.Register("skills.list", func(context.Context, json.RawMessage) (any, error) {
		return app.ListSkills()
	})
	router.Register("skills.categories.list", func(context.Context, json.RawMessage) (any, error) {
		return app.ListCategories()
	})
	router.Register("agents.listEnabled", func(context.Context, json.RawMessage) (any, error) {
		return app.GetEnabledAgents()
	})
	router.Register("backup.providers.list", func(context.Context, json.RawMessage) (any, error) {
		return app.ListCloudProviders(), nil
	})
	router.Register("settings.save", func(_ context.Context, params json.RawMessage) (any, error) {
		cfg, err := decodeNativeAPIParams[config.AppConfig](params)
		if err != nil {
			return nil, err
		}
		return nil, app.SaveConfig(cfg)
	})
	router.Register("proxy.test", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeProxyTestParams](params)
		if err != nil {
			return nil, err
		}
		return app.TestProxyConnection(req.TargetURL, req.Proxy)
	})
	router.Register("logs.openDir", func(context.Context, json.RawMessage) (any, error) {
		return nil, app.OpenLogDir()
	})
	router.Register("app.update.check", func(context.Context, json.RawMessage) (any, error) {
		return app.CheckAppUpdate()
	})
	router.Register("autostart.set", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAutostartSetParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.syncLaunchAtLogin(req.Enabled)
	})
}

func decodeNativeAPIParams[T any](params json.RawMessage) (T, error) {
	var value T
	if len(params) == 0 || string(params) == "null" {
		return value, nil
	}
	if err := json.Unmarshal(params, &value); err != nil {
		return value, err
	}
	return value, nil
}

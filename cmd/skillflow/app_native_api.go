package main

import (
	"context"
	"encoding/json"

	"github.com/shinerio/skillflow/core/platform/nativeapi"
)

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
}

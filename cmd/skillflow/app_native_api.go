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

type nativeSkillsImportParams struct {
	Dir      string `json:"dir"`
	Category string `json:"category"`
}

type nativeSkillsDeleteParams struct {
	SkillID string `json:"skillID"`
}

type nativeSkillsDeleteBatchParams struct {
	SkillIDs []string `json:"skillIDs"`
}

type nativeSkillsMoveCategoryParams struct {
	SkillID  string `json:"skillID"`
	Category string `json:"category"`
}

type nativeSkillsPushParams struct {
	SkillIDs   []string `json:"skillIDs"`
	AgentNames []string `json:"agentNames"`
}

type nativeAgentNameParams struct {
	AgentName string `json:"agentName"`
}

type nativeAgentDeleteSkillParams struct {
	AgentName string `json:"agentName"`
	SkillPath string `json:"skillPath"`
}

type nativeAgentPullParams struct {
	AgentName  string   `json:"agentName"`
	SkillPaths []string `json:"skillPaths"`
	Category   string   `json:"category"`
}

func nativeAPIRouter(app *App) *nativeapi.Router {
	router := nativeapi.NewRouter()
	registerNativeReadOnlyAPI(router, app)
	registerNativeSkillsAPI(router, app)
	registerNativeAgentsAPI(router, app)
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

func registerNativeSkillsAPI(router *nativeapi.Router, app *App) {
	router.Register("skills.importLocal", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsImportParams](params)
		if err != nil {
			return nil, err
		}
		return app.ImportLocal(req.Dir, req.Category)
	})
	router.Register("skills.delete", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsDeleteParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.DeleteSkill(req.SkillID)
	})
	router.Register("skills.deleteBatch", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsDeleteBatchParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.DeleteSkills(req.SkillIDs)
	})
	router.Register("skills.moveCategory", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsMoveCategoryParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.MoveSkillCategory(req.SkillID, req.Category)
	})
	router.Register("skills.updateCheck", func(context.Context, json.RawMessage) (any, error) {
		return nil, app.CheckUpdates()
	})
	router.Register("skills.updateOne", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsDeleteParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.UpdateSkill(req.SkillID)
	})
	router.Register("skills.push", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsPushParams](params)
		if err != nil {
			return nil, err
		}
		return app.PushToAgents(req.SkillIDs, req.AgentNames)
	})
	router.Register("skills.pushForce", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeSkillsPushParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.PushToAgentsForce(req.SkillIDs, req.AgentNames)
	})
}

func registerNativeAgentsAPI(router *nativeapi.Router, app *App) {
	router.Register("agents.list", func(context.Context, json.RawMessage) (any, error) {
		return app.GetEnabledAgents()
	})
	router.Register("agents.scanSkills", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAgentNameParams](params)
		if err != nil {
			return nil, err
		}
		return app.ScanAgentSkills(req.AgentName)
	})
	router.Register("agents.listSkills", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAgentNameParams](params)
		if err != nil {
			return nil, err
		}
		return app.ListAgentSkills(req.AgentName)
	})
	router.Register("agents.deleteSkill", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAgentDeleteSkillParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.DeleteAgentSkill(req.AgentName, req.SkillPath)
	})
	router.Register("agents.pull", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAgentPullParams](params)
		if err != nil {
			return nil, err
		}
		return app.PullFromAgent(req.AgentName, req.SkillPaths, req.Category)
	})
	router.Register("agents.pullForce", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAgentPullParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.PullFromAgentForce(req.AgentName, req.SkillPaths, req.Category)
	})
	router.Register("agents.memoryPreview", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeAgentNameParams](params)
		if err != nil {
			return nil, err
		}
		return app.GetAgentMemoryPreview(req.AgentName)
	})
}

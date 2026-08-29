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

type nativeStarredRepoURLParams struct {
	RepoURL string `json:"repoURL"`
}

type nativeStarredRepoCredentialsParams struct {
	RepoURL  string `json:"repoURL"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type nativeStarredImportParams struct {
	SkillPaths []string `json:"skillPaths"`
	RepoURL    string   `json:"repoURL"`
	Category   string   `json:"category"`
}

type nativePromptNameParams struct {
	Name string `json:"name"`
}

type nativePromptMoveCategoryParams struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type nativePromptCategoryNameParams struct {
	Name string `json:"name"`
}

type nativePromptRenameCategoryParams struct {
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

type nativePromptCreateParams struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	Content          string   `json:"content"`
	ImageURLs        []string `json:"imageURLs"`
	WebLinksMarkdown string   `json:"webLinksMarkdown"`
}

type nativePromptUpdateParams struct {
	OriginalName     string   `json:"originalName"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	Content          string   `json:"content"`
	ImageURLs        []string `json:"imageURLs"`
	WebLinksMarkdown string   `json:"webLinksMarkdown"`
}

type nativePromptCompleteImportParams struct {
	SessionID      string   `json:"sessionID"`
	OverwriteNames []string `json:"overwriteNames"`
}

type nativePromptCancelImportParams struct {
	SessionID string `json:"sessionID"`
}

type nativePromptExportByNamesParams struct {
	Names []string `json:"names"`
}

type nativeMemoryContentParams struct {
	Content string `json:"content"`
}

type nativeMemoryModuleNameParams struct {
	Name string `json:"name"`
}

type nativeMemoryModuleContentParams struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type nativeMemoryModuleEnabledParams struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type nativeMemoryAgentTypeParams struct {
	AgentType string `json:"agentType"`
}

type nativeMemorySavePushConfigParams struct {
	AgentType string `json:"agentType"`
	Mode      string `json:"mode"`
	AutoPush  bool   `json:"autoPush"`
}

type nativeMemoryModulePushTargetsParams struct {
	ModuleName  string   `json:"moduleName"`
	PushTargets []string `json:"pushTargets"`
}

type nativeMemoryPushSelectedParams struct {
	AgentTypes  []string `json:"agentTypes"`
	ModuleNames []string `json:"moduleNames"`
	Mode        string   `json:"mode"`
}

type nativeMemoryOpenEditorParams struct {
	MemoryType string `json:"memoryType"`
	ModuleName string `json:"moduleName"`
}

func nativeAPIRouter(app *App) *nativeapi.Router {
	router := nativeapi.NewRouter()
	registerNativeReadOnlyAPI(router, app)
	registerNativeSkillsAPI(router, app)
	registerNativeAgentsAPI(router, app)
	registerNativeStarredAPI(router, app)
	registerNativePromptsAPI(router, app)
	registerNativeMemoryAPI(router, app)
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

func registerNativeStarredAPI(router *nativeapi.Router, app *App) {
	router.Register("starred.listRepos", func(context.Context, json.RawMessage) (any, error) {
		return app.ListStarredRepos()
	})
	router.Register("starred.addRepo", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeStarredRepoURLParams](params)
		if err != nil {
			return nil, err
		}
		return app.AddStarredRepo(req.RepoURL)
	})
	router.Register("starred.addRepoWithCredentials", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeStarredRepoCredentialsParams](params)
		if err != nil {
			return nil, err
		}
		return app.AddStarredRepoWithCredentials(req.RepoURL, req.Username, req.Password)
	})
	router.Register("starred.removeRepo", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeStarredRepoURLParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.RemoveStarredRepo(req.RepoURL)
	})
	router.Register("starred.updateRepo", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeStarredRepoURLParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.UpdateStarredRepo(req.RepoURL)
	})
	router.Register("starred.updateAll", func(context.Context, json.RawMessage) (any, error) {
		return nil, app.UpdateAllStarredRepos()
	})
	router.Register("starred.listAllSkills", func(context.Context, json.RawMessage) (any, error) {
		return app.ListAllStarSkills()
	})
	router.Register("starred.listRepoSkills", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeStarredRepoURLParams](params)
		if err != nil {
			return nil, err
		}
		return app.ListRepoStarSkills(req.RepoURL)
	})
	router.Register("starred.importSkills", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeStarredImportParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.ImportStarSkills(req.SkillPaths, req.RepoURL, req.Category)
	})
}

func registerNativePromptsAPI(router *nativeapi.Router, app *App) {
	router.Register("prompts.list", func(context.Context, json.RawMessage) (any, error) {
		return app.ListPrompts()
	})
	router.Register("prompts.categories.list", func(context.Context, json.RawMessage) (any, error) {
		return app.ListPromptCategories()
	})
	router.Register("prompts.create", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptCreateParams](params)
		if err != nil {
			return nil, err
		}
		return app.CreatePrompt(req.Name, req.Description, req.Category, req.Content, req.ImageURLs, req.WebLinksMarkdown)
	})
	router.Register("prompts.update", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptUpdateParams](params)
		if err != nil {
			return nil, err
		}
		return app.UpdatePrompt(req.OriginalName, req.Name, req.Description, req.Category, req.Content, req.ImageURLs, req.WebLinksMarkdown)
	})
	router.Register("prompts.moveCategory", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptMoveCategoryParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.MovePromptCategory(req.Name, req.Category)
	})
	router.Register("prompts.delete", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptNameParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.DeletePrompt(req.Name)
	})
	router.Register("prompts.categories.create", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptCategoryNameParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.CreatePromptCategory(req.Name)
	})
	router.Register("prompts.categories.rename", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptRenameCategoryParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.RenamePromptCategory(req.OldName, req.NewName)
	})
	router.Register("prompts.categories.delete", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptCategoryNameParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.DeletePromptCategory(req.Name)
	})
	router.Register("prompts.import.prepare", func(context.Context, json.RawMessage) (any, error) {
		return app.PrepareImportPrompts()
	})
	router.Register("prompts.import.complete", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptCompleteImportParams](params)
		if err != nil {
			return nil, err
		}
		return app.CompleteImportPrompts(req.SessionID, req.OverwriteNames)
	})
	router.Register("prompts.import.cancel", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptCancelImportParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.CancelImportPrompts(req.SessionID)
	})
	router.Register("prompts.export", func(context.Context, json.RawMessage) (any, error) {
		return app.ExportPrompts()
	})
	router.Register("prompts.exportByNames", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativePromptExportByNamesParams](params)
		if err != nil {
			return nil, err
		}
		return app.ExportPromptsByNames(req.Names)
	})
	router.Register("prompts.rootDir", func(context.Context, json.RawMessage) (any, error) {
		return app.PromptRootDir()
	})
}

func registerNativeMemoryAPI(router *nativeapi.Router, app *App) {
	router.Register("memory.main.get", func(context.Context, json.RawMessage) (any, error) {
		return app.GetMainMemory()
	})
	router.Register("memory.main.save", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryContentParams](params)
		if err != nil {
			return nil, err
		}
		return app.SaveMainMemory(req.Content)
	})
	router.Register("memory.modules.list", func(context.Context, json.RawMessage) (any, error) {
		return app.ListModuleMemories()
	})
	router.Register("memory.modules.get", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModuleNameParams](params)
		if err != nil {
			return nil, err
		}
		return app.GetModuleMemory(req.Name)
	})
	router.Register("memory.modules.create", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModuleContentParams](params)
		if err != nil {
			return nil, err
		}
		return app.CreateModuleMemory(req.Name, req.Content)
	})
	router.Register("memory.modules.save", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModuleContentParams](params)
		if err != nil {
			return nil, err
		}
		return app.SaveModuleMemory(req.Name, req.Content)
	})
	router.Register("memory.modules.delete", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModuleNameParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.DeleteModuleMemory(req.Name)
	})
	router.Register("memory.modules.setEnabled", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModuleEnabledParams](params)
		if err != nil {
			return nil, err
		}
		return app.SetModuleMemoryEnabled(req.Name, req.Enabled)
	})
	router.Register("memory.pushConfig.get", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryAgentTypeParams](params)
		if err != nil {
			return nil, err
		}
		return app.GetMemoryPushConfig(req.AgentType)
	})
	router.Register("memory.pushConfig.save", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemorySavePushConfigParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.SaveMemoryPushConfig(req.AgentType, req.Mode, req.AutoPush)
	})
	router.Register("memory.pushConfig.getAll", func(context.Context, json.RawMessage) (any, error) {
		return app.GetAllMemoryPushConfigs()
	})
	router.Register("memory.modulePushTargets.get", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModulePushTargetsParams](params)
		if err != nil {
			return nil, err
		}
		return app.GetModulePushTargets(req.ModuleName)
	})
	router.Register("memory.modulePushTargets.save", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryModulePushTargetsParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.SaveModulePushTargets(req.ModuleName, req.PushTargets)
	})
	router.Register("memory.modulePushTargets.getAll", func(context.Context, json.RawMessage) (any, error) {
		return app.GetAllModulePushTargets()
	})
	router.Register("memory.push", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryAgentTypeParams](params)
		if err != nil {
			return nil, err
		}
		return app.PushMemoryToAgent(req.AgentType)
	})
	router.Register("memory.pushAll", func(context.Context, json.RawMessage) (any, error) {
		return app.PushAllMemory()
	})
	router.Register("memory.pushSelected", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryPushSelectedParams](params)
		if err != nil {
			return nil, err
		}
		return app.PushSelectedMemory(req.AgentTypes, req.ModuleNames, req.Mode)
	})
	router.Register("memory.pushStatus.get", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryAgentTypeParams](params)
		if err != nil {
			return nil, err
		}
		return app.GetMemoryPushStatus(req.AgentType)
	})
	router.Register("memory.pushStatus.getAll", func(context.Context, json.RawMessage) (any, error) {
		return app.GetAllMemoryPushStatuses()
	})
	router.Register("memory.openInEditor", func(_ context.Context, params json.RawMessage) (any, error) {
		req, err := decodeNativeAPIParams[nativeMemoryOpenEditorParams](params)
		if err != nil {
			return nil, err
		}
		return nil, app.OpenMemoryInEditor(req.MemoryType, req.ModuleName)
	})
}

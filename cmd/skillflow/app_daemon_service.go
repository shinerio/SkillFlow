package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	daemonruntime "github.com/shinerio/skillflow/core/platform/daemon"
	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/shinerio/skillflow/core/platform/nativeapi"
)

type BackendClientConfig struct {
	BaseURL string `json:"baseUrl"`
	Token   string `json:"token"`
}

var daemonStreamEventsFn = daemonruntime.StreamEvents

var daemonAppMethodAllowlist = map[string]struct{}{
	"GetGitConflictPending":         {},
	"ResolveGitConflict":            {},
	"SetSkippedUpdateVersion":       {},
	"BackupNow":                     {},
	"GetConfig":                     {},
	"GetLastBackupChanges":          {},
	"GetLastBackupCompletedAt":      {},
	"RestoreFromCloud":              {},
	"ListSkills":                    {},
	"ListCategories":                {},
	"MoveSkillCategory":             {},
	"DeleteSkill":                   {},
	"DeleteSkills":                  {},
	"ImportLocal":                   {},
	"UpdateSkill":                   {},
	"CheckUpdates":                  {},
	"GetSkillMeta":                  {},
	"GetSkillMetaByPath":            {},
	"ReadSkillFileContent":          {},
	"SaveConfig":                    {},
	"CheckMissingAgentPushDirs":     {},
	"PushToAgents":                  {},
	"PushToAgentsForce":             {},
	"CreateCategory":                {},
	"RenameCategory":                {},
	"DeleteCategory":                {},
	"GetEnabledAgents":              {},
	"GetAgentMemoryPreview":         {},
	"ListAgentSkills":               {},
	"DeleteAgentSkill":              {},
	"ScanAgentSkills":               {},
	"PullFromAgent":                 {},
	"PullFromAgentForce":            {},
	"CreateModuleMemory":            {},
	"DeleteModuleMemory":            {},
	"GetAllMemoryPushConfigs":       {},
	"GetAllMemoryPushStatuses":      {},
	"GetMainMemory":                 {},
	"ListModuleMemories":            {},
	"PushSelectedMemory":            {},
	"SaveMainMemory":                {},
	"SaveMemoryPushConfig":          {},
	"SaveModuleMemory":              {},
	"SetModuleMemoryEnabled":        {},
	"CancelImportPrompts":           {},
	"CompleteImportPrompts":         {},
	"CreatePrompt":                  {},
	"DeletePrompt":                  {},
	"ListPromptCategories":          {},
	"ListPrompts":                   {},
	"MovePromptCategory":            {},
	"UpdatePrompt":                  {},
	"CreatePromptCategory":          {},
	"RenamePromptCategory":          {},
	"DeletePromptCategory":          {},
	"ListCloudProviders":            {},
	"TestProxyConnection":           {},
	"CheckAppUpdateAndNotify":       {},
	"ListStarredRepos":              {},
	"AddStarredRepo":                {},
	"AddStarredRepoWithCredentials": {},
	"RemoveStarredRepo":             {},
	"UpdateStarredRepo":             {},
	"UpdateAllStarredRepos":         {},
	"ListAllStarSkills":             {},
	"ListRepoStarSkills":            {},
	"ImportStarSkills":              {},
	"PushStarSkillsToAgents":        {},
	"PushStarSkillsToAgentsForce":   {},
}

func (a *App) GetBackendClientConfig() (*BackendClientConfig, error) {
	endpoint, err := readControlEndpoint(daemonServicePathFn())
	if err != nil {
		return nil, err
	}
	return &BackendClientConfig{
		BaseURL: "http://" + endpoint.Address,
		Token:   endpoint.Token,
	}, nil
}

func (a *App) startDaemonEventForwarder(ctx context.Context) {
	if ctx == nil || !shouldProxyAppMethodsToDaemon() {
		return
	}

	go func() {
		for {
			err := daemonStreamEventsFn(daemonServicePathFn(), ctx, func(evt eventbus.Event) {
				emitRuntimeEvent(ctx, evt)
			})
			if ctx.Err() != nil {
				return
			}
			if err != nil && !isControlEndpointMissing(err) {
				a.logErrorf("daemon event stream failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}()
}

func daemonServiceHandlers(app *App) map[string]daemonruntime.ServiceHandler {
	handlers := make(map[string]daemonruntime.ServiceHandler, len(daemonAppMethodAllowlist))
	for methodName := range daemonAppMethodAllowlist {
		name := methodName
		handlers[name] = func(ctx context.Context, params json.RawMessage) (any, error) {
			return invokeDaemonAppMethod(app, name, params)
		}
	}
	handlers["native.api"] = nativeAPIDaemonHandler(app)
	return handlers
}

func nativeAPIDaemonHandler(app *App) daemonruntime.ServiceHandler {
	router := nativeAPIRouter(app)
	return func(ctx context.Context, params json.RawMessage) (any, error) {
		var req nativeapi.Request
		if err := json.Unmarshal(params, &req); err != nil {
			return nativeapi.Response{
				OK:     false,
				Result: json.RawMessage(`null`),
				Error: &nativeapi.Error{
					Code:       nativeapi.CodeInternalError,
					Message:    err.Error(),
					MessageKey: nativeapi.MessageKeyInternalError,
				},
			}, nil
		}
		return router.Handle(ctx, req), nil
	}
}

func invokeDaemonAppMethod(app *App, methodName string, params json.RawMessage) (any, error) {
	method := reflect.ValueOf(app).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found")
	}

	args, err := decodeDaemonMethodArgs(method.Type(), params)
	if err != nil {
		return nil, err
	}
	results := method.Call(args)
	return unwrapDaemonMethodResults(results)
}

func decodeDaemonMethodArgs(methodType reflect.Type, params json.RawMessage) ([]reflect.Value, error) {
	argCount := methodType.NumIn()
	if argCount == 0 {
		return nil, nil
	}

	raw := bytes.TrimSpace(params)
	if argCount == 1 && (len(raw) == 0 || raw[0] != '[') {
		value, err := decodeDaemonMethodArg(methodType.In(0), raw)
		if err != nil {
			return nil, err
		}
		return []reflect.Value{value}, nil
	}

	var rawArgs []json.RawMessage
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		rawArgs = []json.RawMessage{}
	} else if err := json.Unmarshal(raw, &rawArgs); err != nil {
		return nil, err
	}
	if len(rawArgs) != argCount {
		return nil, fmt.Errorf("invalid argument count")
	}

	values := make([]reflect.Value, 0, argCount)
	for i := 0; i < argCount; i++ {
		value, err := decodeDaemonMethodArg(methodType.In(i), rawArgs[i])
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func decodeDaemonMethodArg(argType reflect.Type, raw json.RawMessage) (reflect.Value, error) {
	value := reflect.New(argType)
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return value.Elem(), nil
	}
	if err := json.Unmarshal(trimmed, value.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return value.Elem(), nil
}

func unwrapDaemonMethodResults(results []reflect.Value) (any, error) {
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		if err, ok := results[0].Interface().(error); ok {
			return nil, err
		}
		return results[0].Interface(), nil
	case 2:
		if err, ok := results[1].Interface().(error); ok && err != nil {
			return nil, err
		}
		return results[0].Interface(), nil
	default:
		return nil, fmt.Errorf("unsupported method signature")
	}
}

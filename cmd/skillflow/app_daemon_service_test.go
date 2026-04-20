package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (a *App) DaemonTestEcho(name string) (string, error) {
	return "hello " + name, nil
}

func (a *App) DaemonTestJoin(name string, count int) (map[string]any, error) {
	return map[string]any{
		"name":  name,
		"count": count,
	}, nil
}

func TestGetBackendClientConfigReadsDaemonEndpoint(t *testing.T) {
	prevDaemonServicePathFn := daemonServicePathFn
	t.Cleanup(func() {
		daemonServicePathFn = prevDaemonServicePathFn
	})

	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-service.json")
	daemonServicePathFn = func() string {
		return statePath
	}
	require.NoError(t, writeControlEndpoint(statePath, controlEndpoint{
		Address: "127.0.0.1:17890",
		Token:   "daemon-token",
		PID:     42,
	}))

	cfg, err := NewApp().GetBackendClientConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "http://127.0.0.1:17890", cfg.BaseURL)
	assert.Equal(t, "daemon-token", cfg.Token)
}

func TestInvokeDaemonAppMethodDecodesSingleArgument(t *testing.T) {
	result, err := invokeDaemonAppMethod(NewApp(), "DaemonTestEcho", json.RawMessage(`"codex"`))
	require.NoError(t, err)
	assert.Equal(t, "hello codex", result)
}

func TestInvokeDaemonAppMethodDecodesMultipleArguments(t *testing.T) {
	result, err := invokeDaemonAppMethod(NewApp(), "DaemonTestJoin", json.RawMessage(`["codex",2]`))
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"name":  "codex",
		"count": 2,
	}, result)
}

func TestDaemonServiceHandlersOnlyExposeAllowlistedMethods(t *testing.T) {
	handlers := daemonServiceHandlers(NewApp())

	require.Contains(t, handlers, "GetConfig")
	require.Contains(t, handlers, "ListSkills")
	require.NotContains(t, handlers, "GetBackendClientConfig")
	require.NotContains(t, handlers, "DaemonTestEcho")
}

func TestDaemonServiceHandlersExposeAgentSkillManagementMethods(t *testing.T) {
	handlers := daemonServiceHandlers(NewApp())

	for _, methodName := range []string{
		"ListAllAgentSkills",
		"ListManagedAgentSkills",
		"CreateAgentSkillGroup",
		"RenameAgentSkillGroup",
		"DeleteAgentSkillGroup",
		"AssignAgentSkillGroup",
		"ClearAgentSkillGroup",
		"SetAgentSkillEnabled",
		"SetAgentSkillGroupEnabled",
	} {
		require.Contains(t, handlers, methodName)
	}
}

func TestStartDaemonEventForwarderStartsStreamInUIRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonStreamEventsFn := daemonStreamEventsFn
	prevDaemonServicePathFn := daemonServicePathFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonStreamEventsFn = prevDaemonStreamEventsFn
		daemonServicePathFn = prevDaemonServicePathFn
	})

	activeProcessRole = processRoleUI
	daemonServicePathFn = func() string { return "/tmp/daemon-service.json" }

	called := make(chan string, 1)
	daemonStreamEventsFn = func(statePath string, ctx context.Context, handle func(eventbus.Event)) error {
		called <- statePath
		<-ctx.Done()
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	NewApp().startDaemonEventForwarder(ctx)

	select {
	case got := <-called:
		assert.Equal(t, "/tmp/daemon-service.json", got)
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for daemon event forwarder to start")
	}
}

func TestStartDaemonEventForwarderSkipsNonUIRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonStreamEventsFn := daemonStreamEventsFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonStreamEventsFn = prevDaemonStreamEventsFn
	})

	activeProcessRole = processRoleDaemon

	started := make(chan struct{}, 1)
	daemonStreamEventsFn = func(statePath string, ctx context.Context, handle func(eventbus.Event)) error {
		started <- struct{}{}
		return nil
	}

	NewApp().startDaemonEventForwarder(context.Background())

	select {
	case <-started:
		t.Fatal("daemon event forwarder should not start outside UI role")
	case <-time.After(200 * time.Millisecond):
	}
}

package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	runUIProcessFn           = runUIProcess
	bootstrapHelperProcessFn = bootstrapHelperProcess
	runDaemonProcessFn       = runDaemonProcess
	runDaemonOnlyProcessFn   = runDaemonOnlyProcess
)

func main() {
	os.Exit(runEntry(os.Args))
}

func runEntry(args []string) int {
	role, filteredArgs := determineProcessRole(args)
	activeProcessRole = role
	if len(filteredArgs) > 0 {
		os.Args = filteredArgs
	}
	switch role {
	case processRoleDaemon:
		if err := runDaemonProcessFn(filteredArgs); err != nil {
			println("Error:", err.Error())
			return 1
		}
	case processRoleDaemonOnly:
		if err := runDaemonOnlyProcessFn(); err != nil {
			println("Error:", err.Error())
			return 1
		}
	case processRoleUI:
		if err := runUIProcessFn(); err != nil {
			println("Error:", err.Error())
			return 1
		}
	default:
		println("Error: unknown process role")
		return 1
	}
	return 0
}

func runUIProcess() error {
	app := NewApp()

	return wails.Run(buildUIOptions(app))
}

func runDaemonProcess(filteredArgs []string) error {
	if !helperBootstrapEnabled() {
		return runDaemonProcessWhenHelperDisabled(filteredArgs)
	}
	uiArgs := []string(nil)
	if len(filteredArgs) > 1 {
		uiArgs = filteredArgs[1:]
	}
	return bootstrapHelperProcessFn(uiArgs)
}

func runDaemonProcessWhenHelperDisabled(_ []string) error {
	return runUIProcessFn()
}

func buildUIOptions(app *App) *options.App {
	return &options.App{
		Title:             "SkillFlow",
		Width:             1360,
		Height:            860,
		MinWidth:          960,
		MinHeight:         680,
		HideWindowOnClose: uiProcessOwnsTrayLifecycle(),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	}
}

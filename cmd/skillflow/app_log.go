package main

import (
	"fmt"
	"path/filepath"

	"github.com/shinerio/skillflow/core/applog"
	"github.com/shinerio/skillflow/core/config"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) logDir() string {
	return filepath.Join(config.AppDataDir(), "logs")
}

func (a *App) initLogger(logLevel string) {
	lg, err := applog.New(a.logDir(), logLevel)
	if err != nil {
		if a.ctx != nil {
			runtime.LogErrorf(a.ctx, "logger init failed: %v", err)
		}
		return
	}
	a.sysLog = lg
	a.logInfof("logger initialized, level=%s dir=%s", lg.LevelString(), lg.Dir())
}

func (a *App) setLoggerLevel(level string) string {
	normalized := config.NormalizeLogLevel(level)
	if a.sysLog != nil {
		a.sysLog.SetLevelString(normalized)
	}
	return normalized
}

func (a *App) logDebugf(format string, args ...any) {
	a.logf(applog.LevelDebug, format, args...)
}

func (a *App) logInfof(format string, args ...any) {
	a.logf(applog.LevelInfo, format, args...)
}

func (a *App) logErrorf(format string, args ...any) {
	a.logf(applog.LevelError, format, args...)
}

func (a *App) logf(level applog.Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if a.sysLog != nil {
		a.sysLog.Logf(level, "%s", msg)
		if !a.sysLog.Enabled(level) {
			return
		}
	}
	if a.ctx == nil {
		return
	}
	switch level {
	case applog.LevelDebug:
		runtime.LogDebug(a.ctx, msg)
	case applog.LevelError:
		runtime.LogError(a.ctx, msg)
	default:
		runtime.LogInfo(a.ctx, msg)
	}
}

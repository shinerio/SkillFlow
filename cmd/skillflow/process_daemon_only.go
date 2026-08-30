package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	daemonipc "github.com/shinerio/skillflow/core/platform/ipc"
)

var daemonOnlySignalFn = signal.Notify

func runDaemonOnlyProcess() error {
	if err := daemonipc.PruneStaleState(daemonServicePathFn()); err != nil {
		return fmt.Errorf("daemon-only endpoint check failed: %w", err)
	}
	if _, err := readControlEndpoint(daemonServicePathFn()); err == nil {
		return nil
	}

	controller := newHelperController(nil)
	controller.logInfof("daemon-only process started, pid=%d", os.Getpid())
	if err := controller.initializeDaemonBackend(); err != nil {
		controller.logErrorf("daemon-only initialization failed: err=%v", err)
		return err
	}
	defer controller.closeDaemonService()
	defer controller.logInfof("daemon-only process stopped, pid=%d", os.Getpid())

	signals := make(chan os.Signal, 1)
	daemonOnlySignalFn(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-signals
	controller.logInfof("daemon-only stop signal received, pid=%d", os.Getpid())
	return nil
}

//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// ensureSingleInstance uses a named mutex to guarantee only one instance runs.
// If another instance is already running, it brings that window to the foreground
// and exits the current process.
func ensureSingleInstance() {
	name, _ := syscall.UTF16PtrFromString("Local\\SkillFlowSingleInstance")
	handle, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		// Could not create mutex — allow startup to proceed.
		return
	}
	if err == syscall.ERROR_ALREADY_EXISTS {
		// Another instance is running: signal its tray window to show the main window.
		// We find the tray window by its known class name so it works even when the
		// main Wails window is hidden (not just minimized).
		className, _ := syscall.UTF16PtrFromString(trayWindowClass)
		hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0)
		if hwnd != 0 {
			procPostMessageW.Call(hwnd, trayShowWindowMsg, 0, 0)
		}
		os.Exit(0)
	}
}

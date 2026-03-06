//go:build windows

package main

import (
	"fmt"
	goruntime "runtime"
	"sync"
	"syscall"
	"unsafe"
)

const (
	trayWindowClass = "SkillFlowTrayWindowClass"
	trayIconID      = 1

	trayMenuCmdShow = 1001
	trayMenuCmdExit = 1002

	wmDestroy = 0x0002
	wmClose   = 0x0010
	wmNull    = 0x0000
	wmCommand = 0x0111
	wmRButton = 0x0205
	wmLButton = 0x0202
	wmLDouble = 0x0203
	wmApp     = 0x8000

	trayCallbackMessage  = wmApp + 1
	trayShowWindowMsg    = wmApp + 2 // sent by a second instance to request window focus

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	nimAdd    = 0x00000000
	nimDelete = 0x00000002

	mfString    = 0x00000000
	mfSeparator = 0x00000800

	tpmLeftAlign   = 0x0000
	tpmBottomAlign = 0x0020
	tpmRightButton = 0x0002

	idcArrow       = 32512
	idiApplication = 32512
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	shell32                 = syscall.NewLazyDLL("shell32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procLoadIconW           = user32.NewProc("LoadIconW")
	procLoadCursorW         = user32.NewProc("LoadCursorW")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procPostMessageW        = user32.NewProc("PostMessageW")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	procAppendMenuW         = user32.NewProc("AppendMenuW")
	procTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procShellNotifyIconW    = shell32.NewProc("Shell_NotifyIconW")
	procCreateMutexW        = kernel32.NewProc("CreateMutexW")
	procFindWindowW         = user32.NewProc("FindWindowW")
	trayWndProcCallback     = syscall.NewCallback(trayWndProc)
	windowsTrayStartOnce    sync.Once
	windowsTrayStartErr     error
	windowsTrayState        struct {
		mu   sync.RWMutex
		app  *App
		hwnd uintptr
		menu uintptr
		nid  notifyIconData
	}
)

type point struct {
	X int32
	Y int32
}

type msg struct {
	HWnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UTimeoutOrVer    uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         guid
	HBalloonIcon     uintptr
}

func setupTray(app *App) error {
	windowsTrayState.mu.Lock()
	windowsTrayState.app = app
	windowsTrayState.mu.Unlock()

	windowsTrayStartOnce.Do(func() {
		started := make(chan error, 1)
		go trayLoop(started)
		windowsTrayStartErr = <-started
	})
	return windowsTrayStartErr
}

func teardownTray() {
	windowsTrayState.mu.Lock()
	windowsTrayState.app = nil
	hwnd := windowsTrayState.hwnd
	windowsTrayState.mu.Unlock()
	if hwnd != 0 {
		procPostMessageW.Call(hwnd, wmClose, 0, 0)
	}
}

func trayLoop(started chan<- error) {
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	className, _ := syscall.UTF16PtrFromString(trayWindowClass)
	hInstance := currentModuleHandle()
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   trayWndProcCallback,
		HInstance:     hInstance,
		HIcon:         loadIcon(0, idiApplication),
		HCursor:       loadCursor(0, idcArrow),
		LpszClassName: className,
	}

	atom, _, regErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 && regErr != syscall.Errno(1410) { // class already exists
		started <- fmt.Errorf("register tray class failed: %w", regErr)
		return
	}

	hwnd, _, createErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0, 0, 0, 0,
		0,
		0,
		hInstance,
		0,
	)
	if hwnd == 0 {
		started <- fmt.Errorf("create tray window failed: %w", createErr)
		return
	}

	windowsTrayState.mu.Lock()
	windowsTrayState.hwnd = hwnd
	windowsTrayState.mu.Unlock()

	if err := addTrayIcon(hwnd); err != nil {
		procDestroyWindow.Call(hwnd)
		started <- err
		return
	}
	started <- nil

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	windowsTrayState.mu.Lock()
	windowsTrayState.hwnd = 0
	windowsTrayState.menu = 0
	windowsTrayState.nid = notifyIconData{}
	windowsTrayState.mu.Unlock()
}

func addTrayIcon(hwnd uintptr) error {
	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              trayIconID,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: trayCallbackMessage,
		HIcon:            loadIcon(0, idiApplication),
	}
	copy(nid.SzTip[:], syscall.StringToUTF16("SkillFlow"))

	ok, _, err := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))
	if ok == 0 {
		return fmt.Errorf("add tray icon failed: %w", err)
	}

	windowsTrayState.mu.Lock()
	windowsTrayState.nid = nid
	windowsTrayState.mu.Unlock()
	return nil
}

func removeTrayIcon() {
	windowsTrayState.mu.RLock()
	nid := windowsTrayState.nid
	windowsTrayState.mu.RUnlock()
	if nid.CbSize != 0 {
		procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
	}
}

func currentModuleHandle() uintptr {
	h, _, _ := procGetModuleHandleW.Call(0)
	return h
}

func loadIcon(instance uintptr, iconID uintptr) uintptr {
	h, _, _ := procLoadIconW.Call(instance, iconID)
	return h
}

func loadCursor(instance uintptr, cursorID uintptr) uintptr {
	h, _, _ := procLoadCursorW.Call(instance, cursorID)
	return h
}

func ensureTrayMenu() uintptr {
	windowsTrayState.mu.Lock()
	defer windowsTrayState.mu.Unlock()
	if windowsTrayState.menu != 0 {
		return windowsTrayState.menu
	}
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return 0
	}
	showText, _ := syscall.UTF16PtrFromString("Show SkillFlow")
	exitText, _ := syscall.UTF16PtrFromString("Exit")
	procAppendMenuW.Call(menu, mfString, trayMenuCmdShow, uintptr(unsafe.Pointer(showText)))
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	procAppendMenuW.Call(menu, mfString, trayMenuCmdExit, uintptr(unsafe.Pointer(exitText)))
	windowsTrayState.menu = menu
	return menu
}

func showTrayMenu(hwnd uintptr) {
	menu := ensureTrayMenu()
	if menu == 0 {
		return
	}
	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(hwnd)
	procTrackPopupMenu.Call(
		menu,
		tpmLeftAlign|tpmBottomAlign|tpmRightButton,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		hwnd,
		0,
	)
	procPostMessageW.Call(hwnd, wmNull, 0, 0)
}

func withWindowsTrayApp(fn func(*App)) {
	windowsTrayState.mu.RLock()
	app := windowsTrayState.app
	windowsTrayState.mu.RUnlock()
	if app != nil {
		fn(app)
	}
}

func trayWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case trayCallbackMessage:
		switch uint32(lParam) {
		case wmRButton, wmLButton:
			showTrayMenu(hwnd)
		case wmLDouble:
			withWindowsTrayApp(func(app *App) {
				go app.showMainWindow()
			})
		}
		return 0
	case wmCommand:
		switch uint16(wParam & 0xFFFF) {
		case trayMenuCmdShow:
			withWindowsTrayApp(func(app *App) {
				go app.showMainWindow()
			})
		case trayMenuCmdExit:
			withWindowsTrayApp(func(app *App) {
				go app.quitApp()
			})
		}
		return 0
	case trayShowWindowMsg:
		withWindowsTrayApp(func(app *App) {
			go app.showMainWindow()
		})
		return 0
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		removeTrayIcon()
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		return ret
	}
}

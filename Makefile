ifeq ($(OS),Windows_NT)
    SHELL := cmd.exe
    .SHELLFLAGS := /C
    WAILS := $(shell where wails.exe 2>nul || echo $(USERPROFILE)\go\bin\wails.exe)
else
    WAILS := $(shell which wails 2>/dev/null || echo $(HOME)/go/bin/wails)
endif

APP_DIR := cmd/skillflow
APP_DIR_WIN := $(subst /,\,$(APP_DIR))
NODE ?= node
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows_NT)

WAILS_BUILD_TAGS ?=
WAILS_BUILD_LDFLAGS ?= -s -w
BUILD_PLAN_CMD = $(NODE) $(APP_DIR)/tools/build-plan.mjs
NORMALIZE_PROVIDERS = $(strip $(subst $(COMMA),$(SPACE),$(1)))
PROVIDER_BUILD_TAGS = $(strip provider_select $(foreach provider,$(call NORMALIZE_PROVIDERS,$(1)),backup_$(provider)))
WAILS_SKIP_FLAGS = $(shell $(BUILD_PLAN_CMD) plan)
WAILS_BUILD_FLAGS = -trimpath -m -nosyncgomod $(WAILS_SKIP_FLAGS) $(if $(strip $(WAILS_BUILD_TAGS)),-tags "$(WAILS_BUILD_TAGS)") $(if $(strip $(WAILS_BUILD_LDFLAGS)),-ldflags "$(WAILS_BUILD_LDFLAGS)")

.PHONY: all dev dev-legacy build build-native build-native-macos build-native-windows build-legacy build-cloud test test-native test-native-build test-native-macos test-native-macos-build test-native-windows test-native-windows-build test-cloud tidy generate install-frontend clean help

all: build-native

## dev: Run the current native client in development mode (macOS only for now)
dev: dev-native

## dev-native: Run the native macOS client in development mode
dev-native:
ifeq ($(UNAME_S),Darwin)
	./native/scripts/build-macos.sh
	open build/native/macos/SkillFlow.app
else
	$(error Native development is only supported on macOS in this Makefile; use make build-native on Windows)
endif

## dev-legacy: Run the legacy Wails UI with hot-reload
dev-legacy:
	cd $(APP_DIR) && $(WAILS) dev

## build: Build the native client for the current machine
build: build-native

## build-native: Build the native client and bundled daemon for the current machine
build-native:
ifeq ($(OS),Windows_NT)
	$(MAKE) build-native-windows
else ifeq ($(UNAME_S),Darwin)
	$(MAKE) build-native-macos
else
	$(error Native builds are supported on macOS and Windows only)
endif

## build-native-macos: Build SkillFlow.app with the bundled skillflowd daemon
build-native-macos:
ifeq ($(UNAME_S),Darwin)
	./native/scripts/build-macos.sh
else
	$(error build-native-macos must run on macOS)
endif

## build-native-windows: Build the WinUI client with the bundled skillflowd.exe daemon
build-native-windows:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File native\scripts\build-windows.ps1
else
	$(error build-native-windows must run on Windows)
endif

## build-legacy: Build the legacy Wails production binary
build-legacy:
	cd $(APP_DIR) && $(WAILS) build $(WAILS_BUILD_FLAGS)
	$(BUILD_PLAN_CMD) mark
ifneq ($(OS),Windows_NT)
	@if [ -f $(APP_DIR)/build/darwin/iconfile.icns ] && [ -d $(APP_DIR)/build/bin/SkillFlow.app ]; then \
		cp $(APP_DIR)/build/darwin/iconfile.icns $(APP_DIR)/build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
	fi
endif

## build-cloud: Build legacy Wails with only selected cloud providers, e.g. make build-cloud PROVIDERS="aws,google"
build-cloud:
	$(if $(strip $(PROVIDERS)),,$(error Usage: make build-cloud PROVIDERS="aws,google"))
	cd $(APP_DIR) && $(WAILS) build -trimpath -m -nosyncgomod $(WAILS_SKIP_FLAGS) -tags "$(strip $(WAILS_BUILD_TAGS) $(call PROVIDER_BUILD_TAGS,$(PROVIDERS)))" $(if $(strip $(WAILS_BUILD_LDFLAGS)),-ldflags "$(WAILS_BUILD_LDFLAGS)")
	$(BUILD_PLAN_CMD) mark
ifneq ($(OS),Windows_NT)
	@if [ -f $(APP_DIR)/build/darwin/iconfile.icns ] && [ -d $(APP_DIR)/build/bin/SkillFlow.app ]; then \
		cp $(APP_DIR)/build/darwin/iconfile.icns $(APP_DIR)/build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
	fi
endif

## test: Run backend tests and current-platform native build checks
test:
	go test ./core/... ./cmd/skillflow
	$(MAKE) test-native-build

## test-native: Run full native client tests for the current platform
test-native:
ifeq ($(OS),Windows_NT)
	$(MAKE) test-native-windows
else ifeq ($(UNAME_S),Darwin)
	$(MAKE) test-native-macos
else
	$(error Native tests are supported on macOS and Windows only)
endif

## test-native-build: Compile the current-platform native client
test-native-build:
ifeq ($(OS),Windows_NT)
	$(MAKE) test-native-windows-build
else ifeq ($(UNAME_S),Darwin)
	$(MAKE) test-native-macos-build
else
	$(error Native build checks are supported on macOS and Windows only)
endif

## test-native-macos: Run Swift package tests (requires full Xcode for XCTest)
test-native-macos:
ifeq ($(UNAME_S),Darwin)
	swift test --package-path native/macos/SkillFlow
else
	$(error test-native-macos must run on macOS)
endif

## test-native-macos-build: Compile the Swift package with SwiftPM
test-native-macos-build:
ifeq ($(UNAME_S),Darwin)
	swift build --package-path native/macos/SkillFlow -c release
else
	$(error test-native-macos-build must run on macOS)
endif

## test-native-windows: Run WinUI and .NET tests
test-native-windows:
ifeq ($(OS),Windows_NT)
	dotnet test native/windows/SkillFlow/SkillFlow.sln -c Debug -p:Platform=x64 --nologo
else
	$(error test-native-windows must run on Windows)
endif

## test-native-windows-build: Compile the WinUI client
test-native-windows-build:
ifeq ($(OS),Windows_NT)
	dotnet build native/windows/SkillFlow/SkillFlow.sln -c Debug -p:Platform=x64 --nologo
else
	$(error test-native-windows-build must run on Windows)
endif

## test-cloud: Run Go tests with only selected cloud providers, e.g. make test-cloud PROVIDERS="aws,google"
test-cloud:
	$(if $(strip $(PROVIDERS)),,$(error Usage: make test-cloud PROVIDERS="aws,google"))
	go test -tags "$(strip $(WAILS_BUILD_TAGS) $(call PROVIDER_BUILD_TAGS,$(PROVIDERS)))" ./core/...

## tidy: Sync Go module dependencies
tidy:
	go mod tidy

## generate: Regenerate legacy Wails TypeScript bindings after App method changes
generate:
	cd $(APP_DIR) && $(WAILS) generate module
	$(BUILD_PLAN_CMD) mark-bindings

## install-frontend: Install legacy Wails frontend npm dependencies
install-frontend:
	cd $(APP_DIR)/frontend && npm install

## clean: Remove native and legacy build artifacts
clean:
ifeq ($(OS),Windows_NT)
	if exist "build\native" rmdir /s /q "build\native"
	if exist "$(APP_DIR_WIN)\build\bin" rmdir /s /q "$(APP_DIR_WIN)\build\bin"
	if exist "$(APP_DIR_WIN)\frontend\dist" rmdir /s /q "$(APP_DIR_WIN)\frontend\dist"
	if exist "$(APP_DIR_WIN)\frontend\.cache" rmdir /s /q "$(APP_DIR_WIN)\frontend\.cache"
	if exist "$(APP_DIR_WIN)\frontend\package.json.md5" del /f /q "$(APP_DIR_WIN)\frontend\package.json.md5"
	if exist "$(APP_DIR_WIN)\.build-cache" rmdir /s /q "$(APP_DIR_WIN)\.build-cache"
else
	rm -rf build/native
	rm -rf $(APP_DIR)/build/bin
	rm -rf $(APP_DIR)/frontend/dist
	rm -rf $(APP_DIR)/frontend/.cache
	rm -f $(APP_DIR)/frontend/package.json.md5
	rm -rf $(APP_DIR)/.build-cache
endif

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

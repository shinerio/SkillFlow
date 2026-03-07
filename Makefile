# Detect OS and set appropriate path
ifeq ($(OS),Windows_NT)
    WAILS := $(shell where wails 2>/dev/null || echo $(USERPROFILE)/go/bin/wails.exe)
else
    WAILS := $(shell which wails 2>/dev/null || echo $(HOME)/go/bin/wails)
endif

APP_DIR := cmd/skillflow

.PHONY: all dev build test tidy generate install-frontend clean help

all: build

## dev: Run in dev mode with hot-reload (Go + frontend)
dev:
	cd $(APP_DIR) && $(WAILS) dev

## build: Build production binary
build:
	cd $(APP_DIR) && $(WAILS) build
ifneq ($(OS),Windows_NT)
	@if [ -f $(APP_DIR)/build/darwin/iconfile.icns ] && [ -d $(APP_DIR)/build/bin/SkillFlow.app ]; then \
		cp $(APP_DIR)/build/darwin/iconfile.icns $(APP_DIR)/build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
	fi
endif

## test: Run all Go tests
test:
	go test ./core/...

## tidy: Sync Go module dependencies
tidy:
	go mod tidy

## generate: Regenerate TypeScript bindings after App method changes
generate:
	cd $(APP_DIR) && $(WAILS) generate module

## install-frontend: Install frontend npm dependencies
install-frontend:
	cd $(APP_DIR)/frontend && npm install

## clean: Remove build artifacts
clean:
ifeq ($(OS),Windows_NT)
	if exist $(APP_DIR)\build\bin rmdir /s /q $(APP_DIR)\build\bin
else
	rm -rf $(APP_DIR)/build/bin
endif

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

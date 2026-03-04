WAILS := $(HOME)/go/bin/wails

.PHONY: all dev build test tidy generate install-frontend clean help

all: build

## dev: Run in dev mode with hot-reload (Go + frontend)
dev:
	$(WAILS) dev
## build: Build production binary
build:
	$(WAILS) build

## test: Run all Go tests
test:
	go test ./core/...

## tidy: Sync Go module dependencies
tidy:
	go mod tidy

## generate: Regenerate TypeScript bindings after App method changes
generate:
	$(WAILS) generate module

## install-frontend: Install frontend npm dependencies
install-frontend:
	cd frontend && npm install

## clean: Remove build artifacts
clean:
	rm -rf build/bin

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

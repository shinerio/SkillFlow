# Contributing

## Prerequisites

- macOS 13+ or Windows 10+
- Go 1.25+
- Native client requirements:
  - macOS: Swift 5.9+ with Xcode or Command Line Tools
  - Windows: .NET 8 SDK and Windows App SDK
- Legacy Wails requirements:
  - Node.js 18+
  - Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build and Test Steps

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make test
make build
```

Output:

- macOS: `build/native/macos/SkillFlow.app`
- Windows: `build/native/windows/`
- Legacy Wails: `cmd/skillflow/build/bin/`

`make test` runs the Go backend/shell tests and compiles the current-platform native client. Full native test suites are available with `make test-native`; Swift XCTest requires a full Xcode installation, while Command Line Tools are sufficient for the Swift build check.

## Legacy Wails Workflow

```bash
make install-frontend
make dev-legacy
make build-legacy
make build-cloud PROVIDERS="aws,google"
```

`make generate` regenerates legacy Wails TypeScript bindings after `App` method changes. Frontend unit tests run from `cmd/skillflow/frontend` during CI.

## Common Make Targets

| Target | Description |
|--------|-------------|
| `make dev` | Build and open the native macOS client |
| `make dev-legacy` | Run the legacy Wails UI with frontend hot reload |
| `make build` | Build the native client for the current machine |
| `make build-native-macos` | Build `SkillFlow.app` with the bundled `skillflowd` daemon |
| `make build-native-windows` | Build the WinUI client with the bundled `skillflowd.exe` daemon |
| `make build-legacy` | Build the legacy Wails production binary |
| `make build-cloud PROVIDERS="aws,google"` | Build legacy Wails with selected cloud providers only |
| `make test` | Run Go tests and current-platform native build checks |
| `make test-native` | Run the full current-platform native test suite |
| `make test-cloud PROVIDERS="aws,google"` | Run Go tests with only selected cloud providers |
| `make tidy` | Sync Go module dependencies |
| `make generate` | Regenerate legacy Wails TypeScript bindings |
| `make install-frontend` | Install legacy Wails frontend npm dependencies |
| `make clean` | Remove native and legacy build artifacts |

For contributor-facing internals, see **[docs/architecture/README.md](docs/architecture/README.md)**.

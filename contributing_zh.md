# 参与贡献

## 环境要求

- macOS 13+ 或 Windows 10+
- Go 1.25+
- 原生客户端要求：
  - macOS：Swift 5.9+，安装 Xcode 或 Command Line Tools
  - Windows：.NET 8 SDK 与 Windows App SDK
- Legacy Wails 要求：
  - Node.js 18+
  - Wails v2 CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## 构建与测试

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make test
make build
```

构建输出：

- macOS：`build/native/macos/SkillFlow.app`
- Windows：`build/native/windows/`
- Legacy Wails：`cmd/skillflow/build/bin/`

`make test` 会运行 Go 后端/壳层测试，并编译当前平台的原生客户端。完整原生测试使用 `make test-native`；Swift XCTest 需要完整 Xcode，Swift 编译检查使用 Command Line Tools 即可。

## Legacy Wails 工作流

```bash
make install-frontend
make dev-legacy
make build-legacy
make build-cloud PROVIDERS="aws,google"
```

`App` 方法变更后，使用 `make generate` 重新生成 legacy Wails TypeScript 绑定。前端单元测试由 CI 在 `cmd/skillflow/frontend` 中运行。

## 常用 Make 目标

| 目标 | 说明 |
|------|------|
| `make dev` | 构建并打开 macOS 原生客户端 |
| `make dev-legacy` | 启动 legacy Wails 热更新开发模式 |
| `make build` | 构建当前平台的原生客户端 |
| `make build-native-macos` | 构建 `SkillFlow.app`，并打包 `skillflowd` daemon |
| `make build-native-windows` | 构建 WinUI 客户端，并打包 `skillflowd.exe` daemon |
| `make build-legacy` | 构建 legacy Wails 生产版本 |
| `make build-cloud PROVIDERS="aws,google"` | 仅选择指定云服务商构建 legacy Wails |
| `make test` | 运行 Go 测试与当前平台原生客户端编译检查 |
| `make test-native` | 运行当前平台完整原生测试 |
| `make test-cloud PROVIDERS="aws,google"` | 仅选择指定云服务商运行 Go 测试 |
| `make tidy` | 同步 Go module 依赖 |
| `make generate` | 重新生成 legacy Wails TypeScript 绑定 |
| `make install-frontend` | 安装 legacy Wails 前端 npm 依赖 |
| `make clean` | 清理原生与 legacy 构建产物 |

如需了解面向贡献者的内部结构，请参阅 **[docs/architecture/README_zh.md](docs/architecture/README_zh.md)**。

# 原生平台重构实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 SkillFlow 从 Wails/React 生产 UI 分批迁移到 macOS Swift 原生客户端和 Windows WinUI 原生客户端，同时保留现有功能、数据格式、双平台支持和 Go 核心业务语义。

**Architecture:** 先把现有 Go `core/` 能力收敛为本机 `daemon` 契约，迁移期保留 Wails UI 作为行为基线；随后分别建立 macOS 与 Windows 原生客户端，并按 Settings、My Skills、My Agents、Starred Repos、Prompts、Memory、Backup 的垂直切片迁移。原生客户端只负责平台 UI 和系统集成，业务读写通过本机契约进入 `daemon`。

**Tech Stack:** Go 1.25, Swift, SwiftUI, AppKit, C#, .NET, WinUI 3, Windows App SDK, local IPC, JSON contract tests, Go test, XCTest, dotnet test, GitHub Actions, Markdown docs

---

## 总体执行规则

- 每个批次都必须能单独合并。
- 每个批次完成后都要保留旧 Wails UI 可用，直到发布切换批次。
- 每个功能切片必须同时完成 macOS 和 Windows。
- 原生客户端不直接读写 `config*.json`、`star_repos*.json`、`skills/`、`prompts/` 或 `meta/`，只能通过 `daemon`。
- 新增 Go 代码不得放在仓库根目录。
- 新增原生客户端代码放在 `native/macos/` 和 `native/windows/`。
- 用户可见功能变更必须同步更新 `docs/features.md` 和 `docs/features_zh.md`；纯迁移且行为不变时只更新迁移计划或架构文档。

## Batch 0: 冻结功能契约和资源基线

**目标:** 在写原生客户端前，先把“功能不变”和“资源改善”变成可验证清单。

**Files:**
- Create: `docs/plans/2026-04-25-native-platform-feature-contract.md`
- Create: `docs/plans/2026-04-25-native-platform-performance-baseline.md`
- Create: `docs/plans/2026-04-25-native-platform-release-checklist.md`
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`

### Task 0.1: 整理功能契约

**Step 1:** 从 `docs/features.md` 逐节复制现有页面和交互，整理到 `docs/plans/2026-04-25-native-platform-feature-contract.md`。

**Step 2:** 为每个功能加上验收列：`Wails baseline`、`macOS Native`、`Windows Native`、`daemon API`。

**Step 3:** 明确所有现有后端入口，以 `cmd/skillflow/app*.go` 导出的 `App` 方法为初始 API 对照。

**Step 4:** 提交文档。

```bash
git add docs/plans/2026-04-25-native-platform-feature-contract.md
git commit -m "docs: add native platform feature contract"
```

### Task 0.2: 记录资源基线

**Step 1:** 在 macOS 记录当前 Wails 构建的冷启动、窗口打开 RSS、关闭窗口后 daemon RSS、Dashboard 首屏时间。

**Step 2:** 在 Windows 记录同样指标。

**Step 3:** 写入 `docs/plans/2026-04-25-native-platform-performance-baseline.md`，记录测试机器、系统版本、构建版本和测量命令。

**Step 4:** 提交文档。

```bash
git add docs/plans/2026-04-25-native-platform-performance-baseline.md
git commit -m "docs: record native refactor performance baseline"
```

### Task 0.3: 建立发布检查清单

**Step 1:** 在 `docs/plans/2026-04-25-native-platform-release-checklist.md` 增加双平台验收项。

**Step 2:** 将功能契约、性能基线、配置兼容、备份兼容、Agent 行为、安装包签名列为阻塞项。

**Step 3:** 在 `docs/architecture/README.md` 和 `docs/architecture/README_zh.md` 加入原生迁移计划入口链接。

**Step 4:** 提交文档。

```bash
git add docs/plans/2026-04-25-native-platform-release-checklist.md docs/architecture/README.md docs/architecture/README_zh.md
git commit -m "docs: add native refactor release checklist"
```

## Batch 1: 收敛 Go daemon 契约

**目标:** 在不破坏 Wails UI 的前提下，让业务能力通过稳定本机契约调用。

**Files:**
- Create: `core/platform/nativeapi/contract.go`
- Create: `core/platform/nativeapi/errors.go`
- Create: `core/platform/nativeapi/router.go`
- Create: `core/platform/nativeapi/router_test.go`
- Create: `core/platform/nativeapi/contract_test.go`
- Modify: `core/platform/ipc/protocol.go`
- Modify: `core/platform/ipc/server.go`
- Modify: `core/platform/ipc/client.go`
- Modify: `cmd/skillflow/app_daemon_service.go`
- Modify: `cmd/skillflow/providers.go`
- Modify: `docs/architecture/runtime-and-storage.md`
- Modify: `docs/architecture/runtime-and-storage_zh.md`

### Task 1.1: 定义契约外壳

**Step 1:** 写失败测试，验证契约请求包含 `version`、`method`、`params`、`requestID`，响应包含 `ok`、`result`、`error`。

**Step 2:** 运行测试。

```bash
go test ./core/platform/nativeapi -run TestContractEnvelope
```

Expected: `FAIL`

**Step 3:** 新增最小 `Request`、`Response`、`Error` 类型。

**Step 4:** 运行测试通过。

```bash
go test ./core/platform/nativeapi -run TestContractEnvelope
```

Expected: `ok`

**Step 5:** 提交。

```bash
git add core/platform/nativeapi
git commit -m "feat: add native api contract envelope"
```

### Task 1.2: 建立 daemon API router

**Step 1:** 写失败测试，验证未知方法返回稳定错误码 `method_not_found`。

**Step 2:** 写失败测试，验证 handler panic 会被转换为 `internal_error`，不能崩溃 server。

**Step 3:** 实现 `Router`、`Register`、`Handle`。

**Step 4:** 运行测试。

```bash
go test ./core/platform/nativeapi
```

Expected: `ok`

**Step 5:** 提交。

```bash
git add core/platform/nativeapi
git commit -m "feat: route native daemon api calls"
```

### Task 1.3: 复用现有 IPC 传输承载契约

**Step 1:** 扩展 `core/platform/ipc` 测试，验证带 token 的 JSON 契约请求可以往返。

**Step 2:** 运行测试确认失败。

```bash
go test ./core/platform/ipc ./core/platform/nativeapi
```

Expected: `FAIL`

**Step 3:** 扩展 `ipc.Request` / `ipc.Response` 或新增通用 payload 字段，保持旧控制命令兼容。

**Step 4:** 运行测试通过。

```bash
go test ./core/platform/ipc ./core/platform/nativeapi
```

Expected: `ok`

**Step 5:** 提交。

```bash
git add core/platform/ipc core/platform/nativeapi
git commit -m "feat: carry native api over local ipc"
```

### Task 1.4: 映射第一组只读 API

**Step 1:** 为以下方法写 daemon 契约测试：`settings.get`、`skills.list`、`skills.categories.list`、`agents.listEnabled`、`backup.providers.list`。

**Step 2:** 运行测试确认失败。

```bash
go test ./cmd/skillflow ./core/platform/nativeapi -run 'TestNativeAPI'
```

Expected: `FAIL`

**Step 3:** 在 `cmd/skillflow/providers.go` 或新的组合根中注册 handler，内部复用现有 app service。

**Step 4:** 运行测试通过。

```bash
go test ./cmd/skillflow ./core/platform/nativeapi -run 'TestNativeAPI'
```

Expected: `ok`

**Step 5:** 提交。

```bash
git add cmd/skillflow core/platform/nativeapi
git commit -m "feat: expose read-only native daemon api"
```

### Task 1.5: 更新运行时架构文档

**Step 1:** 更新 `docs/architecture/runtime-and-storage.md`，说明迁移期存在 Wails UI 与 Native UI 两类客户端。

**Step 2:** 更新 `docs/architecture/runtime-and-storage_zh.md`。

**Step 3:** 运行核心测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
git commit -m "docs: describe native daemon api runtime"
```

## Batch 2: 建立 macOS 原生客户端骨架

**目标:** 先跑通 macOS 原生窗口、导航和 daemon 连接，不迁移复杂业务。

**Files:**
- Create: `native/macos/SkillFlow/SkillFlow.xcodeproj/project.pbxproj`
- Create: `native/macos/SkillFlow/Sources/SkillFlowApp.swift`
- Create: `native/macos/SkillFlow/Sources/AppDelegate.swift`
- Create: `native/macos/SkillFlow/Sources/Daemon/DaemonClient.swift`
- Create: `native/macos/SkillFlow/Sources/Daemon/DaemonModels.swift`
- Create: `native/macos/SkillFlow/Sources/Shell/SidebarView.swift`
- Create: `native/macos/SkillFlow/Sources/Shell/RootView.swift`
- Create: `native/macos/SkillFlow/Tests/DaemonClientTests.swift`
- Modify: `.github/workflows/test.yml`

### Task 2.1: 创建 Swift 工程骨架

**Step 1:** 创建最小 macOS App 工程，入口为 `SkillFlowApp.swift`。

**Step 2:** 增加空的 `RootView` 和 `SidebarView`。

**Step 3:** 本地构建。

```bash
xcodebuild -project native/macos/SkillFlow/SkillFlow.xcodeproj -scheme SkillFlow -configuration Debug build
```

Expected: `BUILD SUCCEEDED`

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow
git commit -m "feat: add macos native app shell"
```

### Task 2.2: 接入 daemon 发现和只读调用

**Step 1:** 写 `DaemonClientTests`，使用 fixture endpoint 验证 token、版本和错误映射。

**Step 2:** 运行测试确认失败。

```bash
xcodebuild -project native/macos/SkillFlow/SkillFlow.xcodeproj -scheme SkillFlow test
```

Expected: `FAIL`

**Step 3:** 实现 `DaemonClient` 和 `DaemonModels`。

**Step 4:** 运行测试通过。

```bash
xcodebuild -project native/macos/SkillFlow/SkillFlow.xcodeproj -scheme SkillFlow test
```

Expected: `ok`

**Step 5:** 提交。

```bash
git add native/macos/SkillFlow
git commit -m "feat: connect macos shell to daemon api"
```

### Task 2.3: 接入 CI smoke build

**Step 1:** 在 `.github/workflows/test.yml` 增加 macOS native smoke job。

**Step 2:** 保持现有 Go 和前端测试不变。

**Step 3:** 提交。

```bash
git add .github/workflows/test.yml
git commit -m "ci: add macos native smoke build"
```

## Batch 3: 建立 Windows 原生客户端骨架

**目标:** 跑通 Windows 原生窗口、导航、托盘占位和 daemon 连接。

**Files:**
- Create: `native/windows/SkillFlow/SkillFlow.sln`
- Create: `native/windows/SkillFlow/SkillFlow/SkillFlow.csproj`
- Create: `native/windows/SkillFlow/SkillFlow/App.xaml`
- Create: `native/windows/SkillFlow/SkillFlow/App.xaml.cs`
- Create: `native/windows/SkillFlow/SkillFlow/MainWindow.xaml`
- Create: `native/windows/SkillFlow/SkillFlow/MainWindow.xaml.cs`
- Create: `native/windows/SkillFlow/SkillFlow/Daemon/DaemonClient.cs`
- Create: `native/windows/SkillFlow/SkillFlow/Daemon/DaemonModels.cs`
- Create: `native/windows/SkillFlow/SkillFlow.Tests/SkillFlow.Tests.csproj`
- Create: `native/windows/SkillFlow/SkillFlow.Tests/DaemonClientTests.cs`
- Modify: `.github/workflows/test.yml`

### Task 3.1: 创建 WinUI 工程骨架

**Step 1:** 创建 WinUI 3 项目和测试项目。

**Step 2:** 添加主窗口、左侧导航和占位页面。

**Step 3:** 本地构建。

```bash
dotnet build native/windows/SkillFlow/SkillFlow.sln
```

Expected: `Build succeeded`

**Step 4:** 提交。

```bash
git add native/windows/SkillFlow
git commit -m "feat: add windows native app shell"
```

### Task 3.2: 接入 daemon 发现和只读调用

**Step 1:** 写 `DaemonClientTests`，覆盖 endpoint 读取、token 注入、错误码映射。

**Step 2:** 运行测试确认失败。

```bash
dotnet test native/windows/SkillFlow/SkillFlow.sln
```

Expected: `FAIL`

**Step 3:** 实现 `DaemonClient` 和 `DaemonModels`。

**Step 4:** 运行测试通过。

```bash
dotnet test native/windows/SkillFlow/SkillFlow.sln
```

Expected: `Passed`

**Step 5:** 提交。

```bash
git add native/windows/SkillFlow
git commit -m "feat: connect windows shell to daemon api"
```

### Task 3.3: 接入 CI smoke build

**Step 1:** 在 `.github/workflows/test.yml` 增加 Windows native smoke job。

**Step 2:** 确保 Linux 上原有 Go / frontend job 不被 Windows SDK 依赖阻塞。

**Step 3:** 提交。

```bash
git add .github/workflows/test.yml
git commit -m "ci: add windows native smoke build"
```

## Batch 4: 迁移 Settings / Shell / 系统能力

**目标:** 先迁移低业务风险但高平台体验价值的设置和系统能力。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app_settings.go`
- Modify: `cmd/skillflow/app_proxy.go`
- Modify: `cmd/skillflow/app_update.go`
- Modify: `cmd/skillflow/app_log.go`
- Modify: `cmd/skillflow/app_autostart*.go`
- Modify: `native/macos/SkillFlow/Sources/Settings/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/Settings/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 4.1: 暴露 Settings daemon API

**Step 1:** 为 `settings.get`、`settings.save`、`proxy.test`、`logs.openDir`、`app.update.check`、`autostart.set` 写 Go 契约测试。

**Step 2:** 实现 handler，复用现有 `App` 方法背后的 service。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow
git commit -m "feat: expose settings native api"
```

### Task 4.2: macOS Settings 页面

**Step 1:** 实现 General、Agents、Cloud、Proxy 基础表单。

**Step 2:** 使用原生文件选择器选择目录。

**Step 3:** 运行 macOS 测试和构建。

```bash
xcodebuild -project native/macos/SkillFlow/SkillFlow.xcodeproj -scheme SkillFlow test
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow
git commit -m "feat: implement macos native settings"
```

### Task 4.3: Windows Settings 页面

**Step 1:** 实现同等 Settings 页面。

**Step 2:** 使用 Windows 原生文件选择器选择目录。

**Step 3:** 运行 Windows 测试和构建。

```bash
dotnet test native/windows/SkillFlow/SkillFlow.sln
```

Expected: `Passed`

**Step 4:** 提交。

```bash
git add native/windows/SkillFlow
git commit -m "feat: implement windows native settings"
```

### Task 4.4: 更新功能文档

**Step 1:** 如果设置行为发生用户可见变化，更新 `docs/features.md`。

**Step 2:** 同步更新 `docs/features_zh.md`。

**Step 3:** 提交。

```bash
git add docs/features.md docs/features_zh.md
git commit -m "docs: document native settings behavior"
```

## Batch 5: 迁移 My Skills

**目标:** 完成技能库核心工作流：列表、搜索、排序、分类、导入、删除、移动、更新、手动推送、自动推送目标。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/skill_state.go`
- Modify: `core/orchestration/*`
- Modify: `core/readmodel/skills/*`
- Modify: `native/macos/SkillFlow/Sources/Skills/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/Skills/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 5.1: 暴露 Skills daemon API

**Step 1:** 为以下方法写契约测试：`skills.list`、`skills.importLocal`、`skills.delete`、`skills.deleteBatch`、`skills.moveCategory`、`skills.updateCheck`、`skills.updateOne`、`skills.push`、`skills.pushForce`。

**Step 2:** 实现 handler。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow core/orchestration core/readmodel/skills
git commit -m "feat: expose skills native api"
```

### Task 5.2: macOS My Skills 页面

**Step 1:** 实现原生列表、分类侧边栏、搜索、排序和空状态。

**Step 2:** 实现导入、删除、移动分类和批量选择。

**Step 3:** 实现手动推送目标选择、缺失目录确认和冲突处理。

**Step 4:** 运行 macOS 测试。

```bash
xcodebuild -project native/macos/SkillFlow/SkillFlow.xcodeproj -scheme SkillFlow test
```

Expected: `ok`

**Step 5:** 提交。

```bash
git add native/macos/SkillFlow
git commit -m "feat: implement macos native skills page"
```

### Task 5.3: Windows My Skills 页面

**Step 1:** 实现同等页面和交互。

**Step 2:** 实现 Windows 文件夹拖放导入。

**Step 3:** 实现冲突和缺失目录对话框。

**Step 4:** 运行 Windows 测试。

```bash
dotnet test native/windows/SkillFlow/SkillFlow.sln
```

Expected: `Passed`

**Step 5:** 提交。

```bash
git add native/windows/SkillFlow
git commit -m "feat: implement windows native skills page"
```

## Batch 6: 迁移 My Agents

**目标:** 完成 Agent 列表、技能扫描、Push/Pull、删除、路径管理和记忆预览。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_agent_memory.go`
- Modify: `core/agentintegration/*`
- Modify: `core/readmodel/agentmemory/*`
- Modify: `native/macos/SkillFlow/Sources/Agents/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/Agents/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 6.1: 暴露 Agents daemon API

**Step 1:** 为 `agents.list`、`agents.scanSkills`、`agents.listSkills`、`agents.deleteSkill`、`agents.pull`、`agents.pullForce`、`agents.memoryPreview` 写契约测试。

**Step 2:** 实现 handler。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow core/agentintegration core/readmodel/agentmemory
git commit -m "feat: expose agents native api"
```

### Task 6.2: macOS My Agents 页面

**Step 1:** 实现 Agent 列表和技能/记忆切换。

**Step 2:** 实现手动拉取、选择未导入、冲突处理。

**Step 3:** 实现 PushDir 和 ScanDirs 展示、打开路径。

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow
git commit -m "feat: implement macos native agents page"
```

### Task 6.3: Windows My Agents 页面

**Step 1:** 实现同等页面和交互。

**Step 2:** 验证 Windows 路径打开、缺失目录和扫描错误显示。

**Step 3:** 提交。

```bash
git add native/windows/SkillFlow
git commit -m "feat: implement windows native agents page"
```

## Batch 7: 迁移 Starred Repos

**目标:** 完成仓库收藏、认证、刷新、候选技能、导入和批量操作。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `core/skillsource/*`
- Modify: `core/orchestration/*`
- Modify: `native/macos/SkillFlow/Sources/StarredRepos/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/StarredRepos/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 7.1: 暴露 Starred Repos daemon API

**Step 1:** 为 `starred.listRepos`、`starred.addRepo`、`starred.addRepoWithCredentials`、`starred.removeRepo`、`starred.updateRepo`、`starred.updateAll`、`starred.listAllSkills`、`starred.listRepoSkills`、`starred.importSkills` 写契约测试。

**Step 2:** 实现 handler。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow core/skillsource core/orchestration
git commit -m "feat: expose starred repos native api"
```

### Task 7.2: 双平台实现 Starred Repos 页面

**Step 1:** 先实现 macOS 页面。

**Step 2:** 再实现 Windows 页面。

**Step 3:** 验证 HTTP 认证、SSH 错误、批量导入、批量 Push。

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow native/windows/SkillFlow
git commit -m "feat: implement native starred repos pages"
```

## Batch 8: 迁移 My Prompts

**目标:** 完成提示词分类、编辑、导入、导出、媒体和链接管理。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app_prompt.go`
- Modify: `core/promptcatalog/*`
- Modify: `native/macos/SkillFlow/Sources/Prompts/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/Prompts/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 8.1: 暴露 Prompts daemon API

**Step 1:** 为 prompt CRUD、分类、导入预览、导入完成、导出写契约测试。

**Step 2:** 实现 handler。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow core/promptcatalog
git commit -m "feat: expose prompts native api"
```

### Task 8.2: 双平台实现 My Prompts 页面

**Step 1:** 实现 macOS prompt 列表和编辑器。

**Step 2:** 实现 Windows prompt 列表和编辑器。

**Step 3:** 验证导入冲突、批量导出和媒体链接。

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow native/windows/SkillFlow
git commit -m "feat: implement native prompts pages"
```

## Batch 9: 迁移 My Memory

**目标:** 完成主记忆、模块记忆、批量推送、自动同步、外部编辑器和 Agent 侧验证。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app_memory.go`
- Modify: `core/memorycatalog/*`
- Modify: `native/macos/SkillFlow/Sources/Memory/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/Memory/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 9.1: 暴露 Memory daemon API

**Step 1:** 为主记忆、模块记忆、推送配置、推送状态、外部编辑器写契约测试。

**Step 2:** 实现 handler。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow core/memorycatalog
git commit -m "feat: expose memory native api"
```

### Task 9.2: 双平台实现 My Memory 页面

**Step 1:** 实现 macOS 主记忆和模块记忆编辑。

**Step 2:** 实现 Windows 主记忆和模块记忆编辑。

**Step 3:** 验证批量推送、自动同步状态和外部编辑器打开。

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow native/windows/SkillFlow
git commit -m "feat: implement native memory pages"
```

## Batch 10: 迁移 Cloud Backup

**目标:** 完成云备份配置、备份、恢复、Git 冲突处理和变更列表。

**Files:**
- Modify: `core/platform/nativeapi/router.go`
- Modify: `cmd/skillflow/app_backup.go`
- Modify: `cmd/skillflow/app_restore.go`
- Modify: `core/backup/*`
- Modify: `native/macos/SkillFlow/Sources/Backup/*.swift`
- Modify: `native/windows/SkillFlow/SkillFlow/Backup/*.cs`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`

### Task 10.1: 暴露 Backup daemon API

**Step 1:** 为 provider list、backup now、restore、remote files、last changes、Git conflict resolve 写契约测试。

**Step 2:** 实现 handler。

**Step 3:** 运行测试。

```bash
go test ./core/... ./cmd/skillflow
```

Expected: `ok`

**Step 4:** 提交。

```bash
git add core/platform/nativeapi cmd/skillflow core/backup
git commit -m "feat: expose backup native api"
```

### Task 10.2: 双平台实现 Cloud Backup 页面

**Step 1:** 实现 macOS backup 页面。

**Step 2:** 实现 Windows backup 页面。

**Step 3:** 验证所有 provider 配置、手动备份、恢复和 Git 冲突。

**Step 4:** 提交。

```bash
git add native/macos/SkillFlow native/windows/SkillFlow
git commit -m "feat: implement native backup pages"
```

## Batch 11: Native Preview 打包和发布

**目标:** 先生成可下载 Native Preview，不替换正式发布。

**Files:**
- Modify: `.github/workflows/build.yml`
- Create: `native/macos/packaging/README.md`
- Create: `native/windows/packaging/README.md`
- Modify: `README.md`
- Modify: `README_zh.md`
- Modify: `contributing.md`
- Modify: `contributing_zh.md`

### Task 11.1: macOS Native Preview DMG

**Step 1:** 在 build workflow 添加 macOS native build job。

**Step 2:** 将 `skillflowd` 与 macOS app bundle 打包在一起。

**Step 3:** 保留现有 Wails artifact。

**Step 4:** 提交。

```bash
git add .github/workflows/build.yml native/macos/packaging
git commit -m "ci: package macos native preview"
```

### Task 11.2: Windows Native Preview 包

**Step 1:** 在 build workflow 添加 Windows native build job。

**Step 2:** 将 `skillflowd` 与 WinUI 客户端打包在一起。

**Step 3:** 保留现有 Wails artifact。

**Step 4:** 提交。

```bash
git add .github/workflows/build.yml native/windows/packaging
git commit -m "ci: package windows native preview"
```

### Task 11.3: 更新构建文档

**Step 1:** 更新 README 的下载说明，标记 Native Preview。

**Step 2:** 更新 contributing 中的原生客户端构建命令。

**Step 3:** 同步中文文档。

**Step 4:** 提交。

```bash
git add README.md README_zh.md contributing.md contributing_zh.md
git commit -m "docs: add native preview build instructions"
```

## Batch 12: 切换默认发布到 Native

**目标:** 在所有功能验收完成后，正式把原生客户端作为默认发布产物。

**Files:**
- Modify: `.github/workflows/build.yml`
- Modify: `.github/workflows/test.yml`
- Modify: `Makefile`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/architecture/README.md`
- Modify: `docs/architecture/README_zh.md`
- Modify: `docs/architecture/runtime-and-storage.md`
- Modify: `docs/architecture/runtime-and-storage_zh.md`

### Task 12.1: 发布前总体验证

**Step 1:** 在 macOS 运行完整原生客户端验收清单。

**Step 2:** 在 Windows 运行完整原生客户端验收清单。

**Step 3:** 更新性能基线文档，写入 Native 指标。

**Step 4:** 所有阻塞项清零后继续。

### Task 12.2: 切换默认构建

**Step 1:** 将 release artifact 默认指向 Native 包。

**Step 2:** 将 Wails 构建改为 legacy artifact 或移出默认发布。

**Step 3:** 更新 Makefile，增加 `make build-native`、`make test-native`，保留 `make build-legacy`。

**Step 4:** 提交。

```bash
git add .github/workflows/build.yml .github/workflows/test.yml Makefile
git commit -m "ci: switch default release to native clients"
```

### Task 12.3: 更新最终文档

**Step 1:** 更新功能文档中的 shell 和发布行为。

**Step 2:** 更新架构文档，说明 Wails 已不再是生产 UI。

**Step 3:** 同步中文文档。

**Step 4:** 提交。

```bash
git add docs/features.md docs/features_zh.md docs/architecture/README.md docs/architecture/README_zh.md docs/architecture/runtime-and-storage.md docs/architecture/runtime-and-storage_zh.md
git commit -m "docs: document native client architecture"
```

## Batch 13: 可选核心原生化评估

**目标:** 只有当 Native UI + Go daemon 仍不满足资源或分发目标时，再评估是否双端重写部分核心。

**Files:**
- Create: `docs/plans/YYYY-MM-DD-native-core-rewrite-assessment.md`

### Task 13.1: 评估是否继续原生化核心

**Step 1:** 对比 Native 版与目标指标。

**Step 2:** 识别 Go daemon 中真实资源热点。

**Step 3:** 如果热点来自业务核心，再评估 Swift/C# 双实现成本。

**Step 4:** 不允许在没有契约测试覆盖的情况下重写核心业务。

## 最终验收命令

迁移完成前，每个 PR 至少运行相关子集；切换默认发布前运行完整集合：

```bash
go test ./core/... ./cmd/skillflow
```

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

```bash
xcodebuild -project native/macos/SkillFlow/SkillFlow.xcodeproj -scheme SkillFlow test
```

```bash
dotnet test native/windows/SkillFlow/SkillFlow.sln
```

```bash
make build
```

发布切换后，`make build` 应构建 Native 默认产物；Legacy Wails 构建应保留显式命令直到回滚窗口结束。

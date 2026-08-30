# 原生平台发布检查清单

## 目的

本文定义 SkillFlow 从 Wails/React 生产 UI 切换到 macOS Swift 原生客户端和 Windows WinUI 原生客户端前必须通过的阻塞检查。任何默认发布切换都必须先清空本文的 `Blocked` 和 `Pending` 项。

> **2026-08-30 状态：本地源码构建切换已完成。** `make` / `make build` 现在构建当前平台 Native 产物，macOS 本地构建与启动冒烟已通过；Wails legacy 保留为显式回退命令。本文描述的正式发布切换仍未完成，签名分发、性能数据、完整功能契约、双平台 CI 产物和回滚验证仍处于 Pending。

## 状态标记

| 标记 | 含义 |
|------|------|
| `Pending` | 尚未开始或尚未验证 |
| `Blocked` | 有明确阻塞，不能发布 |
| `Passed` | 已通过并有证据 |
| `N/A` | 不适用于当前发布类型 |

## 总体发布门槛

| 检查项 | 状态 | 证据位置 | 说明 |
|--------|------|----------|------|
| 功能契约已冻结 | Passed | `docs/plans/2026-04-25-native-platform-feature-contract.md` | Batch 0 已建立迁移矩阵 |
| 性能基线记录已建立 | Passed | `docs/plans/2026-04-25-native-platform-performance-baseline.md` | GUI 数值仍需补录后才能切换默认发布 |
| macOS Wails baseline GUI 指标已补录 | Pending | 性能基线文档 | 默认发布切换前必须补录 |
| Windows Wails baseline GUI 指标已补录 | Pending | 性能基线文档 | 默认发布切换前必须补录 |
| macOS Native Preview GUI 指标已补录 | Pending | 性能基线文档 | 默认发布切换前必须补录 |
| Windows Native Preview GUI 指标已补录 | Pending | 性能基线文档 | 默认发布切换前必须补录 |
| 所有功能面 macOS 原生验收通过 | Pending | 功能契约文档 | 不能只完成单平台 |
| 所有功能面 Windows 原生验收通过 | Pending | 功能契约文档 | 不能只完成单平台 |
| daemon API 契约测试覆盖所有迁移功能 | Pending | Go 测试输出 | 不能由 UI 直接读写业务文件 |
| 配置和数据格式兼容性已验证 | Pending | 本文“数据兼容”章节 | 旧数据目录必须可直接使用 |
| 备份恢复兼容性已验证 | Pending | 本文“备份恢复”章节 | 云备份恢复后必须可用 |
| 安装包签名和分发验证通过 | Pending | CI release job | macOS 和 Windows 均需覆盖 |

## 功能契约检查

默认发布切换前，以下功能面必须全部在 `docs/plans/2026-04-25-native-platform-feature-contract.md` 中标记为 `Done`：

- Navigation & Shell
- My Skills
- Legacy Push Route 或等价深链兼容策略
- Legacy Pull Route 或等价深链兼容策略
- Starred Repos
- Cloud Backup
- Settings
- Skill Card
- Skill Tooltip
- Shared Dialogs
- Backend Events
- App Update Dialog
- My Agents
- My Prompts
- My Memory

验收规则：

- macOS 和 Windows 必须同批达到 `Done`。
- daemon API 必须有契约测试。
- 如果某项只属于平台 UI，daemon API 可标为 `N/A`，但必须说明原因。
- 不允许使用“功能相似”代替“行为一致”。

## 性能检查

默认发布切换前，以下指标必须同时有 Wails baseline 和 Native Preview 数据：

| 平台 | 冷启动到可交互 | UI 进程 RSS | daemon RSS | Dashboard 首屏 | 后台任务影响 |
|------|----------------|-------------|------------|----------------|--------------|
| macOS | Pending | Pending | Pending | Pending | Pending |
| Windows | Pending | Pending | Pending | Pending | Pending |

最低要求：

- 关闭窗口后不得保留 WebView、JS runtime 或 React 页面树。
- Native UI 打开时的 RSS 必须明显低于 Wails UI，或者给出无法降低的技术解释。
- `daemon` RSS 不得因为原生 UI 迁移而增加常驻大对象。
- 仓库刷新、Agent 扫描、云备份期间 UI 不能阻塞主线程。

## 数据兼容检查

| 数据 | 状态 | 验收方式 |
|------|------|----------|
| `config.json` | Pending | 旧版本生成，新版本读取和保存后字段语义不变 |
| `config_local.json` | Pending | 本机路径、密钥、窗口状态、代理设置仍保持本机语义 |
| `star_repos.json` | Pending | 仓库收藏和 sync-safe 元数据保留 |
| `star_repos_local.json` | Pending | 本地 clone 状态不参与同步 |
| `skills/` | Pending | 已安装技能可列出、更新、删除、推送 |
| `meta/` | Pending | 技能 metadata 和逻辑身份保持 |
| `meta_local/` | Pending | 本地更新检查状态保持 local-only |
| `prompts/` | Pending | 提示词内容、图片和链接可读写 |
| `memory/` | Pending | 主记忆和模块记忆可读写、推送、备份 |
| `cache/viewstate/` | Pending | 可忽略并重建，不能作为真相来源 |
| `runtime/` | Pending | endpoint、token、PID 仍为 local-only |

验收规则：

- 原生客户端不能直接读写这些业务文件。
- 所有读写必须通过 Go `daemon`。
- 如果 on-disk schema 改变，必须同步更新 `docs/config.md` 和 `docs/config_zh.md`。

## 备份恢复检查

| 检查项 | 状态 | 说明 |
|--------|------|------|
| Git provider 手动备份 | Pending | 包含冲突检测和选择本地/远端 |
| Git provider 恢复 | Pending | 恢复后技能、提示词、记忆、配置可用 |
| Object storage provider 手动备份 | Pending | AWS、Aliyun、Azure、Google、Huawei、Tencent |
| Object storage provider 恢复 | Pending | Secret 仍只来自 local config |
| 自动备份 | Pending | mutation 后备份和定时备份语义一致 |
| 旧 Wails 备份包恢复到 Native | Pending | 不能要求用户手动迁移 |
| Native 备份包回滚到 Wails legacy | Pending | 回滚窗口内必须验证 |

## Agent 行为检查

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 内置 Agent 默认路径 | Pending | claude-code、codex、gemini-cli、opencode、openclaw |
| 自定义 Agent 创建和删除 | Pending | 配置仍拆分 shared/local |
| ScanDirs 扫描 | Pending | 递归深度、大小写、嵌套技能识别一致 |
| PushDir 推送 | Pending | 缺失目录确认、覆盖/跳过冲突一致 |
| Pull 到 My Skills | Pending | 分类、冲突、逻辑身份一致 |
| Agent 侧删除 | Pending | 只删除目标 Agent 副本 |
| 记忆推送 | Pending | merge/takeover/模块规则一致 |
| Agent 侧记忆预览 | Pending | 主记忆和 rules 文件可验证 |

## 平台发布检查

### macOS

| 检查项 | 状态 | 说明 |
|--------|------|------|
| Swift 原生客户端 build | Pending | SwiftPM release build 与 `make build-native-macos` 在发布 CI 中通过 |
| Swift 原生客户端测试 | Pending | XCTest 通过 |
| `skillflowd` 打包进 `.app` | Pending | 本地打包与启动冒烟已通过；发布 CI 仍需复验 |
| 菜单栏和 Quit 行为 | Pending | 不遗留后台进程 |
| DMG 创建 | Pending | 包含 `.app` 和 Applications 链接 |
| codesign | Pending | 正式发布必须签名 |
| Gatekeeper 验证 | Pending | 不出现 damaged |
| Intel 产物 | Pending | `x86_64` |
| Apple Silicon 产物 | Pending | `arm64` |

### Windows

| 检查项 | 状态 | 说明 |
|--------|------|------|
| WinUI Debug build | Pending | `dotnet build` 通过 |
| WinUI 测试 | Pending | `dotnet test` 通过 |
| `skillflowd` 打包进安装目录 | Pending | 路径稳定 |
| 通知区域图标和 Exit 行为 | Pending | 不遗留后台进程 |
| Installer 或发布包 | Pending | 用户可安装或直接运行 |
| 签名 | Pending | 正式发布必须签名 |
| Windows Defender / SmartScreen 风险记录 | Pending | 发布说明中记录 |

## CI 检查

默认发布切换前，CI 必须覆盖：

```bash
go test ./core/... ./cmd/skillflow
```

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

```bash
./native/scripts/build-macos.sh
swift test --package-path native/macos/SkillFlow
```

```bash
dotnet test native/windows/SkillFlow/SkillFlow.sln
```

当前本地构建已经完成默认切换：

- `make build` 构建 Native 默认产物。
- Wails legacy 使用显式命令，例如 `make build-legacy`。
- CI artifact 名称必须区分 `native` 和 `legacy`，直到回滚窗口结束。

## 回滚检查

正式切换默认发布前，必须确认：

- 用户可以从 Native 回滚到最近一个 Wails legacy 版本。
- 回滚后 `config*.json`、`skills/`、`prompts/`、`memory/` 不损坏。
- 如果 Native 引入 schema cutover，必须先实现 `core/platform/upgrade/` 启动切换，并在配置文档中说明。
- 发布说明中明确 Native 和 legacy 的兼容边界。

## 发布决策

只有当以下条件全部成立，才能把 Native 设为默认发布：

1. 本文所有阻塞项为 `Passed` 或有明确 `N/A` 说明。
2. 功能契约所有功能面为 `Done`。
3. 性能基线中双平台 Wails 和 Native 数据均已补录。
4. CI 完整通过。
5. macOS 和 Windows 安装包均通过手工安装、启动、关闭、重开、卸载验证。
6. 回滚路径已验证。

如果任一条件不满足，只能发布 Native Preview，不能替代正式 Wails 产物。

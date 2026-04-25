# 原生平台性能基线

## 目的

本文记录原生化重构前的 Wails/React 版本资源基线，以及后续 Native Preview 和 Native 默认发布的对照指标。性能目标必须用实测数据判断，不能用主观感受替代。

## 基线版本

| 项 | 值 |
|----|----|
| Git commit | `5e15413` |
| 分支 | `docs/native-platform-batch0` |
| 基线类型 | Wails/React legacy baseline |
| 记录日期 | 2026-04-26 |
| 记录范围 | 文档、命令和可补录表格；GUI 指标待在对应平台实测 |

## 当前可用测试环境

| 项 | 值 |
|----|----|
| 系统 | macOS 15.7.2 (`24G325`) |
| 架构 | `x86_64` |
| CPU | 当前沙箱禁止读取 `machdep.cpu.brand_string` |
| Go | `go version go1.26.0 darwin/amd64` |
| Node | `v24.13.1` |
| 前端依赖状态 | worktree 内已执行 `npm install` |

Windows 指标必须在 Windows 机器或 Windows CI runner 上补录。本文不会用 macOS 数据推断 Windows 数据。

## 指标定义

| 指标 | 定义 | 采样方式 |
|------|------|----------|
| 冷启动到可交互 | 从用户启动应用到主窗口可点击 Dashboard 控件 | 手动计时或平台自动化脚本 |
| UI 进程 RSS | 主窗口打开且 Dashboard 首屏稳定后，UI 进程 RSS | `ps` / 任务管理器 / PowerShell |
| daemon RSS | 关闭窗口后，只保留后台进程时的 RSS | `ps` / 任务管理器 / PowerShell |
| Dashboard 首屏加载 | 应用窗口出现后到 Dashboard 技能卡或空状态稳定 | 手动计时或日志打点 |
| 大列表滚动 | 超过 100 个技能卡时滚动是否掉帧或卡顿 | 手工验收和截图/录屏记录 |
| 后台任务影响 | 仓库刷新、Agent 扫描、备份期间 UI 是否阻塞 | 手工验收和日志时间线 |

## macOS Wails 基线

### 构建命令

```bash
cd /Users/shinerio/Workspace/code/SkillFlow
npm --prefix cmd/skillflow/frontend install
npm --prefix cmd/skillflow/frontend run build
make build
```

如只需生成 `frontend/dist` 后运行 Go 测试：

```bash
cd /Users/shinerio/Workspace/code/SkillFlow
npm --prefix cmd/skillflow/frontend run build
go test ./core/... ./cmd/skillflow
```

### 进程识别命令

```bash
ps -axo pid,ppid,rss,command | grep -E 'SkillFlow|skillflow' | grep -v grep
```

### 待补录指标

| 指标 | Wails baseline | Native Preview | Native default | 备注 |
|------|----------------|----------------|----------------|------|
| 冷启动到可交互 | 待补录 | 待补录 | 待补录 | 需要 GUI 运行 |
| UI 进程 RSS | 待补录 | 待补录 | 待补录 | Dashboard 稳定后采样 |
| daemon RSS | 待补录 | 待补录 | 待补录 | 关闭窗口后采样 |
| Dashboard 首屏加载 | 待补录 | 待补录 | 待补录 | 使用同一数据目录 |
| 大列表滚动 | 待补录 | 待补录 | 待补录 | 记录技能数量 |
| 后台任务影响 | 待补录 | 待补录 | 待补录 | 仓库刷新、Agent 扫描、备份 |

## Windows Wails 基线

### 构建命令

```powershell
npm --prefix cmd/skillflow/frontend install
npm --prefix cmd/skillflow/frontend run build
make build
```

如果 `make` 不可用，使用当前发布 workflow 的 Wails 构建命令：

```powershell
cd cmd/skillflow
wails build -platform windows/amd64
```

### 进程识别命令

```powershell
Get-Process | Where-Object { $_.ProcessName -like '*SkillFlow*' } |
  Select-Object Id, ProcessName, WorkingSet64, CPU, StartTime
```

### 待补录指标

| 指标 | Wails baseline | Native Preview | Native default | 备注 |
|------|----------------|----------------|----------------|------|
| 冷启动到可交互 | 待补录 | 待补录 | 待补录 | Windows 实机或 runner |
| UI 进程 RSS | 待补录 | 待补录 | 待补录 | Dashboard 稳定后采样 |
| daemon RSS | 待补录 | 待补录 | 待补录 | 关闭窗口后采样 |
| Dashboard 首屏加载 | 待补录 | 待补录 | 待补录 | 使用同一数据目录 |
| 大列表滚动 | 待补录 | 待补录 | 待补录 | 记录技能数量 |
| 后台任务影响 | 待补录 | 待补录 | 待补录 | 仓库刷新、Agent 扫描、备份 |

## 数据目录控制

为保证前后对比有效，测量时必须记录：

- `<AppDataDir>` 路径。
- 技能数量。
- 提示词数量。
- 模块记忆数量。
- Agent 数量。
- Starred repo 数量。
- Repo cache 是否已预热。
- 云备份 provider 是否启用。

推荐记录模板：

| 项 | 值 |
|----|----|
| `<AppDataDir>` | 待补录 |
| 技能数量 | 待补录 |
| 提示词数量 | 待补录 |
| 模块记忆数量 | 待补录 |
| Agent 数量 | 待补录 |
| Starred repo 数量 | 待补录 |
| Repo cache 状态 | 冷 / 热 |
| 云备份状态 | 关闭 / 开启 |

## 测量流程

1. 使用同一 commit 构建 Wails baseline。
2. 清理或记录当前 `runtime/` 状态。
3. 首次启动应用，记录冷启动到可交互时间。
4. Dashboard 稳定后记录 UI 进程 RSS。
5. 关闭窗口到托盘/菜单栏，等待 5 秒。
6. 记录后台 daemon RSS。
7. 重新打开窗口，记录二次打开到可交互时间。
8. 在 My Skills、My Agents、Starred Repos、My Prompts、My Memory 和 Cloud Backup 页面各停留一次，记录明显卡顿或阻塞。
9. 执行一次仓库刷新、Agent 扫描和备份，记录 UI 是否阻塞。
10. 将同样流程用于 Native Preview 和最终 Native 默认发布。

## 已完成的非 GUI 基线验证

在 `.worktrees/native-platform-batch0` 中已完成以下非 GUI 验证：

```bash
npm install
npm run test:unit
npm run build
go test ./core/... ./cmd/skillflow
```

结果摘要：

- `npm run test:unit`：101 个测试通过。
- `npm run build`：前端生产构建成功，生成 `frontend/dist`。
- `go test ./core/... ./cmd/skillflow`：通过；第一次直接运行失败是因为 `frontend/dist` 尚未生成，按项目 CI 流程先构建前端后通过。

## 发布阻塞规则

默认发布切换到 Native 前，以下字段不能再是 `待补录`：

- macOS Wails baseline 的冷启动、UI RSS、daemon RSS、Dashboard 首屏加载。
- Windows Wails baseline 的冷启动、UI RSS、daemon RSS、Dashboard 首屏加载。
- macOS Native Preview 对应指标。
- Windows Native Preview 对应指标。

如果某个平台无法实测，应将发布切换标记为 `Blocked`，不能用另一平台数据替代。

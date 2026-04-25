# 原生平台重构设计

## 背景

当前 SkillFlow 是 Go + Wails v2 + React/Vite 桌面应用。业务核心已经按 DDD 风格拆在 `core/` 下，桌面壳、Wails 绑定、进程生命周期、托盘/菜单栏和构建发布集中在 `cmd/skillflow/`。

现有跨平台开发体验和运行体验的主要问题不是单个页面实现，而是技术栈边界：

- UI 依赖 WebView、React 和 Wails，无法充分获得 macOS 与 Windows 的系统原生控件、菜单、窗口和辅助功能体验。
- 当前开发同时受 Go、Wails、Node、TypeScript、Vite、WebView 差异影响，定位问题链路较长。
- 项目已经做过 `daemon` / `ui` 进程分离，但生产 UI 仍然是 WebView 进程；关闭窗口后能释放前端内存，打开窗口时仍要承担 Web 技术栈成本。
- 用户目标是保持功能不变，同时追求平台原生体验、流畅度和更低资源占用。

本设计采用第一性原理拆解：SkillFlow 真正必须共享的是业务语义、配置格式、同步格式和 Agent/Skill/Prompt/Memory 行为；不必须共享的是 UI 技术栈。为了可落地迁移，应先稳定共享业务边界，再分别实现平台原生客户端。

## 目标

- 保留 macOS 与 Windows 双系统支持。
- 生产 UI 从 Wails/React/WebView 迁移到平台原生实现：
  - macOS 使用 Swift + SwiftUI/AppKit。
  - Windows 使用 C# + .NET + WinUI 3 / Windows App SDK。
- 所有现有用户功能保持不变，以 `docs/features.md` / `docs/features_zh.md` 为功能基线。
- 保持现有数据兼容：
  - `config.json`
  - `config_local.json`
  - `star_repos.json`
  - `star_repos_local.json`
  - `skills/`
  - `meta/`
  - `meta_local/`
  - `prompts/`
  - `cache/viewstate/`
  - `runtime/`
- 先复用现有 Go 业务核心，避免 Swift 与 C# 双端重复实现复杂业务规则。
- 支持分批、多次合并，每批都能独立验证和回滚。
- 最终默认发布原生客户端，不再把 Wails/React 作为生产 UI。

## 非目标

- 第一阶段不把全部业务核心同时重写成 Swift 和 C#。
- 第一阶段不改变云备份格式、技能格式、提示词格式、记忆文件格式或 Agent 配置语义。
- 第一阶段不新增远程网络服务；所有客户端到核心服务的通信只限本机。
- 第一阶段不追求关闭窗口后保留页面级临时状态，仍允许重新打开时冷启动页面。
- 不为了迁移引入 Electron、Tauri 或新的跨平台 UI 框架。

## 方案选择

### 推荐方案：原生客户端 + 共享本地核心服务

macOS 与 Windows 分别实现原生客户端，现有 Go `core/` 业务能力逐步收敛到本地 `daemon` 服务。原生客户端通过稳定的本地契约调用 `daemon`。

优点：

- UI 能真正贴近平台原生体验。
- Go 业务核心只保留一份，功能一致性风险低。
- 可以先让现有 Wails UI 继续作为对照基线，再逐页替换。
- 每批迁移都可测试、可回滚。

缺点：

- 进程模型仍包含原生 UI 进程和 Go `daemon` 进程。
- 最终二进制和安装包需要同时处理原生客户端与核心服务。

### 备选方案：原生客户端 + Go 动态库

将 Go 核心编译成动态库，通过 Swift/C# FFI 调用。

该方案减少一个常驻服务进程，但 ABI、线程模型、回调、崩溃隔离、签名和升级复杂度更高，不适合作为第一阶段。

### 备选方案：Swift 与 C# 双端完整重写

macOS 和 Windows 分别重写全部业务核心。

该方案原生纯度最高，但 Git、云备份、Agent 扫描、路径规则、冲突处理和同步语义都要双实现。除非共享核心服务仍无法满足资源目标，否则不应作为初始路线。

## 目标架构

```text
SkillFlow
  native/macos/SkillFlow/        # Swift + SwiftUI/AppKit 原生客户端
  native/windows/SkillFlow/      # C# + WinUI 3 原生客户端
  cmd/skillflowd/                # Go daemon 组合根，长期承载业务核心
  cmd/skillflow/                 # 迁移期 Wails 旧 UI，对照基线，最终移除或归档
  core/                          # 现有业务核心、平台能力、编排和读模型
  docs/
    plans/
    architecture/
```

### Go `daemon`

`daemon` 是本机核心服务，负责：

- 配置加载、升级和保存。
- 技能、提示词、记忆、Agent、仓库收藏和云备份业务能力。
- Git、云存储、Agent 文件系统、外部编辑器等适配器。
- 后台任务、自动更新、自动备份和事件发布。
- 本机 IPC endpoint 管理、鉴权和契约版本协商。
- 托盘/菜单栏是否由 `daemon` 承担，需要按平台实现阶段决定；最终推荐由原生客户端承担用户可见菜单，由 `daemon` 只承担后台服务。

`daemon` 不负责：

- 渲染 UI。
- 保留页面树或前端状态。
- 平台原生窗口、菜单、控件布局。

### macOS 原生客户端

macOS 客户端负责：

- SwiftUI/AppKit 主窗口。
- 菜单栏项和系统菜单。
- 原生文件夹选择器、打开路径、外部编辑器。
- macOS 启动项、通知和应用生命周期。
- 与 `daemon` 的本机契约调用和事件订阅。

### Windows 原生客户端

Windows 客户端负责：

- WinUI 3 主窗口。
- 通知区域图标、窗口显示和退出。
- 原生文件夹选择器、打开路径、外部编辑器。
- Windows 启动项、通知和应用生命周期。
- 与 `daemon` 的本机契约调用和事件订阅。

### 本机契约层

契约层要独立于 UI 技术栈和具体 IPC 传输。推荐先定义稳定 JSON 请求/响应模型，再按阶段选择传输：

1. 迁移早期优先复用现有 loopback IPC 能力，减少首批风险。
2. 原生客户端稳定后，macOS 可迁移到 Unix domain socket，Windows 可迁移到 Named Pipe。
3. 契约字段保持一致，传输替换不影响业务 API。

契约约束：

- 只允许本机访问。
- endpoint、token、进程 ID 和协议版本写入 `<AppDataDir>/runtime/` 的本地文件。
- token 必须每次 daemon 启动重新生成。
- 错误响应必须包含稳定错误码和可本地化消息 key。
- 大型列表查询必须支持轻量摘要和必要时分页或增量刷新。

## 功能边界

以现有页面作为迁移切片：

| 切片 | 核心能力 | 迁移顺序 |
|------|----------|----------|
| Shell / Settings | 导航、语言、主题、日志、代理、启动项、更新检查 | 1 |
| My Skills | 技能列表、分类、导入、删除、移动、更新、手动推送、自动推送 | 2 |
| My Agents | Agent 配置、扫描、推送、拉取、冲突处理、路径管理 | 3 |
| Starred Repos | 仓库添加、认证、刷新、候选技能、导入、批量操作 | 4 |
| My Prompts | 提示词分类、编辑、导入、导出、媒体和链接 | 5 |
| My Memory | 主记忆、模块记忆、批量推送、自动同步、外部编辑 | 6 |
| Cloud Backup | 云配置、备份、恢复、冲突、变更列表 | 7 |

每个切片必须同时满足：

- Go daemon 契约测试通过。
- macOS 原生实现通过手工验收。
- Windows 原生实现通过手工验收。
- 行为与旧 Wails UI 对齐。
- 如果是用户可见功能变化，更新 `docs/features.md` 和 `docs/features_zh.md`。

## 数据兼容设计

第一阶段必须保持现有磁盘格式不变。原生客户端不直接读写业务数据文件，只通过 `daemon` 访问。这样可以避免 Swift/C# 复制路径规范、升级逻辑和备份过滤规则。

仍由 Go 核心持有：

- 路径持久化规则。
- `config.json` / `config_local.json` 拆分语义。
- `star_repos*.json` 同步状态。
- `meta/` 与 `meta_local/` 技能元数据。
- `prompts/` 提示词目录结构。
- `runtime/` 本地进程状态。
- `logs/` 日志轮转规则。

## 发布策略

迁移期间保留两类产物：

- Legacy Wails 构建：用于对照、回滚和验证。
- Native Preview 构建：用于原生客户端逐步验收。

当全部功能切片完成后：

- 默认发布切换到 Native。
- Wails UI 停止新增功能，仅保留短期回滚分支。
- CI 产物变为：
  - macOS Apple Silicon DMG。
  - macOS Intel DMG。
  - Windows 安装包或单文件可执行包。
  - 内嵌或伴随的 `skillflowd`。

## 验收标准

### 功能验收

- `docs/features.md` 中所有功能在 macOS 和 Windows 原生客户端可用。
- 旧数据目录可直接被新客户端使用，无需手工迁移。
- 云备份恢复后的数据可被新客户端正确读取。
- Agent 扫描、推送、拉取、冲突处理行为与旧版本一致。

### 性能验收

每个平台记录以下指标并进入发布检查：

- 冷启动到主窗口可交互时间。
- Dashboard 首屏加载时间。
- 关闭窗口后常驻 `daemon` RSS。
- 打开窗口时 UI 进程 RSS。
- 大列表页面滚动流畅度。
- 备份、仓库刷新、Agent 扫描期间 UI 是否阻塞。

目标方向：

- 关闭窗口后不保留 WebView、JS runtime 或 React 页面树。
- `daemon` 常驻内存只保留业务运行所需状态。
- 原生 UI 页面进入不依赖前端 bundle 解析和 WebView 初始化。

### 工程验收

- 新增原生代码放在 `native/` 下。
- Go 业务代码仍遵守 `core/`、`cmd/` 边界。
- 根目录不新增 Go 源文件。
- Markdown 文档使用简体中文。
- 每个切片独立提交，不混入无关重构。

## 风险与控制

### 风险：原生 UI 与旧 UI 行为不一致

控制方式：

- 先建立功能矩阵和契约测试。
- 每个切片都用旧 Wails UI 做对照。
- 复杂流程保留截图或录屏验收记录。

### 风险：daemon API 过早设计过大

控制方式：

- 按页面切片扩展契约。
- 先覆盖现有 `App` 方法等价能力，不预设未来抽象。
- 大对象查询先保持简单，只有遇到真实性能问题再分页或增量化。

### 风险：两个原生客户端进度不一致

控制方式：

- 每个功能切片要求 macOS 和 Windows 同批完成。
- 不允许某个平台长期停留在缺失功能状态。
- UI 细节可以平台化，业务能力和错误语义必须一致。

### 风险：发布和签名复杂度上升

控制方式：

- 先做 Native Preview，不立即替换正式包。
- 先跑通无签名或开发签名构建，再接入正式签名。
- daemon 与客户端的打包关系在发布切换前固定下来。

## 推荐实施节奏

1. 先冻结功能契约和性能基线。
2. 把现有 Go 核心服务化，形成稳定本机契约。
3. 分别建立 macOS 和 Windows 原生空壳。
4. 按业务切片迁移页面。
5. 切换默认发布产物。
6. 观察资源指标，再决定是否继续把部分核心业务原生化。

该路线不追求一次性重写，而是把风险拆在可验证的边界上。核心原则是：先保持行为不变，再替换用户可见技术栈，最后再讨论业务核心是否值得双端原生化。

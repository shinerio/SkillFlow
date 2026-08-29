# 原生平台功能契约

## 目的

本文把当前 Wails/React 版本的用户能力冻结为原生化重构的行为契约。后续 macOS Swift 客户端、Windows WinUI 客户端和 Go `daemon` 契约都必须以本文为迁移核对表。

功能细节的唯一来源仍是：

- 英文功能文档：`docs/features.md`
- 中文功能文档：`docs/features_zh.md`
- 当前后端入口：`cmd/skillflow/app*.go` 中生产 `App` 导出方法

本文不替代功能文档。它只回答三个问题：

1. 哪些用户能力必须保留。
2. 哪些能力需要 daemon API 支撑。
3. 每个平台迁移完成前要核对哪些验收项。

## 状态标记

| 标记 | 含义 |
|------|------|
| `Baseline` | 当前 Wails/React 版本已具备，用作行为对照 |
| `Pending` | 原生客户端或 daemon API 尚未实现 |
| `N/A` | 该平台或层不直接承担此能力 |
| `Blocked` | 被外部环境、签名、平台机器或未决设计阻塞 |
| `Done` | 已通过本契约对应验收 |

## 功能迁移矩阵

| 功能面 | Wails baseline | macOS Native | Windows Native | daemon API | 验收重点 |
|--------|----------------|---------------|-----------------|------------|----------|
| Navigation & Shell | Baseline | Pending | Pending | Partial | 导航、语言、主题、反馈、关闭/重开、托盘/菜单栏、窗口尺寸 |
| My Skills | Baseline | Pending | Pending | Done | 列表、分类、搜索、排序、导入、删除、移动、更新、手动推送、自动推送 |
| Legacy Push Route | Baseline | N/A | N/A | N/A | 旧路由或深链必须有兼容策略 |
| Legacy Pull Route | Baseline | N/A | N/A | N/A | 旧路由或深链必须有兼容策略 |
| Starred Repos | Baseline | Pending | Pending | Pending | 仓库添加、认证、刷新、候选技能、批量导入、批量推送 |
| Cloud Backup | Baseline | Pending | Pending | Pending | Provider 配置、备份、恢复、Git 冲突、变更列表 |
| Settings | Baseline | Pending | Pending | Done | Agents、Cloud、General、Proxy、保存、启动项、日志 |
| Skill Card | Baseline | Pending | Pending | Pending | 状态语义、操作入口、溢出、右键菜单、拖放 |
| Skill Tooltip | Baseline | Pending | Pending | Pending | 延迟、定位、内容、可读性 |
| Shared Dialogs | Baseline | Pending | Pending | Pending | 覆盖/跳过冲突、缺失目录确认、错误展示 |
| Backend Events | Baseline | Pending | Pending | Pending | 更新事件、备份事件、记忆事件、订阅清理 |
| App Update Dialog | Baseline | Pending | Pending | Pending | 检查、下载、应用、跳过版本、平台差异 |
| My Agents | Baseline | Pending | Pending | Pending | Agent 列表、扫描、PushDir、ScanDirs、手动拉取、删除、记忆预览 |
| My Prompts | Baseline | Pending | Pending | Pending | 分类、卡片、编辑器、导入、导出、媒体、链接 |
| My Memory | Baseline | Pending | Pending | Pending | 主记忆、模块记忆、批量推送、自动同步、外部编辑、Agent 侧校验 |

## 页面级验收清单

### 1. Navigation & Shell

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 固定侧边栏包含四个核心工作面和底部工具区 | Baseline | Pending | Pending | N/A |
| 语言切换能立即生效并持久化 | Baseline | Pending | Pending | Pending |
| 主题切换能立即生效并持久化 | Baseline | Pending | Pending | Pending |
| Feedback 打开 GitHub issue 页面 | Baseline | Pending | Pending | N/A |
| 关闭窗口释放 UI 进程，保留后台能力 | Baseline | Pending | Pending | Pending |
| Show SkillFlow 可重新打开窗口 | Baseline | Pending | Pending | Pending |
| Quit/Exit 同时结束 UI 与后台进程 | Baseline | Pending | Pending | Pending |
| 启动恢复本机窗口尺寸并限制在可用屏幕内 | Baseline | Pending | Pending | Pending |

### 2. My Skills

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 实时搜索、A-Z/Z-A 排序和分类过滤 | Baseline | Pending | Pending | Done |
| 分类创建、重命名、删除和空分类限制 | Baseline | Pending | Pending | Pending |
| 本地文件夹导入技能 | Baseline | Pending | Pending | Done |
| OS 文件夹拖放导入技能 | Baseline | Pending | Pending | N/A |
| 单个删除和批量删除技能 | Baseline | Pending | Pending | Done |
| 拖放或菜单移动技能分类 | Baseline | Pending | Pending | Done |
| 检查 Git-backed 技能更新 | Baseline | Pending | Pending | Done |
| 单卡更新并刷新已有 Agent 副本 | Baseline | Pending | Pending | Done |
| 自动更新开关本机持久化 | Baseline | Pending | Pending | Done |
| 自动推送目标本机持久化并可回填 | Baseline | Pending | Pending | Done |
| 手动推送选择可见技能和目标 Agent | Baseline | Pending | Pending | Done |
| 缺失推送目录确认和冲突覆盖/跳过 | Baseline | Pending | Pending | Done |

### 3. Legacy Push Route

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| `/sync/push` 兼容跳转或深链兼容策略明确 | Baseline | Pending | Pending | N/A |
| 手动推送能力实际入口在 My Skills | Baseline | Pending | Pending | Pending |

### 4. Legacy Pull Route

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| `/sync/pull` 兼容跳转或深链兼容策略明确 | Baseline | Pending | Pending | N/A |
| 手动拉取能力实际入口在 My Agents | Baseline | Pending | Pending | Pending |

### 5. Starred Repos

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| Folder/Flat 两种视图 | Baseline | Pending | Pending | Pending |
| 仓库添加、删除、打开浏览器 | Baseline | Pending | Pending | Pending |
| HTTP 凭证添加仓库 | Baseline | Pending | Pending | Pending |
| SSH 认证错误可读展示 | Baseline | Pending | Pending | Pending |
| 单仓库刷新和全部刷新 | Baseline | Pending | Pending | Pending |
| 仓库详情面包屑和技能搜索排序 | Baseline | Pending | Pending | Pending |
| 批量选择、批量导入、批量推送 | Baseline | Pending | Pending | Pending |
| 内置 starter repos 仅首次初始化 | Baseline | Pending | Pending | Pending |
| Repo cache 与已安装技能更新语义一致 | Baseline | Pending | Pending | Pending |

### 6. Cloud Backup

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| Provider 列表与当前配置展示 | Baseline | Pending | Pending | Pending |
| 手动备份、恢复、远端文件列表 | Baseline | Pending | Pending | Pending |
| 备份变更列表和上次完成时间 | Baseline | Pending | Pending | Pending |
| Git provider 冲突状态和选择本地/远端 | Baseline | Pending | Pending | Pending |
| 自动备份间隔和 mutation 后备份语义 | Baseline | Pending | Pending | Pending |
| Secret 只落入本地配置 | Baseline | Pending | Pending | Pending |

### 7. Settings

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| Agents Tab 可编辑 Agent 路径和启用状态 | Baseline | Pending | Pending | Done |
| Cloud Tab 可编辑 provider 配置和本地密钥 | Baseline | Pending | Pending | Done |
| General Tab 可编辑日志级别、扫描深度、仓库缓存目录 | Baseline | Pending | Pending | Done |
| Proxy Tab 可配置代理并测试连接 | Baseline | Pending | Pending | Done |
| Save 会分发到对应上下文和本地设置 | Baseline | Pending | Pending | Done |
| 打开 AppData、日志目录、Git backup 目录 | Baseline | Pending | Pending | Done |

### 8. Skill Card

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| Installed/Imported/Pushed/Seen/Updatable 状态语义一致 | Baseline | Pending | Pending | Pending |
| 卡片操作入口与上下文菜单一致 | Baseline | Pending | Pending | Pending |
| Agent 状态溢出使用稳定 `+N` 展示 | Baseline | Pending | Pending | N/A |
| 更新中禁用重复点击并展示进度 | Baseline | Pending | Pending | Pending |

### 9. Skill Tooltip

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 300ms hover 后展示 | Baseline | Pending | Pending | N/A |
| 展示技能元数据、描述和源信息 | Baseline | Pending | Pending | Pending |
| 边缘定位不溢出窗口 | Baseline | Pending | Pending | N/A |

### 10. Shared Dialogs

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| Push/Pull/Import 冲突可选择 overwrite/skip | Baseline | Pending | Pending | Pending |
| 缺失目录确认不会静默创建 | Baseline | Pending | Pending | Pending |
| 错误消息不暴露密钥 | Baseline | Pending | Pending | Pending |

### 11. Backend Events

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 技能更新事件可刷新 Dashboard 和 My Agents | Baseline | Pending | Pending | Pending |
| 记忆事件可刷新 Memory 页面和 Agent 记忆预览 | Baseline | Pending | Pending | Pending |
| 备份和恢复事件可展示进度或结果 | Baseline | Pending | Pending | Pending |
| 页面销毁时取消事件订阅 | Baseline | Pending | Pending | Pending |

### 12. App Update Dialog

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 手动检查和启动检查共用语义 | Baseline | Pending | Pending | Pending |
| 下载、应用、跳过版本行为一致 | Baseline | Pending | Pending | Pending |
| 平台下载产物选择正确 | Baseline | Pending | Pending | Pending |
| 版本注入来源可追踪 | Baseline | Pending | Pending | Pending |

### 13. My Agents

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| Agent 列表和启用状态展示 | Baseline | Pending | Pending | Pending |
| Skills/Memory 面板切换 | Baseline | Pending | Pending | Pending |
| PushDir 技能浏览、删除和打开路径 | Baseline | Pending | Pending | Pending |
| ScanDirs 扫描和扫描结果搜索排序 | Baseline | Pending | Pending | Pending |
| 手动拉取、选择未导入、冲突处理 | Baseline | Pending | Pending | Pending |
| 记忆预览能读取主记忆和 rules 文件 | Baseline | Pending | Pending | Pending |

### 14. My Prompts

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 提示词列表、搜索、分类和排序 | Baseline | Pending | Pending | Pending |
| 分类创建、重命名、删除 | Baseline | Pending | Pending | Pending |
| 提示词创建、编辑、移动、删除 | Baseline | Pending | Pending | Pending |
| 编辑器支持描述、正文、图片 URL、Web 链接 | Baseline | Pending | Pending | Pending |
| 导入预览、冲突选择、完成和取消 | Baseline | Pending | Pending | Pending |
| 全量导出和指定导出 | Baseline | Pending | Pending | Pending |

### 15. My Memory

| 验收项 | Wails baseline | macOS Native | Windows Native | daemon API |
|--------|----------------|---------------|-----------------|------------|
| 主记忆读取、编辑、保存 | Baseline | Pending | Pending | Pending |
| 模块记忆列表、创建、编辑、删除、启停 | Baseline | Pending | Pending | Pending |
| 每 Agent 推送模式和自动同步配置 | Baseline | Pending | Pending | Pending |
| 模块推送目标配置 | Baseline | Pending | Pending | Pending |
| 单 Agent、全部、选择性批量推送 | Baseline | Pending | Pending | Pending |
| 推送状态读取和展示 | Baseline | Pending | Pending | Pending |
| 打开外部编辑器 | Baseline | Pending | Pending | Pending |
| 云备份覆盖记忆内容但排除本地配置 | Baseline | Pending | Pending | Pending |

## 当前生产后端入口对照

当前 Wails 生产 `App` 导出方法共 101 个，不含测试专用方法。Batch 1 的 daemon 契约应以这些入口为第一轮能力对照，但不要求保持同名同形参数。原生 API 可以按业务语言重命名，只要行为等价并有契约测试。

### Shell、设置、日志、更新和代理

- `GetConfig`
- `SaveConfig`
- `GetAppDataDir`
- `OpenAppDataDir`
- `GetBackendClientConfig`
- `OpenURL`
- `OpenFolderDialog`
- `OpenPath`
- `GetLogDir`
- `OpenLogDir`
- `OpenGitBackupDir`
- `GetAppVersion`
- `CheckAppUpdate`
- `DownloadAppUpdate`
- `ApplyAppUpdate`
- `GetSkippedUpdateVersion`
- `SetSkippedUpdateVersion`
- `CheckAppUpdateAndNotify`
- `TestProxyConnection`

### 技能库、分类和推送拉取

- `ListSkills`
- `ListCategories`
- `CreateCategory`
- `RenameCategory`
- `DeleteCategory`
- `MoveSkillCategory`
- `DeleteSkill`
- `DeleteSkills`
- `GetSkillMeta`
- `GetSkillMetaByPath`
- `ReadSkillFileContent`
- `ImportLocal`
- `CheckUpdates`
- `UpdateSkill`
- `GetEnabledAgents`
- `ScanAgentSkills`
- `ListAgentSkills`
- `DeleteAgentSkill`
- `CheckMissingAgentPushDirs`
- `PushToAgents`
- `PushToAgentsForce`
- `PullFromAgent`
- `PullFromAgentForce`
- `PushStarSkillsToAgents`
- `PushStarSkillsToAgentsForce`
- `AddCustomAgent`
- `RemoveCustomAgent`

### 仓库收藏

- `AddStarredRepo`
- `AddStarredRepoWithCredentials`
- `RemoveStarredRepo`
- `ListStarredRepos`
- `ListAllStarSkills`
- `ListRepoStarSkills`
- `UpdateStarredRepo`
- `UpdateAllStarredRepos`
- `ImportStarSkills`

### 云备份

- `BackupNow`
- `ListCloudFiles`
- `RestoreFromCloud`
- `ListCloudProviders`
- `GetGitConflictPending`
- `ResolveGitConflict`
- `GetLastBackupChanges`
- `GetLastBackupCompletedAt`

### 提示词

- `ListPrompts`
- `ListPromptCategories`
- `CreatePrompt`
- `UpdatePrompt`
- `MovePromptCategory`
- `DeletePrompt`
- `CreatePromptCategory`
- `RenamePromptCategory`
- `DeletePromptCategory`
- `ImportPrompts`
- `PrepareImportPrompts`
- `CompleteImportPrompts`
- `CancelImportPrompts`
- `ExportPrompts`
- `ExportPromptsByNames`
- `PromptRootDir`

### 记忆

- `GetMainMemory`
- `SaveMainMemory`
- `ListModuleMemories`
- `GetModuleMemory`
- `CreateModuleMemory`
- `SaveModuleMemory`
- `DeleteModuleMemory`
- `SetModuleMemoryEnabled`
- `GetMemoryPushConfig`
- `SaveMemoryPushConfig`
- `GetAllMemoryPushConfigs`
- `GetModulePushTargets`
- `SaveModulePushTargets`
- `GetAllModulePushTargets`
- `PushMemoryToAgent`
- `PushAllMemory`
- `PushSelectedMemory`
- `GetMemoryPushStatus`
- `GetAllMemoryPushStatuses`
- `OpenMemoryInEditor`
- `GetAgentMemoryPreview`

### 兼容或待清理入口

- `Greet`

`Greet` 目前不对应用户功能。迁移 daemon 契约时应先确认是否仍被前端或测试依赖；如果没有依赖，应作为 legacy cleanup 单独处理，不应进入 Native API。

## 原生迁移完成定义

一个功能面只有同时满足以下条件，才能从 `Pending` 改为 `Done`：

1. macOS 原生客户端完成对应用户操作。
2. Windows 原生客户端完成对应用户操作。
3. Go daemon 契约有覆盖该操作的测试。
4. 行为已与 Wails baseline 对照。
5. 如果涉及用户可见行为变化，已更新 `docs/features.md` 和 `docs/features_zh.md`。
6. 如果涉及配置或落盘语义变化，已更新 `docs/config.md` 和 `docs/config_zh.md`。

## Batch 0 结论

截至本文创建时，Wails baseline 已作为完整功能基线冻结；所有原生客户端和 daemon API 迁移项仍为 `Pending`。后续批次应只把已验证的行从 `Pending` 更新为 `Done`，不能用“实现中”替代验收。

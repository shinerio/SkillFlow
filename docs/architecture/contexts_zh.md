# 限界上下文与领域模型

## Context Map

| 上下文 | 类型 | 拥有的真相 |
|------|------|-----------|
| `skillcatalog` | 核心域 | 已安装技能 |
| `promptcatalog` | 核心域 | Prompt 库 |
| `agentintegration` | 核心域 | Agent 定义、推送/拉取语义，以及 agent 侧 skill 启停行为 |
| `skillsource` | 支撑域 | 被跟踪的仓库、逻辑 SkillSource 与来源发现状态 |
| `backup` | 支撑域 | 备份与恢复规划 |

`Settings`、`Dashboard`、`My Agents` 等页面都不是 bounded context，它们只是组合后的读视图。

tray、窗口状态、开机自启、单实例、应用更新等壳层关注点应由 `cmd/skillflow/` 与 `platform/` 处理，而不是单独抽成一个 bounded context。

## `skillcatalog`

### 职责

- 拥有已安装技能库的真相
- 拥有实例身份与逻辑身份的映射
- 拥有分类、安装元数据和实例更新状态

### 聚合根

- `InstalledSkill`
- `SkillCategory`

### 关键值对象

- `SkillID`
- `SkillName`
- `LogicalSkillKey`
- `SkillSourceRef`
- `SkillStorageRef`
- `SkillVersionState`

### 领域规则

- 一个 `InstalledSkill` 永远表示一个已安装的技能实例
- `SkillID` 是实例身份，并保留在 `skillcatalog` 内部
- `LogicalSkillKey` 是跨上下文身份
- 名称和绝对路径都不能作为跨域主标识
- 对于 repo-backed skill，`LogicalSkillKey` 由 `SkillSourceRef` 派生，不应再作为独立持久化字段保存

### 对外发布语言

- `InstalledSkillSummary`
- `InstalledSkillVersionView`
- `SkillCategorySummary`

## `promptcatalog`

### 职责

- 拥有 Prompt 库的真相
- 拥有 Prompt 分类、内容和元数据
- 拥有 Prompt 导入导出行为

### 聚合根

- `Prompt`
- `PromptCategory`

### 关键值对象

- `PromptID`
- `PromptName`
- `PromptContent`
- `PromptStorageRef`
- `PromptLinkSet`
- `PromptMediaSet`

### 领域规则

- Prompt 是一等内容概念，不是 Skill 的子类型
- 分类是领域概念，不只是目录名
- 当前导入冲突语义是“名称已存在”的 name-based conflict

### 对外发布语言

- `PromptSummary`
- `PromptCategorySummary`

## `agentintegration`

### 职责

- 拥有 Agent 配置
- 拥有内建 Agent 配置默认值
- 拥有扫描、推送、拉取语义
- 拥有 agent 侧 skill 启用 / 停用语义
- 拥有推送与拉取冲突判定
- 拥有 Agent 侧存在状态的语义

### 聚合根

- `AgentProfile`
- `AgentPushPolicy`

### 关键值对象

- `AgentID`
- `AgentName`
- `AgentType`
- `ScanDirectorySet`
- `PushDirectory`
- `AgentSkillRef`
- `PushConflict`
- `PullConflict`
- `AgentSkillObservation`

### 领域规则

- 该上下文不拥有 Skill 内容真相
- `seenInAgentScan` 与 `pushed` 是两种不同状态
- 冲突判定属于该上下文，不应散落到 UI 入口代码中

### 对外发布语言

- `AgentSummary`
- `AgentSkillPresence`
- `PushPlan`
- `PullPlan`

## `skillsource`

### 职责

- 拥有 GitHub Star Repo 以及其他被跟踪的外部仓库
- 拥有从仓库中派生出的逻辑 SkillSource
- 拥有来源同步状态
- 拥有来源侧候选技能发现结果
- 为来源驱动的已安装技能提供版本线索

### 聚合根

- `StarRepo`
- `SkillSource`

### 关键值对象

- `StarRepoID`
- `SourceID`
- `RepoSource`
- `SourceSubPath`
- `SourceSyncStatus`
- `SourceCacheRef`
- `SourceSkillCandidate`

### 领域规则

- 该上下文不拥有已安装技能真相
- `StarRepo` 表示一个被收藏并跟踪的 GitHub 仓库
- `SkillSource` 表示一个由 `repo + subpath` 唯一标识的逻辑技能来源
- 一个 `StarRepo` 可以包含多个 `SkillSource`
- 一个已安装逻辑 Skill 应只对应一个 `SkillSource`
- 它拥有“某个 Star Repo 是否被跟踪”以及“该来源发现了哪些候选技能”的真相
- 是否已安装、是否可更新，应通过与 `skillcatalog` 和 `agentintegration` 的 published language 组合得出

### 领域解释

`StarRepo` 是仓库级领域模型，`SkillSource` 是其下的技能级来源模型。两者不应混为一个概念。

### 对外发布语言

- `StarRepoSummary`
- `SkillSourceSummary`
- `SourceSkillCandidateView`
- `SourceVersionHint`

## `backup`

### 职责

- 拥有备份配置
- 拥有备份范围与恢复规划
- 拥有备份快照对比逻辑

### 聚合根

- `BackupProfile`
- `GitBackupProfile`

### 关键值对象

- `BackupTarget`
- `BackupScope`
- `BackupSnapshot`
- `BackupChangeSet`
- `RestorePlan`
- `RestoreConflict`

### 领域规则

- 该上下文不拥有 Skill 或 Prompt 的业务真相
- 它拥有的是如何捕获、恢复和校验这些真相

## 共享内核

只有高度稳定的概念才允许进入 `shared/`：

- `LogicalSkillKey`
- 通用领域错误契约
- 基础领域事件契约

`SkillID`、`PromptID` 这类上下文内部实例 ID 应保留在所属上下文中，除非未来被证明必须跨域共享。

## 统一的 Skill 身份与状态模型

这套规则是强约束。

### 身份层次

- 实例身份：`SkillID`
- 逻辑身份：`LogicalSkillKey`

### 逻辑键生成规则

- 对于 repo-backed skill：
  - `LogicalSkillKey` 由 `SkillSourceRef` 派生
  - 标准形式：`git:<repo-source>#<subpath>`
  - 当已经存在 `SkillSourceRef` 时，不应再把它作为独立字段重复持久化
- 对于手动导入或非仓库来源的 skill：
  - `LogicalSkillKey` 应在导入时生成一个稳定的内容型标识
  - 推荐形式：`content:<hash>`
  - 这个 hash 应基于导入时的规范化 skill 内容快照计算，而不是基于绝对路径或本地专属元数据
  - 由于这类 skill 没有 `SkillSourceRef` 可供后续派生，因此生成后需要持久化保存

### 状态语义

- `installed`：该逻辑键至少存在一个已安装技能实例
- `imported`：外部来源流程中对 `installed` 的文案别名
- `pushed`：该逻辑技能存在于某个 Agent 的配置推送目录中
- `seenInAgentScan`：该逻辑技能在扫描目录中被发现，这不代表一定由 SkillFlow 推送
- `updatable`：已安装实例或 Agent 侧副本落后于当前已知来源状态

### 跨上下文规则

前端页面和跨域读模型都不能只根据名称或绝对路径判断是不是同一个 Skill。逻辑身份必须由后端统一解析。

### 已安装映射约束

对于仓库来源的 Skill，一个由 `repo + subpath` 标识的逻辑来源，在技能库中只应对应一个已安装 Skill。重复导入同一个逻辑来源时，应视为“已安装冲突”或“更新已有实例”，而不是创建第二个已安装实例。

对于手动导入的 Skill，逻辑键来自导入时的内容快照。如果该逻辑键在技能库中已经存在，则应视为“已安装冲突”或显式对齐流程，而不是继续创建第二个拥有同一逻辑身份的已安装 Skill。

## 跨上下文协作

### 写协调

跨上下文写操作必须通过显式的 orchestration，而不是直接共享聚合。

例如：

- 来源导入 -> 创建已安装技能 -> 可选自动推送 -> 可选自动备份
- 已安装技能更新 -> 刷新已推送副本
- 恢复备份 -> 各上下文执行重建步骤

### 读组合

跨上下文 UI 视图应通过 read model 组合多个上下文的 published language。

例如：

- Dashboard
- My Agents
- Settings
- 带安装状态和推送状态的来源候选技能列表

*最后更新：2026-04-12*

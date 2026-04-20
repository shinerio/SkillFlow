# 智能体 Skill 启用/停用与全局分组设计

**日期：** 2026-04-12

## 背景

当前 **我的智能体** 页面以单智能体为中心，支持查看推送目录与扫描目录里的 skill，并支持手动拉取、删除已推送 skill、预览记忆内容。

本次需求新增三类能力：

1. 支持单个 skill 的启用和停用。
2. 增加一个跨所有智能体的全局看板，用于管理所有智能体 skill 集合的分组。
3. 每个智能体支持按全局分组批量启用和停用。

同时，本次虽然只落地 Codex，但架构需要支持后续扩展到其他内置智能体和自定义智能体；不同智能体的启用/停用实现方式可能完全不同。

## 用户目标

用户希望在 **我的智能体** 中完成两层管理：

- 在全局层面，用统一的 skill 名称视角给所有智能体里出现过的 skill 分组。
- 在单智能体层面，按 skill 名称或全局分组控制该智能体中 skill 是否启用，而不必删除目录。

这里的“分组”是纯管理语义，不承担版本管理和内容真值职责。

## 确认过的产品约束

1. 全局分组按 **skill 名称** 管理，不考虑版本差异、逻辑身份差异或描述差异。
2. 一个 skill 名称只能属于一个全局分组。
3. 全局分组入口不应与单智能体的 `Skills | Memory` 并列，而应与各智能体并列，作为左侧导航中的 `全部` 入口。
4. `全部` 视图聚合所有已启用智能体的：
   - `pushDir`
   - 所有 `scanDirs`
5. 单智能体视图仍保留 `Skills | Memory`。
6. 对 Codex，如果存在同名 skill，要一起全部启用或全部停用。

## 设计决策

### 1. 信息架构

将 **我的智能体** 左侧导航改为：

- `全部`
- `codex`
- 其他已启用智能体

右侧内容规则：

- 选中 `全部`：
  - 只展示全局 skill 分组管理视图
  - 不展示 `Memory`
  - 不展示单智能体启停操作
- 选中某个具体智能体：
  - 保持现有 `Skills | Memory` 二段结构
  - `Skills` 面板按全局分组展示该智能体的 skill 名称，并支持单个/整组启停
  - `Memory` 面板保持现状

### 2. 边界归属

按职责拆分如下：

- `agentintegration`
  - 负责智能体 skill 管理语义
  - 负责不同智能体 enable/disable 规则的适配器抽象
  - 负责将 SkillFlow 的期望状态落地到具体智能体配置
- `readmodel`
  - 负责组装：
    - 全局 `全部` 视图
    - 单智能体按组展示视图
- `config`
  - 只负责持久化管理状态
- `skillcatalog`
  - 不承载这套分组语义
  - 继续只拥有“已安装 skill”真值、分类、逻辑身份、更新状态等职责

### 3. 数据模型

新增一个本地配置块，放入 `config_local.json`，建议命名为 `agentSkillManagement`。

原因：

- 它直接作用于本机上的智能体实例与本机路径。
- 它依赖本机真实存在的 agent skill 目录与 agent 配置文件。
- 后续自定义智能体也属于本地配置空间。

建议结构：

```json
{
  "agentSkillManagement": {
    "groups": ["frontend", "backend"],
    "assignments": [
      { "skillName": "react-expert", "groupName": "frontend" },
      { "skillName": "go-reviewer", "groupName": "backend" }
    ],
    "agentStates": [
      {
        "agentName": "codex",
        "disabledSkillNames": ["legacy-lint"],
        "disabledGroupNames": ["backend"]
      }
    ]
  }
}
```

字段语义：

- `groups`
  - 全局分组名集合
  - 允许空分组存在，便于先建组再分配名称
- `assignments`
  - `skillName -> groupName` 映射
  - 一个 skill 名称最多出现一次
- `agentStates`
  - 每个智能体的局部禁用状态
- `disabledSkillNames`
  - 该智能体下单独禁用的 skill 名称
- `disabledGroupNames`
  - 该智能体下整组禁用的分组名

默认语义：

- 未出现在禁用集合里的 skill，视为启用
- 启用不保存正向记录，只通过删除禁用记录表达

### 4. 名称分组与跨上下文身份规则

这次新增的全局分组是 **我的智能体管理视图** 的局部语义，刻意按 `skillName` 折叠。

这不改变系统其他地方的身份规则：

- `skillcatalog`、导入冲突、更新检测、逻辑归并等仍然依赖 `LogicalSkillKey`
- 不能把本次 name-based grouping 反向传播到安装库身份判断

换句话说：

- “分组按名称”只用于 agent skill 管理体验
- “逻辑身份按 `LogicalSkillKey`”仍是跨上下文真实身份规则

### 5. 全局视图聚合规则

`全部` 视图只聚合 **已启用智能体**，并扫描：

- 每个智能体的 `pushDir`
- 每个智能体的所有 `scanDirs`

聚合键为 `skillName`。

聚合后每个全局条目至少包含：

- `skillName`
- `groupName`
- `agents`
- `instanceCount`

内部可保留但默认不主展示：

- `paths`
- `agent -> paths` 明细

不展示：

- 描述
- frontmatter 详细字段
- 版本差异
- 来源详情

### 6. 单智能体启停语义

单智能体页仍然展示具体 agent 下的 skill，但按全局分组组织。

最终状态判定公式：

```text
disabled = skillName ∈ disabledSkillNames
        OR groupName ∈ disabledGroupNames
```

其中：

- `groupName` 由全局 `assignments` 查得
- 未分组 skill 不受组开关影响，只受单 skill 开关影响

UI 展示最终状态即可，不强制暴露状态来源；若需要，可在次级文案显示 “disabled by group”。

### 7. 智能体适配器抽象

在 `agentintegration` 中新增面向 skill enablement 的适配器能力。建议抽象成三类职责：

- 解析当前智能体可管理的 skill 实例集合
- 将 SkillFlow 的期望启停状态翻译到该智能体自己的配置机制
- 可选地读取该智能体当前外部配置状态

建议接口方向：

- `ResolveManagedSkills(profile)`
  - 返回该智能体当前可管理的 `skillName -> paths`
- `ApplySkillEnablement(profile, desiredState)`
  - 把期望状态落地到该智能体配置
- `ReadSkillEnablement(profile)`
  - 可选，用于回读真实外部配置

不同智能体可以各自实现不同方案：

- Codex：改 `config.toml`
- 未来其他智能体：可能改 YAML、JSON、数据库、符号链接或 wrapper 文件

### 8. Codex 落地方案

本次 Codex 以官方文档为准：

- 通过 `~/.codex/config.toml` 的 `[[skills.config]]` 条目控制 skill 启用状态
- `enabled = false` 表示禁用 skill
- 变更后需要重启 Codex 生效

本次实现策略：

- 先扫描 Codex 当前 skill 实例路径
- 以 `skill.Name` 归并同名实例
- 某个名称被判定为禁用时：
  - 为该名称下所有路径写入 `[[skills.config]]` + `enabled = false`
- 某个名称被判定为启用时：
  - 删除这些路径对应的禁用条目
  - 不主动写 `enabled = true`

选择“删除禁用条目表示启用”的原因：

- 官方文档明确给出了 `enabled = false` 的禁用模式
- 未明确说明应长期保留 `enabled = true`
- 删除禁用条目更贴近默认启用语义，也更稳妥

Codex 写入规则要求：

- 只管理当前目标 skill 路径对应的 `[[skills.config]]` 条目
- 保留用户已有其他配置和不相关条目
- 重复禁用不产生重复条目
- 重复启用不误删无关配置

### 9. 交互设计

#### 左侧导航

- `全部`
- 各启用智能体

#### `全部` 视图

提供：

- 新建分组
- 重命名分组
- 删除空分组
- 把某个 skill 名称分配到某个分组
- 把某个 skill 名称移出分组

展示列建议：

- 名称
- 分组
- 出现的智能体
- 实例数量

#### 单智能体 `Skills` 视图

按 “分组 -> skill 名称” 展示：

- 每个分组块：
  - 分组名
  - 名称数量
  - `Enable Group`
  - `Disable Group`
- 每个 skill 名称项：
  - 名称
  - 当前状态
  - 启用/停用开关

未分组项使用固定分组：

- `Ungrouped`

#### 单智能体 `Memory` 视图

- 保持现状，不引入全局分组语义

#### 文案提示

对于 Codex，需要在单智能体操作区域显示提示：

- 修改后需要重启 Codex 才会生效

### 10. 日志要求

所有 enable/disable 相关后端操作都需要遵守仓库日志规则，至少记录：

- 操作名
- agentName
- 目标 skillName 或 groupName
- `started` / `completed` / `failed`
- 失败原因

例如：

- `apply agent skill enablement started: agent=codex skill=react-expert`
- `apply agent group enablement completed: agent=codex group=frontend disabled=true`

### 11. 测试策略

后端至少覆盖：

- `config`
  - `agentSkillManagement` 的 merge/split/default/normalize
- `agentintegration`
  - 名称分组解析
  - 单 skill 禁用
  - 组禁用
  - 单 skill 与组禁用叠加
  - Codex 同名多路径一起启停
  - 启用时删除禁用条目而不是写 `enabled = true`
  - 保留无关 TOML 内容
- `readmodel`
  - `全部` 视图聚合
  - 单智能体按组读模型

前端至少覆盖：

- 左侧 `全部 + agents` 导航切换
- `全部` 视图分组管理
- 单智能体页按组展示
- 单 skill 启停
- 整组启停
- Codex 重启提示

### 12. 文档同步范围

本次属于明确的用户功能变化，必须同步更新：

- `docs/features.md`
- `docs/features_zh.md`
- `docs/config.md`
- `docs/config_zh.md`

如果实现中涉及新的上下文协作接口或 agentintegration 扩展点，还应同步更新：

- `docs/architecture/contexts.md`
- `docs/architecture/contexts_zh.md`

## 不做的事

本次明确不做：

- 为安装库引入按名称的全局分类真值
- 处理版本不一致展示
- 展示 skill 描述、来源、frontmatter 详情
- 同步到除 Codex 之外的其他智能体 enable/disable 机制
- 长期兼容旧 schema 的业务层分支


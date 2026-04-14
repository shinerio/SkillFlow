# 我的智能体全局分组与启停管理重构实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 把“我的智能体”重构为“全部视图负责全局分组，单智能体视图负责来源浏览与启停管理”的交互结构，修复新建分组无反馈问题，并将 skill 管理切换为卡片网格模式。

**Architecture:** 保持现有后端分组与启停模型不变，主要改造 `ToolSkills` 页面结构与配套前端状态辅助函数。`全部` 视图对齐“我的技能”的侧边栏 + 卡片拖拽交互，单智能体 `技能` 页增加启停管理模式，并使用卡片网格按启用状态分区展示。分组新增、重命名、删除全部改为应用内弹窗。

**Tech Stack:** Go, React, TypeScript, Wails, Node test runner, Go test, Markdown docs

---

### Task 1: 先用测试锁定新的前端状态与交互规则

**Files:**
- Modify: `cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs`
- Create: `cmd/skillflow/frontend/tests/toolSkillsEnablementView.test.mjs`
- Modify: `cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts`

**Step 1: Write the failing test**

补测试覆盖以下规则：

- `全部` 页左侧分组侧边栏由：
  - 配置中的分组
  - 实际 skill 上的分组
  合并生成
- `全部` 页支持：
  - `全部`
  - `未分组`
  - 自定义分组
  的筛选
- 拖拽到 `未分组` 时会映射为清空分组动作
- 启停管理视图可以把 skill 分为：
  - 已启用
  - 已禁用

建议测试名：

- `buildToolSkillGroupNavItems merges configured and discovered groups`
- `filterAllSkillsByGroup handles all ungrouped and named groups`
- `resolveToolSkillDropAction clears assignment for ungrouped`
- `partitionManagedSkillsByEnabledState splits enabled and disabled cards`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `FAIL`，因为这些状态辅助函数和测试期望还不存在。

**Step 3: Write minimal implementation**

在 `cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts` 中增加最小辅助函数：

- 构建分组侧边栏项
- 按分组筛选 `全部` 页数据
- 解析拖拽归类动作
- 对管理卡片按启用状态分区

保持函数纯净、无 UI 依赖。

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `PASS`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs cmd/skillflow/frontend/tests/toolSkillsEnablementView.test.mjs
git commit -m "test: define tool skill grouping and enablement view state"
```

### Task 2: 用应用内弹窗替换分组 prompt 交互

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`

**Step 1: Write the failing test**

补测试覆盖：

- 新建分组弹窗默认关闭
- 点击按钮后打开
- 空名称、纯空格、重名时确认按钮不可用
- 提交中按钮禁用
- 提交失败时显示错误消息

如果页面级测试不方便，至少给输入校验与弹窗状态提取一个纯函数测试。

建议测试名：

- `validateToolSkillGroupName rejects blank and duplicate names`
- `createToolSkillGroupDialogState disables submit while invalid or saving`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `FAIL`

**Step 3: Write minimal implementation**

在 `ToolSkills.tsx` 中：

- 删除 `window.prompt(...)`
- 增加应用内弹窗状态：
  - 新建分组
  - 重命名分组
  - 删除分组确认
- 把错误显示放在弹窗内

必要时提取一个局部 helper，但不要过度抽象。

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `PASS`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/src/i18n/en.ts
git commit -m "feat: replace tool skill group prompts with dialogs"
```

### Task 3: 重构 `全部` 视图为侧边栏 + 卡片网格 + 拖拽归类

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts`
- Optionally Create: `cmd/skillflow/frontend/src/components/AgentSkillGroupSidebar.tsx`
- Optionally Create: `cmd/skillflow/frontend/src/components/AllAgentSkillCard.tsx`
- Test: `cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs`

**Step 1: Write the failing test**

补测试覆盖：

- `全部` 页分组筛选后的结果数量
- 拖拽 skill 卡片到分组时调用的是分配动作
- 拖拽到 `未分组` 时调用的是清空动作

如果组件级测试过重，就把拖拽决策逻辑继续保持为纯函数测试。

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `FAIL`

**Step 3: Write minimal implementation**

重排 `全部` 页：

- 左侧改为分组侧边栏
- 右侧改为 skill 卡片网格
- 卡片信息包含：
  - skill 名称
  - 分组
  - 智能体列表
  - 实例数
- 增加拖拽到侧边栏的归组能力

尽量复用“我的技能”的交互模式：

- 高亮 drop target
- 拖拽结束后清理状态

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `PASS`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts cmd/skillflow/frontend/src/components/AgentSkillGroupSidebar.tsx cmd/skillflow/frontend/src/components/AllAgentSkillCard.tsx cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs
git commit -m "feat: redesign all-agent skill grouping with drag and sidebar"
```

### Task 4: 为单智能体技能页增加“启停管理”模式与卡片网格

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts`
- Optionally Create: `cmd/skillflow/frontend/src/components/ManagedAgentSkillCard.tsx`
- Test: `cmd/skillflow/frontend/tests/toolSkillsEnablementView.test.mjs`

**Step 1: Write the failing test**

补测试覆盖：

- 单智能体 `技能` 视图可在：
  - 来源浏览
  - 启停管理
  两个模式间切换
- 启停管理模式下：
  - 已启用 skill 进入已启用区
  - 已禁用 skill 进入已禁用区
- 分组级按钮状态与卡片按钮状态可由数据推导

建议测试名：

- `getDefaultAgentSkillMode returns browse`
- `partitionManagedSkillsByEnabledState keeps disabled cards separate`
- `buildManagedSkillGroupActions exposes enable and disable actions`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `FAIL`

**Step 3: Write minimal implementation**

在单智能体 `技能` 页中：

- 新增 `启停管理` 入口按钮
- 默认保留来源浏览视图
- 点击后切到管理视图
- 管理视图使用多行多列卡片
- 按启用状态分区
- 组头提供整组启用 / 停用
- 卡片提供单 skill 启用 / 停用
- Codex 顶部提示继续保留

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit
```

Expected: `PASS`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts cmd/skillflow/frontend/src/components/ManagedAgentSkillCard.tsx cmd/skillflow/frontend/tests/toolSkillsEnablementView.test.mjs
git commit -m "feat: add agent skill enablement management mode"
```

### Task 5: 跑通回归并修正文案与文档

**Files:**
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Verify: `cmd/skillflow/app_daemon_service.go`
- Verify: `core/readmodel/agentskills/model.go`

**Step 1: Write the failing test**

如果前端新增了任何纯函数文案状态辅助，补对应测试；否则直接进入验证步骤。

**Step 2: Run targeted verification before final edits**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit

cd /Users/shinerio/Workspace/code/SkillFlow/.worktrees/dev-disable-skill
go test ./cmd/skillflow -run 'TestDaemonServiceHandlersExposeAgentSkillManagementMethods|TestListManagedAgentSkillsAndAllAgentSkills|TestAgentSkillManagementMutationMethods'
go test ./core/readmodel/agentskills
```

Expected: 全部通过

**Step 3: Update docs and copy**

同步更新用户可见文档：

- `docs/features.md`
- `docs/features_zh.md`

更新内容：

- `全部` 视图支持侧边栏分组与拖拽归类
- 单智能体支持独立启停管理视图
- skill 管理改为卡片网格

**Step 4: Run final verification**

Run:

```bash
cd cmd/skillflow/frontend
npm run test:unit

cd /Users/shinerio/Workspace/code/SkillFlow/.worktrees/dev-disable-skill
go test ./cmd/skillflow
go test ./core/readmodel/agentskills
```

Expected: `PASS`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/src/i18n/en.ts docs/features.md docs/features_zh.md
git add cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts cmd/skillflow/frontend/src/components/AgentSkillGroupSidebar.tsx cmd/skillflow/frontend/src/components/AllAgentSkillCard.tsx cmd/skillflow/frontend/src/components/ManagedAgentSkillCard.tsx
git add cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs cmd/skillflow/frontend/tests/toolSkillsEnablementView.test.mjs
git commit -m "feat: redesign my agents skill grouping and enablement"
```

### Task 6: 构建桌面应用并做人工检查

**Files:**
- Verify only: `cmd/skillflow/build/bin/SkillFlow.app`

**Step 1: Build the app**

Run:

```bash
cd cmd/skillflow
~/go/bin/wails build
```

Expected: 构建成功，生成新的 `build/bin/SkillFlow.app`

**Step 2: Manual verification checklist**

检查以下路径：

1. 打开“我的智能体”不会空白或崩溃
2. `全部` 页可看到：
   - `全部`
   - `未分组`
   - 自定义分组
3. `新建分组` 弹窗可打开、可校验、可提交
4. skill 卡片可拖到左侧分组
5. 单智能体页可进入 `启停管理`
6. 启停管理中卡片为多行多列
7. 单 skill 和整组启停都生效
8. Codex 仍显示重启提示

**Step 3: Commit**

```bash
git add cmd/skillflow/build/bin/SkillFlow.app
git commit -m "build: refresh app bundle for my agents redesign"
```

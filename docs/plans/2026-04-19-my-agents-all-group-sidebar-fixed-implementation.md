# 我的智能体-全部分组栏固定 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让“我的智能体 -> 全部”页面在浏览长 skill 列表时保持左侧分组栏持续可见，并保证右侧 skill 区独立滚动。

**Architecture:** 在前端页面层把“全部”页从页面级滚动切换为双栏内部滚动。页面只在 `selectedTarget === 'all'` 时启用固定内容壳层，左侧分组栏和右侧 skill 区分别承担各自滚动，避免分组栏随 skill 列表离开视口。

**Tech Stack:** React 18、TypeScript、Tailwind 工具类、Node 内置测试运行器

---

### Task 1: 为“全部”页滚动模式写失败回归测试

**Files:**
- Modify: `cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs`
- Modify: `cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts`

**Step 1: 写失败测试**

在 `toolSkillsManagement.test.mjs` 中新增一个测试，断言“全部”页使用面板内部滚动模式，其它目标仍使用页面滚动模式。

**Step 2: 运行测试并确认失败**

Run: `npm run test:unit`
Expected: 新增测试失败，提示滚动模式 helper 不存在或返回值不符合预期。

**Step 3: 实现最小 helper**

在 `toolSkillsManagement.ts` 中新增一个轻量 helper，用来判断当前页面应使用哪种滚动模式。

**Step 4: 运行测试并确认通过**

Run: `npm run test:unit`
Expected: 新增测试通过，已有测试不回归。

### Task 2: 将“全部”页切换为左右独立滚动布局

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/components/AgentSkillGroupSidebar.tsx`

**Step 1: 调整页面内容容器**

仅在“全部”页把内容区从 `overflow-y-auto` 调整为 `overflow-hidden + min-h-0`，把滚动责任下放到内部双栏面板。

**Step 2: 调整“全部”页主面板布局**

让双栏容器使用 `h-full / min-h-0 / overflow-hidden`，确保右侧 skill 区可以独立滚动。

**Step 3: 调整分组栏组件**

给分组栏增加 `h-full / min-h-0 / overflow-y-auto`，保证分组过多时左侧也能独立滚动，但默认始终固定在左侧。

**Step 4: 保持其它视图不变**

确认 `selectedTarget !== 'all'` 的页面仍沿用原来的页面滚动方式。

### Task 3: 运行验证

**Files:**
- Test: `cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs`
- Test: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Test: `cmd/skillflow/frontend/src/components/AgentSkillGroupSidebar.tsx`

**Step 1: 运行前端单测**

Run: `npm run test:unit`
Expected: 全部通过。

**Step 2: 人工检查布局逻辑**

检查代码中“全部”页是否只让右侧 skill 区滚动、左侧分组栏是否具备独立滚动能力。

**Step 3: 总结影响**

记录本次改动只影响“我的智能体 -> 全部”页面，不更新功能文档，因为它属于交互布局修正，不是新增功能。

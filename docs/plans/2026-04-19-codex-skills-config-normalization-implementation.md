# Codex skills.config 归并与路径标准化 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 Codex 禁用 skill 配置回归 OpenAI 官方 `[[skills.config]]` 格式，并把每个 `path` 规范到 skill 入口 markdown 文件。

**Architecture:** 在 Codex 文件系统适配器中新增一层配置解析与标准化逻辑，读取时兼容官方 `[[skills.config]]` 和历史错误的 `skills.config = [ ... ]`，写回时统一输出官方格式。路径标准化在写入前完成，保证配置引用的是具体 skill 文件而不是目录。

**Tech Stack:** Go 1.25、标准库字符串与路径处理、`stretchr/testify`

---

### Task 1: 为官方格式写回与文件级路径补失败测试

**Files:**
- Modify: `core/agentintegration/infra/gateway/filesystem_adapter_test.go`

**Step 1: 写失败测试**

新增测试覆盖：

- 多个禁用 skill 写回为多个 `[[skills.config]]`
- `path` 指向 `SKILL.md` 或真实存在的 skill markdown 文件

**Step 2: 运行测试并确认失败**

Run: `go test ./core/agentintegration/infra/gateway -run 'TestCodexApplySkillEnablement'`
Expected: 失败，当前实现输出了错误的数组格式，或路径不是入口 markdown 文件。

**Step 3: 写最小实现**

在 `filesystem_adapter.go` 中添加 managed path 标准化和官方格式序列化逻辑。

**Step 4: 运行测试并确认通过**

Run: `go test ./core/agentintegration/infra/gateway -run 'TestCodexApplySkillEnablement'`
Expected: PASS

### Task 2: 为历史错误格式兼容与写回折叠补测试

**Files:**
- Modify: `core/agentintegration/infra/gateway/filesystem_adapter_test.go`
- Modify: `core/agentintegration/infra/gateway/filesystem_adapter.go`

**Step 1: 写失败测试**

新增测试：

- 输入历史错误格式 `skills.config = [ ... ]`
- 写回后折叠为官方 `[[skills.config]]`
- 非 managed 项保持不变

**Step 2: 运行测试并确认失败**

Run: `go test ./core/agentintegration/infra/gateway -run 'TestCodexApplySkillEnablement'`
Expected: 失败，当前解析器不支持把历史错误格式收敛回官方格式。

**Step 3: 写最小实现**

扩展解析器，统一读取两种格式并在写回阶段只输出官方 `[[skills.config]]` 格式。

**Step 4: 运行测试并确认通过**

Run: `go test ./core/agentintegration/infra/gateway -run 'TestCodexApplySkillEnablement'`
Expected: PASS

### Task 3: 运行完整相关验证

**Files:**
- Test: `core/agentintegration/infra/gateway/filesystem_adapter_test.go`

**Step 1: 跑适配器相关测试**

Run: `go test ./core/agentintegration/infra/gateway`
Expected: PASS

**Step 2: 复核无关配置保留**

检查测试结果，确认 `[profiles.default]` 等非 skill 配置仍保留。

**Step 3: 总结影响范围**

说明本次只影响 Codex skill enablement 配置写入格式与 path 精度，不影响其它 agent 适配器。

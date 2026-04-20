# Codex skills.config 归并与路径标准化设计

## 背景

当前 Codex 适配器在写入禁用 skill 配置时存在两个问题：

1. 曾经尝试写成非官方的 `skills.config = [ ... ]` 数组块格式，但这与 OpenAI Codex 官方文档不一致。
2. `path` 直接写入 skill 目录，而不是 skill 的入口 markdown 文件，导致配置粒度过粗，也不满足需要指向 `SKILL.md` 的要求。

## 目标

- 将 SkillFlow 管理的禁用 skill 按 OpenAI 官方文档写成 `[[skills.config]]` 段。
- 每个禁用项的 `path` 必须指向 skill 实际入口 markdown 文件，而不是目录。
- 读取时兼容旧的 `[[skills.config]]` 格式和之前错误写出的数组格式。
- 保留非 SkillFlow 管理的其它 Codex 配置内容，不误删用户手写配置。

## 方案

### 统一输出格式

SkillFlow 写回 Codex 配置时，按 OpenAI 官方文档输出：

```toml
[[skills.config]]
path = "/abs/path/to/skill/SKILL.md"
enabled = false

[[skills.config]]
path = "/abs/path/to/other/skill/skill.md"
enabled = false
```

每个禁用 skill 一个 `[[skills.config]]` 段，避免继续输出与官方不兼容的数组格式。

### 路径标准化

对于 `ManagedAgentSkill.Paths` 中的 skill 目录，写入前先解析实际入口文件：

1. 优先匹配目录下真实存在的 `SKILL.md`
2. 若不存在，则匹配大小写变体 `skill.md`、`Skill.md` 等真实文件名
3. 若目录中找不到入口 markdown，则回退为 `filepath.Join(dir, "SKILL.md")`

这样既满足“指向入口文件”的需求，也避免在大小写敏感文件系统上写入一个并不存在的路径。

### 配置兼容

读取配置时同时识别：

- 官方格式：`[[skills.config]]`
- 历史错误格式：`skills.config = [ ... ]`

写回时统一收敛为官方 `[[skills.config]]` 格式。

### 管理边界

SkillFlow 只替换当前受管理 skill 对应的禁用项：

- 当前 managed skill 若启用，则从写回结果中移除对应禁用项
- 当前 managed skill 若禁用，则在统一数组块中写入对应项
- 不属于当前 managed 集合的已有配置项予以保留

## 取舍

### 方案 A：仅修路径

- 优点：改动小
- 缺点：不能纠正错误的数组格式输出

### 方案 B：回归官方格式并修正路径

- 优点：一次修复格式兼容性和路径精度两个根因，输出格式与官方文档一致
- 缺点：解析逻辑比当前略复杂，需要覆盖旧格式兼容测试

结论：采用方案 B。

## 验证

- 多个禁用 skill 写回后，按官方格式输出多个 `[[skills.config]]` 段
- 每个 `path` 指向 skill 主 markdown 文件，而不是目录
- 历史错误格式 `skills.config = [ ... ]` 在下一次写回后会折叠成官方格式
- 非 SkillFlow 管理的其它配置内容保持原样

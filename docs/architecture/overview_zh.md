# 总体架构

## 架构风格

SkillFlow 的后端架构是一个基于 DDD 的模块化单体，由 macOS/Windows 原生客户端和 legacy Wails 桌面客户端通过同一个 Go daemon 共享。

这套设计的目标是：

1. 明确业务真相的归属。
2. 把领域逻辑和壳层、OS 集成逻辑分开。
3. 用 bounded context 替代按技术能力切分的包边界。
4. 让 daemon 组合与 legacy Wails 相关 transport/壳层代码留在 `cmd/skillflow/`，可复用业务代码沉淀到 `core/`，平台原生客户端放在 `native/`。
5. 让跨上下文写流程进入 `core/orchestration/`，跨上下文读组合进入 `core/readmodel/`。

## 高层结构

```text
native/                macOS Swift 与 Windows WinUI 原生客户端
cmd/skillflow/          Go daemon 组合根、legacy Wails 壳层、进程宿主、网关、OS 集成
core/
  platform/             纯技术能力
  shared/               最小共享内核
  orchestration/        跨上下文写协调
  readmodel/            跨上下文读视图组合
  config/               面向前端的设置门面
  skillcatalog/         核心域
  promptcatalog/        核心域
  agentintegration/     核心域
  skillsource/          支撑域
  backup/               支撑域
```

## 真相归属规则

- `skillcatalog` 拥有已安装技能的真相。
- `promptcatalog` 拥有 Prompt 库的真相。
- `agentintegration` 拥有 Agent 配置、推送/拉取语义和 Agent 侧存在状态的真相。
- `skillsource` 拥有被跟踪的仓库、逻辑上的技能来源以及来源发现状态的真相。
- `backup` 拥有备份与恢复规划的真相，不拥有 Skill 或 Prompt 的业务真相。
- `core/config` 只是 transport / 壳层使用的设置门面，本身不拥有业务真相。
- tray、窗口状态、单实例、开机自启、应用更新等壳层关注点应归于 `cmd/skillflow/` 与 `platform/`，而不是单独的 bounded context。
- Go `daemon` 持有后端运行时和全部业务数据访问权；原生客户端与可销毁的 legacy Wails UI 都通过本地鉴权传输层通信。
- `Settings` 不是 bounded context，而是多个上下文在 UI 层的组合面。

在当前产品形态下，`skillsource` 里有两个不同层级的领域概念：

- `StarRepo`：用户收藏并跟踪的 GitHub 仓库
- `SkillSource`：某个 Skill 的逻辑来源，由 `repo + subpath` 唯一标识

一个 `StarRepo` 可以暴露多个 `SkillSource`。

## 核心原则

### 以真相归属为边界，而不是以页面为边界

Dashboard、Settings、My Agents、Starred Repos 这些页面都不是领域边界。后端边界应由谁拥有业务真相决定，而不是由前端导航决定。

### transport adapter 保持在模块边界

由于 legacy Wails 要求绑定方法必须位于 `cmd/skillflow/package main`，legacy transport adapter 保留在 `cmd/skillflow/`。它们负责把 Wails 请求转换成应用层用例或 read model 调用；原生客户端使用 daemon API 契约。

如果以后增加 CLI 或 API 入口，也应采用相同的模块边界适配角色，而不是把 Wails 专属代码带进 `core/`。

### 先按上下文纵向切分，再在上下文内部做分层

推荐按 bounded context 组织代码，再在上下文内部拆成 `app`、`domain`、`infra`。这样可以避免代码退化成巨大的公共 `service`、`repository` 或 `domain` 包。

### orchestration 与 read model 分离

- 写操作跨上下文时，放到 `orchestration/`
- 读操作跨上下文聚合时，放到 `readmodel/`
- 不要把跨域编排偷偷塞进某个上下文的领域模型里

### 共享内核必须足够小

只有那些跨多个上下文都必须复用、并且语义稳定的概念，才允许进入 `shared/`，例如逻辑键、通用领域错误和基础事件契约。上下文内部的实例 ID 应保留在所属上下文中，除非未来被证明必须跨域共享。

## 核心业务概念

`Skill` 和 `Prompt` 是并列的一等业务概念，不应该在领域层强行抽象成统一的 `Asset` 父模型。

如果需要统一的内容视图，应当通过 `readmodel/` 或应用层查询来实现，而不是通过领域继承体系实现。

*最后更新：2026-08-30*

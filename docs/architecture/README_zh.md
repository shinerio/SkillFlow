# SkillFlow 架构文档集

这个目录保存 SkillFlow 当前后端架构的参考文档。

SkillFlow 的后端是一个基于 DDD 的模块化单体：

- `cmd/skillflow/` 是 Wails 桌面壳层、transport adapter、进程宿主和组合根
- 后端业务代码位于 `core/` 下的各个 bounded context
- 每个上下文内部统一组织为 `app`、`domain`、`infra`
- 跨上下文写协调通过 `core/orchestration/`
- 跨上下文读组合通过 `core/readmodel/`
- `core/config/` 是面向前端的设置门面，承接上下文与 platform 拥有的设置
- 纯技术能力位于 `core/platform/`
- 只有高度稳定的共享内核概念位于 `core/shared/`

## 文档列表

- [总体架构](./overview_zh.md)
  - 高层架构风格、目录结构和真相归属规则
- [分层与依赖规则](./layers_zh.md)
  - transport adapter、`app`、`domain`、`infra`、`orchestration`、`readmodel`、`platform`、`shared` 的职责定义
- [限界上下文与领域模型](./contexts_zh.md)
  - bounded context 地图、聚合根、值对象、published language 和跨上下文身份规则
- [应用层用例](./use-cases_zh.md)
  - 各上下文的 command/query 归属、共享编排以及 read model 组合规则
- [运行时、仓库布局与存储](./runtime-and-storage_zh.md)
  - Wails 壳层约束、daemon/UI 双进程、loopback 网关职责、存储布局，以及 repository 与 gateway 的划分规则

## 迁移计划

- [原生平台重构设计](../plans/2026-04-25-native-platform-refactor-design.md)
  - 从 Wails/React 生产 UI 迁移到 macOS Swift 与 Windows WinUI 原生客户端的目标设计
- [原生平台重构实施计划](../plans/2026-04-25-native-platform-refactor-implementation.md)
  - daemon API 稳定、原生壳、功能切片、打包和发布切换的分批执行计划
- [原生平台功能契约](../plans/2026-04-25-native-platform-feature-contract.md)
  - Wails baseline、macOS 原生客户端、Windows 原生客户端和 daemon API 的行为保持检查表
- [原生平台性能基线](../plans/2026-04-25-native-platform-performance-baseline.md)
  - 资源占用和启动性能对比所需的测量流程和基线表格
- [原生平台发布检查清单](../plans/2026-04-25-native-platform-release-checklist.md)
  - 原生客户端替代 Wails 生产 UI 前必须通过的阻塞发布门槛

## 不变约束

- 仓库根目录不能放 Go 源文件。
- `cmd/skillflow/*.go` 必须保持扁平，因为 Wails 绑定要求单一 `package main` 目录。
- SkillFlow 仍然是 Wails 桌面应用，不引入 REST 服务层。
- 由于 Wails 绑定限制，当前 transport entrypoint 保留在 `cmd/skillflow/`。
- `Skill` 和 `Prompt` 是并列的核心业务概念。
- `Settings` 是 UI 组合视图，不是独立 bounded context。
- `core/config/` 是设置门面，不是业务真相归属的 bounded context。

## 范围

这组文档只覆盖后端架构。用户可见行为仍然以 [`docs/features_zh.md`](../features_zh.md) 为准，落盘配置格式仍然以 [`docs/config_zh.md`](../config_zh.md) 为准。

*最后更新：2026-04-26*

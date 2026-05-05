# SkillFlow Architecture Set

This directory contains the current backend architecture reference for SkillFlow.

SkillFlow's backend is a DDD-oriented modular monolith:

- `cmd/skillflow/` is the Wails desktop shell, transport adapter layer, process host, and composition root
- backend business code lives under bounded contexts in `core/`
- each bounded context is organized as `app`, `domain`, and `infra`
- cross-context write coordination goes through `core/orchestration/`
- cross-context read composition goes through `core/readmodel/`
- `core/config/` is a frontend-facing settings facade over context- and platform-owned settings
- pure technical capabilities live in `core/platform/`
- only highly stable shared kernel concepts live in `core/shared/`

## Documents

- [Overview](./overview.md)
  - high-level architectural style, repository shape, and source-of-truth rules
- [Layers and Dependencies](./layers.md)
  - definitions for transport adapters, `app`, `domain`, `infra`, `orchestration`, `readmodel`, `platform`, and `shared`
- [Bounded Contexts and Domain Model](./contexts.md)
  - bounded context map, aggregate roots, value objects, published language, and cross-context identity rules
- [Application Use Cases](./use-cases.md)
  - command/query ownership by context, shared orchestration, and read-model composition rules
- [Runtime, Repository Layout, and Storage](./runtime-and-storage.md)
  - Wails shell constraints, daemon/UI runtime split, loopback gateway responsibilities, storage layout, and repository vs gateway rules

## Migration Plans

- [Native platform refactor design](../plans/2026-04-25-native-platform-refactor-design.md)
  - target migration from the Wails/React production UI to macOS Swift and Windows WinUI native clients
- [Native platform refactor implementation plan](../plans/2026-04-25-native-platform-refactor-implementation.md)
  - batch-by-batch execution plan for daemon API stabilization, native shells, feature slices, packaging, and release cutover
- [Native platform feature contract](../plans/2026-04-25-native-platform-feature-contract.md)
  - behavior-preservation checklist for the Wails baseline, macOS native client, Windows native client, and daemon API
- [Native platform performance baseline](../plans/2026-04-25-native-platform-performance-baseline.md)
  - required measurement protocol and baseline tables for resource and startup comparisons
- [Native platform release checklist](../plans/2026-04-25-native-platform-release-checklist.md)
  - blocking release gates before native clients can replace the Wails production UI

## Invariants

- The repository root must contain no Go source files.
- `cmd/skillflow/*.go` must stay flat because Wails bindings require a single `package main` directory.
- SkillFlow remains a Wails desktop app with direct bindings, not a REST service.
- Current transport entrypoints live in `cmd/skillflow/` because of Wails binding constraints.
- `Skill` and `Prompt` are parallel core business concepts.
- `Settings` is a UI composition surface, not a bounded context.
- `core/config/` is a settings facade, not a source-of-truth bounded context.

## Scope

These documents cover backend architecture only. User-facing behavior remains documented in [`docs/features.md`](../features.md), and persisted file schemas remain documented in [`docs/config.md`](../config.md).

*Last updated: 2026-04-26*

# Agent Skill Enable/Disable Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add skill-level and group-level enable/disable management for agent skills, introduce an `All` view in **My Agents** for global name-based group management across every enabled agent's push and scan paths, and implement Codex-specific persistence through `~/.codex/config.toml`.

**Architecture:** Keep global grouping and per-agent desired disable state inside local config, model the behavior in `agentintegration`, expose read models for the `All` view and grouped single-agent view, and isolate Codex-specific `config.toml` translation behind an agent skill-management adapter. The frontend updates `ToolSkills` so the left rail becomes `All + enabled agents`, the `All` surface manages global group assignments, and each concrete agent keeps the existing `Skills | Memory` split with grouped enablement controls.

**Tech Stack:** Go, React, TypeScript, Wails, TOML parsing/writing, Node test runner, Go test, Markdown docs

---

### Task 1: Add failing config tests for local agent skill management persistence

**Files:**
- Modify: `core/config/service_test.go`
- Modify: `core/config/model.go`
- Modify: `core/config/service.go`

**Step 1: Write the failing test**

Add tests that prove:

- `config_local.json` can persist and reload `agentSkillManagement`
- defaults keep the new block empty but non-corrupt
- merge/split keeps the data local-only rather than writing it into `config.json`
- normalization removes blank names and duplicate entries where appropriate

Suggested test names:

- `TestServiceSaveAndLoadAgentSkillManagement`
- `TestServiceSplitSharedExcludesAgentSkillManagement`
- `TestServiceNormalizesAgentSkillManagement`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/config/... -run AgentSkillManagement
```

Expected: `FAIL` because the config model and service do not know the new block yet.

**Step 3: Write minimal implementation**

Implement:

- local config structs for:
  - group names
  - `skillName -> groupName` assignments
  - per-agent disabled skill names
  - per-agent disabled group names
- `AppConfig` fields for the same runtime data
- merge/split/default/normalize logic in `core/config/service.go`

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/config/... -run AgentSkillManagement
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/config/model.go core/config/service.go core/config/service_test.go
git commit -m "feat: persist local agent skill management config"
```

### Task 2: Add failing domain and service tests for name-based group and disable-state resolution

**Files:**
- Create: `core/agentintegration/domain/skill_management.go`
- Create: `core/agentintegration/domain/skill_management_test.go`
- Modify: `core/agentintegration/app/service_test.go`
- Modify: `core/agentintegration/app/service.go`

**Step 1: Write the failing test**

Add tests that prove:

- one skill name maps to at most one group
- final disabled state is true when:
  - skill name is directly disabled
  - or its group is disabled
- ungrouped skills are affected only by direct disable
- duplicate paths with the same skill name collapse into one management unit for state evaluation

Suggested test names:

- `TestResolveSkillGroupByName`
- `TestDisabledWhenSkillNameIsDirectlyDisabled`
- `TestDisabledWhenGroupIsDisabled`
- `TestUngroupedSkillIgnoresGroupDisable`
- `TestCollapseSameNameInstancesForManagement`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/agentintegration/... -run 'SkillGroup|Disabled|Collapse'
```

Expected: `FAIL` because no dedicated management model exists yet.

**Step 3: Write minimal implementation**

Implement:

- domain structs/value helpers for:
  - group definitions
  - assignments
  - per-agent disable state
  - final-state resolution
- service helpers that evaluate grouped agent skill states from raw scan results

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/agentintegration/... -run 'SkillGroup|Disabled|Collapse'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/agentintegration/domain/skill_management.go core/agentintegration/domain/skill_management_test.go core/agentintegration/app/service.go core/agentintegration/app/service_test.go
git commit -m "feat: model agent skill groups and disable states"
```

### Task 3: Add failing tests for Codex skill enablement TOML reconciliation

**Files:**
- Create: `core/agentintegration/infra/gateway/codex_skill_manager.go`
- Create: `core/agentintegration/infra/gateway/codex_skill_manager_test.go`
- Modify: `cmd/skillflow/adapters.go`

**Step 1: Write the failing test**

Add tests that prove:

- disabling one skill name adds `[[skills.config]]` `enabled = false` entries for every matching path
- enabling removes those disabled entries instead of writing `enabled = true`
- unrelated TOML sections and unrelated `skills.config` entries remain untouched
- repeated apply operations are idempotent

Suggested test names:

- `TestCodexApplyDisableAddsEntriesForEverySameNamePath`
- `TestCodexApplyEnableRemovesDisabledEntries`
- `TestCodexApplyPreservesUnrelatedConfig`
- `TestCodexApplyIsIdempotent`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/agentintegration/... -run CodexApply
```

Expected: `FAIL` because no Codex-specific skill-management adapter exists.

**Step 3: Write minimal implementation**

Implement:

- a Codex manager that reads and rewrites `~/.codex/config.toml`
- path-targeted reconciliation for `[[skills.config]]`
- registration/wiring from `cmd/skillflow/adapters.go`

Prefer a TOML library already present in the module if available; otherwise add the smallest suitable dependency.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/agentintegration/... -run CodexApply
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/agentintegration/infra/gateway/codex_skill_manager.go core/agentintegration/infra/gateway/codex_skill_manager_test.go cmd/skillflow/adapters.go go.mod go.sum
git commit -m "feat: add codex skill enablement adapter"
```

### Task 4: Add failing read-model tests for the `All` view and grouped single-agent view

**Files:**
- Create: `core/readmodel/agentskills/service.go`
- Create: `core/readmodel/agentskills/service_test.go`
- Modify: `cmd/skillflow/app_services.go`

**Step 1: Write the failing test**

Add tests that prove:

- the `All` view aggregates all enabled agents' push and scan paths by skill name
- the `All` view carries group name, agent list, and instance count
- the single-agent grouped view returns `Ungrouped` for names without assignment
- the single-agent grouped view includes final enabled/disabled status

Suggested test names:

- `TestListAllAgentSkillsAggregatesByNameAcrossEnabledAgents`
- `TestListAllAgentSkillsCarriesGroupAndAgents`
- `TestListGroupedAgentSkillsUsesUngroupedFallback`
- `TestListGroupedAgentSkillsIncludesFinalStatus`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./core/readmodel/... -run 'AllAgentSkills|GroupedAgentSkills'
```

Expected: `FAIL` because no dedicated read model exists yet.

**Step 3: Write minimal implementation**

Implement:

- a read-model service that:
  - scans all enabled agents for the `All` surface
  - groups by `skillName`
  - resolves assigned group and per-agent final status
- wiring in `cmd/skillflow/app_services.go`

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./core/readmodel/... -run 'AllAgentSkills|GroupedAgentSkills'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add core/readmodel/agentskills/service.go core/readmodel/agentskills/service_test.go cmd/skillflow/app_services.go
git commit -m "feat: compose read models for grouped agent skills"
```

### Task 5: Add failing App-layer tests for the new management APIs

**Files:**
- Modify: `cmd/skillflow/app_agent_api_test.go`
- Modify: `cmd/skillflow/app.go`
- Modify: `cmd/skillflow/app_settings.go`

**Step 1: Write the failing test**

Add tests that prove:

- the left-side `All` entry can load aggregated grouped skills
- app methods can create, rename, delete groups
- app methods can assign and unassign a skill name to a group
- app methods can toggle one skill for one agent
- app methods can toggle one group for one agent

Suggested method directions:

- `ListAllAgentSkills()`
- `ListManagedAgentSkills(agentName string)`
- `CreateAgentSkillGroup(name string)`
- `RenameAgentSkillGroup(oldName, newName string)`
- `DeleteAgentSkillGroup(name string)`
- `AssignAgentSkillGroup(skillName, groupName string)`
- `ClearAgentSkillGroup(skillName string)`
- `SetAgentSkillEnabled(agentName, skillName string, enabled bool)`
- `SetAgentSkillGroupEnabled(agentName, groupName string, enabled bool)`

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./cmd/skillflow -run 'AgentSkillGroup|AgentSkillEnabled|AllAgentSkills'
```

Expected: `FAIL` because the app methods do not exist.

**Step 3: Write minimal implementation**

Implement:

- thin App methods
- delegation into config + agentintegration + readmodel services
- logging for each mutation

If exported `App` methods change, note that `make generate` is required after the implementation is complete.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./cmd/skillflow -run 'AgentSkillGroup|AgentSkillEnabled|AllAgentSkills'
```

Expected: `ok`

**Step 5: Commit**

```bash
git add cmd/skillflow/app.go cmd/skillflow/app_settings.go cmd/skillflow/app_agent_api_test.go
git commit -m "feat: expose agent skill management app APIs"
```

### Task 6: Add failing frontend helper tests for `All` navigation and grouped rendering state

**Files:**
- Create: `cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts`
- Create: `cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs`

**Step 1: Write the failing test**

Add tests that prove:

- the left navigation includes a stable `all` entry before enabled agents
- `all` mode disables memory rendering
- grouped single-agent data falls back to `Ungrouped`
- search filters grouped names without requiring descriptions

Suggested test names:

- `buildToolSkillsNavItems prepends all entry`
- `isMemoryPanelAvailable returns false for all`
- `groupManagedSkills uses Ungrouped fallback`
- `filterManagedSkills searches name only`

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` because the helper does not exist.

**Step 3: Write minimal implementation**

Implement small pure helpers for:

- nav-item construction
- `all` vs agent page mode
- grouped rendering preparation

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/lib/toolSkillsManagement.ts cmd/skillflow/frontend/tests/toolSkillsManagement.test.mjs
git commit -m "test: cover tool skills management helpers"
```

### Task 7: Add failing frontend integration tests and implement the `All + agents` UI

**Files:**
- Modify: `cmd/skillflow/frontend/src/pages/ToolSkills.tsx`
- Modify: `cmd/skillflow/frontend/src/lib/backend.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/en.ts`
- Modify: `cmd/skillflow/frontend/src/i18n/zh.ts`
- Modify: `cmd/skillflow/frontend/tests/toolSkillsPanels.test.mjs`
- Modify: `cmd/skillflow/frontend/tests/agentSettings.test.mjs`
- Create: `cmd/skillflow/frontend/tests/toolSkillsManagementUi.test.mjs`

**Step 1: Write the failing test**

Add tests that prove:

- left rail shows `All` before agent entries
- selecting `All` renders the global group-management surface
- selecting an agent keeps `Skills | Memory`
- single skill toggle calls the new backend method
- group toggle calls the new backend method
- Codex shows the restart-required hint

**Step 2: Run test to verify it fails**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `FAIL` until `ToolSkills` uses the new APIs and surfaces.

**Step 3: Write minimal implementation**

Implement:

- `All + enabled agents` left navigation
- `All` page rendering with group CRUD and assignment UI
- grouped single-agent skill rendering
- single skill enable/disable toggles
- group enable/disable actions
- Codex restart hint
- keep existing memory panel behavior for real agents only

**Step 4: Run test to verify it passes**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

**Step 5: Commit**

```bash
git add cmd/skillflow/frontend/src/pages/ToolSkills.tsx cmd/skillflow/frontend/src/lib/backend.ts cmd/skillflow/frontend/src/i18n/en.ts cmd/skillflow/frontend/src/i18n/zh.ts cmd/skillflow/frontend/tests/toolSkillsManagementUi.test.mjs
git commit -m "feat: add all-agent skill grouping and toggles UI"
```

### Task 8: Regenerate Wails bindings after App API changes

**Files:**
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.js`
- Modify: `cmd/skillflow/frontend/wailsjs/go/main/App.d.ts`

**Step 1: Run binding generation**

Run:

```bash
make generate
```

Expected: the Wails JS bindings include the new App methods.

**Step 2: Verify generated files changed as expected**

Run:

```bash
rg -n "ListAllAgentSkills|SetAgentSkillEnabled|SetAgentSkillGroupEnabled|CreateAgentSkillGroup" cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts
```

Expected: matches present in both files.

**Step 3: Commit**

```bash
git add cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts
git commit -m "chore: regenerate wails bindings for agent skill management"
```

### Task 9: Sync feature and config documentation

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `docs/config.md`
- Modify: `docs/config_zh.md`
- Modify: `docs/architecture/contexts.md`
- Modify: `docs/architecture/contexts_zh.md`

**Step 1: Write the failing check**

Create a manual checklist that confirms:

- **My Agents** documents the new `All` entry and grouped enable/disable flows
- docs describe Codex restart requirement
- config docs describe `agentSkillManagement`
- architecture docs mention agentintegration owns agent skill management semantics

**Step 2: Run the check**

Run:

```bash
rg -n "All|全部|agentSkillManagement|enable|disable|Codex|group" docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/contexts.md docs/architecture/contexts_zh.md
```

Expected: missing or incomplete matches before the doc update.

**Step 3: Write minimal documentation updates**

Update:

- features docs in English and Chinese
- config docs in English and Chinese
- architecture context docs in English and Chinese if the new ownership needs explicit coverage
- feature-doc last-updated dates

**Step 4: Run the check again**

Run:

```bash
rg -n "All|全部|agentSkillManagement|enable|disable|Codex|group" docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/contexts.md docs/architecture/contexts_zh.md
```

Expected: relevant matches present in every updated doc.

**Step 5: Commit**

```bash
git add docs/features.md docs/features_zh.md docs/config.md docs/config_zh.md docs/architecture/contexts.md docs/architecture/contexts_zh.md
git commit -m "docs: describe agent skill grouping and enablement"
```

### Task 10: Final verification

**Files:**
- No code changes expected

**Step 1: Run targeted Go tests**

Run:

```bash
go test ./core/config/... ./core/agentintegration/... ./core/readmodel/... ./cmd/skillflow
```

Expected: `ok`

**Step 2: Run frontend unit tests**

Run:

```bash
cd cmd/skillflow/frontend && npm run test:unit
```

Expected: `ok`

**Step 3: Run frontend production build**

Run:

```bash
cd cmd/skillflow/frontend && npm run build
```

Expected: `ok`

**Step 4: Run focused grep checks for docs and bindings**

Run:

```bash
rg -n "ListAllAgentSkills|SetAgentSkillEnabled|SetAgentSkillGroupEnabled|agentSkillManagement" cmd/skillflow/frontend/wailsjs/go/main/App.js cmd/skillflow/frontend/wailsjs/go/main/App.d.ts docs/config.md docs/config_zh.md
```

Expected: matches present.

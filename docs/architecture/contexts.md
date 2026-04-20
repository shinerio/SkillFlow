# Bounded Contexts and Domain Model

## Context Map

| Context | Type | Owns |
|--------|------|------|
| `skillcatalog` | core | installed skill truth |
| `promptcatalog` | core | prompt truth |
| `agentintegration` | core | agent definitions, push/pull semantics, and agent-side skill enablement behavior |
| `skillsource` | supporting | tracked repositories, logical skill sources, and source discovery |
| `backup` | supporting | backup and restore planning |

`Settings`, `Dashboard`, `My Agents`, and similar screens are not bounded contexts. They are composed read surfaces.

Shell concerns such as tray, window state, launch-at-login, single-instance behavior, and app update are handled by `cmd/skillflow/` and `platform/`, not by a separate bounded context.

## `skillcatalog`

### Responsibilities

- own the installed skill library
- own installed instance identity and logical identity mapping
- own categories, installed metadata, and instance update state

### Aggregate Roots

- `InstalledSkill`
- `SkillCategory`

### Important Value Objects

- `SkillID`
- `SkillName`
- `LogicalSkillKey`
- `SkillSourceRef`
- `SkillStorageRef`
- `SkillVersionState`

### Domain Rules

- one `InstalledSkill` always represents one installed library instance
- `SkillID` is the instance identity and remains local to `skillcatalog`
- `LogicalSkillKey` is the cross-context identity
- name and absolute path are never the primary cross-context identifier
- for repository-backed skills, `LogicalSkillKey` is derived from `SkillSourceRef` and should not be persisted as an independent field

### Published Language

- `InstalledSkillSummary`
- `InstalledSkillVersionView`
- `SkillCategorySummary`

## `promptcatalog`

### Responsibilities

- own the prompt library
- own prompt categories, content, and metadata
- own prompt import/export behavior

### Aggregate Roots

- `Prompt`
- `PromptCategory`

### Important Value Objects

- `PromptID`
- `PromptName`
- `PromptContent`
- `PromptStorageRef`
- `PromptLinkSet`
- `PromptMediaSet`

### Domain Rules

- prompts are first-class content, not a sub-type of skill
- category is a domain concept, not just a directory name
- current import conflict semantics are name-based conflicts against existing prompts

### Published Language

- `PromptSummary`
- `PromptCategorySummary`

## `agentintegration`

### Responsibilities

- own agent profiles
- own built-in agent profile defaults
- own scan, push, and pull semantics
- own agent-side skill enable/disable semantics
- own push and pull conflict detection
- own agent-side presence meaning

### Aggregate Roots

- `AgentProfile`
- `AgentPushPolicy`

### Important Value Objects

- `AgentID`
- `AgentName`
- `AgentType`
- `ScanDirectorySet`
- `PushDirectory`
- `AgentSkillRef`
- `PushConflict`
- `PullConflict`
- `AgentSkillObservation`

### Domain Rules

- this context does not own skill content truth
- `seenInAgentScan` and `pushed` are different states
- conflict detection belongs here, not in UI-facing code

### Published Language

- `AgentSummary`
- `AgentSkillPresence`
- `PushPlan`
- `PullPlan`

## `skillsource`

### Responsibilities

- own starred GitHub repositories and other tracked external repositories
- own logical skill sources derived from repositories
- own source synchronization status
- own source-side candidate discovery
- provide version hints for source-backed installed skills

### Aggregate Roots

- `StarRepo`
- `SkillSource`

### Important Value Objects

- `StarRepoID`
- `SourceID`
- `RepoSource`
- `SourceSubPath`
- `SourceSyncStatus`
- `SourceCacheRef`
- `SourceSkillCandidate`

### Domain Rules

- this context does not own installed skill truth
- `StarRepo` represents a tracked GitHub repository at the repository level
- `SkillSource` represents one logical skill source identified by `repo + subpath`
- one `StarRepo` may contain many `SkillSource`
- one installed logical skill should map to exactly one `SkillSource`
- it owns whether a starred repository is tracked and what candidates were discovered from that source
- installed or updatable state should be derived by combining source data with `skillcatalog` and `agentintegration` published language

### Domain Interpretation

`StarRepo` is a repository-level domain model. `SkillSource` is the skill-level source model under that repository. They should not be collapsed into one concept.

### Published Language

- `StarRepoSummary`
- `SkillSourceSummary`
- `SourceSkillCandidateView`
- `SourceVersionHint`

## `backup`

### Responsibilities

- own backup configuration
- own backup scope and restore planning
- own backup snapshot comparison

### Aggregate Roots

- `BackupProfile`
- `GitBackupProfile`

### Important Value Objects

- `BackupTarget`
- `BackupScope`
- `BackupSnapshot`
- `BackupChangeSet`
- `RestorePlan`
- `RestoreConflict`

### Domain Rules

- this context does not own business truth for skills or prompts
- it owns how those truths are captured, restored, and verified

## Shared Kernel

Only highly stable concepts belong in `shared/`:

- `LogicalSkillKey`
- common domain error contracts
- base domain event contracts

Context-local instance IDs such as `SkillID` and `PromptID` stay inside their owning bounded contexts unless future cross-context use proves otherwise.

## Unified Skill Identity and State Model

This rule is normative.

### Identity Layers

- instance identity: `SkillID`
- logical identity: `LogicalSkillKey`

### Logical Key Derivation

- repository-backed skills:
  - `LogicalSkillKey` is derived from `SkillSourceRef`
  - canonical form: `git:<repo-source>#<subpath>`
  - it should not be persisted as a separate field when `SkillSourceRef` already exists
- manually imported or non-repository-backed skills:
  - `LogicalSkillKey` should be generated at import time as a stable content-based key
  - recommended form: `content:<hash>`
  - the hash should be computed from a canonicalized snapshot of the imported skill payload, not from absolute paths or local-only metadata
  - once generated, it should be persisted because there is no `SkillSourceRef` to derive it from later

### Status Semantics

- `installed`: at least one installed skill instance exists for the logical key
- `imported`: wording alias for `installed` in external-source flows
- `pushed`: the logical skill exists in an agent's configured push directory
- `seenInAgentScan`: the logical skill was detected in scan directories, which does not imply SkillFlow pushed it
- `updatable`: the installed skill instance or agent-side copy is behind the currently known source state

### Cross-Context Rule

Frontend pages and cross-context read models must not infer skill sameness from name or absolute path alone. Logical identity must be resolved in backend code.

### Installed Mapping Constraint

For repository-backed skills, one logical source identified by `repo + subpath` should correspond to exactly one installed skill in the library. Re-importing the same logical source should be treated as an already-installed conflict or an update path, not as a second installed instance.

For manually imported skills, the logical key is generated from the imported content snapshot. If that generated key already exists in the installed library, the import should be treated as an already-installed conflict or an explicit reconciliation path rather than creating another installed skill with the same logical identity.

## Cross-Context Collaboration

### Write Coordination

Cross-context writes should use explicit orchestration, not direct aggregate sharing.

Examples:

- source import -> installed skill creation -> optional auto-push -> optional backup
- installed skill update -> pushed copy refresh
- restore backup -> context-specific rebuild steps

### Read Composition

Cross-context UI views should use read models that combine published language from multiple contexts.

Examples:

- Dashboard
- My Agents
- Settings
- source candidate list enriched with installed and pushed status

*Last updated: 2026-04-12*

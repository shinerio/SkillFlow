# SkillFlow Configuration File Reference

> 🌐 [中文版](config_zh.md) | **English**

This document explains the on-disk format of SkillFlow's persisted configuration and metadata files.

Examples below use placeholders such as `<AppDataDir>` and `<RepoCacheDir>`:

- `<AppDataDir>`: the app data directory returned by `config.AppDataDir()`
- `<RepoCacheDir>`: the local repo-clone cache root from `config_local.json.repoCacheDir`; it defaults to `<AppDataDir>/cache/repos`

The actual starred-repository file name is `star_repos.json` (plural), not `star_repo.json`.

All synced app content now lives under `<AppDataDir>`. Only the heavyweight starred-repo clone cache may move to a different local-only `<RepoCacheDir>`.

## Quick Summary

| File | Purpose | Synced |
|------|---------|--------|
| `config.json` | Shared, sync-safe settings | Yes |
| `config_local.json` | Machine-specific paths, secrets, and local runtime state | No |
| `star_repos.json` | Starred repository identity metadata | Yes |
| `star_repos_local.json` | Local-only starred-repo runtime sync state overlay | No |
| `prompts/<category>/<name>/prompt.json` | Prompt metadata such as description, related images, and web links | Yes |
| `meta/<skill-id>.json` | One sidecar metadata file per installed skill | Yes |
| `meta_local/<skill-id>.local.json` | Local-only per-skill volatile metadata overlay | No |
| `cache/viewstate/*.json` | Local derived UI/cache snapshots | No |
| `runtime/*.json`, `runtime/helper.lock` | Local daemon/UI process coordination state | No |

The table describes file roles. Synced content trees stay under `<AppDataDir>`, while starred-repo clone data may live under a separate local-only `<RepoCacheDir>`.

## `cache/viewstate/*.json`

Path: `<AppDataDir>/cache/viewstate/*.json`

These files store local-only derived state used to speed up page entry and reduce repeated directory scans. Typical payloads include installed-skill card snapshots and agent-presence indexes.

Rules:

- They are optimization artifacts, not source-of-truth records.
- They must be rebuilt from `skills/`, `meta/`, agent directories, and other existing truth-layer files.
- They must not be uploaded by cloud backup or written back into synced metadata files.
- Cross-device cache differences are expected and harmless.

## `runtime/*.json`, `runtime/helper.lock`

Path: `<AppDataDir>/runtime/`

These files store local-only daemon/UI process coordination state. Typical files include `helper-control.json`, `ui-control.json`, `daemon-service.json`, and `helper.lock`.

Rules:

- They contain loopback endpoint addresses, random tokens, and process IDs that are valid only on the current machine and current run.
- `helper-control.json` and `helper.lock` are historical file names for the background process host. In the current runtime model, that host is the long-lived `daemon`.
- They must be excluded from cloud backup, Git sync, and cross-device restore.
- Older Git backups that once tracked `runtime/` are cleaned up on the next push.
- They may be deleted while SkillFlow is fully stopped; the app recreates them on the next launch.

## `config.json`

Path: `<AppDataDir>/config.json`

`config.json` stores settings that are safe to move across devices. It must not contain machine-specific absolute paths or sensitive credentials.

Before config is loaded, SkillFlow runs the one-time terminology cutover in `core/platform/upgrade`. Legacy `tools`-based keys are rewritten in place to the new `agents`-based schema, and the removed legacy `skillStatusVisibility` field is deleted in place. Runtime code only reads the latest schema.

### Example

```json
{
  "defaultCategory": "Default",
  "logLevel": "info",
  "repoScanMaxDepth": 5,
  "agents": [
    { "name": "claude-code", "enabled": true },
    { "name": "codex", "enabled": true },
    { "name": "copilot", "enabled": true },
    { "name": "gemini-cli", "enabled": false }
  ],
  "cloud": {
    "provider": "git",
    "enabled": true,
    "syncIntervalMinutes": 30
  },
  "cloudProfiles": {
    "git": {
      "bucketName": "",
      "remotePath": "team-a/backup/skillflow/",
      "credentials": {
        "repo_url": "https://github.com/example/skillflow-backup.git",
        "branch": "main",
        "username": "alice"
      }
    },
    "aws": {
      "bucketName": "skillflow-backup",
      "remotePath": "nightly/skillflow/",
      "credentials": {
        "region": "us-east-1"
      }
    }
  },
  "skippedUpdateVersion": "v1.2.3"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `defaultCategory` | string | Default category used when importing or creating skills. |
| `logLevel` | string | Backend log level. Valid values are `debug`, `info`, and `error`. Invalid values are normalized to `error`. |
| `repoScanMaxDepth` | number | Maximum recursive depth used when scanning agent directories and repos. Values are normalized to the `1-20` range, with `5` as the default. |
| `agents` | object[] | Built-in agent enable/disable state only. Path-related agent settings are stored in `config_local.json`. |
| `agents[].name` | string | Built-in agent name such as `claude-code`, `codex`, `copilot`, `gemini-cli`, `opencode`, or `openclaw`. |
| `agents[].enabled` | boolean | Whether this built-in agent is enabled in the UI and scanning/push flows. |
| `cloud` | object | Active cloud-backup selection and scheduling state. |
| `cloud.provider` | string | Active provider name, such as `git`, `aws`, `aliyun`, `azure`, `google`, `huawei`, or `tencent`. |
| `cloud.enabled` | boolean | Whether cloud backup is enabled. |
| `cloud.syncIntervalMinutes` | number | Automatic backup interval in minutes. `0` means "only back up on mutations". |
| `cloudProfiles` | object | Provider-specific sync-safe settings keyed by provider name. |
| `cloudProfiles.<provider>.bucketName` | string | Bucket/container name for object-storage providers. Usually empty for `git`. |
| `cloudProfiles.<provider>.remotePath` | string | Remote prefix under the provider root. It is normalized to always end in `skillflow/`. |
| `cloudProfiles.<provider>.credentials` | object | Provider settings that are safe to sync, such as endpoints or repo URLs. Secrets are split into `config_local.json`. |
| `skippedUpdateVersion` | string | App version tag the user chose to skip in the startup update prompt. |

For built-in agents, `pushDir` and `scanDirs` may differ. Copilot pushes user-installed skills to `~/.copilot/skills`, while its default scan list also includes shared personal-skill directories and the versioned Copilot CLI `builtin-skills` directory when one is detected locally.

### Sync-safe cloud credential keys

Only the following credential keys are persisted in `config.json`:

| Provider | Keys stored in `config.json` | Meaning |
|----------|------------------------------|---------|
| `aliyun` | `endpoint` | OSS endpoint host. |
| `aws` | `region` | AWS S3 region. |
| `azure` | `account_name`, `service_url` | Azure Storage account identity and service endpoint. |
| `git` | `repo_url`, `branch`, `username` | Remote Git repo address, branch, and optional HTTPS username. |
| `google` | none | Google credentials are always local-only. |
| `huawei` | `endpoint` | OBS endpoint host. |
| `tencent` | `endpoint` | COS endpoint host. |

## `config_local.json`

Path: `<AppDataDir>/config_local.json`

`config_local.json` stores machine-specific paths, local runtime state, and secrets. It is intentionally excluded from cloud backup and Git sync.

The same startup cutover also rewrites legacy local keys such as `autoPushTools` / `tools` to `autoPushAgents` / `agents` before `config_local.json` is read.

### Example

```json
{
  "repoCacheDir": "/Users/demo/Library/Application Support/SkillFlow/cache/repos",
  "autoUpdateSkills": true,
  "autoPushAgents": ["codex", "gemini-cli"],
  "launchAtLogin": true,
  "agents": [
    {
      "name": "claude-code",
      "scanDirs": [
        "/Users/demo/.claude/skills",
        "/Users/demo/.claude/plugins/marketplaces"
      ],
      "pushDir": "/Users/demo/.claude/skills",
      "memoryPath": "/Users/demo/.claude/CLAUDE.md",
      "rulesDir": "/Users/demo/.claude/rules",
      "custom": false,
      "enabled": true
    },
    {
      "name": "copilot",
      "scanDirs": [
        "/Users/demo/.claude/skills",
        "/Users/demo/.agents/skills",
        "/Users/demo/.copilot/pkg/universal/1.0.31/builtin-skills"
      ],
      "pushDir": "/Users/demo/.copilot/skills",
      "memoryPath": "/Users/demo/.copilot/copilot-instructions.md",
      "rulesDir": "",
      "custom": false,
      "enabled": true
    },
    {
      "name": "my-custom-agent",
      "scanDirs": ["/Users/demo/work/my-agent/skills"],
      "pushDir": "/Users/demo/work/my-agent/skills",
      "memoryPath": "/Users/demo/work/my-agent/AGENTS.md",
      "rulesDir": "/Users/demo/work/my-agent/rules",
      "custom": true,
      "enabled": true
    }
  ],
  "cloudCredentialsByProvider": {
    "git": {
      "token": "ghp_xxx"
    },
    "aws": {
      "access_key_id": "AKIA...",
      "secret_access_key": "secret"
    },
    "google": {
      "service_account_json": "{\"type\":\"service_account\",\"project_id\":\"demo\"}"
    }
  },
  "proxy": {
    "mode": "manual",
    "url": "http://127.0.0.1:7890"
  },
  "window": {
    "width": 1440,
    "height": 920
  }
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `repoCacheDir` | string | Absolute local root used for cloned starred-repository caches. Defaults to `<AppDataDir>/cache/repos` when empty. |
| `autoUpdateSkills` | boolean | Whether starred-repo refresh on this device should automatically update matching installed git-backed skills in **My Skills**. |
| `autoPushAgents` | string[] | Agent names that should receive auto-push after import/update flows. Values are trimmed and deduplicated. |
| `launchAtLogin` | boolean | Whether SkillFlow should register itself as a login/startup item on the current machine. |
| `agents` | object[] | Agent path configuration. This includes built-in agents and all custom agents. |
| `agents[].name` | string | Agent identifier. |
| `agents[].scanDirs` | string[] | Local directories scanned for external skills from this agent. |
| `agents[].pushDir` | string | Local target directory used when pushing skills to this agent. |
| `agents[].memoryPath` | string | Local path to the agent's main memory file used by **My Memory** push and **My Agents** memory preview. |
| `agents[].rulesDir` | string | Local path to the agent's rules directory used by memory module push and preview. This may be empty for built-in agents that do not expose a first-party rules directory, such as Copilot. |
| `agents[].custom` | boolean | `true` for user-created custom agents, `false` for built-in agents. |
| `agents[].enabled` | boolean | Stored for every agent, but only meaningful for custom agents in `config_local.json`; built-in enable state comes from `config.json`. |
| `cloudCredentialsByProvider` | object | Sensitive provider credentials keyed by provider name. |
| `cloudCredentialsByProvider.<provider>` | object | Local-only credential map for one provider. |
| `proxy` | object | Local proxy settings used for outbound HTTP requests. |
| `proxy.mode` | string | Proxy mode: `none`, `system`, or `manual`. |
| `proxy.url` | string | Manual proxy URL. Used only when `mode` is `manual`. |
| `window` | object | Last persisted window size. This key is omitted until a valid size has been saved. |
| `window.width` | number | Window width in pixels. |
| `window.height` | number | Window height in pixels. |

### Local-only cloud credential keys

These keys are intentionally kept out of `config.json`:

| Provider | Keys stored in `config_local.json` | Meaning |
|----------|------------------------------------|---------|
| `aliyun` | `access_key_id`, `access_key_secret` | OSS access key pair. |
| `aws` | `access_key_id`, `secret_access_key` | AWS access key pair. |
| `azure` | `account_key` | Azure Storage account secret. |
| `git` | `token` | HTTPS access token. |
| `google` | `service_account_json` | Inline service-account JSON or a local key-file path. |
| `huawei` | `access_key_id`, `secret_access_key` | OBS access key pair. |
| `tencent` | `secret_id`, `secret_key` | COS credential pair. |

## `star_repos.json`

Path: `<AppDataDir>/star_repos.json`

`star_repos.json` stores the synced identity state of tracked starred repositories.

### Example

```json
[
  {
    "url": "https://github.com/example/awesome-skills.git",
    "name": "example/awesome-skills",
    "source": "github.com/example/awesome-skills"
  }
]
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `url` | string | Original Git clone URL entered or seeded for the repository. |
| `name` | string | Human-friendly repo name, usually `<owner>/<repo>` or `<group>/<subgroup>/<repo>`. |
| `source` | string | Canonical repo source key used for matching across modules, usually `<host>/<repo-path>`. |

Runtime `StarRepo.LocalDir` is re-derived from the current `config_local.json.repoCacheDir` plus the repo URL. It is no longer persisted in the synced `star_repos.json` payload.

## `star_repos_local.json`

Path: `<AppDataDir>/star_repos_local.json`

This local-only overlay stores per-repo volatile sync state that should not be synced across devices.

### Example

```json
{
  "repos": {
    "github.com/example/awesome-skills": {
      "lastSync": "2026-03-11T08:15:00Z",
      "syncError": "authentication failed"
    }
  }
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `repos` | object | Map keyed by repo source key (or URL fallback) to local sync state entries. |
| `repos.<key>.lastSync` | string | Last successful sync timestamp on the current device (RFC3339). |
| `repos.<key>.syncError` | string | Last sync error message on the current device. Omitted when empty. |

## `prompts/<category>/<name>/prompt.json`

Path: `<AppDataDir>/prompts/<category>/<name>/prompt.json`

Each prompt keeps its body in the sibling `system.md` file and stores prompt-card metadata in `prompt.json`.

### Example

```json
{
  "name": "Review API",
  "description": "Review backend changes",
  "imageURLs": [
    "https://cdn.example.com/review-1.png",
    "https://cdn.example.com/review-2.png"
  ],
  "webLinks": [
    {
      "label": "PRD",
      "url": "https://docs.example.com/prd"
    },
    {
      "label": "Preview",
      "url": "https://preview.example.com/review"
    }
  ],
  "createdAt": "2026-03-15T13:00:00Z",
  "updatedAt": "2026-03-15T13:05:00Z"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `name` | string | Prompt name. It normally matches the prompt directory name and stays globally unique in the prompt library. |
| `description` | string | Optional short prompt description shown on the card. |
| `imageURLs` | string[] | Optional related image URLs. SkillFlow currently accepts up to 3 entries and requires `http` or `https` URLs. |
| `webLinks` | object[] | Optional structured web links rendered as clickable chips on the prompt card. |
| `webLinks[].label` | string | Visible link text. |
| `webLinks[].url` | string | External URL opened by the card chip. Only `http` and `https` URLs are persisted. |
| `createdAt` | string | Prompt creation timestamp (RFC3339). |
| `updatedAt` | string | Last metadata update timestamp (RFC3339). |

## `meta/<skill-id>.json`

Path: `<AppDataDir>/meta/<skill-id>.json`

Each installed skill gets one JSON sidecar file named after `Skill.ID` rather than the skill name.

### Example

```json
{
  "ID": "0f4b6f23-4f1e-4c56-a1fa-7fa0f7ce1234",
  "Name": "code-review",
  "Path": "skills/Engineering/code-review",
  "Category": "Engineering",
  "Source": "github",
  "SourceURL": "https://github.com/example/skill-collection.git",
  "SourceSubPath": "engineering/code-review",
  "SourceSHA": "8f3d4c2",
  "LatestSHA": "31ad9be",
  "InstalledAt": "2026-03-10T09:30:00Z",
  "UpdatedAt": "2026-03-11T07:45:00Z"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `ID` | string | Stable instance UUID for the installed skill. This value also becomes the metadata file name. |
| `Name` | string | Skill directory name. |
| `Path` | string | Local skill directory path. When the directory is under the synchronized root, it is stored as a forward-slash relative path such as `skills/Engineering/code-review`. |
| `Category` | string | SkillFlow category folder that currently contains the skill. |
| `Source` | string | Install source type. Current values are `github` and `manual`. |
| `SourceURL` | string | Original source repository or source location URL for git-backed installs. Usually empty for manual imports. |
| `SourceSubPath` | string | Relative path inside the source repo when the skill was imported from a subdirectory. |
| `SourceSHA` | string | Commit SHA recorded when the installed skill was last imported or updated. |
| `LatestSHA` | string | Latest remote SHA most recently discovered by the update checker. |
| `InstalledAt` | string | Timestamp when the skill was first imported into SkillFlow. |
| `UpdatedAt` | string | Timestamp when the skill metadata was last changed, such as category moves or updates. |

### Important note

`meta/<skill-id>.json` stores installation state, not the YAML frontmatter parsed from `SKILL.md`. Frontmatter fields such as `name`, `description`, and `allowed-agents` stay in the skill content itself.

## `meta_local/<skill-id>.local.json`

Path: `<AppDataDir>/meta_local/<skill-id>.local.json`

This file stores local-only, high-churn per-skill fields that should not be synced across devices.

### Example

```json
{
  "lastCheckedAt": "2026-03-11T08:00:00Z"
}
```

### Keys

| Key | Type | Meaning |
|-----|------|---------|
| `lastCheckedAt` | string | Timestamp of the most recent update-check attempt on the current device. |

## `memory/memory_local.json`

Local-only memory configuration. Excluded from cloud backup and git sync.

**Location:** `<appDataDir>/memory/memory_local.json`

**Schema:**

```json
{
  "pushConfigs": {
    "<agentType>": {
      "mode": "merge" | "takeover",
      "autoPush": true | false
    }
  },
  "moduleStates": {
    "<moduleName>": {
      "enabled": true | false
    }
  },
  "pushState": {
    "<agentType>": {
      "lastPushedAt": "2026-03-21T10:00:00Z",
      "lastPushedHash": "<sha256-hex>"
    }
  }
}
```

**Fields:**

| Section | Key | Type | Description |
|---------|-----|------|-------------|
| `pushConfigs` | `<agentType>` | object | Per-agent push configuration |
| `pushConfigs.<agent>.mode` | — | string | `"merge"` or `"takeover"` |
| `pushConfigs.<agent>.autoPush` | — | bool | Whether this agent auto-syncs all memories after local edits |
| `moduleStates` | `<moduleName>` | object | Local global state for one module memory |
| `moduleStates.<module>.enabled` | — | bool | Whether the module participates in auto sync and default full-agent push |
| `pushState` | `<agentType>` | object | Per-agent last push tracking |
| `pushState.<agent>.lastPushedAt` | — | RFC3339 string | Timestamp of last successful push |
| `pushState.<agent>.lastPushedHash` | — | string | SHA-256 of the actual content most recently pushed to this agent |

**Notes:**

- There is no persisted per-module push-target list anymore. Manual batch push selections are temporary UI state only.
- `moduleStates` is local-only UI/distribution state. When a module has no stored entry yet, SkillFlow treats it as enabled by default for backward compatibility.
- Disabling a module removes it from auto-sync pushes and from the default full-agent push set, but manual batch push can still select it explicitly.
- A partial batch push stores the pushed snapshot hash, so the same agent can still show `pendingPush` when the current local library contains more modules than the last pushed selection.

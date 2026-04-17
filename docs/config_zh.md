# SkillFlow 配置文件参考

> 🌐 **中文** | [English](config.md)

本文档说明 SkillFlow 持久化到磁盘的配置文件与元数据文件格式。

下文使用了 `<AppDataDir>` 与 `<RepoCacheDir>` 这类占位符：

- `<AppDataDir>`：`config.AppDataDir()` 返回的应用数据目录
- `<RepoCacheDir>`：`config_local.json.repoCacheDir` 中保存的本地仓库克隆缓存根目录；默认为 `<AppDataDir>/cache/repos`

当前所有会参与同步的应用内容都固定保存在 `<AppDataDir>` 下。只有体积较大的收藏仓库 clone cache 可以单独迁移到本地专属的 `<RepoCacheDir>`。

## 快速概览

| 文件 | 作用 | 是否参与同步 |
|------|------|--------------|
| `config.json` | 共享的、可同步的安全配置 | 是 |
| `config_local.json` | 机器相关路径、敏感信息和本地运行状态 | 否 |
| `star_repos.json` | 收藏仓库的身份元数据 | 是 |
| `star_repos_local.json` | 收藏仓库本地同步状态覆盖文件 | 否 |
| `prompts/<category>/<name>/prompt.json` | 提示词卡片元数据，如描述、关联图片和网页链接 | 是 |
| `meta/<skill-id>.json` | 每个已安装 Skill 的 sidecar 元数据 | 是 |
| `meta_local/<skill-id>.local.json` | 每个 Skill 的本地易变元数据覆盖文件 | 否 |
| `cache/viewstate/*.json` | 本地派生 UI / 缓存快照 | 否 |
| `runtime/*.json`、`runtime/helper.lock` | 本地 daemon/UI 进程协调状态 | 否 |

这个表描述的是文件职责。会参与同步的内容目录固定保存在 `<AppDataDir>` 下；而收藏仓库的 clone 数据则允许位于单独的本地 `<RepoCacheDir>`。

## `cache/viewstate/*.json`

路径：`<AppDataDir>/cache/viewstate/*.json`

这些文件保存只属于当前设备的派生状态，用于加快页面进入速度并减少重复目录扫描。典型内容包括已安装 Skill 卡片快照和智能体 presence 索引。

规则：

- 它们只是性能优化产物，不是权威真值。
- 必须从 `skills/`、`meta/`、智能体目录等现有真值层文件重建。
- 不能被云备份上传，也不能反向写回任何可同步元数据文件。
- 不同设备上的缓存内容不同是正常且无害的。

## `runtime/*.json`、`runtime/helper.lock`

路径：`<AppDataDir>/runtime/`

这些文件保存 daemon/UI 双进程在当前设备上的本地协调状态。典型文件包括 `helper-control.json`、`ui-control.json`、`daemon-service.json` 和 `helper.lock`。

规则：

- 文件里保存的是回环地址、随机 token 和 PID，只对当前设备、当前这次运行有效。
- `helper-control.json` 与 `helper.lock` 仍沿用历史文件名，但它们现在对应的是长期驻留的后台 `daemon` 进程宿主。
- 它们必须被排除在云备份、Git 同步和跨设备恢复之外。
- 如果旧版本曾把 `runtime/` 跟踪进 Git 备份，下一次推送时会自动清理掉。
- 在 SkillFlow 完全退出时可以删除；应用下次启动会自动重建。

## `config.json`

路径：`<AppDataDir>/config.json`

`config.json` 只保存可跨设备移动的安全配置，不应包含机器相关绝对路径或敏感凭据。

在真正加载配置之前，SkillFlow 会先执行 `core/platform/upgrade` 里的单次术语割接。旧版基于 `tools` 的键会被原地改写为新版基于 `agents` 的 schema，已经移除的旧字段 `skillStatusVisibility` 也会在启动时被原地删除，运行时代码只读取最新 schema。

### 示例

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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `defaultCategory` | string | 导入或创建 Skill 时默认使用的分类。 |
| `logLevel` | string | 后端日志级别。可选值为 `debug`、`info`、`error`；非法值会被归一化为 `error`。 |
| `repoScanMaxDepth` | number | 扫描智能体目录与仓库时允许的最大递归深度。会被归一化到 `1-20` 范围内，默认值是 `5`。 |
| `agents` | object[] | 只保存内置智能体的启用/停用状态。路径相关设置保存在 `config_local.json`。 |
| `agents[].name` | string | 内置智能体名，例如 `claude-code`、`codex`、`copilot`、`gemini-cli`、`opencode`、`openclaw`。 |
| `agents[].enabled` | boolean | 该内置智能体是否在界面与扫描/推送流程中启用。 |
| `cloud` | object | 当前选中的云备份提供方及调度状态。 |
| `cloud.provider` | string | 当前激活的 provider 名称，例如 `git`、`aws`、`aliyun`、`azure`、`google`、`huawei`、`tencent`。 |
| `cloud.enabled` | boolean | 是否启用云备份。 |
| `cloud.syncIntervalMinutes` | number | 自动备份间隔（分钟）。`0` 表示“仅在状态变更后备份”。 |
| `cloudProfiles` | object | 以 provider 名称为键的 provider 级同步安全配置。 |
| `cloudProfiles.<provider>.bucketName` | string | 对象存储 provider 使用的 bucket/container 名称；`git` 一般为空。 |
| `cloudProfiles.<provider>.remotePath` | string | provider 根目录下的远端前缀。保存时会被规范化为以 `skillflow/` 结尾。 |
| `cloudProfiles.<provider>.credentials` | object | 允许同步的 provider 配置项，例如 endpoint 或 repo URL。敏感项会拆分到 `config_local.json`。 |
| `skippedUpdateVersion` | string | 用户在启动更新提示里选择“跳过此版本”后记录的应用版本号。 |

对于内置智能体，`pushDir` 与 `scanDirs` 可以不同。Copilot 就是一个典型例子：SkillFlow 会把用户安装的 Copilot Skill 推送到 `~/.copilot/skills`，同时默认扫描共享个人 Skill 目录，以及本机存在时自动发现的 Copilot CLI 版本化 `builtin-skills` 目录。

### 会写入 `config.json` 的云配置键

只有下面这些凭据键会出现在 `config.json` 中：

| Provider | 保存在 `config.json` 的键 | 作用 |
|----------|---------------------------|------|
| `aliyun` | `endpoint` | OSS 的 endpoint 主机名。 |
| `aws` | `region` | AWS S3 区域。 |
| `azure` | `account_name`, `service_url` | Azure Storage 账号标识与服务地址。 |
| `git` | `repo_url`, `branch`, `username` | 远端 Git 仓库地址、分支、可选的 HTTPS 用户名。 |
| `google` | 无 | Google 凭据始终只保存在本地。 |
| `huawei` | `endpoint` | OBS 的 endpoint 主机名。 |
| `tencent` | `endpoint` | COS 的 endpoint 主机名。 |

## `config_local.json`

路径：`<AppDataDir>/config_local.json`

`config_local.json` 保存机器相关路径、本地运行状态与敏感凭据。它会被明确排除在云备份和 Git 同步之外。

同一套启动割接逻辑也会在读取 `config_local.json` 之前，把旧版 `autoPushTools` / `tools` 改写成新版 `autoPushAgents` / `agents`。

### 示例

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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `repoCacheDir` | string | 本机用于保存收藏仓库 clone cache 的绝对根路径。为空时会回退到 `<AppDataDir>/cache/repos`。 |
| `autoUpdateSkills` | boolean | 当前设备在刷新收藏仓库后，是否应自动把匹配的已安装 Git Skill 更新到 **我的skills**。 |
| `autoPushAgents` | string[] | 在导入/更新后自动推送到哪些智能体。保存前会去空格并去重。 |
| `launchAtLogin` | boolean | 当前设备上是否把 SkillFlow 注册为开机/登录自启动项。 |
| `agents` | object[] | 智能体路径配置，既包含内置智能体，也包含自定义智能体。 |
| `agents[].name` | string | 智能体标识名。 |
| `agents[].scanDirs` | string[] | 扫描该智能体外部 Skill 的本地目录列表。 |
| `agents[].pushDir` | string | 将 Skill 推送到该智能体时使用的目标目录。 |
| `agents[].memoryPath` | string | 智能体主记忆文件的本地路径，供 **我的记忆** 推送和 **我的智能体** 记忆预览使用。 |
| `agents[].rulesDir` | string | 智能体规则目录的本地路径，供记忆模块推送和预览使用。对于没有官方规则目录语义的内置智能体（如 Copilot），这里可以为空字符串。 |
| `agents[].custom` | boolean | `true` 表示用户创建的自定义智能体，`false` 表示内置智能体。 |
| `agents[].enabled` | boolean | 每个智能体都会带上这个字段，但在 `config_local.json` 中它只对自定义智能体有实际意义；内置智能体的启用状态以 `config.json` 为准。 |
| `cloudCredentialsByProvider` | object | 以 provider 名称为键的敏感凭据集合。 |
| `cloudCredentialsByProvider.<provider>` | object | 某个 provider 的本地专属凭据键值表。 |
| `proxy` | object | 出站 HTTP 请求使用的本地代理设置。 |
| `proxy.mode` | string | 代理模式：`none`、`system`、`manual`。 |
| `proxy.url` | string | 手动代理 URL。仅在 `mode` 为 `manual` 时使用。 |
| `window` | object | 最近一次保存的窗口尺寸。只有保存过有效尺寸后才会出现。 |
| `window.width` | number | 窗口宽度（像素）。 |
| `window.height` | number | 窗口高度（像素）。 |

### 只保存在 `config_local.json` 的云凭据键

下面这些键不会写入 `config.json`：

| Provider | 保存在 `config_local.json` 的键 | 作用 |
|----------|---------------------------------|------|
| `aliyun` | `access_key_id`, `access_key_secret` | OSS 访问密钥对。 |
| `aws` | `access_key_id`, `secret_access_key` | AWS 访问密钥对。 |
| `azure` | `account_key` | Azure Storage 账号密钥。 |
| `git` | `token` | HTTPS 访问令牌。 |
| `google` | `service_account_json` | 内联 service-account JSON，或本地密钥文件路径。 |
| `huawei` | `access_key_id`, `secret_access_key` | OBS 访问密钥对。 |
| `tencent` | `secret_id`, `secret_key` | COS 凭据对。 |

## `star_repos.json`

路径：`<AppDataDir>/star_repos.json`

`star_repos.json` 保存已跟踪收藏仓库的可同步身份状态。

### 示例

```json
[
  {
    "url": "https://github.com/example/awesome-skills.git",
    "name": "example/awesome-skills",
    "source": "github.com/example/awesome-skills"
  }
]
```

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `url` | string | 用户输入或系统内置的原始 Git 克隆地址。 |
| `name` | string | 便于展示的仓库名，通常是 `<owner>/<repo>` 或 `<group>/<subgroup>/<repo>`。 |
| `source` | string | 跨模块关联使用的规范化仓库源键，通常形如 `<host>/<repo-path>`。 |

运行时里的 `StarRepo.LocalDir` 会根据当前 `config_local.json.repoCacheDir` 和仓库 URL 即时推导，不再写入可同步的 `star_repos.json`。

## `star_repos_local.json`

路径：`<AppDataDir>/star_repos_local.json`

这个仅本地文件用于保存每个收藏仓库变化频繁、且不应跨设备同步的同步状态。

### 示例

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

### 字段说明

| 键 | 类型 | 说明 |
|----|------|------|
| `repos` | object | 以仓库 source 键（或 URL 兜底）为 key 的本地同步状态映射。 |
| `repos.<key>.lastSync` | string | 当前设备最近一次成功同步时间（RFC3339）。 |
| `repos.<key>.syncError` | string | 当前设备最近一次同步失败错误信息；为空时省略。 |

## `prompts/<category>/<name>/prompt.json`

路径：`<AppDataDir>/prompts/<category>/<name>/prompt.json`

每个提示词的正文保存在同级的 `system.md` 中，提示词卡片元数据则保存在 `prompt.json`。

### 示例

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

### 字段说明

| 键 | 类型 | 说明 |
|----|------|------|
| `name` | string | 提示词名称。通常与提示词目录名一致，并在整个提示词库内保持全局唯一。 |
| `description` | string | 可选的简短描述，会显示在提示词卡片上。 |
| `imageURLs` | string[] | 可选的关联图片 URL。当前最多支持 3 条，且必须是 `http` 或 `https` URL。 |
| `webLinks` | object[] | 可选的结构化网页链接，会渲染为提示词卡片上的可点击胶囊。 |
| `webLinks[].label` | string | 链接展示文本。 |
| `webLinks[].url` | string | 点击卡片胶囊后打开的外部 URL。持久化时只允许 `http` 和 `https`。 |
| `createdAt` | string | 提示词创建时间（RFC3339）。 |
| `updatedAt` | string | 最近一次元数据更新时间（RFC3339）。 |

## `meta/<skill-id>.json`

路径：`<AppDataDir>/meta/<skill-id>.json`

每个已安装 Skill 都会有一个 sidecar JSON，文件名使用 `Skill.ID`，而不是 Skill 名称。

### 示例

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

### 键说明

| 键 | 类型 | 作用 |
|----|------|------|
| `ID` | string | 已安装 Skill 的稳定实例 UUID，同时也是元数据文件名。 |
| `Name` | string | Skill 目录名。 |
| `Path` | string | Skill 的本地目录路径。如果位于同步根目录内，会以正斜杠相对路径保存，例如 `skills/Engineering/code-review`。 |
| `Category` | string | 当前 Skill 所在的 SkillFlow 分类目录。 |
| `Source` | string | 安装来源类型。当前取值有 `github` 和 `manual`。 |
| `SourceURL` | string | Git 安装来源的原始仓库或来源地址；手动导入通常为空。 |
| `SourceSubPath` | string | 如果 Skill 来自仓库子目录，这里保存其仓库内相对路径。 |
| `SourceSHA` | string | 该 Skill 最近一次导入或更新时记录的提交 SHA。 |
| `LatestSHA` | string | 更新检测最近一次发现的远端最新 SHA。 |
| `InstalledAt` | string | Skill 首次导入到 SkillFlow 的时间。 |
| `UpdatedAt` | string | 最近一次修改元数据的时间，例如移动分类或完成更新。 |

### 重要说明

`meta/<skill-id>.json` 保存的是安装状态，不是 `SKILL.md` 里的 YAML frontmatter。像 `name`、`description`、`allowed-agents` 这类 frontmatter 字段仍然保存在 Skill 内容本身。

## `meta_local/<skill-id>.local.json`

路径：`<AppDataDir>/meta_local/<skill-id>.local.json`

这个文件用于保存当前设备本地、变化频繁且不应跨设备同步的 Skill 字段。

### 示例

```json
{
  "lastCheckedAt": "2026-03-11T08:00:00Z"
}
```

### 字段说明

| 键 | 类型 | 说明 |
|----|------|------|
| `lastCheckedAt` | string | 当前设备最近一次执行更新检查的时间。 |

## `memory/memory_local.json`

本地专用记忆配置，不参与云备份和 git 同步。

**路径：** `<appDataDir>/memory/memory_local.json`

**Schema：**

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

**字段说明：**

| 分区 | 键 | 类型 | 说明 |
|------|----|------|------|
| `pushConfigs` | `<agentType>` | object | 各智能体推送配置 |
| `pushConfigs.<agent>.mode` | — | string | `"merge"`（合并）或 `"takeover"`（覆盖） |
| `pushConfigs.<agent>.autoPush` | — | bool | 该智能体在本地编辑后是否自动同步全部记忆 |
| `moduleStates` | `<moduleName>` | object | 单个模块记忆的本地全局状态 |
| `moduleStates.<module>.enabled` | — | bool | 该模块是否参与自动同步与默认的整智能体推送 |
| `pushState` | `<agentType>` | object | 各智能体最近推送记录 |
| `pushState.<agent>.lastPushedAt` | — | RFC3339 字符串 | 最近一次成功推送的时间戳 |
| `pushState.<agent>.lastPushedHash` | — | string | 最近一次实际推送到该智能体内容的 SHA-256 哈希 |

**说明：**

- 已不再持久化“按模块配置推送目标”的列表；手动批量推送里的选择仅存在于当前页面状态。
- `moduleStates` 属于本地 UI / 分发状态；如果某个模块还没有写入记录，SkillFlow 会为了兼容旧数据将其视为默认启用。
- 停用模块会把它排除在自动同步和默认整智能体推送之外，但手动批量推送仍可显式勾选它。
- 如果一次批量推送只推送了部分模块，`lastPushedHash` 会记录这次实际推送的快照，因此当本地记忆库比该快照包含更多模块时，该智能体仍会显示 `pendingPush`。

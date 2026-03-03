# SkillFlow 产品需求与架构设计文档

**日期**：2026-03-04
**版本**：v0.1
**技术栈**：Go + Wails + React + TypeScript

---

## 一、产品概述

SkillFlow 是一个面向**个人开发者**的桌面客户端工具，用于统一管理在多个 LLM 工具（Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 等）中使用的 SKILLS，支持本地分类管理、跨工具同步、GitHub 安装、更新检测以及云端备份。

---

## 二、Skill 识别规范

- 一个 skill 对应一个目录
- 目录根部必须存在 `SKILLS.md` 文件
- 当前版本仅检测文件存在即视为合法 skill，代码需预留扩展点支持未来更复杂的校验逻辑

---

## 三、技术架构

### 3.1 技术选型

| 层次 | 技术 |
|------|------|
| 后端业务逻辑 | Go（最新稳定版） |
| 桌面 UI 框架 | Wails v2 |
| 前端 | React + TypeScript |
| 前端状态管理 | Zustand / Jotai |
| 系统密钥存储 | zalando/go-keyring |

### 3.2 分层架构

```
┌─────────────────────────────────────────┐
│           Frontend (React + TS)          │
│  Dashboard / Settings / Sync / Backup   │
└──────────────┬──────────────────────────┘
               │ Wails bindings + EventsEmit
┌──────────────▼──────────────────────────┐
│            App Layer (Wails)             │
│  暴露 Go 方法给前端 / 监听 channel 转发事件  │
└──────────────┬──────────────────────────┘
               │ 调用
┌──────────────▼──────────────────────────┐
│             Core Library                 │
│  skill/ sync/ backup/ install/ notify/  │
└──────────────┬──────────────────────────┘
               │ 实现接口
┌──────────────▼──────────────────────────┐
│           Provider Implementations       │
│  adapters/ providers/ installers/       │
└─────────────────────────────────────────┘
```

### 3.3 项目目录结构

```
SkillFlow/
├── core/                    # 核心业务库（无 UI 依赖）
│   ├── skill/               # skill 模型、本地存储、分类管理
│   ├── sync/                # 同步接口 + 各工具适配器
│   ├── backup/              # 云备份接口 + 各云厂商实现
│   ├── install/             # 安装接口 + GitHub/本地实现
│   ├── update/              # 更新检测（仅 GitHub 来源）
│   ├── config/              # 配置读写（工具路径、云凭证等）
│   ├── registry/            # 扩展点注册表
│   └── notify/              # channel-based 通知中心
├── app/
│   └── wails/               # Wails App，桥接 core 与前端
│       ├── main.go
│       ├── app.go           # 暴露给前端的所有方法
│       └── events.go        # 监听 core notify channel，转发到前端
├── frontend/                # React + TypeScript
│   ├── src/
│   │   ├── pages/           # Dashboard, SyncPush, SyncPull, Backup, Settings
│   │   ├── components/      # SkillCard, CategoryPanel, ConflictDialog 等
│   │   └── store/           # 全局状态
│   └── wailsjs/             # 自动生成的 Wails JS bindings
├── docs/
│   └── plans/
└── build/
```

---

## 四、核心数据模型

### 4.1 Skill 模型

```go
type SourceType string

const (
    SourceGitHub SourceType = "github"
    SourceManual SourceType = "manual"
)

type Skill struct {
    ID          string
    Name        string     // 目录名即 skill 名
    Path        string     // 在 SkillFlow 本地存储中的绝对路径
    Category    string     // 分类名，空字符串表示未分类
    Source      SourceType
    SourceURL   string     // GitHub 来源时记录仓库 URL
    SourceSubPath string   // GitHub 仓库内子路径，如 skills/skill-a
    SourceSHA   string     // 安装时记录的 commit SHA（用于更新检测）
    InstalledAt time.Time
    UpdatedAt   time.Time
    LastCheckedAt time.Time
}
```

### 4.2 配置模型

```go
type ToolConfig struct {
    Name      string // "claude-code" | "opencode" | "codex" | "gemini-cli" | "openclaw" | 自定义名称
    SkillsDir string
    Enabled   bool
    Custom    bool   // 是否为用户自定义工具
}

type CloudConfig struct {
    Provider    string
    Enabled     bool
    BucketName  string
    RemotePath  string
    Credentials map[string]string // 明文字段存非敏感信息，敏感信息走系统密钥链
}

type AppConfig struct {
    SkillsStorageDir string
    DefaultCategory  string
    Tools            []ToolConfig
    Cloud            CloudConfig
}
```

### 4.3 通知事件模型

```go
type EventType string

const (
    EventBackupStarted   EventType = "backup.started"
    EventBackupProgress  EventType = "backup.progress"
    EventBackupCompleted EventType = "backup.completed"
    EventBackupFailed    EventType = "backup.failed"
    EventSyncCompleted   EventType = "sync.completed"
    EventUpdateAvailable EventType = "update.available"
    EventSkillConflict   EventType = "skill.conflict"
)

type Event struct {
    Type    EventType
    Payload any
}
```

---

## 五、扩展接口设计

### 5.1 安装接口

```go
type InstallSource struct {
    Type string // "github" | "local"
    URI  string
}

type SkillCandidate struct {
    Name      string
    Path      string
    Installed bool
}

type Installer interface {
    Scan(ctx context.Context, source InstallSource) ([]SkillCandidate, error)
    Install(ctx context.Context, source InstallSource, selected []SkillCandidate, category string) error
    Type() string
}
```

### 5.2 同步适配器接口

```go
type ToolAdapter interface {
    Name() string
    DefaultSkillsDir() string
    // 推送到工具目录（拍平，不带分类结构）
    Push(ctx context.Context, skills []Skill, targetDir string) error
    // 从工具目录扫描所有 skills
    Pull(ctx context.Context, sourceDir string) ([]Skill, error)
}
```

### 5.3 云备份接口

```go
type CredentialField struct {
    Key         string
    Label       string
    Placeholder string
    Secret      bool
}

type RemoteFile struct {
    Path         string
    Size         int64
    LastModified time.Time
    IsDir        bool
}

type CloudProvider interface {
    Name() string
    Init(credentials map[string]string) error
    // 增量同步，保持云端与本地目录结构完全一致
    Sync(ctx context.Context, localDir string, bucket string, remotePath string) error
    Restore(ctx context.Context, bucket string, remotePath string, localDir string) error
    List(ctx context.Context, bucket string, remotePath string) ([]RemoteFile, error)
    // 动态返回该 provider 所需的凭证字段，用于 UI 动态渲染配置表单
    RequiredCredentials() []CredentialField
}
```

### 5.4 注册表

```go
// core/registry/registry.go
// 编译时注入所有实现，无动态加载

func RegisterInstaller(i install.Installer)
func RegisterAdapter(a sync.ToolAdapter)
func RegisterCloudProvider(p backup.CloudProvider)
```

---

## 六、功能模块详述

### 6.1 本地 Skills 存储

SkillFlow 维护自己的中央 skills 存储目录作为 source of truth：

```
~/Library/Application Support/SkillFlow/   (macOS)
%APPDATA%\SkillFlow\                       (Windows)
├── skills/
│   ├── 未分类/
│   ├── 写作/
│   └── 编程/
├── config.json
└── meta/              # skill 元数据，与 skill 目录分离
    ├── {skill-id}.json
    └── ...
```

元数据与 skill 内容分离，确保推送到其他工具时 skill 目录保持干净。

### 6.2 Skills 分类管理

- 支持创建、重命名、删除分类（对应本地子目录）
- Dashboard 左侧分类面板支持拖拽排序
- Skills 卡片支持跨分类拖拽（拖到左侧分类上）
- 安装/导入时可选择目标分类

### 6.3 GitHub 安装流程

1. 用户输入公开 GitHub 仓库 URL
2. 调用 GitHub API 扫描 `skills/` 目录，过滤含 `SKILLS.md` 的子目录
3. 前端展示复选框列表（已安装的标注蓝色标签）
4. 用户选择 skill + 目标分类 → 安装
5. 遇到同名冲突 → 弹出对话框（覆盖/跳过）
6. 安装完成后记录元数据（来源 URL、子路径、commit SHA）
7. 触发自动云备份

### 6.4 手动导入

- 方式一：点击导入按钮，弹出文件夹选择器
- 方式二：拖拽文件夹到 SkillFlow 窗口任意位置
- 导入时选择目标分类
- 元数据来源标记为 `manual`，不支持自动更新检测

### 6.5 更新检测（仅 GitHub 来源）

- 记录安装时 skill 目录对应的最新 commit SHA
- 检测时调用 `GET /repos/{owner}/{repo}/commits?path={subPath}&per_page=1` 对比 SHA
- 触发时机：应用启动时后台静默检测 + 用户手动触发
- 有更新时 SkillCard 显示红色角标，用户确认后重新下载覆盖

### 6.6 跨工具同步

**推送到工具：**
- 支持多选目标工具（已启用的工具）
- 同步范围：全部 / 按分类 / 手动选择
- 推送时拍平目录结构（不带分类子目录）
- 遇到冲突弹出对话框逐个处理（覆盖/跳过）

**从工具拉取：**
- 选择来源工具 → 扫描其 skills 目录列出所有 skill
- 用户勾选需要导入的 skill
- 选择目标分类（默认使用配置中的 DefaultCategory）
- 遇到冲突弹出对话框

**工具管理：**
- 内置工具：Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw
- 每个工具有默认 skills 目录路径，用户可覆盖
- 每个工具有启用/禁用开关，禁用后在同步页面隐藏
- 支持添加多个自定义工具（名称 + skills 目录路径）

### 6.7 云端备份

- 支持云厂商：阿里云 OSS、腾讯云 COS、华为云 OBS（架构支持扩展）
- 触发方式：skills 变动时自动触发 + 用户手动触发
- 同步策略：增量同步，云端与本地目录结构完全镜像（不压缩）
- 凭证安全：敏感凭证存入系统密钥链（macOS Keychain / Windows Credential Manager）
- 配置 UI：根据 `RequiredCredentials()` 动态渲染各云厂商的配置表单

---

## 七、前端页面结构

```
侧边栏
├── 📦 我的 Skills（Dashboard）
├── 🔄 同步管理
│   ├── 推送到工具
│   └── 从工具拉取
├── ☁️ 云备份
└── ⚙️ 设置
    ├── Tab: 工具路径
    ├── Tab: 云存储
    └── Tab: 通用
```

### Dashboard

- 左侧：分类面板（可拖拽排序、右键重命名/删除）
- 右侧：Skills 看板（卡片展示，显示名称、来源标签、更新角标）
- 顶部：搜索栏、GitHub 安装按钮、手动导入按钮
- 支持拖拽文件夹到窗口任意位置触发导入
- Skills 卡片可拖拽到左侧分类面板更改分类
- 右键菜单：更新 / 删除 / 移动到分类

---

## 八、构建与打包

| 平台 | 架构 | 输出产物 |
|------|------|----------|
| macOS | Apple Silicon (arm64) | SkillFlow.app + .dmg |
| macOS | Intel (amd64) | SkillFlow.app + .dmg |
| Windows | amd64 | SkillFlow.exe + NSIS installer |

- 使用 GitHub Actions 矩阵构建，三个平台并行
- macOS 两个架构独立构建，不合并为 Universal Binary

---

## 九、内置工具默认 Skills 目录

| 工具 | macOS 默认路径 | Windows 默认路径 |
|------|--------------|----------------|
| Claude Code | `~/.claude/skills` | `%APPDATA%\claude\skills` |
| OpenCode | `~/.opencode/skills` | `%APPDATA%\opencode\skills` |
| Codex | `~/.codex/skills` | `%APPDATA%\codex\skills` |
| Gemini CLI | `~/.gemini/skills` | `%APPDATA%\gemini\skills` |
| OpenClaw | `~/.openclaw/skills` | `%APPDATA%\openclaw\skills` |

> 以上默认路径需在首次启动时验证是否存在，存在则自动启用对应工具。

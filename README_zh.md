# SkillFlow

> 🌐 **中文** | [English](README.md)

一款跨平台桌面应用，用于统一管理多个 AI 编程工具中的 LLM Skills（提示词库 / 斜杠命令），支持 GitHub 安装、云端备份和跨工具同步。

## 下载安装

从 **[GitHub Releases →](https://github.com/shinerio/SkillFlow/releases/latest)** 下载最新版本

| 平台 | 文件 |
|------|------|
| macOS（Apple Silicon） | `SkillFlow-macos-apple-silicon` |
| macOS（Intel） | `SkillFlow-macos-intel` |
| Windows（x64） | `SkillFlow-windows-amd64.exe` |

## 功能概览

| 功能 | 说明 |
|------|------|
| **Skill 库** | 集中存储所有 Skills，支持分类管理、实时搜索、首字母正序 / 逆序排序、拖拽整理、批量删除，以及仅允许删除空分类的安全删除逻辑 |
| **提示词库** | 将可复用的 prompt 保存为同步的 `prompts/<category>/<name>/system.md` 卡片，支持必填唯一名称、可选描述、分类、导入导出、拖拽移动分类、一键复制，以及 `and` / `or` 关键字搜索 |
| **GitHub 安装** | 克隆任意仓库，递归发现仓库内嵌套的 Skill 候选项，并一键选择安装；候选项状态 badge 会遵循可配置的页面级显示策略，同名候选项仍按规范化仓库来源 + 子路径准确区分 |
| **跨工具同步** | 推送或拉取 Skills 到/从 Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 及自定义工具；推送页默认进入“手动选择”，拉取页默认全部不选并新增“选择未导入”复选框，已推送工具会以紧凑图标列表显示并支持悬浮查看完整列表 |
| **仓库收藏** | 关注 Git 仓库，无需导入即可递归浏览和使用仓库内嵌套的 Skills；支持 GitHub Star 自动捕获（PAT）、全仓深度 `skill.md`/`SKILL.md` 识别，并在仓库卡片展示 `manual`/`star` 来源标记 |
| **云端备份** | 将 Skill 库镜像至阿里云 OSS、AWS S3、Azure Blob Storage、Google Cloud Storage、腾讯云 COS、华为云 OBS 或任意 Git 仓库，支持自定义对象存储远程路径预览、按服务商独立保存配置、敏感云凭据仅保存在本地、Git 冲突时的手动处理入口，以及只显示每次备份/恢复实际改动文件的结果页 |
| **更新检测** | 按规范化仓库来源 + 子路径检测已安装 GitHub Skill 的新提交；实例已是最新时会自动清理过期更新标记，并支持一键更新 |
| **应用自动更新** | 弹出模态对话框提示新版本；Windows 支持一键下载并重启；macOS 链接至 GitHub Releases 页面；用户可跳过当前版本以抑制后续启动弹窗 |
| **托盘驻留** | 点击窗口关闭按钮后应用继续在后台运行；macOS 会隐藏 Dock 图标并仅保留顶部状态栏黑白图标入口，Windows 驻留系统托管区 |
| **桌面框架** | 固定侧边栏提供品牌化 SkillFlow 标题、应用图标、语言 / 主题快捷切换，以及反馈入口 |
| **启动窗口** | 每台设备都会把最近一次手动调整后的窗口尺寸保存在本地 `config_local.json`，下次启动时优先恢复；首次启动仍会按当前显示器自适应计算 |
| **设置** | 设置页会随窗口宽度自适应展开，并支持按工具独立配置启用状态、推送/扫描路径、自定义工具、代理设置、可调的本地/远程扫描深度、按页面控制的卡片状态显示、不会参与同步的本地路径/代理配置，以及设置页里的 `Ctrl+S` / `Cmd+S` 保存快捷键；目录选择器会优先回到当前路径 |
| **中英文界面切换** | 可在侧边栏或设置页中立即切换中文 / English，语言偏好仅保存在本地 |
| **Dark / Young / Light 主题** | 可在重做后的石墨灰 Dark、由旧浅色主题演化而来的纸感浅蓝 Young，以及参考 Messor 配色的新 Light 主题之间切换；跨会话持久化 |

每个按钮、对话框和交互的完整说明，请查阅 **[docs/features_zh.md](docs/features_zh.md)**。


## SkillFlow vs cc-switch

### 1. 核心定位

| | SkillFlow | cc-switch |
|---|---|---|
| **目标** | Skill（提示词库）的专项管理工具 | AI CLI 工具的全能配置助手 |
| **核心价值** | Skill的发现、安装、同步、云备份。**理念**：轻工具，注重资产（skil）的积累和管理 | Provider API切换 + MCP + Skills + Prompts 一站式管理 |


### 2. 支持工具

| | SkillFlow | cc-switch |
|---|---|---|
| **Claude Code** | ✅ | ✅ |
| **OpenCode** | ✅ | ✅ |
| **Codex** | ✅ | ✅ |
| **Gemini CLI** | ✅ | ✅ |
| **OpenClaw** | ✅ | ❌ |
| **自定义工具** | ✅ | ❌ |

### 3. 功能横向比较

| | SkillFlow | cc-switch |
|---|---|---|
| **本地 Skill 库管理** | ✅ 本地Central Library（支持分类、搜索、拖拽排序、批量删除） | ❌ 无本地库概念，接安装/卸载到 `~/.claude/skills/`，无独立本地库概念 |
| **安装来源** | GitHub 仓库克隆，支持 **Starred Repos** 浏览</br> 手动导入 </br> 扫描工具内置skill | GitHub 仓库扫描（3 个预配置仓库 + 自定义） |
| **GitHub 安装** | ✅ 深度递归扫描，冲突处理 | ✅ 基础（预配置仓库 + 自定义） |
| **跨工具同步** | ✅ 精细控制（per-skill 冲突处理） | ✅ 基础（一键安装到 ~/.claude/skills/） |
| **更新检测** | ✅ 逐 Skill 检测新 commit | ❌ |
| **云备份** | ✅ 多对象存储/Git | ❌（靠外部云盘目录同步） |
| **收藏仓库浏览** | ✅ | ❌ |

## 支持的工具

内置适配器：**Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

可在设置中添加自定义工具，指定任意本地目录路径即可。

## Skill 格式

有效的 Skill 目录须在根目录下包含 `skill.md` 文件，满足此要求的目录均可通过本地导入或 GitHub 安装。

```
my-skill/
  skill.md     ← 必须存在
  ...其他文件
```

## 云端备份配置

在**设置 → 云存储**中配置：

- 可同步的设置和元数据保存在应用数据目录中，并使用相对路径保证跨平台恢复正确。
- 仅本机使用的文件系统路径与代理设置（如 `SkillsStorageDir`、工具扫描/推送目录、手动代理地址）保存在 `config_local.json` 中，不参与备份/同步。
- 敏感云凭据（如 Access Key ID、Secret Key、访问令牌）只保存在 `config_local.json` 中按服务商分组的本地配置里；可同步的 `config.json` 仅保留存储桶、Endpoint、仓库地址、分支等非敏感云配置。
- 可复用提示词会与 Skills 一起保存在 `prompts/<category>/<name>/system.md` 下，因此 Git 备份和对象存储都会同步同一份提示词库。
- 对象存储服务商支持自定义父级 `remotePath`；最终备份前缀始终会渲染并保存为 `<存储桶>/<remotePath>/skillflow/`（若父级路径为空，则为 `<存储桶>/skillflow/`）。
- 每个云服务商都会保留各自独立的存储桶 / 路径 / 凭据配置，因此在设置中切换服务商时不会覆盖其他服务商的值。
- Git 同步发生冲突时，可以选择以本地为准、以远端为准，或直接打开备份文件夹手动处理。
- 备份页只显示当前应用会话中最近一次备份或恢复实际涉及的改动文件，不再展示远端全量文件列表。
- 阿里云 OSS、腾讯云 COS、华为云 OBS 使用相同的存储桶 + Endpoint 配置模型。AWS S3 使用存储桶 + Region。Azure Blob Storage 使用容器名称（填写在存储桶字段）+ Account Name + Account Key，并支持可选的 Service URL。Google Cloud Storage 使用存储桶 + Service Account JSON 或本地密钥文件路径。对于腾讯云 COS，存储桶始终来自单独的存储桶输入框，而 Endpoint 字段既可填写纯 Endpoint host，也可填写完整桶域名/URL，并会按用户输入形式保留。
- 应用数据目录：
  - macOS：`~/Library/Application Support/SkillFlow/`
  - Windows：`%USERPROFILE%\.skillflow\`

各云服务商所需配置字段：

| 云服务商 | 必填字段 |
|----------|---------|
| 阿里云 OSS | Access Key ID（仅本地）、Access Key Secret（仅本地）、Endpoint（同步） |
| AWS S3 | Access Key ID（仅本地）、Secret Access Key（仅本地）、Region（同步） |
| Azure Blob Storage | 容器名称（存储桶字段，同步）、Account Name（同步）、Account Key（仅本地）、Service URL（同步，可选） |
| Google Cloud Storage | Service Account JSON 或本地密钥文件路径（仅本地） |
| 腾讯云 COS | SecretId（仅本地）、SecretKey（仅本地）、Endpoint（同步） |
| 华为云 OBS | Access Key ID（仅本地）、Secret Access Key（仅本地）、Endpoint（同步） |
| Git 仓库 | 仓库地址（同步）、分支（同步）、用户名（同步）、访问令牌（仅本地） |

## 参与贡献 & 自行构建

### 环境要求

- macOS 11+ 或 Windows 10+
- Go 1.23+
- Node.js 18+
- Wails v2 CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 构建步骤

```bash
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend   # 安装前端依赖
make dev                # 热重载开发模式
make test               # 运行 Go 测试
make build              # 构建生产版本 → build/bin/
```

常用 `make` 目标：

| 目标 | 说明 |
|------|------|
| `make dev` | 热重载开发模式（Go + 前端） |
| `make build` | 构建生产版本二进制 |
| `make test` | 运行所有 Go 测试 |
| `make tidy` | 同步 Go 模块依赖 |
| `make generate` | App 方法变更后重新生成 TypeScript 绑定 |
| `make clean` | 删除构建产物 |

内部架构详情请查阅 **[docs/architecture_zh.md](docs/architecture_zh.md)**。

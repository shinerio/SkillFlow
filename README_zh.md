# SkillFlow

> 🌐 **中文** | [English](README.md)

一款跨平台桌面应用，用于统一管理多个 AI 编程工具中的 LLM SKILLS（提示词库 / 斜杠命令），支持 GitHub 安装、云端备份和跨工具同步。

## 功能概览

| 功能 | 说明 |
|------|------|
| **Skill 库** | 集中存储所有 Skills，支持分类管理、实时搜索、拖拽整理和批量删除 |
| **GitHub 安装** | 克隆任意仓库，浏览 Skill 候选项，一键选择安装；后续扫描自动拉取更新 |
| **跨工具同步** | 推送（全部 / 按分类 / 手动选择）或从 Claude Code、OpenCode、Codex、Gemini CLI、OpenClaw 及自定义工具拉取；逐条冲突处理 |
| **仓库收藏** | 关注 Git 仓库，无需导入即可浏览和使用其中的 Skills；支持文件夹或平铺视图；可直接批量推送到工具 |
| **云端备份** | 将 Skill 库镜像至阿里云 OSS、腾讯云 COS、华为云 OBS 或任意 Git 仓库；变更后自动备份；支持定时同步间隔；冲突时弹框选择以本地或远端为准 |
| **更新检测** | 自动检测 GitHub 来源 Skills 的新提交；一键更新 |
| **设置** | 按工具独立配置启用状态、推送路径和扫描路径（文字输入 + 文件夹选择器）、自定义工具、云服务凭据、代理设置 |

每个按钮、对话框和交互的完整说明，请查阅 **[feature_zh.md](feature_zh.md)**。

## 支持的工具

内置适配器：**Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

可在设置中添加自定义工具，指定任意本地目录路径即可。

## 环境要求

- macOS 11+ 或 Windows 10+
- Go 1.23+
- Node.js 18+（用于前端构建）
- [Wails v2](https://wails.io) CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## 开发指南

```bash
# 克隆仓库并安装前端依赖
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend

# 开发模式运行（热重载）
make dev

# 运行 Go 测试
make test

# 构建生产版本
make build
```

常用 `make` 目标：

| 目标 | 说明 |
|------|------|
| `make dev` | 热重载开发模式（Go + 前端） |
| `make build` | 构建生产版本二进制 |
| `make test` | 运行所有 Go 测试 |
| `make tidy` | 同步 Go 模块依赖 |
| `make generate` | 重新生成 TypeScript 绑定 |
| `make clean` | 删除构建产物 |

构建产物：`build/bin/SkillFlow`（macOS）/ `build/bin/SkillFlow.exe`（Windows）

## Skill 格式

有效的 Skill 目录须在根目录下包含 `skill.md` 文件，满足此要求的目录均可通过本地导入或 GitHub 安装。

```
my-skill/
  skill.md     ← 必须存在
  ...            ← 其他文件
```

## 云端备份配置

在**设置 → 云存储**中配置，凭据保存在本地配置文件中：

- macOS：`~/Library/Application Support/SkillFlow/config.json`
- Windows：`%APPDATA%\SkillFlow\config.json`

各云服务商所需配置字段：

| 云服务商 | 必填字段 |
|----------|---------|
| 阿里云 OSS | Access Key ID、Access Key Secret、Endpoint |
| 腾讯云 COS | SecretId、SecretKey、Region |
| 华为云 OBS | Access Key、Secret Key、Endpoint |

## CI / 发布

通过 GitHub Actions 在 `v*` tag 上自动构建，生成以下平台的二进制文件：
- macOS（Intel x86_64）
- macOS（Apple Silicon arm64）
- Windows（x86_64）

## 技术架构

Go 核心库（`core/`）采用基于接口的插件架构。Wails v2 通过直接方法绑定（无 HTTP API）将 Go 后端与 React 18 + TypeScript + Tailwind CSS 前端相连。详见 [CLAUDE.md](CLAUDE.md)。


## 支持的工具

内置适配器：**Claude Code** · **OpenCode** · **Codex** · **Gemini CLI** · **OpenClaw**

可在设置中添加自定义工具，指定任意本地目录路径即可。

## 环境要求

- macOS 11+ 或 Windows 10+
- Go 1.23+
- Node.js 18+（用于前端构建）
- [Wails v2](https://wails.io) CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## 开发指南

```bash
# 克隆仓库并安装前端依赖
git clone https://github.com/shinerio/SkillFlow
cd SkillFlow
make install-frontend

# 开发模式运行（热重载）
make dev

# 运行 Go 测试
make test

# 构建生产版本
make build
```

常用 `make` 目标：

| 目标 | 说明 |
|------|------|
| `make dev` | 热重载开发模式（Go + 前端） |
| `make build` | 构建生产版本二进制 |
| `make test` | 运行所有 Go 测试 |
| `make tidy` | 同步 Go 模块依赖 |
| `make generate` | 重新生成 TypeScript 绑定 |
| `make clean` | 删除构建产物 |

构建产物：`build/bin/SkillFlow`（macOS）/ `build/bin/SkillFlow.exe`（Windows）

## Skill 格式

有效的 Skill 目录须在根目录下包含 `skill.md` 文件，满足此要求的目录均可通过本地导入或 GitHub 安装。

```
my-skill/
  skill.md     ← 必须存在
  ...            ← 其他文件
```

## 云端备份配置

在**设置 → 云存储**中配置，凭据保存在本地配置文件中：

- macOS：`~/Library/Application Support/SkillFlow/config.json`
- Windows：`%APPDATA%\SkillFlow\config.json`

各云服务商所需配置字段：

| 云服务商 | 必填字段 |
|----------|---------|
| 阿里云 OSS | Access Key ID、Access Key Secret、Endpoint |
| 腾讯云 COS | SecretId、SecretKey、Region |
| 华为云 OBS | Access Key、Secret Key、Endpoint |

## CI / 发布

通过 GitHub Actions 在 `v*` tag 上自动构建，生成以下平台的二进制文件：
- macOS（Intel x86_64）
- macOS（Apple Silicon arm64）
- Windows（x86_64）

## 技术架构

Go 核心库（`core/`）采用基于接口的插件架构。Wails v2 通过直接方法绑定（无 HTTP API）将 Go 后端与 React 18 + TypeScript + Tailwind CSS 前端相连。详见 [CLAUDE.md](CLAUDE.md)。

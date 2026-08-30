[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shinerio/SkillFlow)

# SkillFlow

> 🌐 **中文** | [English](README.md)

理念：学于一瞬，用于万境。

SkillFlow 是一款跨平台桌面应用，用来在不同智能体环境之间统一管理可复用的 Skill、Prompt 和 Memory。它提供本地资产库、跨智能体同步、仓库来源跟踪、更新检测，以及面向对象存储或 Git 的同步安全备份。

![SkillFlow](docs/skillflow.gif)

## SkillFlow 能做什么

- **📦 资产内化**：解决“收藏即遗忘”的困境。将散落在博客、视频和日常实践里的 Skill、Prompt、Memory 沉淀为可管理、可迭代、可淘汰、可更新的个人资产。
- **🎛️ 透明掌控**：拒绝技能管理的“黑盒”状态。对海量技能按场景分类归档，确保开发者对智能体能力边界拥有完全知情权与控制权。
- **🔄 多端复用**：打破模型孤岛。让同一套 Skill 和 Memory 可以在 Claude Code、Codex、Gemini 等不同智能体工具之间复用。
- **🧠 模块化记忆管理**：同时维护一份主记忆和多份模块记忆，并按需要自动同步到多个 AI 编程助手工具，支持合并和接管两种策略。
- **🖥️ 环境一致性**：构建跨设备云同步体系。屏蔽 macOS / Windows 多设备异构环境差异，确保在任何办公场景和设备上都能获得高度一致的开发环境与可复用资产。
- **⚡ 自动迭代**：减少重复维护成本，通过自动更新 Skill 和自动同步 Memory，避免来源分散后内容逐渐陈旧。
- **📝 提示词工程**：实现 Prompt 的版本化管理。杜绝重复编写造成的实际效果参差不齐，推动提示词的持续优化与标准化迭代。

## 下载安装

从 **[GitHub Releases](https://github.com/shinerio/SkillFlow/releases/latest)** 获取最新版本。

| 平台 | 文件 |
|------|------|
| macOS（Apple Silicon） | `SkillFlow-macos-apple-silicon.dmg` |
| macOS（Intel） | `SkillFlow-macos-intel.dmg` |
| Windows（x64） | `SkillFlow-windows.exe` |

## 功能概览

| 功能 | 说明 |
|------|------|
| **Skill 库** | 本地 Skill 库，支持分类、搜索、排序、拖拽整理、批量删除、更新检测和自动推送目标。 |
| **我的提示词** | 同步提示词卡片，支持分类管理、范围导出、冲突可选导入、一键复制、图片预览和网页链接。 |
| **我的记忆** | 统一的记忆工作台，支持主记忆、模块记忆、编辑预览、批量推送、删除，以及按智能体配置自动同步模式。 |
| **跨智能体同步** | 支持把 Skill 推送到多个智能体，并将 Memory 自动同步到内置或自定义智能体工具。 |
| **仓库收藏** | 跟踪 Git 仓库、浏览发现的 Skill，并在不整仓安装的前提下按需导入或推送。 |
| **云端备份** | 将可同步的 Skill、Prompt、Memory 和元数据备份到对象存储服务商或 Git。 |
| **桌面体验** | macOS 与 Windows 原生客户端内置 daemon 运行时，支持中英文界面、多主题、开机自启与代理测试；legacy Wails 构建保留用于回滚。 |

详细参考：

- UI / UX 行为： [docs/features_zh.md](docs/features_zh.md)
- 落盘文件与配置结构： [docs/config_zh.md](docs/config_zh.md)

## 支持的智能体

内置适配器：

- **Claude Code**
- **OpenCode**
- **Codex**
- **Gemini CLI**
- **OpenClaw**

你也可以在设置页通过配置本地扫描目录和推送目录来添加 **自定义智能体**。

## Skill 格式

有效的 Skill 目录必须在根目录包含 `skill.md` 文件；文件名匹配大小写不敏感。

```text
my-skill/
  skill.md
  ...其他文件
```

参与贡献和本地构建说明请查阅 **[contributing_zh.md](contributing_zh.md)**。

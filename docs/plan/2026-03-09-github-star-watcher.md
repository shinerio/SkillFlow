# GitHub Star Watcher Proposal (User-Level)

## 1. 产品定位与核心价值

- **定位：** 运行在后台的轻量级、自动化知识沉淀能力。
- **价值：** 自动监听用户在 GitHub 上的 Star 动作，全自动检索仓库内是否包含 `skill.md` 或 `SKILL.md` 文件。一旦命中，立即收录到 Starred Repos（与手动 Add Repository 一致），并触发 clone 到本地 cache 目录，便于后续搜索与安装到 My Skills。
- 对不包含技能文件的仓库直接忽略。
- Star 来源与手动来源可区分：
  - 手动添加显示 `manual` 标记
  - GitHub 自动捕获显示 `⭐` 标记
  - 同一仓库若同时满足两种来源，则两个标记同时展示。

## 2. 用户使用路径 (User Journey)

1. **极简配置：** 若用户未配置 PAT，则在 Starred Repos 页面提示前往设置页填写 GitHub Personal Access Token (PAT) 并保存；未配置时不启用自动监听。
2. **无感守护：** 客户端托盘常驻，后台静默轮询，保持低 CPU / 网络开销。
3. **即时反馈：** 当用户 Star 到包含 skill 文件的仓库，应用在捕获后给出通知提示（托盘/桌面通知能力可用时触发），并在 Starred Repos 中可见。

## 3. 核心业务规则 (Business Rules)

- **全局深度搜索：** 只要仓库任意层级存在 `skill.md` / `SKILL.md` 即判定命中。
- **只增不减 (Append-only)：** GitHub 取消 Star 不会自动删除本地已沉淀收藏。
- **零打扰防重试：** 每个新 Star 仓库最多 24 小时扫描一次；无论命中与否，冷静期内不重复深扫。

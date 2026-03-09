# GitHub Star Watcher Technical Design

## 1. 本地存储方案

使用独立 JSON 文件：`github_star_repo.json`（可被 Git/云同步）。

```json
{
  "metadata": {
    "last_etag": "W/\"etag-value\""
  },
  "repos": {
    "12345678": {
      "full_name": "owner/repo",
      "has_skill": true,
      "last_scanned_at": 1715050000
    }
  }
}
```

设计原则：
- 使用 `repos[repo_id]` 字典以实现 O(1) 去重与冷静期判断。
- `metadata.last_etag` 用于增量轮询。
- PAT 仅存 `config_local.json`，不进入同步文件。

## 2. 核心运转引擎

### 模块 A：增量探雷器 (ETag Polling Engine)

- 触发频率：每 5 分钟。
- 请求：`GET /user/starred?per_page=30` + `If-None-Match`。
- 分支：
  - `304 Not Modified`：结束本轮。
  - `200 OK`：表示 Star 集合有变化，解析第一页并筛选尚未记录的 `repo_id`。
- 任务项信息：`repo_id`, `owner`, `repo`, `default_branch`, `full_name`。
- `last_etag` 仅在本轮扫描任务处理并落盘后更新，避免中途退出造成漏扫。

### 模块 B：靶向扫描器 (Git Trees Scanning Engine)

- 输入：模块 A 产生的待扫描仓库。
- 冷静期拦截：若 `now - last_scanned_at < 24h` 则跳过。
- API：`GET /repos/{owner}/{repo}/git/trees/{default_branch}?recursive=1`
- 匹配规则：遍历 `tree`，命中条件为
  - `type == "blob"`
  - 且 `path` 等于 `skill.md` / `SKILL.md`，或以 `/skill.md` / `/SKILL.md` 结尾。
- 落盘：无论命中与否都刷新 `last_scanned_at`；命中则置 `has_skill=true`。
- 命中后动作：加入 Starred Repos（等效手动 Add），并触发 clone/update。

## 3. 流控与异常防御

- 避免 Search API 限额，使用 Git Trees API。
- 每次请求读取 `X-RateLimit-Remaining` / `X-RateLimit-Reset`。
  - 当 remaining ≤ 1，休眠至 reset。
- 捕获 `403/429`：若响应头带 `Retry-After`，按秒休眠后重试。

## 4. 与现有 Starred Repos 的集成

- Starred Repo 记录新增来源标记：
  - `manual`：用户手动添加
  - `starred`：GitHub Star 自动捕获
- UI 同时显示两个来源标记（若都为 true）。
- Append-only：不根据 GitHub unstar 自动删除本地记录。

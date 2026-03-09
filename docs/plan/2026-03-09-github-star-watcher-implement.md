# GitHub Star Watcher Implementation Plan

## 1) Data Model & Persistence

1. Add local-only GitHub config in `core/config`:
   - `AppConfig.GitHub.PAT`
   - persisted only in `config_local.json` (never shared file)
2. Add new storage file in `core/git`:
   - `github_star_repo.json` schema structs
   - atomic load/save helpers
   - methods for cooldown checks, etag update, and repo scan records
3. Extend `core/git.StarredRepo` with source flags:
   - `Manual bool`
   - `Starred bool`
   - backward compatibility migration: old entries default `manual=true`

## 2) GitHub Watcher Engine

1. Add a GitHub watcher service in `cmd/skillflow`:
   - 5-minute ticker
   - skip when PAT missing
2. ETag polling method:
   - call `/user/starred?per_page=30`
   - `If-None-Match` support
   - parse new repos not present in `github_star_repo.json`
3. Tree scanning method:
   - call `/repos/{owner}/{repo}/git/trees/{default_branch}?recursive=1`
   - detect `skill.md` / `SKILL.md`
   - record `last_scanned_at` and `has_skill`
4. Rate-limit handling:
   - inspect remaining/reset headers, sleep when needed
   - handle `403/429 + Retry-After`

## 3) Integration with Existing Flow

1. On skill match:
   - upsert into starred repos list (`starred=true`)
   - preserve existing `manual=true` if present
   - clone/update repository using existing git logic
2. Publish event for capture success:
   - frontend listens and refreshes list
   - trigger UI notification/toast where available

## 4) Frontend Changes

1. Settings page:
   - add GitHub PAT input (masked)
   - save via `SaveConfig`
2. Starred Repos page:
   - show warning when PAT missing
   - show badges/icons for manual + starred source markers
   - listen capture event and show toast
3. i18n:
   - add zh/en translation keys for new labels and hints

## 5) Testing Strategy

1. Unit tests for `core/git` star watcher storage:
   - load/save, cooldown, etag transitions
2. Unit tests for watcher engine HTTP behavior:
   - 304 flow, 200 flow, trees match/no-match, retry-after
3. Regression tests for starred repo storage migration:
   - old data without flags auto-migrates to `manual=true`
4. Run `go test ./core/...` and focused `go test ./cmd/skillflow/...` as available.

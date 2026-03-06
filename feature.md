# SkillFlow — Complete Feature Reference

> 🌐 [中文版](feature_zh.md) | **English**
>
> This document enumerates every feature, button, interaction, and UX detail in SkillFlow.
> **Keep this file in sync whenever features are added, changed, or removed.**

---

## Table of Contents

1. [Navigation & Shell](#1-navigation--shell)
2. [My Skills (Dashboard)](#2-my-skills-dashboard)
3. [Push to Tools](#3-push-to-tools)
4. [Pull from Tools](#4-pull-from-tools)
5. [Starred Repos](#5-starred-repos)
6. [Cloud Backup](#6-cloud-backup)
7. [Settings](#7-settings)
8. [Skill Card](#8-skill-card)
9. [Skill Tooltip](#9-skill-tooltip)
10. [Shared Dialogs](#10-shared-dialogs)
11. [Backend Events](#11-backend-events)
12. [App Update Banner](#12-app-update-banner)

---

## 1. Navigation & Shell

A fixed left sidebar (w-56) provides navigation throughout the app.

| Route | Icon | Label |
|-------|------|-------|
| `/` | Package | My Skills |
| `/sync/push` | ArrowUpFromLine | Push to Tools |
| `/sync/pull` | ArrowDownToLine | Pull from Tools |
| `/starred` | Star | Starred Repos |
| `/backup` | Cloud | Cloud Backup |
| `/settings` | Settings | Settings |

- Active route: highlighted with indigo background.
- Inactive routes: gray text with hover highlight.
- Window close button behavior: clicking the top-left close button hides the main window and keeps the app running in background.
- macOS tray behavior: app remains in the menu-bar status area (`SF` item); use native single-click to open a menu with `Show Window`, `Hide Window`, and `Quit SkillFlow`.
- Windows tray behavior: app remains in the system notification area; click the tray icon to open a menu with `Show SkillFlow` and `Exit`.

---

## 2. My Skills (Dashboard)

Central library for managing your skill collection.

### Toolbar

| Control | Action |
|---------|--------|
| **Search input** | Real-time case-insensitive filter by skill name |
| **Check Updates** (RefreshCw) | Calls backend `CheckUpdates()`; marks updated skills with a red dot |
| **Batch Delete** (CheckSquare) | Toggles multi-select mode |
| **Manual Import** (FolderOpen) | Opens native folder-picker → `ImportLocal(dir)` |
| **Install from Remote** (Github) | Opens the GitHub Install dialog |

### Select Mode (activated by "Batch Delete")

| Control | Action |
|---------|--------|
| **Select All / Deselect All** | Toggles all currently filtered skills |
| **Delete (n)** (Trash2, red) | `DeleteSkills(ids)` — disabled when nothing selected |
| **Cancel** | Exits select mode and clears selections |

### Category Sidebar

- Lists all categories; clicking one filters the skill grid.
- **"All" button** — shows every skill regardless of category.
- **Drag-and-drop target** — dragging a skill card onto a category moves it there.
- **Right-click context menu** on each category:
  - **Rename** — shows inline text input; confirm with Enter, cancel with Escape; calls `RenameCategory()`. (Not available for `Default`.)
  - **Delete** — calls `DeleteCategory()`; skills are moved to the default category. (Not available for `Default`.)
- **New Category** (Plus icon at bottom) — shows inline text input; confirm with Enter or blur, cancel with Escape; calls `CreateCategory()`.

### Skill Grid

- Grid layout: 3 columns, 4 on wide screens.
- **Empty state** — "No Skills found" message with usage hint.
- **Drag-and-drop** — drag a skill card to a category in the sidebar to move it; drag a folder from the OS file manager onto the window to import it directly.
- **Window-level drag overlay** — semi-transparent indigo overlay with "Release to import Skill" message activates when a file is dragged over the window.
- **Hover tooltip** — appears after 300 ms hovering over a card (see [Skill Tooltip](#9-skill-tooltip)).

---

## 3. Push to Tools

Copies skills from your library to external tool directories.

### Tool Selection

- One toggle button per enabled tool (icon + name).
- Multiple tools can be selected simultaneously.

### Sync Scope

Three mutually exclusive modes:

| Mode | Behavior |
|------|----------|
| **All Skills** | Pushes every skill in the library |
| **By Category** | Dropdown appears to pick a category; pushes only those skills |
| **Manual** | Skill grid appears; select individual skills; select-all toggle available |

### Missing Directory Check

Before pushing, the app calls `CheckMissingPushDirs()`. If any target tool directory does not exist yet, a confirmation dialog appears:

- Lists each missing tool name and its full directory path.
- **"Create & Push"** — creates the directory then proceeds.
- **"Cancel"** — aborts without creating anything.

### Conflict Handling

If a skill already exists in the target directory, a conflict dialog appears for each one (see [Conflict Dialog](#101-conflict-dialog)).

### Bottom Bar

- **"Start Push (n)"** button — disabled when no tools selected or skill count is zero; shows "Pushing…" while in progress.
- **"Push complete ✓"** — green success message after all pushes finish.

---

## 4. Pull from Tools

Imports skills from external tool directories into your library.

### Tool Selection

- Same toggle buttons as Push; selecting a different tool resets the scanned list.

### Scan

- **"Scan"** button — calls `ScanToolSkills(toolName)`; recursively searches the tool's configured scan directories for `skill.md` files.
- Shows animated "Scanning…" state while in progress.
- **Error alert** (red) if scan fails; **warning alert** (yellow) if no skills found.

### Skill Grid

- Appears after a successful scan.
- Each card shows whether the skill is already imported (green "Imported" badge).
- Select individual skills or use "Select All / Deselect All".

### Bottom Bar

- **Target Category dropdown** — pick which category to import into (empty selection falls back to fixed category `Default`).
- **"Start Pull (n)"** button — calls `PullFromTool()`.
- **"Pull complete ✓"** — green success message.
- Conflicts handled by the same [Conflict Dialog](#101-conflict-dialog).

---

## 5. Starred Repos

Browse and import skills directly from watched Git repositories without installing them into your library first.

### View Modes

| Mode | Icon | Description |
|------|------|-------------|
| **Folder** | Folder | Grid of repo cards; click a card to drill into its skills |
| **Flat** | LayoutGrid | All skills from all repos shown in a single grid |

### Toolbar (Normal Mode)

| Button | Action |
|--------|--------|
| **Batch Import** (CheckSquare) | Enters select mode |
| **Update All** (RefreshCw) | `UpdateAllStarredRepos()` — clones/pulls all repos in parallel; icon spins while syncing |
| **Add Repo** (Plus, indigo) | Opens "Add Repo" dialog |

### Toolbar (Select Mode)

| Button | Action |
|--------|--------|
| **Select All / Deselect All** | Toggles all visible skills |
| **Push to Tools (n)** | Opens the Push to Tools dialog (see below) |
| **Import to My Skills (n)** | Opens the Import dialog |
| **Cancel** | Exits select mode |

### Repo Card (Folder View)

- Click to open the skill list for that repo.
- **Open in Browser** (ExternalLink icon) — opens repo URL in default browser.
- **Update** (RefreshCw icon) — `UpdateStarredRepo(url)` — pulls latest commits.
- **Delete** (Trash2 icon, red on hover) — removes from starred list.
- Shows last sync time and any sync error below the repo name.

### Repo Detail View (Drill-down)

- Breadcrumb back button (ChevronLeft) to return to the repo grid.
- Skills grid with same select/import behavior as flat view.

### Add Repo Dialog

- URL input (HTTPS or SSH format); Enter key triggers add.
- **"Add"** button — `AddStarredRepo(url)`.
- If the repo requires HTTP authentication, an **HTTP Auth Dialog** appears automatically.
- If SSH auth fails, an **SSH Auth Error Dialog** explains required setup.
- Shows clone-in-progress state ("Cloning…").

### HTTP Auth Dialog

- Username + Password inputs (password is masked); Enter on password field confirms.
- **"Confirm"** — retries with `AddStarredRepoWithCredentials(url, user, pass)`.
- **"Cancel"** — aborts.
- Shows error if credentials are wrong.

### SSH Auth Error Dialog

- Explains SSH key setup checklist:
  - Key generated with `ssh-keygen`
  - Public key added to GitHub / GitLab
  - SSH agent running (`ssh-add`)
  - Suggestion to use HTTPS instead
- **"Close"** button.

### Import Dialog (to My Skills)

- Category selector (dropdown).
- **"Import n"** — `ImportStarSkills(paths, repoURL, category)`.
- **"Cancel"**.

### Push to Tools Dialog

- Description: "Copies skills directly to the tool directory; no need to import first."
- Lists all enabled tools as checkboxes with their push directory paths shown.
- **Empty state** message if no tools are configured.
- **"Push to n tools"** button.
- **"Cancel"**.

### Missing Directory Confirmation

Same behavior as [Push to Tools page](#missing-directory-check): confirms before creating absent push directories.

### Push Conflict Dialog

When skills already exist in the target tool directory:

- Lists all conflicting skill names.
- **"Overwrite All"** (amber) — `PushStarSkillsToToolsForce()`.
- **"Skip Conflicts"** — `PushStarSkillsToTools()` (already resolved; conflicts discarded).

---

## 6. Cloud Backup

Mirror your skill library to cloud storage. Two backend types are supported: **Object Storage** (Aliyun OSS, Tencent COS, Huawei OBS) and **Git Repository**.

### Status

- **Cloud disabled banner** (yellow) — shown when cloud backup is not configured; links to Settings.

### Actions

| Button | Object Storage label | Git label |
|--------|---------------------|-----------|
| **Backup Now** (Upload icon) | 立即备份 | 立即备份 |
| **Restore / Pull** (Download icon) | 从云端恢复 | 拉取远端 |
| **Refresh** (RefreshCw) | Reloads the file list | Same |

- Backup Now and Restore are disabled when cloud is not configured.
- **"Backup complete / Git sync complete"** (green) / **"Backup/sync failed"** (red) status messages.

### Cloud File List

- Object storage: file path (monospace) + size in KB.
- Git: files tracked by `git ls-files`, each showing relative path + size.
- Scrollable, max-height container.
- **Unified backup scope (all providers)** — backup root is the app data root (`skills/`, `meta/`, `config.json`, etc.); `cache/` and `.git/` are excluded.

### Auto-Backup

Triggered automatically after any of these mutations (when cloud is enabled):

- Delete skill / bulk delete
- Manual import
- Install from GitHub
- Pull from tool
- Update skill
- Import from starred repo

Progress events surface in the UI via the Wails event system (`backup.started`, `backup.progress`, `backup.completed`, `backup.failed`).

### Git Sync (Git provider only)

When the **git** provider is selected:

- **Repository bootstrap** — if the Skills directory is not a git repo, SkillFlow auto-initializes it and configures `origin` from the configured repo URL.
- **Remote binding self-heal** — if `origin` is missing or changed, SkillFlow auto-adds/updates it before pull/push.
- **Startup pull** — on every app launch, SkillFlow runs `git pull` on the Git backup root directory to fetch the latest remote changes.
- **Missing branch tolerance** — if the configured remote branch does not exist yet (first-time setup), startup pull is skipped without failing the backup page.
- **Auto-push after mutations** — same post-mutation trigger as object storage; runs `git add -A && git commit && git push`.
- **Periodic auto-sync** — controlled by the "Auto-sync interval" setting (in minutes, 0 = disabled). A background timer fires `autoBackup()` on the configured interval.
- **Manual actions with conflict detection** — both **Backup Now** and **Restore / Pull** detect git conflicts/divergence and emit `git.conflict` when user action is required.
- **Conflict resolution dialog** — if `git pull` or `git push` detects a conflict or diverged history, a modal appears:
  - The dialog includes a conflict file list when available.
  - **"以本地为准"** (Keep Local) — aborts the merge, force-pushes local state to remote. Calls `ResolveGitConflict(true)`.
  - **"以远端为准"** (Use Remote) — aborts the merge, resets local to `origin/<branch>`. Calls `ResolveGitConflict(false)`.
  - Both options reload app state from disk (skills/meta/config) and emit `git.sync.completed` on success.
- **State refresh after pull** — after successful startup pull or manual pull, app state is immediately reloaded from disk so changed `meta/` and config files take effect.
- If a conflict is detected during startup (before the UI loads), it is stored as a pending flag and surfaced when the Backup page mounts (`GetGitConflictPending()`).

---

## 7. Settings

Configuration panel with four tabs.

### Tools Tab

For each built-in or custom tool:

| Control | Description |
|---------|-------------|
| **Enable toggle** | Enables or disables the tool across the app |
| **Push directory** | Single directory where skills are copied on push; supports both manual text entry and folder-picker button (FolderOpen icon) |
| **Scan directories** | Multiple directories searched when pulling; each row has a folder-picker button and a delete button; new directories added with an input + folder-picker + "Add" button |
| **Delete tool** (custom tools only) | Removes the custom tool entry |

**Add Custom Tool** section (dashed border):

- Tool name input.
- Push directory input with folder-picker button.
- **"Add"** button — `AddCustomTool(name, pushDir)`.

### Cloud Tab

| Control | Description |
|---------|-------------|
| **Provider buttons** | Select cloud provider: Aliyun OSS / Tencent COS / Huawei OBS / **git** |
| **Bucket name** | Object storage bucket name (hidden when git provider is selected) |
| **Credential fields** | Dynamically rendered from `RequiredCredentials()` — text or password inputs per provider. Git fields: repo URL, branch, username, access token |
| **Auto-sync interval** | Number input (minutes); 0 = sync only after mutations; positive value starts a background periodic timer |
| **Enable auto backup toggle** | Turns on/off automatic post-mutation backups and the periodic timer |

### General Tab

| Control | Description |
|---------|-------------|
| **Skills storage directory** | Root path where all skills are stored on disk; manual text entry + folder-picker button |
| **Default category** | Fixed system fallback category `Default` (read-only), used when pulling/importing without specifying a category |
| **Log level buttons** | Toggle runtime log level between `debug`, `info`, and `error` (default: `error`); takes effect after saving settings |
| **Open log directory** | One-click open the local log folder in system file manager |

Log files are stored under the app log directory, with rolling limits:
- At most **2 files** are kept: `skillflow.log` and `skillflow.log.1`.
- Each file is capped at **1MB**.
- When `skillflow.log` reaches the limit, it rotates and overwrites the older backup file.

### Network Tab

Proxy settings for all remote operations (repo scan, GitHub install, update check):

| Mode | Description |
|------|-------------|
| **No proxy** | Direct connection |
| **System proxy** | Reads `HTTP_PROXY` / `HTTPS_PROXY` environment variables |
| **Manual** | Custom proxy URL (http://, https://, socks5://) |

When Manual is selected, a URL input appears with format hint.

### Save Button

- **"Save Settings"** — `SaveConfig(cfg)`; disabled while saving.

---

## 8. Skill Card

Reusable card component shown in the My Skills grid and Sync pages.

### Variants

**Dashboard card** (`SkillCard`):

| Element | Description |
|---------|-------------|
| **Source badge** | GitHub (blue) or Manual (gray) with icon |
| **Skill name** | Truncated; padded to avoid overlap with action buttons |
| **Update dot** (red, top-right) | Shown when `hasUpdate = true` and not in select mode |
| **Open folder button** (FolderOpenDot, top-right) | `OpenPath(skill.path)` — opens directory in OS file manager; visible on hover only |
| **Select checkbox** (top-left) | Visible in select mode only |
| **Hover actions** (bottom-right) | Update (if available) · Copy · Delete — all hidden until hover |
| **Copy button** | Reads `skill.md` content, copies to clipboard, shows "Copied ✓" for 2 s |
| **Drag handle** | Cards are draggable in normal mode; dragged `skillId` moves skill to drop target category |
| **Right-click context menu** | Update (if available) · Move to [Category] (one item per other category) · Delete (red) |

**Sync card** (`SyncSkillCard`):

| Element | Description |
|---------|-------------|
| **Source badge** | Same as above |
| **"Imported" badge** (green) | Shown when skill already exists in the library |
| **Skill name** | Truncated |
| **Subtitle** | Category or repo name |
| **Copy button** (hover) | Same clipboard behavior |
| **Open folder button** (hover) | Same as dashboard card |
| **Selection checkbox** (bottom-right) | Shown when `showSelection = true` |

---

## 9. Skill Tooltip

A floating info panel that appears 300 ms after hovering over any skill card (dashboard only).

### Positioning

- Fixed position, 300 px wide, max 400 px tall.
- Prefers right side of card; falls back to left if near the right window edge.
- Shifts up if near the bottom of the window.

### Content

| Section | Fields shown |
|---------|-------------|
| **Header** | Icon (GitHub / folder) · skill name · source badge · category |
| **Description** | Parsed from `skill.md` YAML frontmatter; shows "No description" if absent; "Loading…" while fetching |
| **Frontmatter fields** | `argument_hint` (Tag icon) · `allowed_tools` (Wrench icon) · `context` (GitBranch icon) |
| **Metadata** | Repository URL (trimmed, opens on click) · installed SHA · available update SHA (amber) · installed date · updated date |

---

## 10. Shared Dialogs

### 10.1 Conflict Dialog

Shown one conflict at a time during push or pull when a skill already exists at the destination.

- Displays: "[Skill name] already exists. How to handle?"
- **"Skip"** — leaves existing file untouched, moves to next conflict.
- **"Overwrite"** — calls the `*Force` variant, replaces the existing file.
- Auto-closes when the conflict queue is empty.

### 10.2 GitHub Install Dialog

Opened from Dashboard toolbar.

| Control | Action |
|---------|--------|
| **URL input** | Git repo URL (HTTPS or SSH); Enter triggers scan |
| **"Scan"** button | `ScanGitHub(url)` — clones or pulls the repo, lists skill candidates |
| **Candidate checkboxes** | Select which skills to install; already-installed skills show a badge |
| **Category dropdown** | Destination category |
| **"Install n Skills"** button | `InstallFromGitHub(url, selected, category)` |

- Info text: "First scan clones the repo; subsequent scans auto-pull."
- Separate error alerts for scan errors and install errors.

### 10.3 Missing Directory Dialog

Appears before any push when target directories do not exist.

- Lists each affected tool name and full directory path.
- **"Create & Push"** — auto-creates directories then proceeds.
- **"Cancel"** — aborts.

---

## 11. Backend Events

Events emitted from the Go backend to the frontend via Wails runtime:

| Event | When fired | Payload |
|-------|-----------|---------|
| `backup.started` | Auto-backup begins | — |
| `backup.progress` | Each file uploaded | `{ currentFile: string }` |
| `backup.completed` | Backup finished | — |
| `backup.failed` | Backup error | — |
| `update.available` | New commit found for a skill | `{ skillID, skillName, currentSHA, latestSHA }` |
| `star.sync.progress` | One repo synced | `{ repoURL, repoName, syncError }` |
| `star.sync.done` | All repos synced | — |
| `git.sync.started` | Git pull/push begins | — |
| `git.sync.completed` | Git sync succeeded | — |
| `git.sync.failed` | Git sync error | — |
| `git.conflict` | Git merge conflict detected | `{ message: string, files?: string[] }` |

The Dashboard listens to `update.available` and marks affected skill cards with a red update dot in real time.
The Backup page listens to all `git.*` events and surfaces the conflict resolution dialog on `git.conflict`.
`App.tsx` listens to all three `app.update.*` events and drives the update banner state machine.

---

## 12. App Update Banner

A fixed top banner that appears when a new app version is detected at startup. Driven by a four-state machine:

| State | Trigger | Banner content |
|-------|---------|---------------|
| `available` | `app.update.available` event | Version label + platform-specific action |
| `downloading` | User clicks "立即下载" (Windows only) | Spinner + version label |
| `ready_to_restart` | `app.update.download.done` event | Completion message + "立即重启" button |
| `download_failed` | `app.update.download.fail` event | Error message + link to release page |

### Platform Behavior

- **Windows** — Full auto-update flow: download → bat script replaces exe → restart.
- **macOS** — Notification only: "查看详情" link opens the GitHub Releases page in the browser.

### Manual Check Button (Settings Page)

A **"检测更新"** button in the top-right corner of the Settings page header:

- Displays current app version (`vX.Y.Z`) next to the button.
- Click → calls `CheckAppUpdate()`; button shows a spinner while checking.
- Result shown inline: "已是最新版本 (vX.Y.Z)" or "发现新版本 vX.Y.Z，请查看顶部横幅".
- On error: "检测失败，请检查网络".
- If a new version is found, the top banner in `App.tsx` is activated via the `app.update.available` event (emitted by `checkAppUpdateOnStartup` on next launch, or the user can trigger the banner manually via the startup flow).

### Controls

| Control | Action |
|---------|--------|
| **查看详情** (macOS, `available`) | Opens `releaseUrl` in system browser |
| **立即下载** (Windows, `available`) | `DownloadAppUpdate(downloadUrl)` — starts async download |
| **立即重启** (`ready_to_restart`) | `ApplyAppUpdate()` — writes bat script and exits; bat replaces exe and relaunches |
| **前往下载页** (`download_failed`) | Opens `releaseUrl` in system browser |
| **×** (all states except `downloading`) | Dismisses banner for the current session |

### Backend Methods

| Method | Description |
|--------|-------------|
| `GetAppVersion()` | Returns current version string (injected by `-ldflags` at build time; `"dev"` in local dev) |
| `CheckAppUpdate()` | Queries GitHub Releases API; returns `AppUpdateInfo` with platform-matched download URL |
| `DownloadAppUpdate(url)` | Downloads new exe to temp file asynchronously; emits `app.update.download.done` or `app.update.download.fail` |
| `ApplyAppUpdate()` | Windows only — writes bat script for post-exit exe replacement, then calls `os.Exit(0)` |

### Version Injection (CI)

GitHub Actions builds inject the tag name at compile time:
```
wails build -ldflags "-X main.Version=${{ github.ref_name }}"
```
The startup check is skipped when `Version == "dev"` (local development).

---

*Last updated: 2026-03-06*

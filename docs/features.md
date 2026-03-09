# SkillFlow — Complete Feature Reference

> 🌐 [中文版](features_zh.md) | **English**
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
12. [App Update Dialog](#12-app-update-dialog)
13. [My Tools](#13-my-tools)
14. [My Prompts](#14-my-prompts)

---

## 1. Navigation & Shell

A fixed left sidebar (w-56) provides navigation throughout the app.

| Route | Icon | Label |
|-------|------|-------|
| `/` | Package | My Skills |
| `/tools` | Wrench | My Tools |
| `/prompts` | FileText | My Prompts |
| `/sync/push` | ArrowUpFromLine | Push to Tools |
| `/sync/pull` | ArrowDownToLine | Pull from Tools |
| `/starred` | Star | Starred Repos |
| `/backup` | Cloud | Cloud Backup |
| `/settings` | Settings | Settings |

- Active route: highlighted with a subtle theme-tinted surface, soft border, and restrained elevation shadow.
- Inactive routes: gray text with hover highlight.
- Top-left of sidebar: the `SkillFlow` wordmark shows the app icon immediately to the left; the icon is slightly taller than the text for clarity and preserves its aspect ratio.
- Top-right of sidebar: **Languages** shortcut button; toggles immediately between **Chinese** and **English**, and persists the preference to `localStorage`.
- Next to it: **Palette** theme shortcut button; cycles immediately through **Dark → Young → Light**.
- Bottom-left **Feedback** button: opens the GitHub "new issue" page in the default browser.
- Window close button behavior: clicking the top-left close button hides the main window and keeps the app running in background.
- Initial window sizing: on launch, SkillFlow first restores the most recently saved window size from local-only `config_local.json`; if none is saved yet, it sizes itself against the current display with a larger desktop-friendly default, clamps to the available screen, and centers the window.
- macOS tray behavior: the app creates a monochrome status icon in the menu-bar status area on startup; after the main window is hidden, the Dock icon is removed and only the menu-bar icon remains. Use native single-click to open a menu with `Show SkillFlow`, `Hide SkillFlow`, and `Quit SkillFlow`.
- Windows tray behavior: app remains in the system notification area with the app's own icon; click the tray icon to open a menu with `Show SkillFlow` and `Exit`.

---

## 2. My Skills (Dashboard)

Central library for managing your skill collection.

### Toolbar

| Control | Action |
|---------|--------|
| **Search input** | Wide search field for real-time case-insensitive filter by skill name; the toolbar wraps on narrower window widths so controls stay visible |
| **Sort toggle** | Two-button toggle for alphabetical order by skill name: **A-Z** or **Z-A** |
| **Update** (RefreshCw) | Calls backend `CheckUpdates()`; groups installed Git-backed skills by normalized repo source + subpath, refreshes `LatestSHA` / `LastCheckedAt`, marks updatable cards with a red dot and Update action, and clears stale markers when an installed instance is already current. This checks only — it does not overwrite local files by itself |
| **Batch Delete** (CheckSquare) | Toggles multi-select mode |
| **Import** (FolderOpen) | Opens native folder-picker → `ImportLocal(dir)` |
| **Remote Install** (Github) | Opens the GitHub Install dialog |

### Auto Push Targets

- A compact single-row strip under the toolbar shows the **Auto Push Targets** title and tool chips together, using the same icon-chip selection style as **Push to Tools**.
- The selection is persisted locally on the current device and reused for future imports into **My Skills**.
- Turning a tool on here immediately backfills the current library to that tool, so existing My Skills entries are pushed right away instead of waiting for the next import.
- Any newly added skill in **My Skills** is automatically copied to the selected tools after the library import succeeds. This applies to local folder import, GitHub install, Pull from Tools, and Starred Repo import.
- Auto-push is non-destructive: if a selected tool already contains a same-name skill in its `PushDir`, SkillFlow skips that target instead of overwriting it.
- Turning a tool off in this strip does not delete anything that was already pushed earlier; removing tool copies still requires manual deletion from **My Tools** or the tool directory.

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
- **Drop highlight** — the target category is highlighted more prominently while dragging a skill card over it.
- **Right-click context menu** on each category:
  - **Rename** — shows inline text input; confirm with Enter, cancel with Escape; calls `RenameCategory()`. (Not available for `Default`.)
  - **Delete** — deletes the category immediately when it is empty; if it still contains skills, shows a blocking dialog telling the user to clear the category first. (`Default` remains undeletable.)
- **New Category** (Plus icon at bottom) — shows inline text input; confirm with Enter or blur, cancel with Escape; calls `CreateCategory()`.

### Skill Grid

- Grid layout: 3 columns, 4 on wide screens.
- **Empty state** — "No Skills found" message with usage hint.
- **Right-click skill menu** — includes move-to-category actions for every category other than the current one, plus delete and update where applicable.
- **Drag-and-drop** — drag a skill card to a category in the sidebar to move it; when drag starts, a smaller floating card follows the cursor; once a sidebar category is targeted, the original card collapses into a thin line until drag ends. Dragging a folder from the OS file manager onto the window imports it directly.
- **Window-level drag overlay** — semi-transparent indigo overlay with "Release to import Skill" message activates when a file is dragged over the window.
- **Hover tooltip** — appears after 300 ms hovering over a card (see [Skill Tooltip](#9-skill-tooltip)).

### Skill Update Flow

- Toolbar **Update** performs a remote update check for installed Git-backed skills and only marks cards as updatable.
- Card-level **Update** is the action that actually downloads the latest files and overwrites the installed copy in **My Skills**.
- Remote checks are grouped by the same logical git key used elsewhere in the app: normalized repo source + repo subpath.
- When multiple installed instances point at the same logical git skill, one remote SHA lookup updates their check state together, but each installed instance still decides its own `updatable` state by comparing its own `SourceSHA`.

```text
[Dashboard toolbar Update]
          |
          v
     CheckUpdates()
          |
          v
Compare installed SourceSHA with remote latest SHA
          |
          +--> no newer SHA  -> keep card unchanged
          |
          +--> newer SHA     -> store LatestSHA -> show red dot / Update action

[Dashboard card Update]
          |
          v
   UpdateSkill(skillID)
          |
          v
Download latest repo subpath again
          |
          v
Overwrite installed folder in My Skills
          |
          v
SourceSHA = LatestSHA -> clear update marker
```

---

## 3. Push to Tools

Copies skills from your library to external tool directories.

### Layout

- Uses a two-column layout similar to My Skills.
- Left sidebar shows category filters: **All** plus every existing category.
- Right side shows the tool selector, search + A-Z/Z-A sort controls, push mode controls, and a skill-card grid for the current category scope.
- Spacing is tuned so the adaptive startup window can usually show the header controls, the current skill grid, and the bottom push action together on common laptop/desktop displays before scrolling is needed.

### Tool Selection

- One toggle button per enabled tool (icon + name).
- Multiple tools can be selected simultaneously.
- Active category, tool, and scope buttons use a brighter theme-tinted background, a lighter border, and a clearer glow so selection remains obvious in dark mode without adding extra symbols.

### Sync Scope

Two push behaviors based on the current left-sidebar category filter:

| Mode | Behavior |
|------|----------|
| **Manual Select** | Shown to the left of **Push All / Push Current Category** and selected by default; uses the current sidebar filter as the candidate list, shows selection checkboxes on cards, and allows select-all for the visible list |
| **Push All / Push Current Category** | If the sidebar is on **All**, pushes the whole library; if a category is selected, pushes only that category |

### Missing Directory Check

Before pushing, the app calls `CheckMissingPushDirs()`. If any target tool directory does not exist yet, a confirmation dialog appears:

- Lists each missing tool name and its full directory path.
- **"Create & Push"** — creates the directory then proceeds.
- **"Cancel"** — aborts without creating anything.

### Conflict Handling

If a skill already exists in the target directory, a conflict dialog appears for each one (see [Conflict Dialog](#101-conflict-dialog)).

### Skill Grid

- Library cards surface only push-relevant state on this page: which tools already contain that logical skill in their `PushDir`.
- The pushed-tool indicator uses compact tool icons with ellipsis overflow and hover-to-reveal full lists.

### Bottom Bar

- **"Start Push (n)"** button — disabled when no tools selected or skill count is zero; shows "Pushing…" while in progress.
- **"Push complete ✓"** — green success message after all pushes finish.

---

## 4. Pull from Tools

Imports skills from external tool directories into your library.

### Layout

- Uses the same two-column shell as Push to Tools.
- Left sidebar lists all categories and controls the import target category.
- Right side contains the source-tool selector, scan feedback, search + A-Z/Z-A sort controls, selectable skill grid, and bottom action bar.

### Tool Selection

- Same toggle buttons as Push; selecting a different tool resets the scanned list.
- The active import target category and selected source tool use the same brighter background, lighter border, and glow treatment as Push so the current choice stays visually distinct in dark mode.

### Scan

- **"Scan"** button — calls `ScanToolSkills(toolName)`; recursively searches the tool's configured scan directories for `skill.md` files.
- Local tool scanning uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- Shows animated "Scanning…" state while in progress.
- **Error alert** (red) if scan fails; **warning alert** (yellow) if no skills found.
- Tool-scan candidates are deduplicated and correlated by logical key first; same-name items are kept distinct when their content-derived keys differ.

### Skill Grid

- Appears after a successful scan.
- Search field filters the scanned skill list by name in real time.
- Two-button sort toggle switches between **A-Z** and **Z-A** ordering by skill name.
- Each card shows only the state relevant to pull decisions: whether that skill is already imported into My Skills.
- After each scan, all skills start unchecked by default.
- Select individual skills, use "Select All / Deselect All" for the currently visible list, or use the matching square-style "Select Not Imported" toggle to bulk-select only visible skills that are not yet imported.
- Selection and pull conflicts are tracked by scanned path, so same-name skills from different tool folders remain independent.

### Bottom Bar

- **"Start Pull (n)"** button — calls `PullFromTool()`.
- **"Pull complete ✓"** — green success message.
- Conflicts handled by the same [Conflict Dialog](#101-conflict-dialog).

---

## 5. Starred Repos

Browse and import skills directly from watched Git repositories without installing them into your library first.

- Repo scanning is recursive across the full clone, so nested skill folders such as `plugins/<plugin>/skills/<name>` are included; `skill.md` matching is case-insensitive.
- Recursive repo scanning is bounded by the configurable **Remote Repo Recursive Scan Depth** setting in **Settings → General** (default `5`, saved range `1-20`).

### View Modes

| Mode | Icon | Description |
|------|------|-------------|
| **Folder** | Folder | Grid of repo cards; click a card to drill into its skills |
| **Flat** | LayoutGrid | All skills from all repos shown in a single grid |

### Toolbar (Normal Mode)

| Button | Action |
|--------|--------|
| **Search input** | Visible in flat view and repo detail view; filters the current skill grid by name |
| **Sort toggle** | Visible in flat view and repo detail view; switches the current skill grid between **A-Z** and **Z-A** |
| **Batch Import** (CheckSquare) | Enters select mode when a skill grid is visible |
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

### Builtin Starter Repos Initialization

- Builtin starter repos are only seeded when `star_repos.json` does not exist yet (first initialization on a device/profile).
- Current builtin starter repos:
  - `https://github.com/anthropics/skills.git`
  - `https://github.com/ComposioHQ/awesome-claude-skills.git`
  - `https://github.com/affaan-m/everything-claude-code.git`
- Once the user deletes a builtin repo, the persisted list remains authoritative and later app launches will not auto-add that repo again.

### Repo Detail View (Drill-down)

- Breadcrumb back button (ChevronLeft) to return to the repo grid.
- Skills grid with same select/import behavior as flat view.
- Repo skill cards show only imported and pushed-tool state on this page.
- Imported badges are resolved from normalized repo source + subpath, so same-name skills from different repos are not conflated.
- Pushed state is rendered as tool-brand icons with hover-to-reveal full tool lists.

### Repo Sync vs Installed Skill Update

- Repo-card **Update** and toolbar **Update All** refresh the locally cached clone for the starred repo.
- This makes the latest repo contents visible in **Starred Repos** so the user can browse or import newer skill files.
- It does **not** overwrite the already installed copy in **My Skills**.
- If a skill has already been imported into the library, updating that installed copy still happens from **My Skills (Dashboard)** via the card-level **Update** action.

```text
                    Cross-page update flow

[Starred Repo card Update]                  [My Skills card Update]
            |                                          |
            v                                          v
   UpdateStarredRepo(url)                       UpdateSkill(skillID)
            |                                          |
            v                                          v
Refresh cached repo clone                      Overwrite installed library copy
            |                                          |
            v                                          v
Refresh Starred Repos list                     Clear update marker in My Skills
            |
            +--> enables newer import candidates
            x--> does NOT change installed My Skills files
```

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
- If conflicts exist, a follow-up dialog lists the exact `skill → tool` pairs that were skipped; **Overwrite All** only overwrites those listed pairs.
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

Mirror your skill library to cloud storage. Two backend types are supported: **Object Storage** (Aliyun OSS, AWS S3, Azure Blob Storage, Google Cloud Storage, Tencent COS, Huawei OBS) and **Git Repository**.

### Status

- **Cloud disabled banner** (yellow) — shown when cloud backup is not configured; links to Settings.

### Actions

| Button | Object Storage label | Git label |
|--------|---------------------|-----------|
| **Backup Now** (Upload icon) | 立即备份 | 立即备份 |
| **Restore / Pull** (Download icon) | 从云端恢复 | 拉取远端 |
| **Refresh** (RefreshCw) | Reloads the latest backup change result | Same |

- Backup Now and Restore are disabled when cloud is not configured.
- **"Backup complete / Git sync complete"** (green) / **"Backup/sync failed"** (red) status messages.

### Backup Change List

- After each successful backup or restore, the page shows only the files involved in that operation, not the full remote file set.
- Backup page also shows **Last sync completed at** based on the most recent successful backup/restore event in the current app session.
- Both object storage and Git render the same per-run list with action badges: `Added`, `Modified`, `Deleted`.
- Deleted entries show a deletion label instead of a file size.
- Refresh reloads the latest in-memory backup result for the current app session.
- Scrollable, max-height container.
- **Unified backup scope (all providers)** — backup root is the app data root (`skills/`, `meta/`, `prompts/`, `config.json`, etc.); `cache/` and `.git/` are excluded.
- **Custom object-storage prefix** — object storage providers let the user choose a parent `remotePath`; SkillFlow always writes under `<bucket>/<remotePath>/skillflow/` (or `<bucket>/skillflow/` when the parent path is empty).
- **Provider-specific cloud profiles** — each cloud provider keeps its own bucket/path/credential set; switching providers in Settings restores that provider's saved values instead of overwriting another provider's form state.
- **Portable synced paths** — local paths persisted inside synced metadata (such as `meta/*.json` and `star_repos.json`) are stored as forward-slash relative paths under the synchronized root, so restores continue to work across macOS and Windows.
- **Local-only path config** — `config_local.json` stores machine-specific filesystem paths such as external `SkillsStorageDir` values, tool `ScanDirs` / `PushDir`, and proxy settings; it is excluded from backup and git sync.
- **Local-only cloud secrets** — sensitive cloud credentials (for example access key IDs, secret keys, and access tokens) are stored only in per-provider entries inside `config_local.json`; synced `config.json` keeps only non-sensitive cloud settings such as provider, bucket name, remote path, endpoint, repo URL, or branch.
- **Git backup compatibility** — when Git backup uses a parent directory as the working tree, SkillFlow automatically moves any legacy nested `skills/.git` metadata aside so actual skill files remain trackable.

### Provider Coverage

- Object storage now also supports AWS S3 (bucket + region), Azure Blob Storage (container + account name + optional service URL), and Google Cloud Storage (bucket + service-account JSON or local key file path), in addition to Aliyun OSS, Tencent COS, and Huawei OBS.
- Sync-safe connection fields now include `region`, `account_name`, and `service_url` alongside existing endpoint / repo URL / branch settings. Sensitive values such as account keys and service-account JSON remain local-only in `config_local.json`.

### Auto-Backup

Triggered automatically after any of these mutations (when cloud is enabled):

- Delete skill / bulk delete
- Create / update / delete prompt
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
  - **"手动解决"** (Resolve Manually) — opens the Git backup root in the system file manager so the user can inspect and fix conflicted files directly.
  - The keep-local and keep-remote actions both reload app state from disk (skills/meta/config) and emit `git.sync.completed` on success.
- **State refresh after pull** — after successful startup pull or manual pull, app state is immediately reloaded from disk so changed `meta/` and config files take effect.
- If a conflict is detected during startup (before the UI loads), it is stored as a pending flag and surfaced when the Backup page mounts (`GetGitConflictPending()`).

---

## 7. Settings

Configuration panel with four tabs in this order: Tools, Cloud, Proxy, General.

The Settings page content expands with the window up to a wider desktop-friendly maximum width instead of staying pinned to a narrow fixed column.

### Tools Tab

For each built-in or custom tool:

| Control | Description |
|---------|-------------|
| **Enable toggle** | Enables or disables the tool across the app |
| **Push directory** | Single directory where skills are copied on push; supports both manual text entry and folder-picker button (FolderOpen icon), which opens at the current path or nearest existing parent |
| **Scan directories** | Multiple directories searched when pulling; each row has a folder-picker button and a delete button; new directories added with an input + folder-picker + "Add" button, with the picker reopening at the current path or nearest existing parent |
| **Delete tool** (custom tools only) | Removes the custom tool entry |

**Add Custom Tool** section (dashed border):

- Tool name input.
- Push directory input with folder-picker button that reopens at the current path or nearest existing parent.
- **"Add"** button — `AddCustomTool(name, pushDir)`.

### Cloud Tab

| Control | Description |
|---------|-------------|
| **Provider buttons** | Responsive provider cards shown in a wrapping grid. The order is fixed to **git** first, then the remaining providers in their backend order. Each provider restores its own saved bucket/path/credential draft when selected |
| **Bucket name** | Object storage bucket name (hidden when git provider is selected) |
| **Remote path** | Object storage parent path (optional). Users enter the parent folder only; SkillFlow always appends `/skillflow/` to build the final backup prefix |
| **Final backup path preview** | Real-time rendered object-storage destination shown as `<bucket>/<remotePath>/skillflow/` so users can verify the exact remote backup location before saving |
| **Credential fields** | Dynamically rendered from `RequiredCredentials()` — text or password inputs per provider. Sync-safe connection fields such as endpoint / repo URL / branch remain in `config.json`, while sensitive credentials such as access keys and tokens are stored only in `config_local.json`. Git fields: repo URL, branch, username, access token |
| **Input normalization** | Aliyun OSS and Huawei OBS bucket fields accept either a plain bucket name or a common full bucket host/URL and normalize it automatically. Tencent COS uses the same bucket + endpoint model as the other object-storage providers. The bucket value always comes from the dedicated bucket field, while the endpoint field may contain either a plain endpoint host or a full bucket host/URL and is preserved as entered for display |
| **Additional provider details** | AWS S3 trims the region field before saving. Azure Blob Storage uses the bucket field as the container name, stores account name and optional service URL as sync-safe fields, and defaults the service URL to `https://<account>.blob.core.windows.net/` when empty. Google Cloud Storage accepts either an inline service-account JSON string or a local key file path, and keeps that credential local-only in `config_local.json`. |
| **Auto-sync interval** | Number input (minutes); 0 = sync only after mutations; positive value starts a background periodic timer |
| **Enable auto backup toggle** | Turns on/off automatic post-mutation backups and the periodic timer |

### General Tab

| Control | Description |
|---------|-------------|
| **Language** | Two buttons, **中文** and **English**, switch the entire frontend language immediately; shares the same state as the sidebar **Languages** button and persists to `localStorage` |
| **Appearance theme** | Three visual presets shown as preview cards: **Dark** (default, refined graphite with muted mist-blue accents), **Young** (a softened paper-blue evolution of the previous sky-blue Light palette), and **Light** (new low-saturation gray-white palette inspired by Messor); persisted to `localStorage`; changes apply immediately without restart; legacy stored `Light` preference auto-migrates to `Young` |
| **Card status visibility** | A compact per-page row list that lets users hide or show only the statuses that page supports by default. Unsupported statuses are not offered for that page. Default policy: **My Skills** = updatable + pushed tools; **My Tools** = imported + updatable + pushed tools; **Push to Tool** = pushed tools; **Pull from Tool** = imported; **Starred Repos** = imported + pushed tools; **GitHub Install** = imported + updatable + pushed tools |
| **Skills storage directory** | Root path where all skills are stored on disk; manual text entry + folder-picker button that opens at the current path or nearest existing parent |
| **Skill recursive scan depth** | Maximum recursion depth used when scanning local tool directories, starred repos, and GitHub-install repos; default `5`; saved values are clamped to `1-20` to avoid pathological nested trees |
| **Default category** | Fixed system fallback category `Default` (read-only), used when pulling/importing without specifying a category |
| **Log level buttons** | Toggle runtime log level between `debug`, `info`, and `error` (default: `error`); takes effect after saving settings |
| **Launch at login toggle** | Enables/disables OS login-item registration so SkillFlow auto-starts after sign-in on the current device; stored only in local `config_local.json` |
| **Open log directory** | One-click open the local log folder in system file manager; missing targets fall back to the nearest existing parent directory |

Log files are stored under the app log directory, with rolling limits:
- At most **2 files** are kept: `skillflow.log` and `skillflow.log.1`.
- Each file is capped at **1MB**.
- When `skillflow.log` reaches the limit, it rotates and overwrites the older backup file.

### Proxy Tab

Proxy settings for all remote operations (repo scan, GitHub install, update check):

| Mode | Description |
|------|-------------|
| **No proxy** | Direct connection |
| **System proxy** | Reads `HTTP_PROXY` / `HTTPS_PROXY` environment variables |
| **Manual** | Custom proxy URL (http://, https://, socks5://) |

When Manual is selected, a URL input appears with format hint. Proxy settings are persisted in `config_local.json`, so manual proxy values survive restart and are not included in backup/git sync.

### Save Button

- **"Save Settings"** — `SaveConfig(cfg)`; disabled while saving.
- **Keyboard shortcut** — while the Settings page is open, `Ctrl+S` on Windows/Linux and `Cmd+S` on macOS trigger the same save action and suppress the browser/webview default save shortcut.

---

## 8. Skill Card

Reusable card component shown in the My Skills grid and Sync pages.

### Variants

**Dashboard card** (`SkillCard`):

| Element | Description |
|---------|-------------|
| **Status strip** | Source badge plus any coexisting state badges (for example Update available + pushed-tool icons) rendered in one compact header row on the card; the strip prefers a single line when space allows, then automatically wraps instead of clipping badges away when cards become too narrow |
| **Skill name** | Two-line clamp; padded to avoid overlap with action buttons |
| **Open folder button** (FolderOpenDot, top-right) | `OpenPath(skill.path)` — opens directory in OS file manager; visible on hover only |
| **Select checkbox** (top-left) | Visible in select mode only |
| **Hover actions** (bottom-right) | Update (if available) · Copy · Delete — all hidden until hover |
| **Copy button** | Reads `skill.md` content, copies to clipboard, shows "Copied ✓" for 2 s |
| **Drag handle** | Cards are draggable in normal mode; dragged `skillId` moves skill to drop target category |
| **Right-click context menu** | Update (if available) · Move to [Category] (one item per other category) · Delete (red) |

**Sync card** (`SyncSkillCard`):

| Element | Description |
|---------|-------------|
| **Status strip** | Source badge plus only the page-selected state badges (for example imported, update-available, or pushed-tool icons) rendered together in one compact header row; cards keep imported and pushed-tool indicators on the same line when they fit, and automatically wrap them to a second row when the card width is too tight |
| **Pushed-tool indicator** | Shows the exact tools whose `PushDir` already contains this logical skill via small tool-brand icons without an extra arrow prefix; overflows collapse into a compact count badge while hover still reveals the full list |
| **Skill name** | Two-line clamp |
| **Subtitle** | Category or repo name |
| **Copy button** (hover) | Same clipboard behavior |
| **Open folder button** (hover) | Same as dashboard card |
| **Selection checkbox** (bottom-right) | Shown when `showSelection = true` |

### Unified Status Semantics

| State | Meaning |
|------|---------|
| **installed** | The logical skill already exists in **My Skills** as at least one installed instance |
| **imported** | External-page wording for **installed**; on GitHub / Starred Repos / tool views it means “already in My Skills” |
| **pushed** | The logical skill already exists in a tool's configured **PushDir** |
| **pushedTools** | The exact tool names whose configured **PushDir** currently contains that logical skill; used for icon rendering on cards |
| **seenInToolScan** | The logical skill was detected in one of a tool's configured **ScanDirs**; this means the tool already has it somewhere, but not necessarily because SkillFlow pushed it. This is currently used for grouping and correlation, not shown as a card badge. |
| **updatable** | An installed Git-backed skill has a newer remote SHA than its installed `SourceSHA` |

These statuses are resolved by the backend from a unified logical-key model; frontend pages no longer infer them independently from `Name` or `Path`.

```text
                 Unified skill status picture

[GitHub candidate] [Starred skill] [Tool scan candidate]
         \              |               /
          \             |              /
           +---- same logical skill ----+
                        |
                        v
                 [My Skills instance]
                  installed = true
           external pages show imported = true
                        |
                        +---- copy to tool PushDir ----> pushed = true
                        |
                        +---- remote newer SHA -------> updatable = true

[Tool ScanDirs] ---- detect same logical skill ----> seenInToolScan = true
```

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
- **"Overwrite"** — calls the `*Force` variant, replaces only the exact conflicting skill-target pair.
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
- Candidate discovery is recursive across the cloned repo, so nested layouts such as `plugins/<plugin>/skills/<name>` are also listed; `skill.md` matching is case-insensitive.
- Recursive candidate discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- Already-installed badges are resolved from normalized repo source + subpath instead of `Name`, and checkbox state is tracked by candidate path so same-name candidates remain independent.
- Candidate rows can show imported / update-available / pushed-tool state according to the **Settings → General → Card status visibility** policy for GitHub Install, and display the candidate subpath so same-name entries stay distinguishable in the dialog.
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
| `backup.completed` | Backup finished | `{ files: Array<{ path, size, action }> }` |
| `backup.failed` | Backup error | — |
| `update.available` | New commit found for a skill | `{ skillID, skillName, currentSHA, latestSHA }` |
| `star.sync.progress` | One repo synced | `{ repoURL, repoName, syncError }` |
| `star.sync.done` | All repos synced | — |
| `git.sync.started` | Git pull/push begins | — |
| `git.sync.completed` | Git sync succeeded | `{ files: Array<{ path, size, action }> }` when triggered by backup/restore/conflict resolution |
| `git.sync.failed` | Git sync error | — |
| `git.conflict` | Git merge conflict detected | `{ message: string, files?: string[] }` |

The Dashboard listens to `update.available` and marks affected skill cards with a red update dot in real time.
The Backup page listens to all `git.*` events and surfaces the conflict resolution dialog on `git.conflict`.
`App.tsx` listens to all three `app.update.*` events, parses the emitted `AppUpdateInfo` payload, and drives the update dialog state machine.

---

## 12. App Update Dialog

A centered modal dialog that appears when a new app version is detected. Triggered by both the automatic startup check and the manual check in Settings. Driven by a four-state machine:

| State | Trigger | Dialog content |
|-------|---------|---------------|
| `available` | `app.update.available` event | Version labels + release notes + platform-specific action buttons |
| `downloading` | User clicks "下载并自动重启更新" (Windows only) | Spinner + progress message |
| `ready_to_restart` | `app.update.download.done` event | Completion message + "立即重启" / "稍后重启" buttons |
| `download_failed` | `app.update.download.fail` event | Error message + "前往下载页" button |

### Platform Behavior

Both startup and manual checks surface the same modal dialog.

- **Windows** — Three choices in the `available` state:
  1. **下载并自动重启更新** — downloads new exe in the background, then prompts restart.
  2. **前往 Release 页面手动下载** — opens the GitHub Releases page in the system browser.
  3. **跳过此版本（下次启动不再提示）** — persists the skipped version; the startup check will not prompt for this version again. The manual check always shows the dialog regardless.
- **macOS** — Two choices in the `available` state (auto-download not supported):
  1. **前往 Release 页面手动下载** — opens the GitHub Releases page.
  2. **跳过此版本（下次启动不再提示）** — same skip behavior as Windows.
- The `available` dialog always renders the current version, latest version, and the Release-page action from the same `AppUpdateInfo` payload on both platforms.

### Skip Version Behavior

- The skipped version tag is stored in `AppConfig.SkippedUpdateVersion` and persisted in the shared `config.json` file, so it survives app restarts.
- On app startup, if `latestVersion == skippedUpdateVersion` the `app.update.available` event is **not** emitted and no dialog appears.
- When the user manually clicks "检测更新" in Settings, `CheckAppUpdateAndNotify` always emits the event, bypassing the skip — the dialog always appears for manual checks.
- Clicking "跳过此版本" calls `SetSkippedUpdateVersion(latestVersion)`.

### Manual Check Button (Settings Page)

A **"检测更新"** button in the top-right corner of the Settings page header:

- Displays current app version (`vX.Y.Z`) next to the button.
- Click → calls `CheckAppUpdateAndNotify()`; button shows a spinner while checking.
- If a new version is found, the update dialog opens automatically via the `app.update.available` event.
- If already up-to-date, inline text shows "已是最新版本 (vX.Y.Z)".
- On error: "检测失败: …" shown inline.

### Controls

| Control | Action |
|---------|--------|
| **下载并自动重启更新** (Windows, `available`) | `DownloadAppUpdate(downloadUrl)` — starts async download |
| **前往 Release 页面手动下载** (`available`) | `OpenURL(releaseUrl)` — opens release page in browser; closes dialog |
| **跳过此版本** (`available`) | `SetSkippedUpdateVersion(latestVersion)` — persists skip; closes dialog |
| **立即重启** (`ready_to_restart`) | `ApplyAppUpdate()` — writes bat script and exits; bat replaces exe and relaunches |
| **稍后重启** (`ready_to_restart`) | Closes dialog without restarting |
| **前往下载页** (`download_failed`) | `OpenURL(releaseUrl)` — opens release page in browser; closes dialog |
| **×** (all states except `downloading`) | Closes dialog for the current session |

### Backend Methods

| Method | Description |
|--------|-------------|
| `GetAppVersion()` | Returns current version string (injected by `-ldflags` at build time; `"dev"` in local dev) |
| `CheckAppUpdate()` | Queries GitHub Releases API; returns `AppUpdateInfo` with platform-matched download URL |
| `CheckAppUpdateAndNotify()` | Calls `CheckAppUpdate()` and emits `EventAppUpdateAvailable` if an update is found; always notifies regardless of skipped version |
| `GetSkippedUpdateVersion()` | Returns the version tag stored in config as the user-skipped version |
| `SetSkippedUpdateVersion(version)` | Persists the skipped version tag to config; pass `""` to clear |
| `DownloadAppUpdate(url)` | Downloads new exe to temp file asynchronously; emits `app.update.download.done` or `app.update.download.fail` |
| `ApplyAppUpdate()` | Windows only — writes bat script for post-exit exe replacement, then calls `os.Exit(0)` |

### Version Injection (CI)

GitHub Actions builds inject the tag name at compile time:
```
wails build -ldflags "-X main.Version=${{ github.ref_name }}"
```
The startup check is skipped when `Version == "dev"` (local development).

---

## 13. My Tools

Browse the skills currently present inside each enabled tool.

### Layout

- Left sidebar lists enabled tools.
- Main area shows one toolbar plus two skill-card sections: **Push Path** and **Scan Path**.

### Toolbar

| Control | Description |
|---------|-------------|
| **Search input** | Filters both Push Path and Scan Path skill cards by name in real time |
| **Sort toggle** | Switches both sections between **A-Z** and **Z-A** ordering by skill name |
| **Batch Delete** | Available when the visible Push Path list is non-empty; enters select mode |

### Push Path Section

- Shows deletable tool-local skills under the configured push directory.
- Push-path discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- In select mode, **Select All / Deselect All** applies to the currently visible filtered Push Path cards only.
- Cards show only imported, update-available, and "pushed to other tools" states. The current tool is excluded from the pushed-tool icon list so the card only surfaces cross-tool distribution that adds information.

### Scan Path Section

- Shows read-only skills discovered only from scan directories.
- Scan-path discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- Shares the same search and sort controls as Push Path.
- Scan-path cards use the same compact strip for imported / update-available / pushed-to-other-tools states; the fact that they were found via scan paths is conveyed by the section itself instead of repeating a detected badge on every card.

---

## 14. My Prompts

Store reusable system prompts inside the synced `prompts/` directory.

### Navigation & Storage

- Sidebar adds **My Prompts** directly below **My Tools**.
- Each prompt is stored as `prompts/<category>/<name>/system.md` under the backup root, so both object-storage providers and the Git provider sync the same prompt files automatically.
- Prompt names are required, globally unique in the library, and used as the folder key.

### Layout

- The page mirrors **My Skills** with a left category sidebar and a right content pane.
- Categories support **Default** as the built-in fallback group.
- Prompt cards are filtered by selected category, keyword search, and A-Z / Z-A sorting.
- Search supports logical `and` / `or` syntax, for example `review and golang` or `summary or changelog`.

### Category Management

- Categories can be created from the sidebar.
- Non-default categories support rename and delete from the context menu.
- Prompt cards can be dragged onto categories to move them, matching the category-drop behavior from **My Skills**.

### Prompt Cards

- Cards reuse the same visual language as Skill cards (`card-base`, hover glow, compact actions).
- Each card shows **name**, optional **description**, and the opening content excerpt by default.
- When the content is longer than the preview window, the card displays **Click to view more**.
- Top-right action button copies the full prompt content to the desktop clipboard in one click, preserving multi-line content on Windows.
- Bottom-right **Delete** action opens a confirmation dialog before removing both the card and the underlying prompt folder.

### Prompt Editor

- Clicking **Add Prompt** opens a built-in editor with fields for **name** (required), **description** (optional), **category**, and full `system.md` content.
- Clicking an existing prompt card opens the same editor pre-filled for editing and rename operations.
- Saving writes the prompt back to `prompts/<category>/<name>/system.md`.

### Import / Export

- Toolbar **Import** reads a JSON prompt library file and creates or updates prompts by name.
- Toolbar **Export** writes the full prompt library to a JSON file selected by the user.

---

*Last updated: 2026-03-09*

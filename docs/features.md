# SkillFlow — Complete Feature Reference

> 🌐 [中文版](features_zh.md) | **English**
>
> This document enumerates every feature, button, interaction, and UX detail in SkillFlow.
> **Keep this file in sync whenever features are added, changed, or removed.**

---

## Table of Contents

1. [Navigation & Shell](#1-navigation--shell)
2. [My Skills (Dashboard)](#2-my-skills-dashboard)
3. [Legacy Push Route](#3-legacy-push-route)
4. [Legacy Pull Route](#4-legacy-pull-route)
5. [Starred Repos](#5-starred-repos)
6. [Cloud Backup](#6-cloud-backup)
7. [Settings](#7-settings)
8. [Skill Card](#8-skill-card)
9. [Skill Tooltip](#9-skill-tooltip)
10. [Shared Dialogs](#10-shared-dialogs)
11. [Backend Events](#11-backend-events)
12. [App Update Dialog](#12-app-update-dialog)
13. [My Agents](#13-my-agents)
14. [My Prompts](#14-my-prompts)
15. [My Memory](#15-my-memory)

---

## 1. Navigation & Shell

A fixed left sidebar (w-56) provides navigation throughout the app.

| Route | Icon | Label |
|-------|------|-------|
| `/` | Package | My Skills |
| `/prompts` | FileText | My Prompts |
| `/memory` | Brain | My Memory |
| `/tools` | Wrench | My Agents |
| `/starred` | Star | Starred Repos |
| `/backup` | Cloud | Cloud Backup |
| `/settings` | Settings | Settings |

- Active route: highlighted with a subtle theme-tinted surface, soft border, and restrained elevation shadow.
- Inactive routes: gray text with hover highlight.
- Top-left of sidebar: the `SkillFlow` wordmark shows the app icon immediately to the left; the icon is slightly taller than the text for clarity and preserves its aspect ratio.
- Top-right of sidebar: **Languages** shortcut button; toggles immediately between **Chinese** and **English**, and persists the preference to `localStorage`.
- Next to it: **Palette** theme shortcut button; cycles immediately through **Dark → Young → Light → Sport**.
- Primary sidebar section now contains only the four core work surfaces: **My Skills**, **My Prompts**, **My Memory**, and **My Agents**.
- The lower utility section keeps **Starred Repos**, **Cloud Backup**, **Settings**, and **Feedback**.
- Legacy routes remain valid for compatibility:
  - `/sync/push` redirects to `/`
  - `/sync/pull` redirects to `/tools`
- Bottom-left **Feedback** button: opens the GitHub "new issue" page in the default browser.
- Window close button behavior: clicking the top-left close button closes the current UI process instead of only hiding it. The background `daemon` remains in the tray/menu bar, so `Show SkillFlow` or launching the app again opens a fresh window.
- Primary-page refresh behavior: route transitions remount the page subtree keyed by `location.pathname`, so re-entering pages such as **My Skills**, **My Agents**, **My Prompts**, and **Starred Repos** fetches fresh backend state again without requiring a manual in-page refresh.
- Startup smoothing behavior: after the shell is ready, background startup jobs are staggered instead of launching together, so the first interactive second is less likely to compete with skill update checks, starred refresh, app update checks, and git startup pull at the same time.
- Process residency behavior: when the main window is active, SkillFlow runs a background `daemon` plus a visible `ui` process. When the window is hidden to the tray/menu bar, only the `daemon` remains resident; the `ui` process exits and releases its page memory.
- Window reactivation behavior: when the window is hidden to the tray/menu bar, the UI process exits and releases its in-memory page state. Showing SkillFlow again starts a fresh UI process, so the window reopens from a cold start instead of restoring the previous routed page tree and in-progress interactions.
- Initial window sizing: on launch, SkillFlow first restores the most recently saved window size from local-only `config_local.json`; if none is saved yet, it sizes itself against the current display with a larger desktop-friendly default, clamps to the available screen, and centers the window.
- macOS tray behavior: the background `daemon` keeps a monochrome status icon in the menu-bar status area. Closing the main window or choosing `Hide SkillFlow` exits the `ui` process, but the menu-bar item remains available with `Show SkillFlow`, `Hide SkillFlow`, and `Quit SkillFlow`; `Show SkillFlow` relaunches the UI window from a fresh start when no UI process is active.
- Windows tray behavior: the background `daemon` remains in the system notification area with the app icon. Closing the main window exits the `ui` process; the tray menu can still `Show SkillFlow` to relaunch/focus the UI, or `Exit` to terminate both `daemon` and `ui`.

---

## 2. My Skills (Dashboard)

Central library for managing your skill collection.

### Toolbar

| Control | Action |
|---------|--------|
| **Search input** | Wide search field for real-time case-insensitive filter by skill name; the toolbar wraps on narrower window widths so controls stay visible |
| **Sort toggle** | Two-button toggle for alphabetical order by skill name: **A-Z** or **Z-A** |
| **Update** (RefreshCw) | Calls backend `CheckUpdates()`; groups installed Git-backed skills by normalized repo source + subpath, compares installed `SourceSHA` against the same subpath inside the local repo cache, refreshes `LatestSHA` (synced) and `LastCheckedAt` (local-only), marks updatable cards with a red dot and Update action, and clears stale markers when an installed instance is already current. This checks only — it does not overwrite local files by itself |
| **Manual Push** (ArrowUpFromLine) | Enters inline manual-push mode inside **My Skills** without leaving the page |
| **Batch Delete** (CheckSquare) | Toggles multi-select mode |
| **Import** (FolderOpen) | Opens native folder-picker → `ImportLocal(dir)` |
| **Auto Update** (ToggleLeft / ToggleRight) | Uses a clear on/off state button in the old remote-install slot: when auto update is off, the control shows a muted "Turn Auto Update On" action; when it is on, it switches to a highlighted "Turn Auto Update Off" action. Successful toggles show a confirmation notice, and enabled state keeps the same local-only automatic update behavior after startup or manual starred-repo refresh |

### Auto Push Targets

- A compact single-row strip under the toolbar shows the **Auto Push Targets** title and agent chips together, using the same icon-chip selection style as the inline manual-push mode.
- The selection is persisted locally on the current device and reused for future imports into **My Skills**.
- The toolbar **Auto Update** toggle is also persisted locally on the current device, alongside the auto-push target selection.
- The button now exposes the next action instead of a static label, and also uses left/right toggle icons plus contrastful styling so the current state is visible before you click it.
- Turning an agent on here immediately backfills the current library to that agent, so existing My Skills entries are pushed right away instead of waiting for the next import.
- Any newly added skill in **My Skills** is automatically copied to the selected agents after the library import succeeds. This applies to local folder import, Pull from Agents, and Starred Repo import.
- If a cloud restore brings skills onto the current device, SkillFlow auto-pushes any newly restored or newly updated library skills to the selected agents on this device.
- Import auto-push remains non-destructive: if a selected agent already contains a same-name skill in its `PushDir`, SkillFlow skips that target instead of overwriting it.
- Turning an agent off in this strip does not delete anything that was already pushed earlier; removing agent copies still requires manual deletion from **My Agents** or the agent directory.

### Manual Push Mode

- Clicking **Manual Push** turns the dashboard into an inline batch-push surface instead of navigating to a separate page.
- The current category filter, search query, and sort order define the candidate skill list for this push.
- In manual-push mode:
  - skill cards become selectable
  - the strip below the toolbar switches from **Auto Push Targets** to **Manual Push Targets**
  - users can select one or more enabled agents
  - users can **Select All / Deselect All** for the currently visible filtered skills
  - the toolbar shows **Start Push (n)** and **Cancel**
- Manual push is enabled only when at least one target agent and one visible skill are selected.
- Before pushing, SkillFlow still performs the same missing-directory confirmation used by the old dedicated push page.
- Push conflicts still use the existing overwrite/skip dialog behavior, but the flow now completes inside **My Skills**.

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
- List performance behavior: page entry now prefers a local derived snapshot for installed skills when dependencies still match, and route transitions use a lighter fade-only animation to reduce navigation jank.
- Dense-list animation fallback: when the visible Dashboard list exceeds 18 cards, staggered per-card entry motion is disabled automatically and cards render immediately instead of animating one-by-one.
- Card status-strip overflow: status badges and pushed-agent icons now use deterministic truncation with compact `+N` overflow instead of per-card runtime measurement.
- **Empty state** — "No Skills found" message with usage hint.
- **Right-click skill menu** — includes move-to-category actions for every category other than the current one, plus delete and update where applicable.
- **Drag-and-drop** — drag a skill card to a category in the sidebar to move it; when drag starts, a smaller floating card follows the cursor; once a sidebar category is targeted, the original card collapses into a thin line until drag ends. Dragging a folder from the OS file manager onto the window imports it directly.
- **Window-level drag overlay** — semi-transparent indigo overlay with "Release to import Skill" message activates when a file is dragged over the window.
- **Hover tooltip** — appears after 300 ms hovering over a card (see [Skill Tooltip](#9-skill-tooltip)).

### Skill Update Flow

- Toolbar **Update** compares installed Git-backed skills against their locally cached repo clones and only marks cards as updatable.
- Card-level **Update** is the action that copies the latest files from the local repo cache, overwrites the installed copy in **My Skills**, refreshes any same-skill copies that already exist in agent `PushDir`s, and force-updates all agents currently selected in **Auto Push Targets** (creating missing copies and overwriting existing copies there).
- When **Auto Update** is enabled on the Dashboard, the same installed-skill update flow runs automatically after startup starred-repo refresh and after manual single-repo or refresh-all actions in **Starred Repos**.
- Automatic updates still only create or overwrite agent copies for the agents selected in **Auto Push Targets**. Other agents keep their existing copies and surface **Updatable** when those copies fall behind the installed library version.
- While a card update is running, that card keeps the Update action visible, disables repeat clicks, and spins the Refresh icon until the request finishes.
- The Dashboard also shows a temporary top-of-page status banner for skill update progress, success, or failure so users can tell whether the click really triggered work.
- Cache-based checks are grouped by the same logical git key used elsewhere in the app: normalized repo source + repo subpath.
- When multiple installed instances point at the same logical git skill, one cached SHA lookup updates their check state together, but each installed instance still decides its own `updatable` state by comparing its own `SourceSHA`.
- If the corresponding local repo cache is missing or stale, SkillFlow does not fall back to a direct GitHub API lookup here; users need the normal repo-cache sync path to refresh that clone first.

```text
[Dashboard toolbar Update]
          |
          v
     CheckUpdates()
          |
          v
Compare installed SourceSHA with cached repo subpath SHA
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
Copy latest repo subpath from local cache clone
          |
          v
Overwrite installed folder in My Skills
          |
          v
Refresh existing agent PushDir copies for that skill
          |
          v
SourceSHA = LatestSHA -> clear update marker
```

---

## 3. Legacy Push Route

The standalone **Push to Agents** page is no longer a primary navigation destination.

- `/sync/push` is preserved as a compatibility route.
- Opening it now redirects to **My Skills** (`/`).
- Manual push behavior now lives in the dashboard's inline **Manual Push** mode.

---

## 4. Legacy Pull Route

The standalone **Pull from Agents** page is no longer a primary navigation destination.

- `/sync/pull` is preserved as a compatibility route.
- Opening it now redirects to **My Agents** (`/tools`).
- Manual pull behavior now lives in the **My Agents** skills panel through the inline **Manual Pull** mode for the currently selected agent.

---

## 5. Starred Repos

Browse and import skills directly from watched Git repositories without installing them into your library first.

- Repo scanning is recursive across the full clone, so nested skill folders such as `plugins/<plugin>/skills/<name>` are included; `skill.md` matching is case-insensitive.
- Repos whose root directory itself contains `SKILL.md` / `skill.md` are also treated as a single skill candidate.
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
| **Push to Agents (n)** | Opens the push-to-agents dialog (legacy flow, still available from Starred Repos) |
| **Import to My Skills (n)** | Opens the Import dialog |
| **Cancel** | Exits select mode |

### Repo Card (Folder View)

- Click to open the skill list for that repo.
- Card header now highlights the repo host (for example `github.com`) and shows a compact skill count so users can quickly compare repo scale before drill-down.
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
- While repo skills are still loading, both the toolbar count label and the content pane show **Loading...** instead of a transient `0 skills` / `No Skills found` empty state.
- Repo skill cards show only imported and pushed-agent state on this page.
- Imported badges are resolved from normalized repo source + subpath, so same-name skills from different repos are not conflated.
- Pushed state is rendered as agent-brand icons with hover-to-reveal full agent lists.

### Repo Sync vs Installed Skill Update

- Repo-card **Update** and toolbar **Update All** refresh the locally cached clone for the starred repo.
- This makes the latest repo contents visible in **Starred Repos** so the user can browse or import newer skill files.
- If cloud restore syncs a newly starred repo onto this device, SkillFlow immediately clones that repo locally so search, import, and direct push-to-agents are ready without waiting for a manual update.
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

### Push to Agents Dialog

- Description: "Copies skills directly to the agent directory; no need to import first."
- Lists all enabled agents as checkboxes with their push directory paths shown.
- **Empty state** message if no agents are configured.
- **"Push to n agents"** button.
- If conflicts exist, a follow-up dialog lists the exact `skill → agent` pairs that were skipped; **Overwrite All** only overwrites those listed pairs.
- **"Cancel"**.

### Missing Directory Confirmation

Same behavior as the current manual-push flow in **My Skills**: confirms before creating absent push directories.

### Push Conflict Dialog

When skills already exist in the target agent directory:

- Lists all conflicting skill names.
- **"Overwrite All"** (amber) — `PushStarSkillsToAgentsForce()`.
- **"Skip Conflicts"** — `PushStarSkillsToTools()` (already resolved; conflicts discarded).

---

## 6. Cloud Backup

Mirror your skill library to cloud storage. Two backend types are supported: **Object Storage** (Aliyun OSS, AWS S3, Azure Blob Storage, Google Cloud Storage, Tencent COS, Huawei OBS) and **Git Repository**.

### Status

- **Cloud disabled banner** (yellow) — shown when cloud backup is not configured; links to Settings.

### Actions

| Button | Object Storage label | Git label |
|--------|---------------------|-----------|
| **Backup Now** (Upload icon) | Backup Now | Backup Now |
| **Restore / Pull** (Download icon) | Restore from Cloud | Pull Remote |
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
- **Unified backup scope (all providers)** — backup always uses `AppDataDir` as the Git/cloud working root. Synced content lives there under `skills/`, `meta/`, and `prompts/`, while local-only `cache/`, `runtime/`, `logs/`, `meta_local/`, `.git/`, `config_local.json`, and `star_repos_local.json` are excluded. The starred-repo clone cache at `repoCacheDir` is local-only and never part of backup scope.
- **Custom object-storage prefix** — object storage providers let the user choose a parent `remotePath`; SkillFlow always writes under `<bucket>/<remotePath>/skillflow/` (or `<bucket>/skillflow/` when the parent path is empty).
- **Provider-specific cloud profiles** — each cloud provider keeps its own bucket/path/credential set; switching providers in Settings restores that provider's saved values instead of overwriting another provider's form state.
- **Portable synced paths** — local paths persisted inside synced metadata (such as `meta/*.json`) are stored as forward-slash relative paths under `AppDataDir`, so restores continue to work across macOS and Windows.
- **Local-only volatile skill metadata** — high-churn per-skill check timestamps (`LastCheckedAt`) are stored in local-only `meta_local/*.local.json` overlays, so they do not create cross-device git/cloud merge churn.
- **Local-only starred-repo sync runtime state** — per-repo `lastSync` and `syncError` are stored in local-only `star_repos_local.json`, so background sync attempts on one device do not churn synced repo metadata on other devices.
- **Local-only path config** — `config_local.json` stores machine-specific filesystem paths such as `repoCacheDir`, agent `ScanDirs` / `PushDir` / `MemoryPath` / `RulesDir`, and proxy settings; it is excluded from backup and git sync.
- **Other local-only device state** — per-device choices and runtime shell state such as auto-push targets, launch-at-login registration, and the last saved window size remain in `config_local.json`, so restoring on one machine does not overwrite another machine's local behavior.
- **Local-only cloud secrets** — sensitive cloud credentials (for example access key IDs, secret keys, and access tokens) are stored only in per-provider entries inside `config_local.json`; synced `config.json` keeps only non-sensitive cloud settings such as provider, bucket name, remote path, endpoint, repo URL, or branch.
- **Git backup compatibility** — when Git backup uses a parent directory as the working tree, SkillFlow automatically moves any legacy nested `skills/.git` metadata aside so actual skill files remain trackable.
- **Post-restore device compensation** — after a successful cloud restore, newly restored library skills are auto-pushed to this device's selected auto-push agents, and newly restored starred repos are cloned locally right away.

### Provider Coverage

- Object storage now also supports AWS S3 (bucket + region), Azure Blob Storage (container + account name + optional service URL), and Google Cloud Storage (bucket + service-account JSON or local key file path), in addition to Aliyun OSS, Tencent COS, and Huawei OBS.
- Sync-safe connection fields now include `region`, `account_name`, and `service_url` alongside existing endpoint / repo URL / branch settings. Sensitive values such as account keys and service-account JSON remain local-only in `config_local.json`.

### Auto-Backup

Triggered automatically after any of these mutations (when cloud is enabled):

- Delete skill / bulk delete
- Create / update / delete prompt
- Manual import
- Install from GitHub
- Pull from agent
- Update skill
- Import from starred repo

Progress events surface in the UI via the Wails event system (`backup.started`, `backup.progress`, `backup.completed`, `backup.failed`).

### Git Sync (Git provider only)

When the **git** provider is selected:

- **Repository bootstrap** — if the Skills directory is not a git repo, SkillFlow auto-initializes it and configures `origin` from the configured repo URL.
- **Remote binding self-heal** — if `origin` is missing or changed, SkillFlow auto-adds/updates it before pull/push.
- **Startup pull** — on every app launch, SkillFlow runs `git pull` on the Git backup root directory to fetch the latest remote changes.
- **Staggered startup background work** — the startup pull is still automatic, but it no longer starts in the same burst as every other startup check; SkillFlow spreads those jobs out after the UI becomes interactive.
- **Missing branch tolerance** — if the configured remote branch does not exist yet (first-time setup), startup pull is skipped without failing the backup page.
- **Auto-push after mutations** — same post-mutation trigger as object storage; runs `git add -A && git commit && git push`.
- **Excluded-path cleanup** — on every Git backup push, local-only excluded directories such as `cache/` and `runtime/` are removed from the Git index if an older version ever tracked them.
- **Periodic auto-sync** — controlled by the "Auto-sync interval" setting (in minutes, 0 = disabled). A background timer fires `autoBackup()` on the configured interval.
- **Manual actions with conflict detection** — both **Backup Now** and **Restore / Pull** detect git conflicts/divergence and emit `git.conflict` when user action is required.
- **Conflict resolution dialog** — if `git pull` or `git push` detects a conflict or diverged history, a modal appears:
  - The dialog includes a conflict file list when available.
  - **"Keep Local"** — aborts the merge, force-pushes local state to remote. Calls `ResolveGitConflict(true)`.
  - **"Keep Remote"** — aborts the merge, resets local to `origin/<branch>`. Calls `ResolveGitConflict(false)`.
  - **"Resolve Manually"** — opens the Git backup root in the system file manager so the user can inspect and fix conflicted files directly.
  - The keep-local and keep-remote actions both reload app state from disk (skills/meta/config) and emit `git.sync.completed` on success.
- **State refresh after pull** — after successful startup pull or manual pull, app state is immediately reloaded from disk so changed `meta/` and config files take effect.
- If a conflict is detected during startup (before the UI loads), it is stored as a pending flag and surfaced when the Backup page mounts (`GetGitConflictPending()`).

---

## 7. Settings

Configuration panel with four tabs in this order: Agents, Cloud, Proxy, General.

The Settings page content expands with the window up to a wider desktop-friendly maximum width instead of staying pinned to a narrow fixed column.

### Agents Tab

For each built-in or custom agent:

| Control | Description |
|---------|-------------|
| **Enable toggle** | Enables or disables the agent across the app |
| **Skill Paths section** | Dedicated grouped block for skill distribution paths |
| **Push directory** | Single directory where skills are copied on push; supports both manual text entry and folder-picker button (FolderOpen icon), which opens at the current path or nearest existing parent |
| **Scan directories** | Multiple directories searched when pulling; each row has a folder-picker button and a delete button; new directories added with an input + folder-picker + "Add" button, with the picker reopening at the current path or nearest existing parent. Built-in agents may use scan directories that differ from the push directory; Copilot defaults to `~/.copilot/skills` for push and scans its other shared sources separately |
| **Memory Paths section** | Dedicated grouped block for memory-sync paths |
| **Memory file** | Path to the agent's main memory file |
| **Rules directory** | Path to the agent's rules directory. Built-in agents without a first-party rules-directory concept, such as Copilot, keep this field empty by default |
| **Delete agent** (custom agents only) | Removes the custom agent entry |

**Add Custom Agent** entry point (dashed border):

- Opens a centered modal dialog instead of using an inline form.
- The dialog collects:
  - agent name
  - initial skill path (`pushDir`)
  - memory file (`memoryPath`)
  - rules directory (`rulesDir`)
- Creating the agent seeds `scanDirs` with the same initial skill path.
- Additional scan paths are still added later on the agent card inside **Skill Paths**.
- The dialog updates the current in-memory Settings draft; final persistence still happens through **Save Settings**.

### Cloud Tab

| Control | Description |
|---------|-------------|
| **Provider buttons** | Responsive provider cards shown in a wrapping grid. **Git Repo** is always pinned first; the remaining providers keep the backend-returned order. Each provider restores its own saved bucket/path/credential draft when selected |
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
| **Appearance theme** | Four visual presets shown as preview cards: **Young** (default, a softened paper-blue evolution of the previous sky-blue Light palette), **Dark** (refined graphite with muted mist-blue accents), **Light** (a low-saturation gray-white palette inspired by Messor), and **Sport** (mint shell layers with field-green accents for a sharper athletic feel); persisted to `localStorage`; changes apply immediately without restart; legacy stored `Light` preference auto-migrates to `Young` |
| **App data directory** | Fixed root for installed skills, prompts, metadata, and backup working state on the current device; shown read-only with a one-click open action |
| **Repository cache directory** | Local-only root path for cloned starred repositories; manual text entry + folder-picker button that opens at the current path or nearest existing parent; defaults to `<AppDataDir>/cache/repos` |
| **Skill recursive scan depth** | Maximum recursion depth used when scanning local agent directories and starred repos; default `5`; saved values are clamped to `1-20` to avoid pathological nested trees |
| **Default category** | Fixed system fallback category `Default` (read-only), used when pulling/importing without specifying a category |
| **Log level buttons** | Toggle runtime log level between `debug`, `info`, and `error` (default: `error`); takes effect after saving settings |
| **Launch at login toggle** | Enables/disables OS login-item registration so SkillFlow auto-starts after sign-in on the current device; stored only in local `config_local.json`. Reconcile runs on startup and Settings save: already-missing disabled entries are treated as a no-op, while the enabled path is refreshed to the current executable so moved or updated app installs keep working on macOS and Windows |
| **Open log directory** | One-click open the local log folder in system file manager; missing targets fall back to the nearest existing parent directory |

Log files are stored under the app log directory, with rolling limits:
- At most **2 files** are kept: `skillflow.log` and `skillflow.log.1`.
- Each file is capped at **1MB**.
- When `skillflow.log` reaches the limit, it rotates and overwrites the older backup file.

### Proxy Tab

Proxy settings for all remote operations (repo scan, repo-cache sync):

| Mode | Description |
|------|-------------|
| **No proxy** | Direct connection |
| **System proxy** | Reads `HTTP_PROXY` / `HTTPS_PROXY` environment variables |
| **Manual** | Custom proxy URL (http://, https://, socks5://) |

When Manual is selected, a URL input appears with format hint. Proxy settings are persisted in `config_local.json`, so manual proxy values survive restart and are not included in backup/git sync.

- A dedicated **Test Connection** block lets users verify proxy reachability without saving first.
- The target URL input defaults to `https://github.com`, but users can replace it with any `http` / `https` URL.
- The test uses the current in-memory proxy form state, so unsaved edits are honored immediately; if nothing changed, that state naturally matches the saved config.
- Each test enforces a **5-second timeout** and shows inline feedback with target URL, elapsed time, and either the received HTTP status or the transport error message.
- Receiving an HTTP response still counts as successful connectivity even when the status code is not `2xx` (for example `301` or `403`).

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
| **Status strip** | Source badge plus any coexisting state badges (for example Update available + pushed-agent icons) rendered in one compact header row on the card; the strip prefers a single line when space allows, then automatically wraps instead of clipping badges away when cards become too narrow |
| **Skill name** | Two-line clamp; padded to avoid overlap with action buttons |
| **Open folder button** (FolderOpenDot, top-right) | `OpenPath(skill.path)` — opens directory in OS file manager; visible on hover only |
| **Select checkbox** (top-left) | Visible in select mode only |
| **Hover actions** (bottom-right) | Update (if available) · Copy · Delete — hidden until hover, except the Update action stays visible while that card is actively updating |
| **Update action feedback** | Hover adds a lift/highlight animation; clicking switches the action into a disabled spinner state until the update finishes |
| **Copy button** | Reads `skill.md` content, copies to clipboard, shows "Copied ✓" for 2 s |
| **Drag handle** | Cards are draggable in normal mode; dragged `skillId` moves skill to drop target category |
| **Right-click context menu** | Update (if available) · Move to [Category] (one item per other category) · Delete (red) |

**Sync card** (`SyncSkillCard`):

| Element | Description |
|---------|-------------|
| **Status strip** | Source badge plus only the page-selected state badges (for example imported, update-available, or pushed-agent icons) rendered together in one compact header row; cards keep imported and pushed-agent indicators on the same line when they fit, and automatically wrap them to a second row when the card width is too tight |
| **Pushed-agent indicator** | Shows the exact agents whose `PushDir` already contains this logical skill via small agent-brand icons without an extra arrow prefix; overflows collapse into a compact count badge while hover still reveals the full list |
| **Skill name** | Two-line clamp |
| **Subtitle** | Category or repo name |
| **Copy button** (hover) | Same clipboard behavior |
| **Open folder button** (hover) | Same as dashboard card |
| **Selection checkbox** (bottom-right) | Shown when `showSelection = true` |

### Unified Status Semantics

| State | Meaning |
|------|---------|
| **installed** | The logical skill already exists in **My Skills** as at least one installed instance |
| **imported** | External-page wording for **installed**; on GitHub / Starred Repos / agent views it means “already in My Skills” |
| **pushed** | The logical skill already exists in an agent's configured **PushDir** |
| **pushedAgents** | The exact agent names whose configured **PushDir** currently contains that logical skill; used for icon rendering on cards |
| **seenInAgentScan** | The logical skill was detected in one of an agent's configured **ScanDirs**; this means the agent already has it somewhere, but not necessarily because SkillFlow pushed it. This is currently used for grouping and correlation, not shown as a card badge. |
| **updatable** | An installed Git-backed skill has a newer remote SHA than its installed `SourceSHA` |

These statuses are resolved by the backend from a unified logical-key model; frontend pages no longer infer them independently from `Name` or `Path`.

```text
                 Unified skill status picture

[GitHub candidate] [Starred skill] [Agent scan candidate]
         \              |               /
          \             |              /
           +---- same logical skill ----+
                        |
                        v
                 [My Skills instance]
                  installed = true
           external pages show imported = true
                        |
                        +---- copy to agent PushDir ---> pushed = true
                        |
                        +---- remote newer SHA -------> updatable = true

[Agent ScanDirs] --- detect same logical skill ----> seenInAgentScan = true
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

### 10.2 Missing Directory Dialog

Appears before any push when target directories do not exist.

- Lists each affected agent name and full directory path.
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
| `update.available` | New cached commit found for a skill | `{ skillID, skillName, currentSHA, latestSHA }` |
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
| `downloading` | User clicks "Download & Auto-restart" (Windows only) | Spinner + progress message |
| `ready_to_restart` | `app.update.download.done` event | Completion message + "Restart Now" / "Later" buttons |
| `download_failed` | `app.update.download.fail` event | Error message + "Go to Downloads" button |

### Platform Behavior

Both startup and manual checks surface the same modal dialog.

- **Windows** — Three choices in the `available` state:
  1. **Download & Auto-restart** — downloads the new exe in the background, then prompts restart.
  2. **Open Release Page** — opens the GitHub Releases page in the system browser.
  3. **Skip this version (don't remind on next start)** — persists the skipped version; the startup check will not prompt for this version again. The manual check always shows the dialog regardless.
- **macOS** — Two choices in the `available` state (auto-download not supported):
  1. **Open Release Page** — opens the GitHub Releases page.
  2. **Skip this version (don't remind on next start)** — same skip behavior as Windows.
- The `available` dialog always renders the current version, latest version, and the Release-page action from the same `AppUpdateInfo` payload on both platforms.

### Skip Version Behavior

- The skipped version tag is stored in `AppConfig.SkippedUpdateVersion` and persisted in the shared `config.json` file, so it survives app restarts.
- On app startup, if `latestVersion == skippedUpdateVersion` the `app.update.available` event is **not** emitted and no dialog appears.
- When the user manually clicks "Check for Updates" in Settings, `CheckAppUpdateAndNotify` always emits the event, bypassing the skip — the dialog always appears for manual checks.
- Clicking "Skip this version" calls `SetSkippedUpdateVersion(latestVersion)`.

### Manual Check Button (Settings Page)

A **"Check for Updates"** button in the top-right corner of the Settings page header:

- Displays current app version (`vX.Y.Z`) next to the button.
- Click → calls `CheckAppUpdateAndNotify()`; button shows a spinner while checking.
- If a new version is found, the update dialog opens automatically via the `app.update.available` event.
- If already up-to-date, inline text shows "Already up to date (vX.Y.Z)".
- On error: "Check failed: …" shown inline.

### Controls

| Control | Action |
|---------|--------|
| **Download & Auto-restart** (Windows, `available`) | `DownloadAppUpdate(downloadUrl)` — starts async download |
| **Open Release Page** (`available`) | `OpenURL(releaseUrl)` — opens release page in browser; closes dialog |
| **Skip this version** (`available`) | `SetSkippedUpdateVersion(latestVersion)` — persists skip; closes dialog |
| **Restart Now** (`ready_to_restart`) | `ApplyAppUpdate()` — writes bat script and exits; bat replaces exe and relaunches |
| **Later** (`ready_to_restart`) | Closes dialog without restarting |
| **Go to Downloads** (`download_failed`) | `OpenURL(releaseUrl)` — opens release page in browser; closes dialog |
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

## 13. My Agents

Browse the skills currently present inside each enabled agent.

### Layout

- Left sidebar lists enabled agents.
- Main area shows one shared toolbar plus a top segmented control for the skills and memory surfaces.
- The segmented control defaults to **Skills** when the page first loads.
- In Chinese mode the segmented labels are localized instead of mixing English into the primary workflow.
- Switching to another agent keeps the current panel selection instead of forcing the page back to **Skills**.

### Toolbar

| Control | Description |
|---------|-------------|
| **Skills / Memory segmented control** | Switches the right-side content surface between skill management and memory preview |
| **Search input** | Filters only the currently active panel in real time |
| **Sort toggle** | Stays shared in the toolbar; in **Skills** it sorts Push Path and Scan Path cards, and in **Memory** it reorders the visible memory entries within that panel |
| **Batch Delete** | Available only in the **Skills** panel when the visible Push Path list is non-empty; enters select mode |
| **Manual Pull** | Available in the **Skills** panel; enters inline scan-and-pull mode for the currently selected agent |

The result counter in the toolbar is panel-scoped:
- **Skills** counts visible Push Path plus Scan Path cards only.
- **Memory** counts visible memory preview entries only.
- **Manual Pull mode** counts only the currently visible scanned candidates.

### Memory Panel

- Shown only when the top segmented control is switched to **Memory**.
- Shows the selected agent's currently configured main memory file and rules directory content directly from that agent's filesystem paths.
- Header includes a **Refresh** action so users can re-read the agent's current on-disk memory state without leaving **My Agents**.
- **Memory File** card shows the configured file path, an **Open File** action, and the current file content preview. If the path is missing or the file does not exist yet, the card keeps the configured path area visible and shows the corresponding empty state instead of failing the whole page.
- **Rules Directory** area shows the configured directory path, an **Open Directory** action, and one preview card per flat `.md` file found inside that directory.
- Rule cards mark `sf-*.md` files as **Managed** so users can distinguish SkillFlow-managed module memories from other agent-local rule files.
- The shared search input filters memory preview items only while **Memory** is active, and when nothing matches the panel shows a dedicated empty-search message instead of falling back to the skill empty states.

### Skills Panel

- Shown only when the top segmented control is switched to **Skills**.

### Manual Pull Mode

- Clicking **Manual Pull** does not open a separate page or modal. It converts the current skills panel into an inline pull surface for the selected agent.
- Entering the mode triggers a fresh `ScanAgentSkills(agentName)` scan against the selected agent's configured scan directories.
- The target category is chosen directly in the toolbar instead of a dedicated left category sidebar.
- While in manual-pull mode, the toolbar exposes:
  - target category selector
  - **Select All / Deselect All**
  - **Select Not Imported**
  - **Start Pull (n)**
  - **Cancel**
- Search and sorting apply only to the scanned results while this mode is active.
- Scan errors still surface as a red alert and empty scans still surface as a yellow warning state.
- Pull conflicts still use the shared overwrite/skip dialog behavior, but the resolution flow now completes inside **My Agents**.
- After a successful pull, the page exits manual-pull mode and shows a green completion notice at the top of the skills panel.

### Push Path Section

- Shows deletable agent-local skills under the configured push directory.
- Push-path discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- In select mode, **Select All / Deselect All** applies to the currently visible filtered Push Path cards only.
- When a agent-local skill correlates to an installed My Skills entry, the card also shows that installed entry's source badge so users can tell manual imports from git-backed installs.
- Cards continue showing imported, update-available, and "pushed to other agents" states. The current agent is excluded from the pushed-agent icon list so the card only surfaces cross-agent distribution that adds information.

### Scan Path Section

- Shows read-only skills discovered only from scan directories.
- Scan-path discovery uses the same configurable depth limit from **Settings → General** (default `5`, saved range `1-20`).
- Shares the same panel-scoped search and sort controls as Push Path.
- Scan-path cards use the same compact strip for source / imported / update-available / pushed-to-other-agents states whenever the scanned item can be correlated to an installed My Skills entry; the fact that they were found via scan paths is conveyed by the section itself instead of repeating a detected badge on every card.

---

## 14. My Prompts

Store reusable system prompts inside the synced `prompts/` directory.

### Navigation & Storage

- Sidebar adds **My Prompts** directly below **My Agents**.
- Each prompt is stored as `prompts/<category>/<name>/system.md` plus a sidecar `prompts/<category>/<name>/prompt.json` under the backup root, so both object-storage providers and the Git provider sync the same prompt files automatically.
- Prompt names are required, globally unique in the library, and used as the folder key.

### Layout

- The page mirrors **My Skills** with a left category sidebar and a right content pane.
- Categories support **Default** as the built-in fallback group.
- Just like the other primary routes, re-entering **My Prompts** remounts the page and reloads current prompt data instead of keeping a long-lived stale page instance.
- Prompt cards are filtered by selected category, keyword search, and A-Z / Z-A sorting.
- Search supports logical `and` / `or` syntax, for example `review and golang` or `summary or changelog`.

### Category Management

- Categories can be created from the sidebar.
- Non-default categories support rename and delete from the context menu.
- Prompt cards can be dragged onto categories to move them, matching the category-drop behavior from **My Skills**.

### Prompt Cards

- Cards reuse the same visual language as Skill cards (`card-base`, hover glow, compact actions).
- Each card shows **name**, optional **description**, and the opening content excerpt by default.
- Cards no longer show related image thumbnails or saved web-link chips in the list, keeping the prompt grid focused on text scanning.
- When the content is longer than the preview window, the card displays **Click to view more**.
- Top-right action button copies the full prompt content to the desktop clipboard in one click, preserving multi-line content on Windows.
- Prompt copy now falls back across desktop runtime, browser clipboard, and document copy APIs so the action still succeeds when one clipboard path is unavailable.
- Bottom-right **Delete** action opens a confirmation dialog before removing both the card and the underlying prompt folder.

### Prompt Editor

- Clicking **Add Prompt** opens a built-in editor with fields for **name** (required), **description** (optional), **category**, and full `system.md` content.
- The editor supports up to **3 related image URLs**. Saved images stay visible in the image panel, while the image add-input row moves to the shared attachment area after the main prompt content for cleaner reading flow.
- Each saved image thumbnail exposes a small delete action in the top-right corner.
- Clicking an editor thumbnail opens an enlarged in-app preview overlay above the editor instead of sending the user to the external browser.
- The editor now keeps the footer reachable inside smaller desktop windows by constraining the body and prompt content area with internal scrolling instead of growing beyond the app window.
- The editor supports **web links** through a single-line markdown input placed in the shared attachment area after the main prompt content. Clicking **Add** appends the parsed link above as a clickable chip using the markdown label text from `[Label](https://example.com/doc)`, then clears the input field.
- Clicking an existing prompt card opens the same editor pre-filled for editing and rename operations.
- Saving writes the prompt body back to `prompts/<category>/<name>/system.md` and writes metadata such as description, image URLs, and web links to `prompts/<category>/<name>/prompt.json`.

### Import / Export

- Toolbar **Import** reads a JSON prompt library file, preserves the imported category for each prompt, and when a prompt name already exists locally it asks whether to **Skip** or **Overwrite** before writing anything.
- The import conflict dialog includes **Apply the same action to the remaining {count} conflicts**, so one decision can be reused for the rest of the current import run without turning it into a saved global preference.
- Toolbar **Export** now expands an inline export bar instead of exporting immediately.
- The export bar supports **All**, the currently selected left-sidebar category, and **Pick**.
- **Pick** supports multi-select export within the current left-sidebar scope. When the sidebar is on **All**, selection spans the whole prompt library; when the sidebar is on a concrete category, selection is limited to that category.
- Exported JSON still preserves each prompt's own category, image URLs, and web links.

---

## 15. My Memory

SkillFlow provides a unified memory management surface for authoring personal AI coding assistant memories once and distributing them to multiple agent tools. The page manages both the shared main memory and reusable module memories, while the agent side can preview the exact synced result.

### Memory Types

- **Main Memory**: A single `main.md` file containing global instructions shared across all configured agents.
- **Module Memories**: Individual topic-focused markdown files (e.g., `coding-style`, `testing-rules`) stored under `rules/` and referenced from the main memory when pushed.

### Page Layout

The My Memory page shows:
- A top toolbar with **Search**, **New Module**, and **Batch Push**.
- A prominent **Main Memory card** with per-agent push status chips.
- A **two-column module grid** where every module card shows a content preview, a global enabled/disabled badge, an enable/disable action, and the module-ref hint.
- A top **Auto Sync** panel where each enabled agent can be set to `Off`, `Auto Merge`, or `Auto Takeover`.
- A **Batch Push** mode that turns the page into an inline multi-select flow, keeps main memory required, and lets users choose target agents plus one shared push mode for the current push.
- A right-side **Edit Drawer** anchored as a fixed overlay, widened to roughly 72% of the viewport up to a 960 px cap, with Edit / Preview tabs, save, delete, and open-in-editor actions only.
- A **New Module** dialog for creating module memories inline. Module names must match the lowercase `a-z`, `0-9`, `-` format used for exported agent rule files.

### Batch Push Flow

- Clicking **Batch Push** does not open a modal.
- The page enters an inline selection state:
  - the main-memory card is selected and required
  - each module card shows a checkbox in the top-right corner
  - the top panel switches from **Auto Sync** to **Push Target Agents**
  - users multi-select target agents
  - users choose one push mode for the whole operation: **Merge** or **Takeover**
- The push writes the selected module set to the selected agents and rebuilds main-memory module references from that same selection.
- Any previously SkillFlow-managed module files on the target agent that are not part of the current selection are removed during that push, so the agent rules directory matches the chosen module set.
- Explicit module references are rendered as one markdown link per line inside a dedicated `<skillflow-module>...</skillflow-module>` block, for example `[testing](rules/sf-testing.md)`.
- Those module links always use forward-slash relative paths from the agent main-memory file so synced configs stay portable across different machines.
- Because push status compares against the full current local memory set, a partial batch push keeps that agent in **Pending** until it matches the current main memory plus the full local module set again.

### Auto Sync Behaviour

- The **Auto Sync** panel is per agent and exposes three modes: `Off`, `Auto Merge`, and `Auto Takeover`.
- Enabling either auto-sync mode immediately pushes the current main memory plus all currently enabled module memories to that agent using the selected mode.
- After auto sync is enabled, editing the main memory, creating a module, saving a module, toggling a module enabled state, or deleting a module automatically syncs the change to every enabled auto-sync agent.
- Disabling a module removes it from the auto-sync baseline: SkillFlow deletes the managed `sf-<name>.md` file from each auto-sync agent and rewrites the main-memory module reference block to exclude it.
- When a module is deleted locally, SkillFlow also removes the corresponding managed `sf-<name>.md` file from each auto-sync agent and rewrites the main-memory module reference block to match.

### Editing Behaviour

- Closing the drawer with unsaved changes shows a confirmation with **Discard**, **Save and Close**, and **Keep Editing**.
- The preview tab renders Markdown instead of raw text.
- Each module drawer includes a delete action that opens a dedicated confirmation dialog with the module name and a short content preview before removal.

### Push Modes (per agent)

| Mode | Behaviour |
|------|-----------|
| **Merge** | SkillFlow writes the main memory inside `<skillflow-managed>...</skillflow-managed>` and writes explicit module refs in a separate `<skillflow-module>...</skillflow-module>` block. Content outside those SkillFlow-managed sections is preserved. Before writing, SkillFlow repairs incomplete managed tags so the next sync can rebuild a clean managed block. |
| **Takeover** | SkillFlow owns the entire memory file and rules directory for the managed memory content. |

### Push Status

Each agent shows one of three statuses:

| Status | Meaning |
|--------|---------|
| ✓ Synced | Last push matches current local main memory plus the current local enabled module set. |
| ⚠ Pending | Local memory has changed since the last full push, or the agent currently reflects only a partial module selection. |
| Never Pushed | The agent has never received a memory push from SkillFlow. |

### Module File Naming

Module memories are written with an `sf-` prefix to the agent rules directory (e.g., `coding-style` → `sf-coding-style.md`).

### Open in External Editor

Each memory drawer provides an **Open in Editor** button that opens the file in the system default text editor. SkillFlow detects file changes and refreshes the content automatically.

### Agent Settings

In **Settings → Agents**, each agent now exposes:
- **Skill Paths** – grouped block containing the push directory plus the editable multi-row scan directory list.
- **Memory Paths** – grouped block containing the main memory file and rules directory.
- **Memory File** – path to the agent's main memory file (default shown; user-overridable).
- **Rules Directory** – path to the agent's rules directory. Built-in agents without a first-party rules-directory concept, such as Copilot, keep this empty by default.

### Agent-side Verification

The **My Agents → Memory** panel lets users inspect the selected agent's current main memory file and flat rules-directory markdown files directly from disk, including `sf-*.md` files marked as **Managed**. This acts as the verification surface for memory auto sync and batch push results.

### Cloud Backup

Memory content files (`memory/main.md` and `memory/rules/*.md`) are included in cloud backup. The local configuration file (`memory/memory_local.json`) is excluded from backup.

---

*Last updated: 2026-04-17*

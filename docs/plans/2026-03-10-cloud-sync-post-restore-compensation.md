# Cloud Sync Post-Restore Compensation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Automatically compensate after cloud restore by auto-pushing newly restored skills to enabled auto-push tools and immediately cloning newly restored starred repositories.

**Architecture:** Capture a lightweight pre-restore snapshot, then reuse a single post-restore reconciliation path from both startup git pull and manual restore. The reconciliation step reloads disk state, detects newly restored skills and newly restored starred repos, auto-pushes only the newly added skills, and clones only the newly added starred repos so search and direct push are immediately usable.

**Tech Stack:** Go 1.23, Wails desktop backend, local filesystem storage, git-based starred repo cache, Go tests with `testify`

---

### Task 1: Add failing restore-compensation tests

**Files:**
- Create: `cmd/skillflow/app_restore_test.go`
- Reuse: `cmd/skillflow/app_autopush_test.go`

**Step 1: Write the failing test**

- Add one test that snapshots app state, simulates a cloud-restored skill by writing it into storage directly, runs the new post-restore reconciliation method, and expects the selected tool push dir to receive the skill.
- Add one test that snapshots app state, simulates a cloud-restored starred repo by saving it into `star_repos.json`, runs the reconciliation method, and expects the repo clone directory to exist.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/skillflow -run 'TestHandleRestoredCloudState'`

Expected: FAIL because the reconciliation helpers do not exist yet.

### Task 2: Implement post-restore reconciliation

**Files:**
- Modify: `cmd/skillflow/app.go`

**Step 1: Write minimal implementation**

- Add a small restore snapshot type for installed-skill identity keys and starred repo URLs present before restore.
- Add a helper to build the snapshot before restore.
- Add a helper to run after restore success that:
  - reloads app state from disk
  - loads current skills and auto-pushes only those missing from the pre-restore snapshot
  - loads current starred repos and clones only repos whose URLs were absent from the pre-restore snapshot
  - records clone `LastSync` / `SyncError` updates back into star storage

**Step 2: Reuse it from both restore entrypoints**

- Capture the snapshot before startup git restore and before manual `RestoreFromCloud()`.
- Replace the direct `reloadStateFromDisk()` calls with the shared reconciliation helper.

**Step 3: Run tests to verify they pass**

Run: `go test ./cmd/skillflow -run 'TestHandleRestoredCloudState'`

Expected: PASS

### Task 3: Verify no regression in existing auto-push behavior

**Files:**
- Reuse: `cmd/skillflow/app_autopush_test.go`

**Step 1: Run focused backend tests**

Run: `go test ./cmd/skillflow -run 'Test(ImportLocalAutoPushesToSelectedTools|ImportLocalAutoPushSkipsExistingToolSkill|SaveConfigAutoPushesExistingSkillsToNewTool|HandleRestoredCloudState)'`

Expected: PASS

### Task 4: Update product docs for the new behavior

**Files:**
- Modify: `docs/features.md`
- Modify: `docs/features_zh.md`
- Modify: `README.md`
- Modify: `README_zh.md`

**Step 1: Document the behavior**

- Update backup/sync and auto-push descriptions to state that newly restored skills are auto-pushed on the receiving machine.
- Update starred repo descriptions to state that newly restored repos are immediately cloned locally after cloud sync.
- Refresh the last-updated footer in both feature docs.

**Step 2: Run verification**

Run: `go test ./cmd/skillflow`

Expected: PASS or a clear unrelated failure to report.

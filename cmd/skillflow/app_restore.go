package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/skill"
)

type cloudRestoreState struct {
	installedSkillKeys map[string]struct{}
	starredRepoURLs    map[string]struct{}
}

func (a *App) captureCloudRestoreState() cloudRestoreState {
	state := cloudRestoreState{
		installedSkillKeys: map[string]struct{}{},
		starredRepoURLs:    map[string]struct{}{},
	}

	if a.storage != nil {
		if skills, err := a.storage.ListAll(); err == nil {
			for _, sk := range skills {
				if key := cloudRestoreSkillKey(sk); key != "" {
					state.installedSkillKeys[key] = struct{}{}
				}
			}
		}
	}

	if a.starStorage != nil {
		if repos, err := a.starStorage.Load(); err == nil {
			for _, repo := range repos {
				if key := cloudRestoreRepoKey(repo.URL); key != "" {
					state.starredRepoURLs[key] = struct{}{}
				}
			}
		}
	}

	return state
}

func (a *App) handleRestoredCloudState(before cloudRestoreState, source string) error {
	a.logInfof("restore compensation started: source=%s", source)
	a.reloadStateFromDisk()

	newlyRestoredSkills, err := a.restoredSkillsSince(before)
	if err != nil {
		a.logErrorf("restore compensation failed: source=%s load restored skills failed: %v", source, err)
		return err
	}
	if len(newlyRestoredSkills) > 0 {
		a.autoPushImportedSkills(source, newlyRestoredSkills)
	}

	clonedRepos, failedRepos, err := a.cloneNewlyRestoredStarredRepos(before, source)
	if err != nil {
		a.logErrorf("restore compensation failed: source=%s clone restored starred repos failed: %v", source, err)
		return err
	}

	a.logInfof("restore compensation completed: source=%s restoredSkills=%d clonedRepos=%d failedRepos=%d", source, len(newlyRestoredSkills), clonedRepos, failedRepos)
	return nil
}

func (a *App) restoredSkillsSince(before cloudRestoreState) ([]*skill.Skill, error) {
	if a.storage == nil {
		return nil, nil
	}

	skills, err := a.storage.ListAll()
	if err != nil {
		return nil, err
	}
	restored := make([]*skill.Skill, 0, len(skills))
	for _, sk := range skills {
		key := cloudRestoreSkillKey(sk)
		if key == "" {
			continue
		}
		if _, existed := before.installedSkillKeys[key]; existed {
			continue
		}
		restored = append(restored, sk)
	}
	return restored, nil
}

func (a *App) cloneNewlyRestoredStarredRepos(before cloudRestoreState, source string) (int, int, error) {
	if a.starStorage == nil {
		return 0, 0, nil
	}

	repos, err := a.starStorage.Load()
	if err != nil {
		return 0, 0, err
	}

	cloned := 0
	failed := 0
	changed := false
	for i := range repos {
		repoKey := cloudRestoreRepoKey(repos[i].URL)
		if repoKey == "" {
			continue
		}
		if _, existed := before.starredRepoURLs[repoKey]; existed {
			continue
		}
		if hasClonedRepo(repos[i].LocalDir) {
			continue
		}

		a.logInfof("restore starred repo clone started: source=%s repo=%s localDir=%s", source, repos[i].URL, repos[i].LocalDir)
		if err := coregit.CloneOrUpdate(a.cloneContext(), repos[i].URL, repos[i].LocalDir, a.gitProxyURL()); err != nil {
			failed++
			changed = true
			repos[i].SyncError = err.Error()
			a.logErrorf("restore starred repo clone failed: source=%s repo=%s localDir=%s err=%v", source, repos[i].URL, repos[i].LocalDir, err)
			continue
		}

		cloned++
		changed = true
		repos[i].SyncError = ""
		repos[i].LastSync = time.Now()
		a.logInfof("restore starred repo clone completed: source=%s repo=%s localDir=%s", source, repos[i].URL, repos[i].LocalDir)
	}

	if changed {
		if err := a.starStorage.Save(repos); err != nil {
			return cloned, failed, err
		}
	}
	return cloned, failed, nil
}

func cloudRestoreSkillKey(sk *skill.Skill) string {
	if sk == nil {
		return ""
	}
	if logicalKey, err := skill.LogicalKey(sk); err == nil && strings.TrimSpace(logicalKey) != "" {
		return "logical:" + logicalKey
	}
	if strings.TrimSpace(sk.ID) == "" {
		return ""
	}
	return "instance:" + strings.TrimSpace(sk.ID)
}

func cloudRestoreRepoKey(repoURL string) string {
	if normalized, err := coregit.CanonicalRepoURL(repoURL); err == nil && strings.TrimSpace(normalized) != "" {
		return normalized
	}
	return strings.TrimSpace(repoURL)
}

func hasClonedRepo(localDir string) bool {
	if strings.TrimSpace(localDir) == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(localDir, ".git")); err == nil {
		return true
	}
	return false
}

func (a *App) cloneContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

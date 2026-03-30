package discovery

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

const defaultMaxRecursiveScanDepth = 5

type RepoScanner struct{}

func NewRepoScanner() *RepoScanner {
	return &RepoScanner{}
}

func scanTree(repoDir, dir, repoURL, repoName, source string, depth, maxDepth int) ([]sourcedomain.SourceSkillCandidate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var result []sourcedomain.SourceSkillCandidate
	if hasSkillMd(entries) {
		rel, err := filepath.Rel(repoDir, dir)
		if err != nil {
			return nil, err
		}
		result = append(result, sourcedomain.SourceSkillCandidate{
			Name:     filepath.Base(dir),
			Path:     dir,
			SubPath:  filepath.ToSlash(rel),
			RepoURL:  repoURL,
			RepoName: repoName,
			Source:   source,
		})
	}
	if depth >= maxDepth {
		return result, nil
	}

	for _, e := range entries {
		if !e.IsDir() || shouldSkipScanDir(e.Name()) {
			continue
		}
		skills, err := scanTree(repoDir, filepath.Join(dir, e.Name()), repoURL, repoName, source, depth+1, maxDepth)
		if err != nil {
			return nil, err
		}
		result = append(result, skills...)
	}
	return result, nil
}

func hasSkillMd(entries []os.DirEntry) bool {
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(e.Name(), "skill.md") {
			return true
		}
	}
	return false
}

func shouldSkipScanDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (RepoScanner) ScanSkills(repoDir, repoURL, repoName, source string) ([]sourcedomain.SourceSkillCandidate, error) {
	return RepoScanner{}.ScanSkillsWithMaxDepth(repoDir, repoURL, repoName, source, defaultMaxRecursiveScanDepth)
}

func (RepoScanner) ScanSkillsWithMaxDepth(repoDir, repoURL, repoName, source string, maxDepth int) ([]sourcedomain.SourceSkillCandidate, error) {
	if maxDepth < 0 {
		maxDepth = 0
	}
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if hasSkillMd(entries) {
		return scanTree(repoDir, repoDir, repoURL, repoName, source, 0, maxDepth)
	}

	var roots []string
	skillsRoot := filepath.Join(repoDir, "skills")
	if info, err := os.Stat(skillsRoot); err == nil && info.IsDir() {
		roots = append(roots, skillsRoot)
	}
	for _, e := range entries {
		if !e.IsDir() || shouldSkipScanDir(e.Name()) {
			continue
		}
		dir := filepath.Join(repoDir, e.Name())
		if dir == skillsRoot {
			continue
		}
		roots = append(roots, dir)
	}

	var result []sourcedomain.SourceSkillCandidate
	for _, root := range roots {
		skills, err := scanTree(repoDir, root, repoURL, repoName, source, 0, maxDepth)
		if err != nil {
			return nil, err
		}
		result = append(result, skills...)
	}
	return result, nil
}

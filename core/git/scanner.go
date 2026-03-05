package git

import (
	"errors"
	"os"
	"path/filepath"
)

// scanDir scans a single directory for subdirs containing SKILLS.md.
// subPathPrefix is prepended to the subdir name to form SubPath (e.g. "skills/").
func scanDir(root, subPathPrefix, repoURL, repoName string) ([]StarSkill, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var result []StarSkill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillDir := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(skillDir, "SKILLS.md")); err == nil {
			result = append(result, StarSkill{
				Name:     e.Name(),
				Path:     skillDir,
				SubPath:  subPathPrefix + e.Name(),
				RepoURL:  repoURL,
				RepoName: repoName,
			})
		}
	}
	return result, nil
}

// ScanSkills looks for skill directories (subdirs containing SKILLS.md) in the
// given repo clone. It first checks <repoDir>/skills/; if no skills are found
// there, it falls back to scanning <repoDir>/ directly (for repos whose root IS
// the skills collection, e.g. github.com/anthropics/skills).
func ScanSkills(repoDir, repoURL, repoName string) ([]StarSkill, error) {
	// Try <repoDir>/skills/ first.
	result, err := scanDir(filepath.Join(repoDir, "skills"), "skills/", repoURL, repoName)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result, nil
	}

	// Fallback: scan repo root directly.
	return scanDir(repoDir, "", repoURL, repoName)
}

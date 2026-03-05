package git

import (
	"errors"
	"os"
	"path/filepath"
)

// ScanSkills walks <repoDir>/skills/ and returns entries that contain a SKILLS.md file.
func ScanSkills(repoDir, repoURL, repoName string) ([]StarSkill, error) {
	skillsRoot := filepath.Join(repoDir, "skills")
	entries, err := os.ReadDir(skillsRoot)
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
		skillDir := filepath.Join(skillsRoot, e.Name())
		if _, err := os.Stat(filepath.Join(skillDir, "SKILLS.md")); err == nil {
			result = append(result, StarSkill{
				Name:     e.Name(),
				Path:     skillDir,
				SubPath:  "skills/" + e.Name(),
				RepoURL:  repoURL,
				RepoName: repoName,
			})
		}
	}
	return result, nil
}

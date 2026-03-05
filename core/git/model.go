package git

import "time"

type StarredRepo struct {
	URL       string    `json:"url"`
	Name      string    `json:"name"`     // "owner/repo"
	LocalDir  string    `json:"localDir"` // absolute path under cache/
	LastSync  time.Time `json:"lastSync"`
	SyncError string    `json:"syncError,omitempty"`
}

type StarSkill struct {
	Name     string `json:"name"`
	Path     string `json:"path"`     // absolute local path to skill directory
	SubPath  string `json:"subPath"`  // relative path within repo, e.g. "skills/my-skill"
	RepoURL  string `json:"repoUrl"`
	RepoName string `json:"repoName"` // "owner/repo"
	Imported bool   `json:"imported"` // already exists in My Skills
}

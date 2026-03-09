package git

import "time"

type StarredRepo struct {
	URL       string    `json:"url"`
	Name      string    `json:"name"`     // e.g. "owner/repo" or "group/subgroup/repo"
	Source    string    `json:"source"`   // canonical source key: "<host>/<repo-path>"
	LocalDir  string    `json:"localDir"` // absolute path under cache/
	Manual    bool      `json:"manual,omitempty"`
	Starred   bool      `json:"starred,omitempty"`
	LastSync  time.Time `json:"lastSync"`
	SyncError string    `json:"syncError,omitempty"`
}

type StarSkill struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`    // absolute local path to skill directory
	SubPath     string   `json:"subPath"` // relative path within repo, e.g. "skills/my-skill"
	RepoURL     string   `json:"repoUrl"`
	RepoName    string   `json:"repoName"`   // e.g. "owner/repo"
	Source      string   `json:"source"`     // canonical source key: "<host>/<repo-path>"
	LogicalKey  string   `json:"logicalKey"` // stable cross-module identity
	Installed   bool     `json:"installed"`  // already exists in My Skills
	Imported    bool     `json:"imported"`   // external-source wording alias for installed
	Updatable   bool     `json:"updatable"`
	Pushed      bool     `json:"pushed"`
	PushedTools []string `json:"pushedTools"`
}

package notify

type EventType string

const (
	EventBackupStarted   EventType = "backup.started"
	EventBackupProgress  EventType = "backup.progress"
	EventBackupCompleted EventType = "backup.completed"
	EventBackupFailed    EventType = "backup.failed"
	EventSyncCompleted   EventType = "sync.completed"
	EventUpdateAvailable EventType = "update.available"
	EventSkillConflict   EventType = "skill.conflict"
	EventStarSyncProgress EventType = "star.sync.progress" // one repo finished syncing
	EventStarSyncDone     EventType = "star.sync.done"      // all repos finished

	EventGitSyncStarted   EventType = "git.sync.started"
	EventGitSyncCompleted EventType = "git.sync.completed"
	EventGitSyncFailed    EventType = "git.sync.failed"
	EventGitConflict      EventType = "git.conflict" // local ↔ remote conflict requires user decision
)

type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}

type BackupProgressPayload struct {
	FilesTotal    int    `json:"filesTotal"`
	FilesUploaded int    `json:"filesUploaded"`
	CurrentFile   string `json:"currentFile"`
}

type UpdateAvailablePayload struct {
	SkillID    string `json:"skillId"`
	SkillName  string `json:"skillName"`
	CurrentSHA string `json:"currentSha"`
	LatestSHA  string `json:"latestSha"`
}

type ConflictPayload struct {
	SkillName  string `json:"skillName"`
	TargetPath string `json:"targetPath"`
}

type StarSyncProgressPayload struct {
	RepoURL   string `json:"repoUrl"`
	RepoName  string `json:"repoName"`
	SyncError string `json:"syncError,omitempty"`
}

type GitConflictPayload struct {
	Message string `json:"message"`
}

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

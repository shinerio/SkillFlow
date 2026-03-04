package skill

import "time"

type SourceType string

const (
	SourceGitHub SourceType = "github"
	SourceManual SourceType = "manual"
)

type Skill struct {
	ID            string
	Name          string
	Path          string
	Category      string
	Source        SourceType
	SourceURL     string
	SourceSubPath string
	SourceSHA     string
	LatestSHA     string
	InstalledAt   time.Time
	UpdatedAt     time.Time
	LastCheckedAt time.Time
}

func (s *Skill) IsGitHub() bool { return s.Source == SourceGitHub }
func (s *Skill) IsManual() bool { return s.Source == SourceManual }
func (s *Skill) HasUpdate() bool {
	return s.IsGitHub() && s.LatestSHA != "" && s.LatestSHA != s.SourceSHA
}

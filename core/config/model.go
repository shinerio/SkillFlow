package config

type ToolConfig struct {
	Name      string `json:"name"`
	SkillsDir string `json:"skillsDir"`
	Enabled   bool   `json:"enabled"`
	Custom    bool   `json:"custom"`
}

type CloudConfig struct {
	Provider    string            `json:"provider"`
	Enabled     bool              `json:"enabled"`
	BucketName  string            `json:"bucketName"`
	RemotePath  string            `json:"remotePath"`
	Credentials map[string]string `json:"credentials"`
}

type AppConfig struct {
	SkillsStorageDir string       `json:"skillsStorageDir"`
	DefaultCategory  string       `json:"defaultCategory"`
	Tools            []ToolConfig `json:"tools"`
	Cloud            CloudConfig  `json:"cloud"`
}

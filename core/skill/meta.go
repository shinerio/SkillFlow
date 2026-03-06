package skill

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMeta holds parsed YAML frontmatter from a SKILL.md file.
// All fields are optional — missing fields remain zero values.
type SkillMeta struct {
	Name                   string `yaml:"name"                     json:"Name"`
	Description            string `yaml:"description"              json:"Description"`
	ArgumentHint           string `yaml:"argument-hint"            json:"ArgumentHint"`
	AllowedTools           string `yaml:"allowed-tools"            json:"AllowedTools"`
	Context                string `yaml:"context"                  json:"Context"`
	DisableModelInvocation bool   `yaml:"disable-model-invocation" json:"DisableModelInvocation"`
}

// ReadMeta locates skill.md / skill.md (case-insensitive) inside skillPath,
// extracts the YAML frontmatter block (between the first pair of --- delimiters),
// and unmarshals it. Returns an empty SkillMeta when no frontmatter is present.
func ReadMeta(skillPath string) (*SkillMeta, error) {
	entries, err := os.ReadDir(skillPath)
	if err != nil {
		return &SkillMeta{}, nil
	}

	var mdPath string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if lower == "skill.md" {
			mdPath = filepath.Join(skillPath, e.Name())
			break
		}
	}
	if mdPath == "" {
		return &SkillMeta{}, nil
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		return &SkillMeta{}, nil
	}

	frontmatter := extractFrontmatter(string(data))
	if frontmatter == "" {
		return &SkillMeta{}, nil
	}

	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return &SkillMeta{}, nil
	}
	return &meta, nil
}

// extractFrontmatter returns the YAML content between the first pair of "---" lines.
// Returns empty string if no valid frontmatter block is found.
func extractFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	var end int
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == 0 {
		return ""
	}
	return strings.Join(lines[1:end], "\n")
}

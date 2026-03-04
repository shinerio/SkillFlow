package skill

import (
	"errors"
	"os"
	"strings"
)

var ErrNoSKILLSmd = errors.New("skill.md not found in skill directory")

// ValidationRule is the extension point for future complex validators.
type ValidationRule func(dir string) error

type Validator struct {
	rules []ValidationRule
}

func NewValidator(extraRules ...ValidationRule) *Validator {
	rules := []ValidationRule{requireSkillMd}
	return &Validator{rules: append(rules, extraRules...)}
}

func (v *Validator) Validate(dir string) error {
	for _, rule := range v.rules {
		if err := rule(dir); err != nil {
			return err
		}
	}
	return nil
}

// requireSkillMd accepts any casing of "skill.md" or "skills.md".
func requireSkillMd(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if lower == "skill.md" || lower == "skills.md" {
			return nil
		}
	}
	return ErrNoSKILLSmd
}

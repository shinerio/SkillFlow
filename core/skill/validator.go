package skill

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNoSKILLSmd = errors.New("SKILLS.md not found in skill directory")

// ValidationRule is the extension point for future complex validators.
type ValidationRule func(dir string) error

type Validator struct {
	rules []ValidationRule
}

func NewValidator(extraRules ...ValidationRule) *Validator {
	rules := []ValidationRule{requireSKILLSmd}
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

func requireSKILLSmd(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return err
	}
	mdPath := filepath.Join(dir, "SKILLS.md")
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		return ErrNoSKILLSmd
	}
	return nil
}

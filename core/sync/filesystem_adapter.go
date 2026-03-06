package sync

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/shinerio/skillflow/core/skill"
)

// FilesystemAdapter works for all tools — they all share the same file-based skills directory model.
type FilesystemAdapter struct {
	name             string
	defaultSkillsDir string
}

func NewFilesystemAdapter(name, defaultSkillsDir string) *FilesystemAdapter {
	return &FilesystemAdapter{name: name, defaultSkillsDir: defaultSkillsDir}
}

func (f *FilesystemAdapter) Name() string             { return f.name }
func (f *FilesystemAdapter) DefaultSkillsDir() string { return f.defaultSkillsDir }

func (f *FilesystemAdapter) Push(_ context.Context, skills []*skill.Skill, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	for _, sk := range skills {
		dst := filepath.Join(targetDir, sk.Name)
		if err := copyDir(sk.Path, dst); err != nil {
			return err
		}
	}
	return nil
}

func (f *FilesystemAdapter) Pull(_ context.Context, sourceDir string) ([]*skill.Skill, error) {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", sourceDir)
	}
	var skills []*skill.Skill
	var walk func(dir string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		// Check if this directory itself contains a skill.md file.
		for _, e := range entries {
			if !e.IsDir() && isSkillMd(e.Name()) {
				skills = append(skills, &skill.Skill{
					Name:   filepath.Base(dir),
					Path:   dir,
					Source: skill.SourceManual,
				})
				return // found skill here — don't recurse deeper
			}
		}
		// No skill.md found — recurse into subdirectories.
		for _, e := range entries {
			if e.IsDir() {
				walk(filepath.Join(dir, e.Name()))
			}
		}
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			walk(filepath.Join(sourceDir, e.Name()))
		}
	}
	return skills, nil
}

func isSkillMd(name string) bool {
	lower := strings.ToLower(name)
	return lower == "skill.md"
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

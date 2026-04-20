package gateway

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/config"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type FilesystemAdapter struct {
	name             string
	defaultSkillsDir string
	codexConfigPath  string
}

func NewFilesystemAdapter(name, defaultSkillsDir string) *FilesystemAdapter {
	adapter := &FilesystemAdapter{name: name, defaultSkillsDir: defaultSkillsDir}
	if name == "codex" {
		home, _ := os.UserHomeDir()
		adapter.codexConfigPath = filepath.Join(home, ".codex", "config.toml")
	}
	return adapter
}

func (f *FilesystemAdapter) TestSetCodexConfigPath(path string) {
	f.codexConfigPath = path
}

func (f *FilesystemAdapter) Name() string             { return f.name }
func (f *FilesystemAdapter) DefaultSkillsDir() string { return f.defaultSkillsDir }

func (f *FilesystemAdapter) Push(_ context.Context, skills []*skilldomain.InstalledSkill, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	for _, skill := range skills {
		dst := filepath.Join(targetDir, skill.Name)
		if err := copyDir(skill.Path, dst); err != nil {
			return err
		}
	}
	return nil
}

func (f *FilesystemAdapter) Pull(ctx context.Context, sourceDir string) ([]*skilldomain.InstalledSkill, error) {
	return f.PullWithMaxDepth(ctx, sourceDir, config.DefaultRepoScanMaxDepth)
}

func (f *FilesystemAdapter) PullWithMaxDepth(_ context.Context, sourceDir string, maxDepth int) ([]*skilldomain.InstalledSkill, error) {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", sourceDir)
	}
	if maxDepth < 0 {
		maxDepth = 0
	}
	var skills []*skilldomain.InstalledSkill
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() && isSkillMd(entry.Name()) {
				skills = append(skills, &skilldomain.InstalledSkill{
					Name:   filepath.Base(dir),
					Path:   dir,
					Source: skilldomain.SourceManual,
				})
				return
			}
		}
		if depth >= maxDepth {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				walk(filepath.Join(dir, entry.Name()), depth+1)
			}
		}
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			walk(filepath.Join(sourceDir, entry.Name()), 0)
		}
	}
	return skills, nil
}

func (f *FilesystemAdapter) ApplySkillEnablement(profile agentdomain.AgentProfile, skills []agentdomain.ManagedAgentSkill) error {
	if f.name != "codex" {
		return nil
	}
	configPath := f.codexConfigPath
	if strings.TrimSpace(configPath) == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".codex", "config.toml")
	}
	return applyCodexSkillEnablement(configPath, skills)
}

func applyCodexSkillEnablement(configPath string, skills []agentdomain.ManagedAgentSkill) error {
	managedPaths := map[string]bool{}
	managedAliases := map[string]struct{}{}
	for _, skill := range skills {
		for _, path := range skill.Paths {
			rawPath := strings.TrimSpace(path)
			if rawPath == "" {
				continue
			}
			managedAliases[rawPath] = struct{}{}
			normalizedPath := normalizeCodexSkillConfigPath(rawPath)
			if normalizedPath != "" {
				managedAliases[normalizedPath] = struct{}{}
				managedPaths[normalizedPath] = skill.Enabled
			}
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	segments := parseCodexConfigSegments(string(data))
	existingEntries := collectCodexSkillConfigEntries(segments)
	finalEntries := mergeCodexSkillConfigEntries(existingEntries, managedAliases, managedPaths)
	filtered := make([]codexConfigSegment, 0, len(segments))
	insertedSkillConfig := false
	for _, segment := range segments {
		if segment.kind == codexConfigSegmentSkillsConfig {
			if !insertedSkillConfig && len(finalEntries) > 0 {
				filtered = append(filtered, codexConfigSegment{
					kind:  codexConfigSegmentSkillsConfig,
					lines: renderCodexSkillConfigBlocks(finalEntries),
				})
				insertedSkillConfig = true
			}
			continue
		}
		filtered = append(filtered, segment)
	}
	if !insertedSkillConfig && len(finalEntries) > 0 {
		filtered = append(filtered, codexConfigSegment{
			kind:  codexConfigSegmentSkillsConfig,
			lines: renderCodexSkillConfigBlocks(finalEntries),
		})
	}

	var output []string
	for index, segment := range filtered {
		output = append(output, segment.lines...)
		if index < len(filtered)-1 && len(segment.lines) > 0 && strings.TrimSpace(segment.lines[len(segment.lines)-1]) != "" {
			output = append(output, "")
		}
	}
	content := strings.TrimRight(strings.Join(output, "\n"), "\n")
	if len(output) > 0 {
		content += "\n"
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(configPath, []byte(content), 0o644)
}

type codexConfigSegmentKind string

const (
	codexConfigSegmentRaw          codexConfigSegmentKind = "raw"
	codexConfigSegmentSkillsConfig codexConfigSegmentKind = "skills.config"
)

type codexConfigSegment struct {
	kind    codexConfigSegmentKind
	entries []codexSkillConfigEntry
	lines   []string
}

func parseCodexConfigSegments(content string) []codexConfigSegment {
	if content == "" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var segments []codexConfigSegment
	var raw []string
	flushRaw := func() {
		if len(raw) == 0 {
			return
		}
		segments = append(segments, codexConfigSegment{
			kind:  codexConfigSegmentRaw,
			lines: append([]string(nil), raw...),
		})
		raw = nil
	}
	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "[[skills.config]]" {
			flushRaw()
			block := []string{lines[i]}
			i++
			for i < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i])
				if strings.HasPrefix(nextTrimmed, "[[") || (strings.HasPrefix(nextTrimmed, "[") && strings.HasSuffix(nextTrimmed, "]")) {
					break
				}
				block = append(block, lines[i])
				i++
			}
			segments = append(segments, codexConfigSegment{
				kind:    codexConfigSegmentSkillsConfig,
				entries: parseCodexInlineTableSkillConfigEntries(block),
				lines:   block,
			})
			continue
		}
		if isCodexSkillConfigArrayStart(trimmed) {
			flushRaw()
			block, nextIndex := readCodexSkillConfigArrayBlock(lines, i)
			segments = append(segments, codexConfigSegment{
				kind:    codexConfigSegmentSkillsConfig,
				entries: parseCodexArraySkillConfigEntries(block),
				lines:   block,
			})
			i = nextIndex
			continue
		}
		raw = append(raw, lines[i])
		i++
	}
	flushRaw()
	return segments
}

func isCodexSkillConfigArrayStart(trimmed string) bool {
	if !strings.HasPrefix(trimmed, "skills.config") {
		return false
	}
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) != 2 {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(parts[1]), "[")
}

func readCodexSkillConfigArrayBlock(lines []string, start int) ([]string, int) {
	block := make([]string, 0, 4)
	depth := 0
	for index := start; index < len(lines); index++ {
		line := lines[index]
		block = append(block, line)
		depth += strings.Count(line, "[")
		depth -= strings.Count(line, "]")
		if depth <= 0 {
			return block, index + 1
		}
	}
	return block, len(lines)
}

type codexSkillConfigEntry struct {
	Path    string
	Enabled bool
}

func parseCodexInlineTableSkillConfigEntries(lines []string) []codexSkillConfigEntry {
	path := parseCodexSkillConfigPath(lines)
	if path == "" {
		return nil
	}
	return []codexSkillConfigEntry{{
		Path:    path,
		Enabled: parseCodexSkillConfigEnabled(lines),
	}}
}

func parseCodexArraySkillConfigEntries(lines []string) []codexSkillConfigEntry {
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "skills.config"))
		content = strings.TrimSpace(strings.TrimPrefix(content, "="))
		if len(lines) > 1 {
			content = strings.Join(append([]string{content}, lines[1:]...), "\n")
		}
	}
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")
	if strings.TrimSpace(content) == "" {
		return nil
	}

	parts := splitCodexArrayInlineTables(content)
	entries := make([]codexSkillConfigEntry, 0, len(parts))
	for _, part := range parts {
		path := parseCodexInlineFieldValue(part, "path")
		if path == "" {
			continue
		}
		entries = append(entries, codexSkillConfigEntry{
			Path:    path,
			Enabled: parseCodexInlineFieldBool(part, "enabled"),
		})
	}
	return entries
}

func parseCodexSkillConfigPath(lines []string) string {
	for _, line := range lines {
		if value := parseCodexInlineFieldValue(line, "path"); value != "" {
			return value
		}
	}
	return ""
}

func parseCodexSkillConfigEnabled(lines []string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, "enabled") {
			return parseCodexInlineFieldBool(line, "enabled")
		}
	}
	return false
}

func parseCodexInlineFieldValue(content, field string) string {
	parts := strings.Split(content, ",")
	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			continue
		}
		if strings.TrimSpace(strings.Trim(keyValue[0], "{} ")) != field {
			continue
		}
		value := strings.TrimSpace(strings.Trim(keyValue[1], "{} "))
		if unquoted, err := strconv.Unquote(value); err == nil {
			return unquoted
		}
		return strings.Trim(value, `"'`)
	}
	return ""
}

func parseCodexInlineFieldBool(content, field string) bool {
	parts := strings.Split(content, ",")
	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			continue
		}
		if strings.TrimSpace(strings.Trim(keyValue[0], "{} ")) != field {
			continue
		}
		return strings.EqualFold(strings.TrimSpace(strings.Trim(keyValue[1], "{} ")), "true")
	}
	return false
}

func splitCodexArrayInlineTables(content string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	for _, r := range content {
		current.WriteRune(r)
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				part := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(current.String()), ","))
				if part != "" {
					parts = append(parts, part)
				}
				current.Reset()
			}
		}
	}
	return parts
}

func collectCodexSkillConfigEntries(segments []codexConfigSegment) []codexSkillConfigEntry {
	var entries []codexSkillConfigEntry
	for _, segment := range segments {
		if segment.kind != codexConfigSegmentSkillsConfig {
			continue
		}
		entries = append(entries, segment.entries...)
	}
	return entries
}

func mergeCodexSkillConfigEntries(existing []codexSkillConfigEntry, managedAliases map[string]struct{}, managedPaths map[string]bool) []codexSkillConfigEntry {
	merged := make([]codexSkillConfigEntry, 0, len(existing)+len(managedPaths))
	seen := make(map[string]struct{}, len(existing)+len(managedPaths))
	for _, entry := range existing {
		path := strings.TrimSpace(entry.Path)
		if path == "" {
			continue
		}
		if _, ok := managedAliases[path]; ok {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		merged = append(merged, codexSkillConfigEntry{
			Path:    path,
			Enabled: entry.Enabled,
		})
	}

	disabledPaths := make([]string, 0, len(managedPaths))
	for path, enabled := range managedPaths {
		if enabled {
			continue
		}
		disabledPaths = append(disabledPaths, path)
	}
	sort.Strings(disabledPaths)
	for _, path := range disabledPaths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		merged = append(merged, codexSkillConfigEntry{
			Path:    path,
			Enabled: false,
		})
	}
	return merged
}

func renderCodexSkillConfigBlocks(entries []codexSkillConfigEntry) []string {
	if len(entries) == 0 {
		return nil
	}
	lines := make([]string, 0, len(entries)*4)
	for index, entry := range entries {
		lines = append(lines,
			"[[skills.config]]",
			"path = "+strconv.Quote(entry.Path),
			"enabled = "+strconv.FormatBool(entry.Enabled),
		)
		if index < len(entries)-1 {
			lines = append(lines, "")
		}
	}
	return lines
}

func normalizeCodexSkillConfigPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return resolveCodexSkillMarkdownPath(path)
		}
		return path
	}
	if isSkillMd(filepath.Base(path)) {
		return path
	}
	return filepath.Join(path, "SKILL.md")
}

func resolveCodexSkillMarkdownPath(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return filepath.Join(dir, "SKILL.md")
	}
	for _, preferred := range []string{"SKILL.md", "skill.md", "Skill.md"} {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if entry.Name() == preferred {
				return filepath.Join(dir, entry.Name())
			}
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if isSkillMd(entry.Name()) {
			return filepath.Join(dir, entry.Name())
		}
	}
	return filepath.Join(dir, "SKILL.md")
}

func isSkillMd(name string) bool {
	return strings.ToLower(name) == "skill.md"
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

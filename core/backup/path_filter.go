package backup

import (
	"path/filepath"
	"strings"
)

// excludedDirs lists directory names excluded from backup at any path depth.
// A path matches if it equals the dir name or starts with "<dir>/".
var excludedDirs = []string{
	"cache",
	"logs",
	".git",
}

// excludedFiles lists file base names excluded from backup wherever they appear.
// A path matches if its base name equals the entry.
var excludedFiles = []string{
	".DS_Store",
	"config_local.json",
}

// ShouldSkipBackupPath reports whether a relative path should be excluded from backup sync/list/restore.
// Rules are shared across cloud providers to keep backup content consistent with git backup behavior.
func ShouldSkipBackupPath(rel string) bool {
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel)))
	if normalized == "." || normalized == "" {
		return false
	}
	for _, dir := range excludedDirs {
		if normalized == dir || strings.HasPrefix(normalized, dir+"/") {
			return true
		}
	}
	base := normalized[strings.LastIndex(normalized, "/")+1:]
	for _, file := range excludedFiles {
		if base == file {
			return true
		}
	}
	return false
}

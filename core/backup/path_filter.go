package backup

import (
	"path/filepath"
	"strings"
)

// ShouldSkipBackupPath reports whether a relative path should be excluded from backup sync/list/restore.
// Rules are shared across cloud providers to keep backup content consistent with git backup behavior.
func ShouldSkipBackupPath(rel string) bool {
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel)))
	switch {
	case normalized == ".", normalized == "":
		return false
	case normalized == "cache" || strings.HasPrefix(normalized, "cache/"):
		return true
	case normalized == ".git" || strings.HasPrefix(normalized, ".git/"):
		return true
	default:
		return false
	}
}

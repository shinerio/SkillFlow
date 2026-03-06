package applog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevelFilter(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir, "info")
	require.NoError(t, err)

	lg.Debugf("debug line")
	lg.Infof("info line")

	data, err := os.ReadFile(filepath.Join(dir, activeLogName))
	require.NoError(t, err)
	text := string(data)
	assert.NotContains(t, text, "debug line")
	assert.Contains(t, text, "info line")
}

func TestRotateWithTwoFilesOnly(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir, "debug")
	require.NoError(t, err)
	lg.maxBytes = 128

	for i := 0; i < 30; i++ {
		lg.Infof("line %d abcdefghijklmnopqrstuvwxyz", i)
	}

	matches, err := filepath.Glob(filepath.Join(dir, "skillflow.log*"))
	require.NoError(t, err)
	assert.LessOrEqual(t, len(matches), 2)
	assert.FileExists(t, filepath.Join(dir, activeLogName))
	assert.FileExists(t, filepath.Join(dir, backupLogName))
}

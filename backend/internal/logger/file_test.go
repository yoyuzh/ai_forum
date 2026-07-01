package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aiconfig "ai-forum/backend/internal/config"
)

// TestNewWithFile_WritesToRotatedFile asserts that when a FilePath is
// configured, log entries are written to that file via lumberjack.
func TestNewWithFile_WritesToRotatedFile(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	logPath := filepath.Join(dir, "forum.log")
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	fc := FileConfig{FilePath: logPath, MaxSizeMB: 1, MaxBackups: 1, MaxAgeDays: 1}

	// Act
	l, err := NewWithFile(cfg, fc)
	require.NoError(t, err)
	l.Info("file output works")
	require.NoError(t, l.Sync())

	// Assert
	data, err := os.ReadFile(logPath)
	require.NoError(t, err, "log file must be created")
	content := string(data)
	assert.Contains(t, content, "file output works")
	// JSON encoding to file.
	assert.Contains(t, content, `"msg":"file output works"`)
}

// TestNewWithFile_NoFilePathSucceeds asserts empty FilePath does not error
// (rotation disabled path).
func TestNewWithFile_NoFilePathSucceeds(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	l, err := NewWithFile(cfg, FileConfig{})
	require.NoError(t, err)
	require.NotNil(t, l)
}

// TestNewWithFile_RedactionStillWorks asserts redaction composes with file
// output — the token must never reach the rotated file.
func TestNewWithFile_RedactionStillWorks(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	logPath := filepath.Join(dir, "forum.log")
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	fc := FileConfig{FilePath: logPath, MaxSizeMB: 1, MaxBackups: 1, MaxAgeDays: 1}
	rawToken := strings.Repeat("cd", 32) // 64 hex chars

	// Act
	base, err := NewWithFile(cfg, fc)
	require.NoError(t, err)
	l := base.WithRedaction("token")
	l.Info("auth", FieldToken(rawToken))
	require.NoError(t, l.Sync())

	// Assert
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "***")
	assert.NotContains(t, content, rawToken, "raw token must not reach the file")
}

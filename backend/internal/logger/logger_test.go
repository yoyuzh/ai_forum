package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	aiconfig "ai-forum/backend/internal/config"
)

// captureLogger builds a zap.Logger writing to an in-memory buffer with the
// given config, so tests can assert on emitted output.
func captureLogger(t *testing.T, cfg aiconfig.LogConfig) (*Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	// NewWithWriter lets us capture output instead of writing to stderr.
	l, err := NewWithWriter(cfg, buf)
	require.NoError(t, err)
	return l, buf
}

// parseJSONLines parses newline-delimited JSON log entries from buf.
func parseJSONLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	scanner := bufio.NewScanner(buf)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m), "log line must be valid JSON: %s", line)
		entries = append(entries, m)
	}
	require.NoError(t, scanner.Err())
	return entries
}

// TestNew_ProductionJSONEncoding asserts JSON-encoded output containing
// message and level fields.
func TestNew_ProductionJSONEncoding(t *testing.T) {
	// Arrange
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "info", Encoding: "json"})

	// Act
	l.Info("hello world")

	// Assert
	entries := parseJSONLines(t, buf)
	require.Len(t, entries, 1)
	assert.Equal(t, "hello world", entries[0]["msg"])
	// zap's JSON encoder emits level as a lowercase string by default.
	assert.Equal(t, "info", entries[0]["level"])
}

// TestNew_ConsoleEncodingWhenNotJSON asserts non-json encoding uses the
// console encoder (human-readable, not pure JSON).
func TestNew_ConsoleEncodingWhenNotJSON(t *testing.T) {
	// Arrange
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "info", Encoding: "console"})

	// Act
	l.Info("hello console")

	// Assert: console output is NOT valid JSON (it's tab-separated human text).
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "hello console")
	err := json.Unmarshal([]byte(out), &map[string]any{})
	assert.Error(t, err, "console encoding must not produce JSON")
}

// TestWith_ChildLoggerCarriesFields asserts a child logger binds contextual
// fields that appear on every subsequent entry.
func TestWith_ChildLoggerCarriesFields(t *testing.T) {
	// Arrange
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "info", Encoding: "json"})

	// Act: create a child with user_id and request_id, then log twice.
	child := l.With(
		FieldUserID(12),
		FieldRequestID("r1"),
	)
	child.Info("first")
	child.Info("second")

	// Assert: both entries carry both bound fields.
	entries := parseJSONLines(t, buf)
	require.Len(t, entries, 2)
	for _, e := range entries {
		// zap encodes integer fields as JSON numbers.
		assert.EqualValues(t, 12, e["user_id"], "user_id must be bound on every entry")
		assert.Equal(t, "r1", e["request_id"], "request_id must be bound on every entry")
	}
}

// TestWith_AllContextualFields asserts every documented contextual field
// helper exists and is attachable.
func TestWith_AllContextualFields(t *testing.T) {
	// Arrange
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "info", Encoding: "json"})

	// Act
	child := l.With(
		FieldEventID("e1"),
		FieldTaskID("t1"),
		FieldUserID(1),
		FieldRequestID("r1"),
		FieldPostID(2),
		FieldCommentID(3),
		FieldAIAgentID("a1"),
		FieldTriggerType("manual"),
	)
	child.Info("all fields")

	// Assert
	entries := parseJSONLines(t, buf)
	require.Len(t, entries, 1)
	assert.Equal(t, "e1", entries[0]["event_id"])
	assert.Equal(t, "t1", entries[0]["task_id"])
	assert.EqualValues(t, 1, entries[0]["user_id"])
	assert.Equal(t, "r1", entries[0]["request_id"])
	assert.EqualValues(t, 2, entries[0]["post_id"])
	assert.EqualValues(t, 3, entries[0]["comment_id"])
	assert.Equal(t, "a1", entries[0]["ai_agent_id"])
	assert.Equal(t, "manual", entries[0]["trigger_type"])
}

// TestRedact_TokenFieldMasked asserts a redaction-enabled logger masks the
// token field to *** and never emits the raw value.
func TestRedact_TokenFieldMasked(t *testing.T) {
	// Arrange: a 64-char hex token value that must never appear in logs.
	rawToken := strings.Repeat("ab", 32) // 64 hex chars
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "info", Encoding: "json"})
	l = l.WithRedaction("token", "password", "secret", "api_key")

	// Act
	l.Info("auth event", FieldToken(rawToken))

	// Assert
	out := buf.String()
	assert.Contains(t, out, "***", "token field must be masked to ***")
	assert.NotContains(t, out, rawToken, "raw token value must never appear in output")
}

// TestRedact_NonRedactedFieldsPassThrough asserts non-redacted fields are
// emitted unchanged.
func TestRedact_NonRedactedFieldsPassThrough(t *testing.T) {
	// Arrange
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "info", Encoding: "json"})
	l = l.WithRedaction("token")

	// Act
	l.Info("event", FieldUserID(42), FieldRequestID("req-9"))

	// Assert
	entries := parseJSONLines(t, buf)
	require.Len(t, entries, 1)
	assert.EqualValues(t, 42, entries[0]["user_id"])
	assert.Equal(t, "req-9", entries[0]["request_id"])
}

// TestNew_LevelRespected asserts a level setting filters lower-priority logs.
func TestNew_LevelRespected(t *testing.T) {
	// Arrange: level=warn should suppress info logs.
	l, buf := captureLogger(t, aiconfig.LogConfig{Level: "warn", Encoding: "json"})

	// Act
	l.Info("should be suppressed")
	l.Warn("should appear")

	// Assert
	entries := parseJSONLines(t, buf)
	require.Len(t, entries, 1, "info should be filtered at warn level")
	assert.Equal(t, "should appear", entries[0]["msg"])
}

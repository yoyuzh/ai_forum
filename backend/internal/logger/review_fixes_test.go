package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	aiconfig "ai-forum/backend/internal/config"
)

// captureStderr swaps os.Stderr for the duration of fn and returns whatever was
// written to it. Used to assert the default logger sink is a real stream, not
// io.Discard.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()
	require.NoError(t, w.Close())

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

// TestNew_DefaultSinkIsNotDiscard asserts New() writes to a real sink (stderr),
// not io.Discard. A logger that silently drops every line is an operational
// footgun (review HIGH finding).
func TestNew_DefaultSinkIsNotDiscard(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}

	out := captureStderr(t, func() {
		l, err := New(cfg)
		require.NoError(t, err)
		l.Info("must reach a real sink")
	})

	assert.Contains(t, out, "must reach a real sink",
		"New() must emit to stderr (or another real sink), not io.Discard")
}

// TestNewWithFile_EmptyPathEmitsToStderr asserts NewWithFile with no FilePath
// still emits to stderr (not io.Discard).
func TestNewWithFile_EmptyPathEmitsToStderr(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}

	out := captureStderr(t, func() {
		l, err := NewWithFile(cfg, FileConfig{})
		require.NoError(t, err)
		l.Info("file-disabled still logs")
	})

	assert.Contains(t, out, "file-disabled still logs")
}

// TestNew_RedactionOnByDefault asserts that a fresh logger from New() already
// redacts known secret field names (token, password, secret, api_key), so the
// "INTERNAL_API_TOKEN must never appear in logs" guarantee holds even when a
// caller forgets WithRedaction (review HIGH finding).
func TestNew_RedactionOnByDefault(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	rawToken := strings.Repeat("ab", 32) // 64 hex chars

	var buf bytes.Buffer
	l, err := NewWithWriter(cfg, &buf)
	require.NoError(t, err)

	// No WithRedaction call — redaction must already be active.
	l.Info("auth", FieldToken(rawToken))

	out := buf.String()
	assert.Contains(t, out, "***", "token must be redacted by default")
	assert.NotContains(t, out, rawToken, "raw token must never appear even without WithRedaction")
}

// TestNew_DefaultRedactsAllKnownSecretKeys asserts every documented secret
// field name is redacted by default.
func TestNew_DefaultRedactsAllKnownSecretKeys(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	var buf bytes.Buffer
	l, err := NewWithWriter(cfg, &buf)
	require.NoError(t, err)

	l.Info("secrets",
		FieldToken("tok-val"),
		zap.String("password", "pw-val"),
		zap.String("secret", "sec-val"),
		zap.String("api_key", "key-val"),
	)

	out := buf.String()
	for _, v := range []string{"tok-val", "pw-val", "sec-val", "key-val"} {
		assert.NotContains(t, out, v, "secret value %q must be redacted by default", v)
	}
}

// TestWithThenWithRedaction_PreservesFields asserts that binding contextual
// fields via With and THEN enabling redaction keeps the earlier fields. This
// is the review MEDIUM finding: previously WithRedaction rebuilt from the bare
// core, dropping With-bound fields like request_id.
func TestWithThenWithRedaction_PreservesFields(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	var buf bytes.Buffer
	l, err := NewWithWriter(cfg, &buf)
	require.NoError(t, err)

	// With first, then WithRedaction.
	child := l.With(FieldRequestID("r1"), FieldUserID(7)).
		WithRedaction("token")
	child.Info("event", FieldToken("supersecret"))

	entries := parseJSONLinesHelper(t, &buf)
	require.Len(t, entries, 1)
	assert.Equal(t, "r1", entries[0]["request_id"], "With-bound request_id must survive WithRedaction")
	assert.EqualValues(t, 7, entries[0]["user_id"], "With-bound user_id must survive WithRedaction")
	assert.Equal(t, "***", entries[0]["token"])
}

// TestWithRedactionThenWith_PreservesFields asserts the reverse order also
// keeps fields and keeps redaction active.
func TestWithRedactionThenWith_PreservesFields(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	var buf bytes.Buffer
	l, err := NewWithWriter(cfg, &buf)
	require.NoError(t, err)

	child := l.WithRedaction("token").
		With(FieldRequestID("r2"))
	child.Info("event", FieldToken("supersecret2"))

	entries := parseJSONLinesHelper(t, &buf)
	require.Len(t, entries, 1)
	assert.Equal(t, "r2", entries[0]["request_id"])
	assert.Equal(t, "***", entries[0]["token"])
}

// TestNew_UnknownLevelErrorWrapped asserts an unknown log level returns an
// error wrapped with constructor context (review LOW finding). The raw
// parseLevel error is "unknown log level %q"; the constructor must add a
// "failed to build logger" prefix so callers get actionable context.
func TestNew_UnknownLevelErrorWrapped(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "bogus", Encoding: "json"}
	_, err := New(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus", "error should mention the bad level value")
	assert.Contains(t, err.Error(), "failed to build logger",
		"constructor must wrap parseLevel error with builder context")
}

// --- helpers ---

// parseJSONLinesHelper parses newline-delimited JSON from buf.
func parseJSONLinesHelper(t *testing.T, buf *bytes.Buffer) []map[string]any {
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

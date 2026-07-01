package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	aiconfig "ai-forum/backend/internal/config"
)

// TestDefaultRedactKeys_CoversAliasedTokenNames asserts the default redact set
// covers the field names a developer is likely to use for the project's most
// sensitive secrets, not just the bare "token". This closes the review MEDIUM
// finding: zap.String("internal_api_token", val) (the viper config key is
// internal_api.token) previously passed through unredacted because exact-key
// matching only caught "token".
func TestDefaultRedactKeys_CoversAliasedTokenNames(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	var buf strings.Builder
	l, err := NewWithWriter(cfg, &buf)
	require.NoError(t, err)

	rawInternalToken := strings.Repeat("ab", 32)
	rawAuth := "Bearer " + strings.Repeat("c", 40)

	// No WithRedaction call — defaults must already cover these aliased keys.
	l.Info("request",
		zap.String("internal_api_token", rawInternalToken),
		zap.String("authorization", rawAuth),
	)

	out := buf.String()
	assert.NotContains(t, out, rawInternalToken, "internal_api_token must be redacted by default")
	assert.NotContains(t, out, rawAuth, "authorization must be redacted by default")
	assert.Contains(t, out, "***")
}

// TestFieldInternalAPI_RedactedByDefault asserts the FieldInternalAPI helper
// exists and its field is redacted by the default set.
func TestFieldInternalAPI_RedactedByDefault(t *testing.T) {
	cfg := aiconfig.LogConfig{Level: "info", Encoding: "json"}
	var buf strings.Builder
	l, err := NewWithWriter(cfg, &buf)
	require.NoError(t, err)

	rawToken := strings.Repeat("de", 32)
	l.Info("internal call", FieldInternalAPI(rawToken))

	out := buf.String()
	assert.NotContains(t, out, rawToken)
	assert.Contains(t, out, "***")
}

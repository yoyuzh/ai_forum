package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidate_UnresolvedPlaceholderFails asserts that a secret left as a
// literal "${VAR}" placeholder (because the env var was never set) is rejected
// by Validate, not silently accepted as a real secret. This is the fail-fast
// enforcement point for design D4 — startup must NOT succeed with placeholder
// strings standing in for secrets.
func TestValidate_UnresolvedPlaceholderFails(t *testing.T) {
	// Arrange: secrets carry unresolved ${VAR} placeholders, as Viper leaves
	// them when the env var is absent.
	cfg := &Config{
		Server:     ServerConfig{Mode: "release"},
		JWT:        JWTConfig{Secret: "${JWT_SECRET}"},
		InternalAPI: InternalAPIConfig{Token: "${INTERNAL_API_TOKEN}"},
		MySQL:      MySQLConfig{Password: "${MYSQL_PASSWORD}"},
	}

	// Act
	err := Validate(cfg)

	// Assert: every placeholder secret is reported.
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "jwt.secret", "unresolved jwt.secret placeholder must be reported")
	assert.Contains(t, msg, "internal_api.token", "unresolved internal_api.token placeholder must be reported")
	assert.Contains(t, msg, "mysql.password", "unresolved mysql.password placeholder must be reported")
}

// TestValidate_RealSecretValuePasses asserts a non-empty, non-placeholder
// secret value is accepted — the placeholder check must not over-reject.
func TestValidate_RealSecretValuePasses(t *testing.T) {
	cfg := &Config{
		Server:     ServerConfig{Mode: "release"},
		JWT:        JWTConfig{Secret: "a-real-jwt-secret-value"},
		InternalAPI: InternalAPIConfig{Token: "a-real-internal-token-value"},
		MySQL:      MySQLConfig{Password: "a-real-db-password"},
		AI:         AIConfig{APIKey: "a-real-ai-key"},
	}

	err := Validate(cfg)
	assert.NoError(t, err)
}

// TestValidate_PlaceholderRelaxedInDebug asserts that in debug mode, an
// unresolved MySQL.Password placeholder is allowed (matching the existing
// debug relaxation for empty passwords), while JWT/InternalAPI placeholders
// are still rejected (they are always required).
func TestValidate_PlaceholderRelaxedInDebug(t *testing.T) {
	cfg := &Config{
		Server:     ServerConfig{Mode: "debug"},
		JWT:        JWTConfig{Secret: "${JWT_SECRET}"},
		InternalAPI: InternalAPIConfig{Token: "${INTERNAL_API_TOKEN}"},
		MySQL:      MySQLConfig{Password: "${MYSQL_PASSWORD}"}, // relaxed in debug
	}

	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt.secret")
	assert.Contains(t, err.Error(), "internal_api.token")
	assert.NotContains(t, err.Error(), "mysql.password", "debug mode should relax mysql.password placeholder")
}

// TestValidate_AIKeyRequiredInRelease asserts AI.APIKey is a required secret in
// non-debug mode (the AI client must not start with an empty/unresolved key
// and fail at request time — design D4 fail-fast). Mirrors the MySQL.Password
// relaxation: in debug mode the key is optional for local dev.
func TestValidate_AIKeyRequiredInRelease(t *testing.T) {
	// Arrange: release mode, all other secrets set, AI.APIKey unresolved placeholder.
	cfg := &Config{
		Server:     ServerConfig{Mode: "release"},
		JWT:        JWTConfig{Secret: "real-jwt"},
		InternalAPI: InternalAPIConfig{Token: "real-token"},
		MySQL:      MySQLConfig{Password: "real-pw"},
		AI:         AIConfig{APIKey: "${AI_API_KEY}"}, // unresolved placeholder
	}

	// Act
	err := Validate(cfg)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ai.api_key",
		"release mode must fail fast on missing/unresolved AI.APIKey")
}

// TestValidate_AIKeyRelaxedInDebug asserts debug mode relaxes the AI.APIKey
// requirement (local dev without an AI key), while JWT/InternalAPI stay required.
func TestValidate_AIKeyRelaxedInDebug(t *testing.T) {
	cfg := &Config{
		Server:     ServerConfig{Mode: "debug"},
		JWT:        JWTConfig{Secret: "real-jwt"},
		InternalAPI: InternalAPIConfig{Token: "real-token"},
		// AI.APIKey intentionally empty — relaxed in debug.
	}

	err := Validate(cfg)
	assert.NoError(t, err, "debug mode should not require ai.api_key")
}

// TestValidate_AllSecretsPresentPasses asserts a fully-populated release config
// passes, including AI.APIKey — the new check must not over-reject.
func TestValidate_AllSecretsPresentPasses(t *testing.T) {
	cfg := &Config{
		Server:     ServerConfig{Mode: "release"},
		JWT:        JWTConfig{Secret: "real-jwt"},
		InternalAPI: InternalAPIConfig{Token: "real-token"},
		MySQL:      MySQLConfig{Password: "real-pw"},
		AI:         AIConfig{APIKey: "real-ai-key"},
	}

	err := Validate(cfg)
	assert.NoError(t, err)
}



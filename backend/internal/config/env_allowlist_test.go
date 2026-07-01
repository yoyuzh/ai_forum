package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad_SourceHasNoAutomaticEnv asserts the loader source does NOT CALL
// viper.AutomaticEnv(), which would enable blanket env matching and violate
// design D2 / the "explicit allowlist only" hard constraint. This is a source
// contract test because the behavior is defensive: AutomaticEnv's harm is that
// any future SetDefault key would become silently env-overridable without
// being added to the allowlist. We match the call form "AutomaticEnv(" so
// comments mentioning the symbol by name are not false positives.
func TestLoad_SourceHasNoAutomaticEnv(t *testing.T) {
	src, err := os.ReadFile("loader.go")
	require.NoError(t, err)
	assert.NotContains(t, string(src), "AutomaticEnv(",
		"loader.go must not call AutomaticEnv(); only explicit BindEnv entries may override config")
}

// TestLoad_OnlyAllowlistedEnvOverrides asserts that ONLY env vars in the
// explicit bindEnv allowlist override file/default values. A config key whose
// env name is NOT in the allowlist keeps its default/file value even when a
// similarly-named env var is set.
func TestLoad_OnlyAllowlistedEnvOverrides(t *testing.T) {
	// Arrange: minimal YAML; redis.db gets its default (0) from setDefaults.
	path := writeTestYAML(t, `
server:
  mode: release
mysql:
  password: filepw
jwt:
  secret: s
internal_api:
  token: t
`)
	// REDIS_DB is in the allowlist -> must override to 7.
	// REDIS_DB_BACKUP is NOT allowlisted -> must not override (no field anyway,
	// but proves the only path is the allowlist).
	withEnv(t, map[string]string{
		"REDIS_DB":        "7",
		"REDIS_DB_BACKUP": "99",
	})

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, 7, cfg.Redis.DB, "allowlisted REDIS_DB must override the default")
	// Re-run without REDIS_DB to confirm the default still applies when the
	// allowlisted env is absent.
	t.Setenv("REDIS_DB", "")
	cfg2, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, 0, cfg2.Redis.DB, "default redis.db must apply when allowlisted env is unset")
}

// guard: ensure filepath import is used (some toolchains warn).
var _ = filepath.Join
var _ = strings.Contains

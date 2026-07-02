package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aiconfig "ai-forum/backend/internal/config"
	ailogger "ai-forum/backend/internal/logger"
)

// TestLoadDevConfigAndInstantiateLogger is the P0 task-5.2 end-to-end check:
// it loads the committed config/config.dev.yaml (with env vars resolving the
// ${VAR} placeholders) and instantiates a real zap logger, asserting concrete
// config values and the logger level/encoding, without panicking.
func TestLoadDevConfigAndInstantiateLogger(t *testing.T) {
	// Arrange: path to the committed dev config (test runs from internal/config).
	devPath := filepath.Join("..", "..", "config", "config.dev.yaml")
	t.Setenv("MYSQL_PASSWORD", "dev-db-pw")
	t.Setenv("JWT_SECRET", "dev-jwt-secret")
	t.Setenv("INTERNAL_API_TOKEN", "dev-internal-token")
	t.Setenv("AI_API_KEY", "dev-ai-key")

	// Act: load + validate.
	cfg, err := aiconfig.Load(devPath)
	require.NoError(t, err)
	require.NoError(t, aiconfig.Validate(cfg))

	// Assert specific config values.
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, "mysql", cfg.MySQL.Host)
	assert.Equal(t, "ai_forum", cfg.MySQL.Database)
	assert.Equal(t, "dev-db-pw", cfg.MySQL.Password, "env must override ${MYSQL_PASSWORD}")
	assert.Equal(t, "dev-jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, "dev-internal-token", cfg.InternalAPI.Token)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Encoding)

	// Instantiate a real logger from the loaded config — must not panic.
	logger, err := ailogger.New(cfg.Log)
	require.NoError(t, err)
	require.NotNil(t, logger)
	// Smoke-test: emit one line at the configured level (goes to io.Discard
	// since New uses the default sink; this only asserts no panic).
	logger.Info("p0 foundation ready",
		ailogger.FieldUserID(1),
		ailogger.FieldRequestID("p0-smoke"),
	)
}

func TestLoadESAddressesFromEnvCSV(t *testing.T) {
	devPath := filepath.Join("..", "..", "config", "config.dev.yaml")
	t.Setenv("ES_ADDRESSES", "http://es1:9200,http://es2:9200")

	cfg, err := aiconfig.Load(devPath)

	require.NoError(t, err)
	assert.Equal(t, []string{"http://es1:9200", "http://es2:9200"}, cfg.Elasticsearch.Addresses)
}

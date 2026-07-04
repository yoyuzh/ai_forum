package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestYAML writes a YAML config file to a temp dir and returns its path.
func writeTestYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

// withEnv sets env vars for the duration of the test and restores them after.
func withEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	original := map[string]string{}
	for k, v := range vars {
		original[k] = os.Getenv(k)
		require.NoError(t, os.Setenv(k, v))
	}
	t.Cleanup(func() {
		for k, v := range original {
			require.NoError(t, os.Setenv(k, v))
		}
	})
}

// minimalYAML is a config with all required secrets set as placeholders so Load
// produces a fully-populated Config (after env resolution).
const minimalYAML = `
server:
  port: 9090
  mode: release
mysql:
  host: db
  port: 3306
  username: root
  password: filevalue
  database: ai_forum
redis:
  addr: redis:6379
  password: ""
  db: 0
rabbitmq:
  url: amqp://guest:guest@rabbitmq:5672/
elasticsearch:
  addresses:
    - http://elasticsearch:9200
jwt:
  secret: ${JWT_SECRET}
  expire_hours: 168
internal_api:
  token: ${INTERNAL_API_TOKEN}
ai:
  provider: openai
  base_url: https://api.openai.com
  model: gpt-4o-mini
  api_key: ${AI_API_KEY}
  max_concurrency: 4
  request_per_second: 2
  burst: 2
worker:
  ai_reply_concurrency: 4
  tagging_concurrency: 2
  search_index_concurrency: 2
  notification_concurrency: 4
hot_score:
  refresh_interval_seconds: 30
  batch_size: 200
log:
  level: info
  encoding: json
`

// TestLoad_DefaultsAppliedWhenAbsent asserts that when neither file nor env
// sets server.port, the default 8080 is used.
func TestLoad_DefaultsAppliedWhenAbsent(t *testing.T) {
	// Arrange: a YAML with server section removed.
	yaml := `
server:
  mode: release
jwt:
  secret: s
internal_api:
  token: t
`
	path := writeTestYAML(t, yaml)
	// Ensure MYSQL_PASSWORD is unset so Validate won't trip on it.
	withEnv(t, map[string]string{"MYSQL_PASSWORD": "p"})

	// Act
	cfg, err := Load(path)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port, "default server.port should be 8080")
}

// TestLoad_EnvOverridesFileValue asserts env > file precedence for an allowlisted key.
func TestLoad_EnvOverridesFileValue(t *testing.T) {
	// Arrange: YAML sets mysql.password=filevalue; env sets MYSQL_PASSWORD=envvalue.
	path := writeTestYAML(t, minimalYAML)
	withEnv(t, map[string]string{
		"MYSQL_PASSWORD":     "envvalue",
		"JWT_SECRET":         "jwtval",
		"INTERNAL_API_TOKEN": "tokenval",
	})

	// Act
	cfg, err := Load(path)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "envvalue", cfg.MySQL.Password, "env must override file value")
}

// TestLoad_DevConfigParsesAllFields asserts the architecture §14.2 shape fully populates.
func TestLoad_DevConfigParsesAllFields(t *testing.T) {
	// Arrange
	path := writeTestYAML(t, minimalYAML)
	withEnv(t, map[string]string{
		"MYSQL_PASSWORD":     "mp",
		"JWT_SECRET":         "js",
		"INTERNAL_API_TOKEN": "it",
		"AI_BASE_URL":        "https://api.iamhc.cn",
		"AI_API_KEY":         "ak",
	})

	// Act
	cfg, err := Load(path)

	// Assert every field from the spec is non-zero.
	require.NoError(t, err)
	assert.Equal(t, "db", cfg.MySQL.Host)
	assert.Equal(t, 3306, cfg.MySQL.Port)
	assert.Equal(t, "ai_forum", cfg.MySQL.Database)
	assert.Equal(t, "redis:6379", cfg.Redis.Addr)
	assert.Equal(t, "amqp://guest:guest@rabbitmq:5672/", cfg.RabbitMQ.URL)
	require.Len(t, cfg.Elasticsearch.Addresses, 1)
	assert.Equal(t, "http://elasticsearch:9200", cfg.Elasticsearch.Addresses[0])
	assert.Equal(t, 168, cfg.JWT.ExpireHours)
	assert.Equal(t, "openai", cfg.AI.Provider)
	assert.Equal(t, "https://api.iamhc.cn", cfg.AI.BaseURL)
	assert.Equal(t, "gpt-4o-mini", cfg.AI.Model)
	assert.Equal(t, 4, cfg.AI.MaxConcurrency)
	assert.Equal(t, 2, cfg.AI.RequestPerSecond)
	assert.Equal(t, 2, cfg.AI.Burst)
	assert.Equal(t, 4, cfg.Worker.AiReplyConcurrency)
	assert.Equal(t, 2, cfg.Worker.TaggingConcurrency)
	assert.Equal(t, 2, cfg.Worker.SearchIndexConcurrency)
	assert.Equal(t, 4, cfg.Worker.NotificationConcurrency)
	assert.Equal(t, 30, cfg.HotScore.RefreshIntervalSeconds)
	assert.Equal(t, 200, cfg.HotScore.BatchSize)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Encoding)
}

// TestValidate_MissingRequiredSecretsFailsLoud asserts Validate aggregates all
// missing required secrets into one error message.
func TestValidate_MissingRequiredSecretsFailsLoud(t *testing.T) {
	// Arrange: both JWT.Secret and InternalAPI.Token empty.
	cfg := &Config{
		Server: ServerConfig{Mode: "release"},
	}

	// Act
	err := Validate(cfg)

	// Assert: error mentions both keys.
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "jwt.secret", "error must list missing jwt.secret")
	assert.Contains(t, msg, "internal_api.token", "error must list missing internal_api.token")
}

// TestValidate_ReleaseModeRequiresDBPassword asserts non-debug mode fails on empty MySQL.Password.
func TestValidate_ReleaseModeRequiresDBPassword(t *testing.T) {
	cfg := &Config{
		Server:      ServerConfig{Mode: "release"},
		JWT:         JWTConfig{Secret: "s"},
		InternalAPI: InternalAPIConfig{Token: "t"},
		// MySQL.Password intentionally empty
	}

	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mysql.password")
}

// TestValidate_DebugModeRelaxesDBPassword asserts debug mode allows empty MySQL.Password.
func TestValidate_DebugModeRelaxesDBPassword(t *testing.T) {
	cfg := &Config{
		Server:      ServerConfig{Mode: "debug"},
		JWT:         JWTConfig{Secret: "s"},
		InternalAPI: InternalAPIConfig{Token: "t"},
		// MySQL.Password intentionally empty
	}

	err := Validate(cfg)
	assert.NoError(t, err, "debug mode should not require mysql.password")
}

// TestValidate_AlwaysRequiresJWTAndToken asserts JWT.Secret and InternalAPI.Token
// are required even in debug mode.
func TestValidate_AlwaysRequiresJWTAndToken(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Mode: "debug"},
	}

	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt.secret")
	assert.Contains(t, err.Error(), "internal_api.token")
}

// TestDevConfigFileHasNoLiteralSecrets asserts the committed config/config.dev.yaml
// uses ${VAR} placeholders for every secret field, never literal values.
func TestDevConfigFileHasNoLiteralSecrets(t *testing.T) {
	// Arrange: repo-relative path to the committed dev config.
	// This test runs from internal/config, so the file is ../../config/config.dev.yaml.
	path := filepath.Join("..", "..", "config", "config.dev.yaml")

	data, err := os.ReadFile(path)
	require.NoError(t, err, "config/config.dev.yaml must exist")
	content := string(data)

	// Each secret-bearing field, when it carries a value, must reference a ${VAR}
	// placeholder rather than a literal secret. Empty values (e.g. Redis password
	// in dev) are allowed — they are not leaked secrets.
	secretFields := []string{"password", "secret", "token", "api_key"}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, field := range secretFields {
			if !strings.HasPrefix(trimmed, field+":") {
				continue
			}
			// Extract the value after "field:".
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, field+":"))
			value = strings.Trim(value, `"'`)
			if value == "" {
				continue // empty is allowed (optional secret)
			}
			assert.Contains(t, value, "${",
				"field %q in config.dev.yaml must use a ${VAR} placeholder, got: %q", field, line)
		}
	}
}

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// envBindings maps environment variable names to config keys. Only these env
// vars override file values — there is no blanket AutomaticEnv, so unrelated
// env vars cannot silently leak into config (architecture §14.1, design D2).
var envBindings = map[string]string{
	"SERVER_PORT":          "server.port",
	"SERVER_MODE":          "server.mode",
	"MYSQL_HOST":           "mysql.host",
	"MYSQL_PORT":           "mysql.port",
	"MYSQL_USERNAME":       "mysql.username",
	"MYSQL_PASSWORD":       "mysql.password",
	"MYSQL_DATABASE":       "mysql.database",
	"REDIS_ADDR":           "redis.addr",
	"REDIS_PASSWORD":       "redis.password",
	"REDIS_DB":              "redis.db",
	"RABBITMQ_URL":         "rabbitmq.url",
	"ES_ADDRESSES":         "elasticsearch.addresses",
	"JWT_SECRET":           "jwt.secret",
	"JWT_EXPIRE_HOURS":     "jwt.expire_hours",
	"INTERNAL_API_TOKEN":   "internal_api.token",
	"AI_PROVIDER":          "ai.provider",
	"AI_MODEL":             "ai.model",
	"AI_API_KEY":           "ai.api_key",
	"AI_MAX_CONCURRENCY":   "ai.max_concurrency",
	"AI_REQUEST_PER_SECOND": "ai.request_per_second",
	"AI_BURST":             "ai.burst",
	"LOG_LEVEL":            "log.level",
	"LOG_ENCODING":         "log.encoding",
}

// Load reads the YAML file at path, applies defaults, overlays file values,
// then overlays allowlisted environment variables, returning a populated Config.
// Precedence: env > file > default.
func Load(path string) (*Config, error) {
	v := viper.New()

	// 1. Defaults first (lowest precedence).
	setDefaults(v)

	// 2. File values.
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// 3. Environment overrides (highest precedence) via explicit allowlist.
	// We deliberately do not enable viper's blanket env matching: only the
	// env vars enumerated in envBindings above may override file/default
	// values (architecture §14.1, design D2).
	for env, key := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			return nil, fmt.Errorf("failed to bind env %s to %s: %w", env, key, err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return cfg, nil
}

// setDefaults registers the baseline values used when neither file nor env
// supplies a key.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.expire_hours", 168)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")
}

// Package config provides typed, validated configuration loading for the
// ai-forum backend. Precedence is env > file > default; required secrets are
// enforced by Validate at startup.
package config

// Config mirrors the configuration shape fixed by architecture §14.2.
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	MySQL        MySQLConfig        `mapstructure:"mysql"`
	Redis        RedisConfig        `mapstructure:"redis"`
	RabbitMQ     RabbitMQConfig     `mapstructure:"rabbitmq"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	InternalAPI  InternalAPIConfig  `mapstructure:"internal_api"`
	AI           AIConfig           `mapstructure:"ai"`
	Worker       WorkerConfig       `mapstructure:"worker"`
	HotScore     HotScoreConfig     `mapstructure:"hot_score"`
	Log          LogConfig          `mapstructure:"log"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// MySQLConfig holds the strong-consistency source-of-truth connection settings.
type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// RedisConfig holds cache/counters/queue-infra settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RabbitMQConfig holds the domain event bus connection URL.
type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

// ElasticsearchConfig holds the eventually-consistent search read model endpoints.
type ElasticsearchConfig struct {
	Addresses []string `mapstructure:"addresses"`
}

// JWTConfig holds auth token settings.
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// InternalAPIConfig holds the token for worker→api-server internal calls.
type InternalAPIConfig struct {
	Token string `mapstructure:"token"`
}

// AIConfig holds the AI provider client settings.
type AIConfig struct {
	Provider         string `mapstructure:"provider"`
	Model            string `mapstructure:"model"`
	APIKey           string `mapstructure:"api_key"`
	MaxConcurrency   int    `mapstructure:"max_concurrency"`
	RequestPerSecond int    `mapstructure:"request_per_second"`
	Burst            int    `mapstructure:"burst"`
}

// WorkerConfig holds Asynq handler concurrency per task kind.
type WorkerConfig struct {
	AiReplyConcurrency      int `mapstructure:"ai_reply_concurrency"`
	TaggingConcurrency      int `mapstructure:"tagging_concurrency"`
	SearchIndexConcurrency  int `mapstructure:"search_index_concurrency"`
	NotificationConcurrency int `mapstructure:"notification_concurrency"`
}

// HotScoreConfig holds the hot-score refresh loop settings.
type HotScoreConfig struct {
	RefreshIntervalSeconds int `mapstructure:"refresh_interval_seconds"`
	BatchSize              int `mapstructure:"batch_size"`
}

// LogConfig holds structured logging settings.
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

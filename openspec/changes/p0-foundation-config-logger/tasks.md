# P0 Tasks — Foundation: deps, config, logger

## 1. Dependencies

- [x] 1.1 Add pinned v1.0 dependencies to `backend/go.mod`: `jmoiron/sqlx`, `go-sql-driver/mysql`, `redis/go-redis/v9`, `rabbitmq/amqp091-go`, `elastic/go-elasticsearch/v8`, `hibiken/asynq`, `casbin/casbin/v2`, `spf13/viper`, `go.uber.org/zap`, `natefinch/lumberjack.v2`, `golang-migrate/migrate/v4`, `google/uuid`, `stretchr/testify`
- [x] 1.2 Run `go mod tidy` and verify `go build ./...` is green
- [x] 1.3 Run `go vet ./...` clean
- [x] 1.4 Establish `govulncheck` baseline: `govulncheck ./...` runs clean (or documents accepted advisories)

## 2. Config package

- [x] 2.1 Define `Config` struct in `internal/config/config.go` mirroring architecture §14.2 (Server/MySQL/Redis/RabbitMQ/Elasticsearch/JWT/InternalAPI/AI/Worker/HotScore/Log)
- [x] 2.2 Implement `Load(path string) (*Config, error)` in `internal/config/loader.go` with defaults → file → env precedence using an explicit `bindEnv` allowlist (no blanket `AutomaticEnv`)
- [x] 2.3 Implement `Validate(*Config) error` in `internal/config/validate.go` aggregating missing required secrets (`JWT.Secret`, `InternalAPI.Token` always; `MySQL.Password` only when `Server.Mode != "debug"`)
- [x] 2.4 Create `backend/config/config.dev.yaml` with architecture §14.2 values, all secrets as `${VAR}` placeholders
- [x] 2.5 Align `.env.example` with the bound env names; add note that `INTERNAL_API_TOKEN` is generated via `openssl rand -hex 32`

## 3. Logger package

- [x] 3.1 Implement `logger.New(cfg config.Log) (*zap.Logger, error)` in `internal/logger/logger.go` (JSON encoder when `Encoding=="json"`, console otherwise; level from `Level`)
- [x] 3.2 Add `With(fields ...zap.Field)` child-logger helper for `event_id`/`task_id`/`user_id`/`request_id`/`post_id`/`comment_id`/`ai_agent_id`/`trigger_type`
- [x] 3.3 Add `Redact(keys ...string)` option masking named fields to `***` (enforcement point for "never log full token")
- [x] 3.4 Wire lumberjack file rotation only when a log file path is configured (optional)

## 4. Tests (AAA pattern)

- [x] 4.1 `internal/config/config_test.go`: defaults applied when absent; env overrides file; `Validate` fails loud listing all missing keys; debug mode relaxes `MySQL.Password`
- [x] 4.2 `internal/logger/logger_test.go`: JSON encoding produces valid JSON; child logger carries bound fields; redaction masks `token` field to `***`
- [x] 4.3 `config/config.dev.yaml` has no literal secrets (grep test asserts `${...}` placeholders only)
- [x] 4.4 `go test ./internal/config/... ./internal/logger/...` green

## 5. Verification

- [x] 5.1 `go build ./...` green with no business code beyond config/logger
- [x] 5.2 Standalone program/test can `Load()` dev config and instantiate a zap logger without panicking, asserting specific config values and logger level/encoding
- [x] 5.3 `govulncheck ./...` re-run clean

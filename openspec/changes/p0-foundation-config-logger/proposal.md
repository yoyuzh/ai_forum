## Why

The backend skeleton has every module directory and `AGENTS.md` boundary file in place, but `go.mod` declares zero dependencies and every `.go` file is a 2-line `package <name>` placeholder. Before any business module can be written, the two cross-cutting primitives every later phase depends on must exist: a typed, validated configuration loader and a structured logger. Locking the dependency set now also prevents drift across phases.

## What Changes

- Pin the v1.0 backend dependency set in `backend/go.mod` (Go 1.22): `jmoiron/sqlx` + `go-sql-driver/mysql`, `redis/go-redis/v9`, `rabbitmq/amqp091-go`, `elastic/go-elasticsearch/v8`, `hibiken/asynq`, `casbin/casbin/v2`, `spf13/viper`, `go.uber.org/zap`, `natefinch/lumberjack.v2`, `golang-migrate/migrate/v4`, `google/uuid`, `stretchr/testify`. Run `go mod tidy`.
- Implement `internal/config`: a `Config` struct mirroring architecture §14.2, a `Load(path)` with precedence **env > file > default**, and a `Validate` that fails loud when required secrets (`JWT_SECRET`, `INTERNAL_API_TOKEN`, `MYSQL_PASSWORD`) are absent in non-debug mode.
- Implement `internal/logger`: a `zap`-based `New(cfg)` with JSON (prod) / console (dev) encoders, field helpers for `event_id`/`task_id`/`user_id`/`request_id`, and a `Redact` helper that masks named secret fields.
- Add `backend/config/config.dev.yaml` (architecture §14.2 example with `${VAR}` placeholders) and align `.env.example` with the config env names, including a note that `INTERNAL_API_TOKEN` should be generated via `openssl rand -hex 32`.
- Add unit tests (AAA pattern) for config defaults, env-override precedence, missing-secret failure, and logger redaction.

## Capabilities

### New Capabilities
- `app-config`: Typed, validated configuration loading with env > file > default precedence and required-secret enforcement at startup.
- `app-logging`: Structured zap logging with secret redaction and contextual field helpers.

### Modified Capabilities
<!-- None — no specs exist yet. -->

## Impact

- **Code**: `backend/go.mod`, `backend/go.sum`, `backend/internal/config/*`, `backend/internal/logger/*`, `backend/config/config.dev.yaml`, `.env.example`.
- **Dependencies**: First real external dependencies introduced to the Go module.
- **Systems**: No runtime system changes; no business code references these yet beyond config/logger. Later phases (`P1` onward) consume `config.Load` and `logger.New` directly.
- **Non-functional**: `govulncheck` baseline established here and re-run each subsequent phase.

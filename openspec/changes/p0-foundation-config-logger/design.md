## Context

The backend skeleton (`backend/internal/*`, `backend/cmd/*`) exists as 2-line `package <name>` placeholders. `go.mod` declares only `go 1.22` with no dependencies. Architecture §14.2 fixes the configuration shape and §13.1 mandates zap structured logs with fields like `trace_id`, `request_id`, `user_id`, `task_id`, `event_id`. Every subsequent phase (P1 migrations onward) calls `config.Load` and `logger.New`, so these two primitives must land first and be stable.

Constraints (from CLAUDE.md / architecture doc / AGENTS.md):
- Secrets (`JWT_SECRET`, `INTERNAL_API_TOKEN`, `MYSQL_PASSWORD`, `AI_API_KEY`) MUST be injected via environment variables, never hardcoded.
- `INTERNAL_API_TOKEN` is generated via `openssl rand -hex 32`; the token must never appear in logs.
- Config precedence: **env > file > default**.

## Goals / Non-Goals

**Goals:**
- Pin a conservative, version-locked dependency set for v1.0.
- Provide `config.Load(path) (*Config, error)` with env > file > default precedence and explicit required-secret validation.
- Provide `logger.New(cfg) (*zap.Logger, error)` with JSON/console encoders, contextual field helpers, and secret redaction.
- Establish the `govulncheck` baseline that every later phase re-runs.

**Non-Goals:**
- No business modules, no migrations, no HTTP routes, no DI wiring (P1–P3).
- No Casbin enforcement logic — only the dependency is pinned (RBAC lands in P4).
- No log file rotation tuning beyond plumbing lumberjack optionally (v1 ships stderr/log file optional).

## Decisions

### D1: sqlx over GORM
The outbox pattern (P5) requires inserting `outbox_events` **inside the same `*sqlx.Tx`** as business writes. An ORM that owns transaction/session lifecycle obscures that boundary. `jmoiron/sqlx` gives thin, explicit control over `Tx` and keeps outbox writes raw and co-transactional. Alternative considered: GORM — rejected for hidden Tx semantics.

### D2: viper for config with explicit key mapping
`viper.SetConfigFile(path)` reads the YAML, `viper.AutomaticEnv()` + an explicit `bindEnv` key map (e.g. `MYSQL_PASSWORD` → `mysql.password`) gives env override without ambiguous matching. Defaults via `viper.SetDefault`. Validation is a separate explicit pass (not relying on `Unmarshal` tag errors that can swallow failures). Alternative: hand-rolled loader — rejected, viper handles env+file+default precedence natively.

### D3: zap with production JSON / development console
`zap.NewProduction` encoder when `log.encoding == json`; `zap.NewDevelopment` (console) otherwise. Level from `log.level`. A `With(fields ...zap.Field)` helper lets callers attach `event_id`/`task_id`/`user_id`/`request_id`. A `Redact(keys ...string)` option wraps the core so any field whose name is in the redact set is masked to `***` — this is the enforcement point for "never log the full token" (§13.4). Lumberjack is wired only when a file path is configured.

### D4: Required-secret validation as a distinct Validate step
`Validate(*Config) error` aggregates all missing required secrets (`JWT.Secret`, `InternalAPI.Token`, `MySQL.Password`) and returns one error listing every missing key. In `server.mode == debug`, `MySQL.Password` and `AI_API_KEY` may be optional for local dev, but `JWT.Secret` and `InternalAPI.Token` are always required. Fail-fast at startup is the security enforcement point.

### D5: config.dev.yaml uses ${VAR} placeholders, never literals
The committed dev config file contains `${MYSQL_PASSWORD}` etc., never real secrets. `.env.example` enumerates every env var with the same names the key map binds.

## Risks / Trade-offs

- **[Risk] viper's `AutomaticEnv` can silently match unintended env vars** → Mitigation: use an explicit `bindEnv` allowlist mapping rather than blanket `AutomaticEnv`, so only documented env keys override config.
- **[Risk] `Unmarshal` swallows type errors** → Mitigation: `Validate` runs structural checks after unmarshal; tests assert a missing-required-secret path fails loud.
- **[Risk] Redaction misses nested/renamed token fields** → Mitigation: redact by both key name and a known value-prefix guard is out of scope for P0; P0 redacts by exact field name. Documented as a known limitation to harden in P3 when the internal-API middleware logs token checks.
- **[Risk] Dependency bloat slows builds** → Mitigation: pin minor versions; the set is exactly what v1.0 needs, no extras.

## Migration Plan

1. Add deps to `go.mod`, run `go mod tidy`, verify `go build ./...` is green.
2. Implement `config` then `logger` with tests.
3. Commit `.env.example` and `config/config.dev.yaml`.
4. Rollback: revert `go.mod`/`go.sum` and delete the two packages; nothing else depends on them yet.

## Open Questions

- None for P0. (Casbin persistence adapter choice — sqlx adapter vs. in-memory model — is deferred to P4 RBAC.)

## Why

The outbox pattern (P5) requires writing `outbox_events` **inside the same transaction** as business writes, and every consumer needs idempotency via `processed_events`. Neither table exists yet — `backend/migrations/` is empty (only `.gitkeep`). Before any business module can write to MySQL, we need a connection, a transaction abstraction that the outbox pattern depends on, and baseline schema plus dev admin seed data.

## What Changes

- Implement `internal/database/mysql.go`: `NewMySQL(cfg)` returning a `*sqlx.DB` with utf8mb4/parseTime/collation set, sane pool defaults.
- Implement `internal/database/tx.go`: `RunInTx(ctx, db, fn func(tx *sqlx.Tx) error) error` — begin, run fn, commit on nil / rollback on error. This is the outbox-friendly primitive; **no RabbitMQ publish inside** (per `database/AGENTS.md` Must Not).
- Adopt `golang-migrate` (file-based, sequential). Add Makefile targets `migrate-up` / `migrate-down` / `migrate-create NAME=` reading the DSN from env.
- Add migrations (each with `.up.sql` + `.down.sql`):
  - `000001_init_schema` — baseline (utf8mb4, InnoDB; golang-migrate owns its version table).
  - `000002_outbox_events` — verbatim schema from architecture §8.4 (`idx_outbox_status_created_at`).
  - `000003_processed_events` — verbatim schema from architecture §9.2 (`uk_processed_event_consumer`, `idx_processed_events_processed_at`).
  - `000004_seed_dev` — dev-only seed: an admin user with a known dev bcrypt hash. Dev-only by convention (documented in the migration and `backend/AGENTS.md`); production bootstrap is a separate path not built in v1 (P13 scope is docker-compose, no prod deploy automation), so `000004` must not be applied against a non-dev DSN. AI agent seed data moves to P6, after P6 creates the `ai_agents` and `ai_agent_tag_preferences` tables.
- Add an integration test (build tag `integration`) proving migrations apply cleanly on a fresh MySQL 8.4 container and `RunInTx` commits/rolls back correctly.

## Capabilities

### New Capabilities
- `mysql-data-access`: sqlx-based MySQL connection and transaction abstraction with outbox-friendly `RunInTx` semantics.
- `db-migrations`: Versioned, reversible database migrations managed by golang-migrate, owning the baseline, outbox, and processed_events tables.
- `dev-seed-data`: Dev-only admin seed data enabling authenticated/admin flows to be tested without manual data entry.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `backend/internal/database/{mysql.go,tx.go,*_test.go}`, `backend/migrations/*`, `backend/Makefile`, `backend/internal/database/AGENTS.md` (document migrate targets).
- **Ownership rule established**: `outbox_events` and `processed_events` migrations are owned by **this phase (P1)**. Later phases (P4 forum core) MUST reference these tables, never re-create them. A CI check will be added in P5 to enforce single-table migration ownership.
- **Systems**: Requires a running MySQL 8.4 (docker-compose). No business tables yet beyond infra tables + seed.
- **Dependencies**: `golang-migrate/migrate/v4` (pinned in P0).

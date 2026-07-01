# P1 Tasks — DB, migrations, tx, seed

## 1. MySQL connection

- [x] 1.1 Implement `internal/database/mysql.go`: `NewMySQL(cfg config.MySQL) (*sqlx.DB, error)` with DSN `parseTime=true,loc=Local,charset=utf8mb4,collation=utf8mb4_unicode_ci,timeout=10s`; set `MaxOpenConns`/`MaxIdleConns`/`ConnMaxLifetime` defaults
- [x] 1.2 Define `DBTX` interface (`ExecContext`/`GetContext`/`SelectContext`/`QueryxContext` etc.) satisfied by both `*sqlx.DB` and `*sqlx.Tx`

## 2. Transaction primitive

- [x] 2.1 Implement `internal/database/tx.go`: `RunInTx(ctx, db, fn func(tx *sqlx.Tx) error) error` (begin → fn → commit on nil / rollback on error; wrap rollback errors)
- [x] 2.2 Confirm `database/AGENTS.md` Must Not ("no MQ publish inside a tx") is respected — `RunInTx` performs no MQ side effects

## 3. Migrations (golang-migrate, sequential)

- [x] 3.1 `migrations/000001_init_schema.up.sql` + `.down.sql` (utf8mb4, InnoDB baseline; golang-migrate owns version table) — also creates the `users` baseline table owned by P1 (P4 extends via ALTER; see ownership note below)
- [x] 3.2 `migrations/000002_outbox_events.up.sql` + `.down.sql` — verbatim architecture §8.4 (incl. `idx_outbox_status_created_at`)
- [x] 3.3 `migrations/000003_processed_events.up.sql` + `.down.sql` — verbatim architecture §9.2 (incl. `uk_processed_event_consumer`, `idx_processed_events_processed_at`)
- [x] 3.4 `migrations/000004_seed_dev.up.sql` + `.down.sql` — dev-only admin user (bcrypt hash); down removes only the seeded admin row by fixed ID/name. Do not insert AI agents here; P6 owns AI tables and AI seed rows.

## 4. Makefile

- [x] 4.1 Add `migrate-up` / `migrate-down` / `migrate-create NAME=` targets reading `MYSQL_DSN` from env (shared CI + local path) — uses a local `backend/cmd/migrate` wrapper that blank-imports the mysql+file drivers (the upstream `go run .../cmd/migrate@v4.18.1` CLI does not compile driver subpackages)
- [x] 4.2 Document the targets in `backend/AGENTS.md`

## 5. Integration tests (build tag `integration`)

- [x] 5.1 `internal/database/database_integration_test.go` (tag `integration`): against docker-compose MySQL — assert `Ping`, `RunInTx` commit persists, `RunInTx` rollback removes row, migrations apply cleanly
- [x] 5.2 Assert `outbox_events` columns/indexes match §8.4 and `processed_events` match §9.2 via `information_schema` introspection
- [x] 5.3 Assert `uk_processed_event_consumer` rejects a duplicate `(event_id, consumer_name)` insert
- [x] 5.4 Assert `000004_seed_dev` leaves 1 admin user and no AI rows; down reverses cleanly

## 6. Verification

- [x] 6.1 `make migrate-up` green on fresh MySQL 8.4; `make migrate-down` reverses cleanly
- [x] 6.2 `go test -tags=integration ./internal/database/...` green
- [x] 6.3 `go build ./...` and `go vet ./...` clean (with and without `-tags=integration`)
- [x] 6.4 No literal secrets in any migration (grep asserts `${...}`/bcrypt only)

## Ownership note (resolved during implementation)

P1's `000004_seed_dev` needed to insert an admin user, but `users` was originally P4's `000005_users` CREATE. To respect single-owner, `users` baseline ownership moved to P1 (`000001_init_schema`); P4's `000005_users` is now an `ALTER TABLE users` adding business columns (P4 tasks.md updated). `outbox_events`/`processed_events` remain P1-only.

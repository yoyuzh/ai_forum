# Module Instructions

## Responsibility

Own the Go backend codebase for the three v1.0 processes: `api-server`, `worker-service`, and `outbox-publisher`.

## Owns

- `cmd/api`, `cmd/worker`, `cmd/outbox-publisher`.
- `internal/` domain and infrastructure modules.
- `migrations/`, backend tests, backend Dockerfile, and Go module files.

## Must Not

- Do not create additional Go processes for v1.0.
- Do not split modules into networked microservices.
- Do not let a process bypass module service/repository boundaries.

## Allowed Dependencies

- Go standard library and backend dependencies selected by the architecture docs.
- Internal modules through explicit interfaces and dependency injection.

## Communication Rules

- `api-server` handles HTTP, auth, RBAC, synchronous writes, and SSE.
- `worker-service` consumes RabbitMQ events and runs Asynq handlers.
- `outbox-publisher` scans and publishes `outbox_events` only.
- Same-process modules do not call each other via HTTP.

## Data Rules

- Write core business state to MySQL first.
- Redis and Elasticsearch are derived/rebuildable infrastructure.
- Write reliable domain events through the outbox pattern.

## Testing Rules

- Add `go test ./...` coverage as implementation appears.
- Use integration tests for migrations and infrastructure adapters.

## Notes for Codex

- Keep placeholders thin until real business code is requested.
- Prefer package names that match directory names, using underscores only when Go identifiers require it.

## Migration Targets

Schema is managed with `golang-migrate` (sequential `000NNN_*.up.sql`/`.down.sql` pairs in `backend/migrations/`). Run from the repo root so CI and local share one path:

```bash
make migrate-up                 # apply all pending migrations
make migrate-down               # roll back the most recent migration
make migrate-create NAME=fix    # scaffold a new migration pair
make migrate-force V=3          # force a version (recovery only)
```

The DSN is read from env: set `MYSQL_DSN` directly to override, otherwise it is built from `MYSQL_HOST`/`MYSQL_PORT`/`MYSQL_USERNAME`/`MYSQL_PASSWORD`/`MYSQL_DATABASE` (same vars the config loader uses). `golang-migrate` is invoked via `go run` against the pinned CLI source — no global install needed.

### Migration ownership

- `outbox_events` and `processed_events` are owned by `000002`/`000003` (P1). No later migration SHALL `CREATE` or `DROP` them; extend with `ALTER` under a new migration number.
- `users` baseline is owned by `000001` (P1, seed admin in `000004`). P4 extends `users` via `ALTER` (its `000005_users` no longer `CREATE`s the table).

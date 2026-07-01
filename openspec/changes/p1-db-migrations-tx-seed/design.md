## Context

Architecture §8.4 fixes `outbox_events` and §9.2 fixes `processed_events`. The outbox pattern mandates that `outbox_events` be written in the same transaction as the business write, then published asynchronously by a separate process (P5). This demands an explicit transaction primitive that domain services can call and insert an outbox row onto the same `*sqlx.Tx`. `database/AGENTS.md` forbids publishing MQ inside a transaction.

`backend/migrations/` is currently empty. P1 can only seed tables it owns or that already exist. Therefore P1 seeds the dev admin user only; P6 seeds AI agents and tag preferences after creating their tables.

## Goals / Non-Goals

**Goals:**
- A `*sqlx.DB` connection with correct charset/collation and pool defaults.
- A `RunInTx` primitive that is the sole sanctioned place to do multi-statement business+outbox writes.
- Reversible migrations for baseline, outbox, processed_events, and dev seed, runnable via Makefile from the same DSN in CI and local.
- Integration tests proving migrations apply cleanly and `RunInTx` commits/rolls back.

**Non-Goals:**
- No business domain tables (posts/comments/etc.) — those are owned by P4.
- No `ai_reply_tasks`, `ai_reply_decisions`, `decision_logs` tables — owned by P6/P7.
- No migration CI gate enforcement — that lands in P5/P13 (contract ownership and migration ownership checks). P1 establishes the ownership rule in docs only.

## Decisions

### D1: sqlx raw SQL + explicit RunInTx
Chosen in P0 (D1). `RunInTx(ctx, db, fn)` begins a tx, passes `*sqlx.Tx` to fn, commits on nil return, rolls back on non-nil. Domain services accept a `DBTX` interface (`ExecContext`/`GetContext`/`SelectContext` etc.) so repositories work on both `*sqlx.DB` and `*sqlx.Tx`. This is what makes outbox-in-tx natural.

### D2: golang-migrate, sequential file-based
`golang-migrate` with sequential numeric prefixes (`000001_…`) over timestamps — easier to reason about ordering and dependency. Up and down files paired. The DSN comes from env (`MYSQL_DSN`) so `make migrate-up` works in CI and local identically. Alternative: `goose` embedding — rejected, golang-migrate's CLI + library split keeps migrations runnable without compiling.

### D3: Verbatim outbox/processed_events schema
Copy §8.4 and §9.2 column-for-column (names, types, indexes). This is the contract every later phase's SQL depends on. Any deviation must be an explicit architecture-doc amendment, not a silent migration change.

### D4: Dev seed as a guarded migration
`000004_seed_dev.up.sql` inserts an admin user (bcrypt-hashed). The down migration deletes only the seeded admin row by fixed ID/name so it is safe to re-run. This is dev-only data; production deploy skips it (documented). AI agent rows and tag preferences are seeded in P6 after those tables exist.

## Risks / Trade-offs

- **[Risk] Two phases touch the same table** → Mitigation: ownership rule documented here (outbox/processed_events owned by P1); P5 adds a CI check that no two migrations alter the same table.
- **[Risk] Seed migration leaks into prod** → Mitigation: document that `000004` is dev-only; production uses a separate seed path or skips it. No secrets in seed (admin password is a known dev bcrypt hash, not a real credential).
- **[Risk] `RunInTx` callers forget to insert the outbox row** → Mitigation: P4 specs require every forum write to assert one outbox row in-tx; enforced by test, not by this primitive.
- **[Risk] Migration down destroys data** → Mitigation: down migrations are reversible and only drop infra/seed objects; P13 verifies a `migrate-down`+`migrate-up` cycle on a populated DB.

## Migration Plan

1. Implement `mysql.go` + `tx.go` + repository `DBTX` interface.
2. Write the four migration pairs.
3. Add Makefile targets.
4. Add integration test (build tag `integration`) against docker-compose MySQL.
5. Rollback: `make migrate-down` reverses all four; delete the package.

## Open Questions

- Exact set of seeded `ai_agents` and their tag preferences — to be drawn from `stitch_ai_forum/design_cohere.md` and `ai_forum_requirements_v2.md` §6.5 during implementation; the count and roles are not fixed by this spec.

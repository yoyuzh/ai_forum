## ADDED Requirements

### Requirement: CI pipeline runs all gates
A GitHub Actions pipeline SHALL run on every change: backend `go test`/`go vet`/`govulncheck` + migrate-on-fresh-DB + the P5 contract-ownership test + P13 implementation-completeness check + a single-table migration-ownership check; web+admin `npm lint`/`build`; and the Playwright e2e suite against the compose stack.

#### Scenario: CI gates run green
- **WHEN** a PR is opened
- **THEN** all backend, frontend, and e2e gates run and must pass to merge

### Requirement: Single-table migration ownership enforced in CI
CI SHALL include a check that no two migrations create or drop the same table, enforcing the P1/P4 ownership rule for `outbox_events`/`processed_events` and domain tables.

#### Scenario: Duplicate table ownership fails CI
- **WHEN** two migrations attempt to create the same table
- **THEN** the ownership check fails CI

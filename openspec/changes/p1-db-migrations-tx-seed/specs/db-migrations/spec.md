## ADDED Requirements

### Requirement: Reversible migrations via golang-migrate
The project SHALL manage schema with `golang-migrate` using sequential numeric migration files, each paired with `.up.sql` and `.down.sql`. Makefile targets `migrate-up`, `migrate-down`, and `migrate-create NAME=` SHALL read the DSN from a single env var so CI and local share one path.

#### Scenario: Fresh apply and reverse
- **WHEN** `make migrate-up` runs against a fresh MySQL 8.4 container
- **THEN** all migrations apply with zero errors, and a subsequent `make migrate-down` reverses them cleanly

### Requirement: outbox_events schema matches §8.4
Migration `000002_outbox_events` SHALL create the table with columns `id BIGINT PK AUTO_INCREMENT`, `event_id VARCHAR(64) NOT NULL UNIQUE`, `event_type VARCHAR(100) NOT NULL`, `aggregate_type VARCHAR(50) NOT NULL`, `aggregate_id BIGINT NOT NULL`, `payload JSON NOT NULL`, `status VARCHAR(20) DEFAULT 'PENDING'`, `retry_count INT DEFAULT 0`, `created_at DATETIME`, `published_at DATETIME`, and index `idx_outbox_status_created_at(status, created_at)`.

#### Scenario: Schema introspection matches
- **WHEN** `information_schema` is queried for `outbox_events` columns and indexes
- **THEN** the names, types, and the `idx_outbox_status_created_at` index match architecture §8.4 exactly

### Requirement: processed_events schema matches §9.2
Migration `000003_processed_events` SHALL create the table with `id BIGINT PK AUTO_INCREMENT`, `event_id VARCHAR(64) NOT NULL`, `consumer_name VARCHAR(100) NOT NULL`, `processed_at DATETIME`, unique key `uk_processed_event_consumer(event_id, consumer_name)`, and index `idx_processed_events_processed_at(processed_at)`.

#### Scenario: Unique key prevents duplicate processing
- **WHEN** the same `(event_id, consumer_name)` pair is inserted twice
- **THEN** the second insert fails with a unique-key violation

### Requirement: Migration ownership is single-owner
The `outbox_events` and `processed_events` tables SHALL be created only by migrations owned in this phase. No later migration SHALL `CREATE` or `DROP` these tables; later phases that need columns SHALL use `ALTER` under a new migration number and document the dependency.

#### Scenario: No duplicate table ownership
- **WHEN** the migration set is scanned
- **THEN** exactly one migration creates `outbox_events` and exactly one creates `processed_events`

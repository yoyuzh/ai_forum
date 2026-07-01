# Module Instructions

## Responsibility

Own MySQL connection setup, transaction helpers, and migration integration helpers.

## Owns

- Database connection factories.
- Transaction abstractions.
- Migration integration helpers.

## Must Not

- Do not own domain repositories.
- Do not publish RabbitMQ messages inside transaction helpers.

## Allowed Dependencies

- Config, logger, and selected SQL/ORM/migration libraries.

## Communication Rules

- Provide transaction boundaries to domain services.
- Do not use database helpers to bypass repository ownership.

## Data Rules

- MySQL is the source of truth.
- Domain modules own their tables and repositories.

## Testing Rules

- Integration-test connection and transaction behavior when implemented.

## Notes for Codex

- Keep transaction APIs friendly to outbox writes.

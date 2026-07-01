# Module Instructions

## Responsibility

Own backend internal modules and enforce domain/infrastructure boundaries.

## Owns

- All packages under `backend/internal`.
- Interfaces and adapters used by the three backend processes.

## Must Not

- Do not expose packages for import outside the backend module.
- Do not create hidden HTTP calls between same-process modules.
- Do not mix RabbitMQ event definitions with Asynq task definitions.

## Allowed Dependencies

- Lower-level shared packages such as config, logger, common, database, cache, mq, task, event, and outbox when appropriate.
- Domain modules through interfaces only.

## Communication Rules

- Domain modules coordinate through injected interfaces.
- Async flows go through `outbox_events`, RabbitMQ domain events, and Asynq tasks.
- RabbitMQ consumers and Asynq handlers must be idempotent and retry-safe.

## Data Rules

- `processed_events` records RabbitMQ consumer idempotency.
- Unique-key conflicts that mean duplicate work should be treated as idempotent success.
- Keep table ownership with the module that owns the domain concept.

## Testing Rules

- Unit-test services with mocked interfaces.
- Integration-test repositories, outbox, queues, cache, and search adapters.

## Notes for Codex

- Add new packages only when they map to a clear domain or infrastructure responsibility.

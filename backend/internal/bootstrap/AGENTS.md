# Module Instructions

## Responsibility

Own process bootstrapping, dependency construction, lifecycle hooks, and graceful shutdown.

## Owns

- Files under this directory.
- Process wiring helpers shared by `api-server`, `worker-service`, and `outbox-publisher`.

## Must Not

- Do not implement business logic.
- Do not hide module boundary violations inside bootstrap wiring.

## Allowed Dependencies

- Backend config, logger, database, cache, mq, task, outbox, router, and module constructors as needed.

## Communication Rules

- Compose dependencies through constructors and interfaces.
- Do not create same-process HTTP clients between internal modules.

## Data Rules

- Bootstrap does not own persistent data.

## Testing Rules

- Test lifecycle and dependency wiring once real constructors exist.

## Notes for Codex

- Keep startup code explicit and boring.

# Module Instructions

## Responsibility

Own the `outbox-publisher` process entrypoint.

## Owns

- `main.go` wiring for scanning `outbox_events` and publishing RabbitMQ domain events.

## Must Not

- Do not run business handlers or Asynq task logic.
- Do not create or infer domain events outside persisted `outbox_events` rows.

## Allowed Dependencies

- Backend bootstrap/config/logger/database/mq/outbox modules required to start this process.

## Communication Rules

- Entrypoints wire dependencies only; business behavior belongs in internal modules.

## Data Rules

- Entrypoints must not own business tables directly.

## Testing Rules

- Keep startup wiring testable through small constructors once implementation starts.

## Notes for Codex

- Keep `main.go` minimal and delegate setup to bootstrap packages.

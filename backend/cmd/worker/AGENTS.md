# Module Instructions

## Responsibility

Own the `worker-service` process entrypoint.

## Owns

- `main.go` wiring for RabbitMQ consumers, Asynq workers, and background service dependencies.

## Must Not

- Do not serve public user/admin HTTP APIs.
- Do not mutate business tables except through owning service/repository boundaries.

## Allowed Dependencies

- Backend bootstrap/config/logger/database/mq/task modules required to start this process.

## Communication Rules

- Entrypoints wire dependencies only; business behavior belongs in internal modules.

## Data Rules

- Entrypoints must not own business tables directly.

## Testing Rules

- Keep startup wiring testable through small constructors once implementation starts.

## Notes for Codex

- Keep `main.go` minimal and delegate setup to bootstrap packages.

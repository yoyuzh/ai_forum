# Module Instructions

## Responsibility

Uniformly define Asynq task types, payloads, enqueue helpers, and handler registration.

## Owns

- Asynq task constants, payload structs, enqueue APIs, and handler registry wiring.

## Must Not

- Do not define RabbitMQ domain events here.
- Do not put concrete business logic inside task definitions or generic handler registration.

## Allowed Dependencies

- Asynq, `logger`, `config`, `event` references only when converting event facts into task requests.

## Communication Rules

- Asynq tasks express what should execute next.
- Handlers delegate concrete work to the owning service.
- All task handlers must support idempotency, retries, and structured logs.

## Data Rules

- Payloads should contain task IDs/business IDs needed to reload MySQL state.
- Do not store full business objects in task payloads.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

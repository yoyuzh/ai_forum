# Module Instructions

## Responsibility

Own RabbitMQ connection, exchange, queue, consumer, and producer infrastructure.

## Owns

- RabbitMQ adapters, queue bindings, consumer registration helpers, producer helpers.

## Must Not

- Do not define business event payloads; use `event`.
- Do not execute domain business logic in infrastructure callbacks.

## Allowed Dependencies

- `event`, `logger`, `config`, and shared retry/error helpers.

## Communication Rules

- RabbitMQ only carries domain events: facts that happened.
- Consumers must support idempotency through `processed_events` or equivalent owner logic.

## Data Rules

- Broker state is transport infrastructure, not durable business truth.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

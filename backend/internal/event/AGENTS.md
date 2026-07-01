# Module Instructions

## Responsibility

Uniformly define domain event names, envelopes, and payload contracts.

## Owns

- RabbitMQ domain event type constants and payload structs.

## Must Not

- Do not define Asynq tasks here.
- Do not put business handlers or producer implementations here.
- Do not include full business objects in event payloads.

## Allowed Dependencies

- Standard library and small shared primitives only.

## Communication Rules

- RabbitMQ events express what happened.
- Consumers must fetch complete business data from MySQL when needed.

## Data Rules

- Payloads should contain necessary business IDs and metadata only.
- Keep event schemas stable and versionable.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

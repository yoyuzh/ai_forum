# Module Instructions

## Responsibility

Own scanning, publishing, and status updates for `outbox_events`.

## Owns

- Outbox publisher services/repositories and outbox status transitions.

## Must Not

- Do not process business logic.
- Do not generate business event contents.
- Do not directly modify posts, comments, or AI tables.

## Allowed Dependencies

- `database`, `mq`, `event`, `logger`, and config.

## Communication Rules

- Read persisted outbox rows and publish them to RabbitMQ.
- Publishing retries must be safe and observable.

## Data Rules

- Own only `outbox_events` scanning/locking/published/failed status fields.
- Do not infer domain state outside stored rows.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

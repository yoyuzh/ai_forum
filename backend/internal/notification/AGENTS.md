# Module Instructions

## Responsibility

Own notification generation and delivery state.

## Owns

- Notification tables, services, task handlers, DTOs, and repositories.

## Must Not

- Do not own comment/post creation.
- Do not synchronously block forum writes.

## Allowed Dependencies

- `event`, `task`, `database`, user/forum read interfaces.

## Communication Rules

- Generate notifications from events/tasks such as `comment.created`, `ai.reply.completed`, and `user.mentioned`.

## Data Rules

- Notification records are derived from source domain data.
- Consumers must use `processed_events` idempotency where applicable.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

# Module Instructions

## Responsibility

Own content moderation policy and review execution for user and AI content.

## Owns

- Moderation rules, task handlers, review DTOs/services/repositories.

## Must Not

- Do not create posts/comments itself.
- Do not generate AI replies.

## Allowed Dependencies

- `task`, `database`, `modelclient` or provider interfaces when needed, domain read interfaces.

## Communication Rules

- Expose moderation interfaces to AI/forum modules.
- Async moderation work uses Asynq tasks.

## Data Rules

- Store moderation decisions and audit data owned by moderation.
- Treat unchecked model/user content as unsafe.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

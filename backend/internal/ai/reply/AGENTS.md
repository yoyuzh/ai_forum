# Module Instructions

## Responsibility

Own AI reply task execution, prompt assembly, generation, moderation handoff, and idempotency.

## Owns

- `ai_reply_tasks` access.
- Reply task services, repositories, handlers, DTOs, and idempotency checks.

## Must Not

- Do not skip business de-duplication.
- Do not write unmoderated generated content into comments.
- Do not treat unique-key conflicts as hard task failures when they mean duplicate work.

## Allowed Dependencies

- `modelclient` for generation.
- `moderation` for output review.
- `forum/comment` repository/service boundaries for comment writes.
- `outbox`, `database`, `logger`, and `task`.

## Communication Rules

- Execute only AI reply tasks.
- Use `modelclient` for generation, `moderation` before persistence, and `forum/comment` boundaries for comment writes.
- On success, write `outbox_events(ai.reply.completed)`.

## Data Rules

- Check for an existing business reply before creating work.
- Rely on `ai_reply_tasks` unique constraints as the database fallback.
- Unique-key conflicts are idempotent success.
- If `parent_comment_id` participates in uniqueness, use `parent_comment_id_norm = COALESCE(parent_comment_id, 0)` to avoid MySQL `NULL != NULL`.

## Testing Rules

- Test duplicate task handling, unique-key conflicts, moderation rejection, successful comment write, and completed-event outbox write.

## Notes for Codex

- This package is orchestration for reply tasks, not generic AI business logic.

# Module Instructions

## Responsibility

Own comments, replies, comment trees, and comment-like domain behavior.

## Owns

- `comments` table access through the comment repository.
- Comment DTOs, handlers, services, models, tree helpers, and comment events.

## Must Not

- Do not generate AI replies directly.
- Do not send notifications directly except by writing domain events.
- Do not bypass AI reply idempotency rules when writing AI comments.

## Allowed Dependencies

- Forum/post read interfaces when needed, `database`, `common`, `outbox`, and explicit moderation/user interfaces.

## Communication Rules

- User comments may emit `comment.created` through outbox.
- AI replies are written as comments only through explicit repository/service boundaries requested by `ai/reply`.

## Data Rules

- Preserve comment tree invariants.
- AI comments must include `ai_agent_id`, `comment_type=AI`, and `trigger_type` when written through the approved boundary.

## Testing Rules

- Test tree building, parent/child validation, deletion visibility, and event creation.

## Notes for Codex

- Keep user comment behavior separate from AI reply orchestration.

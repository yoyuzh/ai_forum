# Module Instructions

## Responsibility

Own AI roles, tagging, reply decisions, reply generation, follow-up judgment, preferences, and model provider calls.

## Owns

- AI submodules and AI-owned tables such as agents, preferences, decisions, and reply tasks.
- AI task orchestration boundaries for generated replies.

## Must Not

- Do not own the post creation main flow.
- Do not bypass forum/comment boundaries when writing AI replies.
- Do not output unmoderated AI content to comments.

## Allowed Dependencies

- `forum/comment` repository/service boundary for writing AI replies.
- `moderation` for AI output review.
- `event`, `outbox`, `task`, `cache`, `database`, and `modelclient` interfaces as needed.

## Communication Rules

- AI work is triggered by RabbitMQ events and Asynq tasks.
- AI replies must be persisted as comments through explicit repository/service boundaries.
- AI replies must bind `ai_agent_id`, `comment_type=AI`, and `trigger_type`.
- AI output must pass moderation before comment persistence.

## Data Rules

- Decision attempts must be explainable and persisted where required.
- Reply tasks must be idempotent.
- Do not treat model output as trusted data until moderation succeeds.

## Testing Rules

- Test decision scoring, task idempotency, moderation failure paths, and repository boundary interactions.

## Notes for Codex

- Keep model provider concerns in `modelclient`; keep reply orchestration in `reply`.

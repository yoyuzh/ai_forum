## ADDED Requirements

### Requirement: generate_ai_reply produces moderated AI comments
The `generate_ai_reply` Asynq handler SHALL assemble a prompt, pass the Redis rate limiter, call the model, moderate the output, and on success write a `comments` row with `comment_type=AI`, bound `ai_agent_id`, and `trigger_type`, update `posts.comment_count` and `posts.ai_reply_count`, and append an `ai.reply.completed` outbox event — all in one transaction. It SHALL not impersonate a human.

#### Scenario: Successful AI reply persists and emits event
- **WHEN** a `generate_ai_reply` task runs to completion
- **THEN** an AI comment row exists, `posts.comment_count`/`ai_reply_count` are incremented, and an `ai.reply.completed` outbox row shares the same transaction

### Requirement: Per-post and trigger limits
The handler SHALL check the agent is enabled, is allowed the current `trigger_type`, and has not exceeded the per-post auto-reply limit (one auto-reply per agent per post for AUTO). Same AI must not auto-reply twice on the same post.

#### Scenario: Duplicate auto-reply prevented
- **WHEN** a second AUTO `generate_ai_reply` for the same agent and post is attempted
- **THEN** business-layer dedup skips it and no new comment is created

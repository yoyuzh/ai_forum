## Why

This is the AI reply half of the chain and the moment the user sees AI respond. The `generate_ai_reply` handler assembles a prompt, calls the model through a rate limiter, moderates the output, writes the AI comment + `ai.reply.completed` outbox event in one transaction, and notifies the api-server SSE hub via the internal API (P3). It also owns the `ai_reply_tasks` table with the **four-column** generated-column unique key the architecture mandates (§6.4.1) — the critique's #4 risk.

## What Changes

- `internal/ai/modelclient`: model client interface + OpenAI-compatible implementation; prompt assembly per agent persona.
- `internal/ai/reply`: `generate_ai_reply` Asynq handler — business-layer dedup query, create `ai_reply_tasks`, check enabled/trigger-permission/per-post limits, assemble prompt, pass rate limiter, call model, moderate, write `comments` (comment_type=AI, ai_agent_id, trigger_type) + update `posts.comment_count`/`ai_reply_count` + append `ai.reply.completed` outbox, all in-tx.
- `internal/moderation`: moderation interface + rule-based v1 (sensitive-word + risk-tag); AI replies blocked by moderation are NOT persisted as visible comments.
- Migration `ai_reply_tasks` with generated column `parent_comment_id_norm GENERATED ALWAYS AS (COALESCE(parent_comment_id,0)) STORED` and unique key `uk_ai_reply_task(post_id, parent_comment_id_norm, ai_agent_id, trigger_type)` (4 columns, §6.4.1).
- SSE bridge: **extend** (not recreate) P3's `internal/internalapi` Hub — implement real dispatch so `worker → POST /internal/posts/{postId}/events` pushes `ai_reply_completed` to the post's SSE clients. Also implement `GET /api/posts/{postId}/events` SSE endpoint and `GET /api/posts/{postId}/ai-status` (§6.1.1).
- Rate limiter (Redis, per §6.4 AI API limiter).
- Tests: AI comment + `ai.reply.completed` in-tx; 4-col unique key concurrent insert → exactly one conflict treated as idempotent success; moderation block not persisted; SSE status delivered; rate-limit retry.

## Capabilities

### New Capabilities
- `ai-reply-generation`: Async `generate_ai_reply` task producing moderated AI comments with trigger-type semantics and per-post limits.
- `ai-reply-task-idempotency`: `ai_reply_tasks` with the 4-column generated-column unique key; business dedup + DB conflict as idempotent success.
- `ai-moderation`: Moderation gate that blocks AI replies from becoming visible comments on failure.
- `sse-ai-status`: SSE endpoint + ai-status endpoint + internal-API hub dispatch delivering AI reply status to clients.

### Modified Capabilities
- `internal-api-gateway`: EXTENDS P3's Hub interface with real SSE dispatch (the P3 no-op Hub becomes a real Hub). No route/middleware changes.

## Impact

- **Code**: `backend/internal/ai/{reply,modelclient}/*.go`, `backend/internal/moderation/*.go`, `backend/internal/sse/*.go` (real Hub), `backend/internal/internalapi/*.go` (extend dispatch), `backend/migrations/000016_ai_reply_tasks` (+down), `backend/internal/router` (SSE + ai-status routes).
- **Critique risk #4 closed**: 4-column unique key asserted by concurrent-insert test.
- **Ownership**: extends P3 internalapi/sse files (does not recreate).
- **Consumes**: P6 decision (enqueued tasks), P5 contracts + processed_events, P4 comments table.
- **Feeds**: P11 web real SSE, P9 notification (on `ai.reply.completed`).

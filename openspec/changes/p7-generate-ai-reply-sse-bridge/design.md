## Context

Architecture §6.4 (AI reply), §6.4.1 (idempotency two layers), §6.9 (SSE), §6.1.1 (ai-status), §12.3 (moderation), §6.5 (@AI — P8). P3 shipped an internalapi Hub interface + token middleware; P6 enqueues `generate_ai_reply`; P5 ships `ai.reply.completed` event + `processed_events`. This phase generates the reply, moderates, persists, and bridges to SSE. Critique #4: the unique key MUST be 4-column `(post_id, parent_comment_id_norm, ai_agent_id, trigger_type)`, not single-column.

## Goals / Non-Goals

**Goals:**
- `generate_ai_reply` handler with business dedup + 4-col unique backstop.
- Model client + prompt assembly + Redis rate limiter.
- Moderation gate (AI reply blocked → not persisted visible).
- Real SSE Hub dispatch + `GET /api/posts/{postId}/events` + `GET /api/posts/{postId}/ai-status`.
- Extend P3 Hub (no file recreation).

**Non-Goals:**
- No @AI mention path (P8) — but `trigger_type` column supports MENTION/FOLLOWUP for P8.
- No followup judge (P8).
- No notification generation (P9 consumes `ai.reply.completed`).
- No search index write here (P9).

## Decisions

### D1: Two-layer idempotency (§6.4.1)
Layer 1 (business): before creating a task, query `ai_reply_tasks WHERE post_id=? AND COALESCE(parent_comment_id,0)=COALESCE(?,0) AND ai_agent_id=? AND trigger_type=?`. Existing PENDING/RUNNING/SUCCESS/BLOCKED/SKIPPED → skip; FAILED → no auto-retry. Layer 2 (DB): the 4-column generated-column unique key. A unique-conflict on insert is treated as idempotent success (not failure) per §6.4.1/§9.4.

### D2: 4-column generated-column unique key
`parent_comment_id_norm BIGINT GENERATED ALWAYS AS (COALESCE(parent_comment_id,0)) STORED`; `UNIQUE KEY uk_ai_reply_task (post_id, parent_comment_id_norm, ai_agent_id, trigger_type)`. A concurrent-insert test asserts exactly one row survives and the conflict is treated as idempotent success.

### D3: Moderation before persist
Model output → moderation gate → if blocked, write `ai_reply_tasks` status BLOCKED, append `ai.reply.failed`/no visible comment. Do NOT write a `comments` row. Moderation is an interface so P9/v1 rule-based impl swaps for richer later.

### D4: In-tx comment + outbox + counters
`generate_ai_reply` writes the `comments` row (comment_type=AI, ai_agent_id, trigger_type), updates `posts.comment_count`/`ai_reply_count`, and appends `ai.reply.completed` — all in one `RunInTx`. No in-tx MQ publish.

### D5: SSE Hub extends P3
P3's `Hub` interface gains a real in-memory implementation: `Subscribe(postId) chan Event`, `Unsubscribe`, `Publish(postId, event)`. The `POST /internal/posts/{postId}/events` handler (P3 skeleton) calls `Hub.Publish`. `GET /api/posts/{postId}/events` is an SSE handler streaming from `Hub.Subscribe`. Single api-server instance for v1 (§6.9). `GET /api/posts/{postId}/ai-status` returns the §6.1.1 aggregate.

### D6: Redis rate limiter
Token-bucket per §6.4 (`ai.request_per_second`, `burst`); over-limit → Asynq retry with backoff (§9.4 retry rule).

## Risks / Trade-offs

- **[Risk] Wrong unique key → wrong dedup** → Mitigation: D2 + concurrent-insert test asserting the 4-column key.
- **[Risk] Moderation failure still writes comment** → Mitigation: D3 test asserts no `comments` row + BLOCKED task on moderation fail.
- **[Risk] SSE Hub not extended but recreated (P7 vs P3 file collision)** → Mitigation: D5 extends P3 files; ownership documented.
- **[Risk] Model timeout/failure loses the task** → Mitigation: Asynq retry; `ai.reply.failed` outbox on terminal failure.
- **[Risk] Single-instance SSE Hub doesn't scale** → Mitigation: v1 scope (§6.9); multi-instance is v2.

## Migration Plan

1. migration → modelclient → reply handler → moderation → SSE Hub extension → routes → tests.
2. Rollback: drop `ai_reply_tasks`; revert Hub to no-op; SSE endpoints 404.

## Open Questions

- Model provider default (OpenAI gpt-4o-mini per §14.2) — configurable; v1 ships OpenAI-compatible client.

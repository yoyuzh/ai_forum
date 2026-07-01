## Context

Architecture §6.7 (search), §6.8 (notification), §3.3 (ES final-consistency read-model), §7.4 (`q.search.index`, `q.notification`). Critique found `comment.deleted`/`ai.reply.failed`/`post.moderated` unowned. P5 ships event contracts + the ownership test; this phase implements the two downstream handlers and ensures every relevant event has a consumer, keeping the ownership metadata accurate and enabling P13 implementation-completeness to pass.

## Goals / Non-Goals

**Goals:**
- `sync_search_index` re-fetching from MySQL, upsert/delete in ES, including moderation/deletes.
- `send_notification` writing notification rows.
- ES index mapping with IK for Chinese.
- ES-down chaos: MySQL writes unaffected.
- Full rebuild == incremental.

**Non-Goals:**
- No search ranking/relevance tuning beyond IK (v1).
- No notification delivery push (SSE is for AI status only; notifications are DB rows read on demand).
- No hot-score (P10).

## Decisions

### D1: Search worker re-fetches from MySQL (§6.7)
Event payload carries only IDs; the handler re-fetches the full row from MySQL before writing ES. Prevents stale-payload corruption. Document type (`post`/`comment`/`ai_comment`) derived from event type + aggregate.

### D2: Deletes and moderation sync
`post.deleted`/`comment.deleted` → delete ES document. `post.moderated` → update or remove the document per moderation outcome. `ai.reply.failed` does not write ES (no comment was created) but is consumed to ack and log.

### D3: ES-down does not block MySQL (chaos)
The search consumer is async and best-effort; if ES is down, the Asynq task retries (§9.4) but the MySQL write path (P4) is unaffected. A chaos test kills the ES container and asserts a post create still returns 200 and persists.

### D4: Full rebuild == incremental
A rebuild utility re-indexes from MySQL using the same document-assembly code path as the incremental handler. A test asserts a rebuild produces identical documents to incremental sync for a known dataset.

### D5: Notification recipients
`comment.created` → notify post author (and mentioned users via `user.mentioned`). `ai.reply.completed` → notify post author. `send_notification` writes `notifications` rows; no push delivery in v1.

## Risks / Trade-offs

- **[Risk] ES lag hides a just-created post from search** → Mitigation: post detail reads MySQL (immediate); search reads ES (1–3s lag) — by design (§6.7).
- **[Risk] Stale payload written to ES** → Mitigation: D1 re-fetch.
- **[Risk] Unowned events break ownership/implementation checks** → Mitigation: D2 consumes them and P4/P7 produce them.
- **[Risk] Notification spam** → Mitigation: dedup via `processed_events`; recipient rules bounded.

## Migration Plan

1. notifications migration → ES mapping → search handler → notification handler → chaos + rebuild tests.
2. Rollback: ES index can be dropped and rebuilt; notifications table reversible.

## Open Questions

- ES index refresh interval v1 — 1s (near-real-time) default.

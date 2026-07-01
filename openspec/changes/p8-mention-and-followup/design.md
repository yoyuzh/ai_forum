## Context

Architecture §6.5 (@AI) and §6.6 (followup). P7's `generate_ai_reply` already supports `trigger_type` via the 4-col unique key, so MENTION/FOLLOWUP are new trigger paths, not new task types. The mention path is synchronous in the request (creates the task directly, §6.5), unlike AUTO which is fully outbox-driven. Followup uses a lightweight model with strict safe-default semantics (§6.6 anomaly table).

## Goals / Non-Goals

**Goals:**
- @AI mention: validate, rate-limit, create MENTION task + enqueue.
- Followup judge: structured JSON, safe-default false on every anomaly class.
- Guards: AI→AI blocked, real-user-only FOLLOWUP, ≤3 followup/agent/post.

**Non-Goals:**
- No new `generate_ai_reply` variant — reuses P7.
- No rule-based fallback for followup (§6.6 explicitly excludes v1 rule fallback).
- No notification for mentions (P9).

## Decisions

### D1: Mention is synchronous-in-request
Per §6.5, the comment create handler validates mentions, writes `comment_mentions`, checks Redis rate limits, and creates `ai_reply_tasks`(MENTION) + enqueues `generate_ai_reply` directly. This is the one place AI task creation happens in the api-server request path (vs outbox-driven AUTO). The comment write itself still uses the P4 in-tx + outbox pattern for `comment.created`.

### D2: Followup safe-default false (§6.6 anomaly table)
Any of: model timeout, non-JSON response, missing `should_reply` field, non-boolean value, or call failure → `should_reply=false` and the handler ends without enqueueing. No retry on anomaly (transient; user can re-@AI). A unit test per anomaly class.

### D3: Guards
AI comments do not trigger followup (parent must be AI, author must be real user). Same agent ≤3 followup replies per post (query `ai_reply_tasks` WHERE trigger_type=FOLLOWUP AND ai_agent_id AND post_id). The 4-col unique key already prevents duplicate FOLLOWUP for the same parent.

### D4: Mention rate limits (§6.5)
≤3 AI mentions per comment; ≤5 @AI per user per minute (Redis sliding window). Disabled agents or agents with `allowMention=false` are skipped.

## Risks / Trade-offs

- **[Risk] Mention path blocks the request on rate-limit/agent checks** → Mitigation: checks are fast (Redis + indexed query); the actual reply is still async.
- **[Risk] Followup anomaly silently swallows a legit continuation** → Mitigation: §6.6 rationale (miss > false-trigger); user can re-@AI; anomaly logged.
- **[Risk] AI→AI followup loop** → Mitigation: D3 author-must-be-real-user guard.

## Migration Plan

1. Extend comment create → followup handler → mention validation → tests.
2. Rollback: mentions create no tasks; followup consumer stops.

## Open Questions

- Lightweight followup model choice — same provider, smaller model; configurable.

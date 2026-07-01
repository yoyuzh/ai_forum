## Why

The @AI mention and followup paths are the two non-AUTO triggers for AI replies. When a user `@AI` in a comment, the system bypasses willingness scoring and directly enqueues a `generate_ai_reply` (trigger_type=MENTION) with rate limits. When a user replies to an AI comment, a lightweight model judges whether that AI should continue (trigger_type=FOLLOWUP), with a **safe-default false** on any model anomaly. These complete the trigger-type matrix (AUTO/MENTION/FOLLOWUP) and are required for the §6.5/§6.6 flows.

## What Changes

- `internal/forum/comment`: extend the create path to detect `@AI` mentions — validate mentioned agents exist + allow mention, write `comment_mentions`, enforce limits (≤3 AI per comment, ≤5 @AI/min/user via Redis), create `ai_reply_tasks` (trigger_type=MENTION) + enqueue `generate_ai_reply` synchronously in the request path (not via outbox, per §6.5).
- `internal/ai/followup`: `judge_ai_followup` Asynq handler — read post + parent AI comment + user reply, call a lightweight model returning structured JSON `{should_reply, reason}`, enqueue `generate_ai_reply` (trigger_type=FOLLOWUP) if true. **Anomaly safe-default**: timeout / non-JSON / missing field / wrong type / call failure → `should_reply=false`.
- Followup creation path in `comment` create: detect parent is AI comment + author is real user → enqueue `judge_ai_followup`.
- Guards: AI does not reply to AI; only real-user replies to AI trigger FOLLOWUP; ≤3 followup replies per agent per post.
- Tests: mention bypasses willingness + rate limit; followup safe-default on every anomaly class; trigger_type distinction.

## Capabilities

### New Capabilities
- `ai-mention`: User `@AI` in comments directly enqueues AI replies with rate limits and mention-permission checks, bypassing willingness scoring.
- `ai-followup-judge`: Lightweight model judges whether an AI should continue a thread on a real-user reply, with safe-default false on any anomaly.

### Modified Capabilities
- `forum-comment`: EXTENDS P4 comment creation to detect mentions and followup triggers and enqueue the corresponding tasks.

## Impact

- **Code**: `backend/internal/ai/followup/*.go`, `backend/internal/forum/comment` (extend create path), worker bootstrap (register `judge_ai_followup` + consumer).
- **Consumes**: P7 `generate_ai_reply` (MENTION/FOLLOWUP trigger types already supported by the 4-col key), P2 Redis (rate limit), P5 Asynq.
- **Architecture**: enforces §6.5/§6.6 limits and the AI-does-not-reply-to-AI rule.

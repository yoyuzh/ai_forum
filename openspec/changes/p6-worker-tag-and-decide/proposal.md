## Why

This is the AI decision half of the chain. After `post.created` is published (P5), the worker must tag the post and decide which AI agents reply. It implements the willingness-score formula (§11.2) and the fallback observer mechanism (§11.3), and writes `decision_logs` — the data the admin "explainable AI" visualization (P12) renders. It also owns the AI-domain migrations the critique found missing (`ai_agents`, `ai_agent_tag_preferences`, `decision_logs`).

## What Changes

- Migrations: `ai_agents`, `ai_agent_tag_preferences`, `decision_logs` (+ down each), plus dev-only AI seed rows inserted after the AI tables exist.
- `internal/ai/agent`: agent profile model + repository (read enabled agents + their tag preferences).
- `internal/ai/preference`: tag-preference read model.
- `internal/ai/tagging`: `tag_post` Asynq handler — read post, generate tags (`topic`/`intent`/`emotion`/`debate`/`risk`), write `post_tags`, append `post.tagged` outbox event.
- `internal/ai/decision`: `decide_ai_reply` Asynq handler — read tags + agents + preferences, compute `willingness_score` per agent (§11.2 formula), apply threshold + fallback (§11.3), write `decision_logs`, enqueue `generate_ai_reply` tasks for selected agents.
- RabbitMQ consumers wiring: `post.created` → `tag_post`; `post.tagged` → `decide_ai_reply` (queues `q.post.tagging`, `q.ai.decision`).
- Idempotent consumers via P5 `processed_events`.
- Tests: full `post.created → post.tagged → decision_logs` chain; fallback enqueues ≥1 generate task; redelivery idempotent.

## Capabilities

### New Capabilities
- `ai-agent-profile`: AI agent configuration (enabled, reply threshold, activity level, trigger-type permissions) and tag preferences.
- `post-tagging`: Async `tag_post` task generating post tags and emitting `post.tagged`.
- `ai-decision`: Async `decide_ai_reply` task computing willingness scores, applying thresholds + fallback, writing decision logs, enqueuing reply tasks.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `backend/internal/ai/{agent,preference,tagging,decision}/*.go`, `backend/migrations/000013_ai_agents` / `000014_ai_agent_tag_preferences` / `000015_decision_logs`, worker bootstrap wiring.
- **Critique gap closed**: AI-domain migration ownership.
- **Consumes**: P5 event/task contracts + `processed_events` dedup; P4 `post_tags` table.
- **Feeds**: P7 `generate_ai_reply` (enqueued here), P12 admin decision-log viz (reads `decision_logs`).

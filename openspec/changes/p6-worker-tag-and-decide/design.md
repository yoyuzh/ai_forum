## Context

Architecture §6.2 (tagging), §6.3 (decision), §11.1 (preferences), §11.2 (willingness formula), §11.3 (fallback), §11.4 (decision log fields). P5 ships the event/task contracts and `processed_events` dedup; P4 ships `post_tags`. This phase implements the first two worker handlers, owns the AI-domain migrations, and seeds dev AI agents/tag preferences after the AI tables exist. Decision logs are the substrate for P12's explainable-AI screen, so their shape is pinned here.

## Goals / Non-Goals

**Goals:**
- `tag_post` handler producing 5 tag types + `post.tagged` outbox.
- `decide_ai_reply` handler computing willingness per §11.2, applying threshold + fallback §11.3, writing `decision_logs`, enqueuing `generate_ai_reply`.
- AI-domain migrations and dev AI seed rows.
- Idempotent consumers.

**Non-Goals:**
- No AI reply generation (P7), no model client beyond a tagging interface (P7 model client).
- No @AI/followup (P8).
- No admin UI (P12) — but the `decision_logs` schema must carry everything P12 needs (willingness, threshold, hit tags, reason, fallback flag, decision).

## Decisions

### D1: Willingness formula exactly per §11.2
`willingness_score = topic*0.35 + intent*0.25 + emotion*0.15 + debate*0.15 + activity*0.10 - risk_penalty - frequency_penalty`, where per-tag-type `tag_score = max_score*0.7 + avg_score*0.3`. Computed in pure functions with unit tests for each coefficient.

### D2: Fallback observer per §11.3
Compute scores → candidate pool = agents over threshold → if empty, pick highest → if highest < 0.35, invoke fallback observer agent. Guarantee ≥1 reply task enqueued. The fallback flag is recorded in `decision_logs`.

### D3: decision_logs schema carries full explainability
Columns: `post_id`, `comment_id` (nullable until P7), `ai_agent_id`, `trigger_type`, `willingness_score`, `threshold_value`, `decision` (REPLY/IGNORE/FALLBACK), `reason`, `hit_tags` (JSON), `created_at`. This is the contract P12 renders.

### D4: tagging uses a pluggable tagger interface
`tagging.Tagger` interface (rule-based or lightweight model). v1 ships a rule-based implementation; the model-client integration is P7. Keeps `tag_post` testable without a live model.

### D5: Consumers reuse P5 idempotency
`post.created` consumer → enqueue `tag_post` Asynq task (after `MarkProcessed`). `post.tagged` consumer → enqueue `decide_ai_reply`. Both idempotent.

## Risks / Trade-offs

- **[Risk] Formula drift from doc** → Mitigation: D1 unit tests assert exact coefficients against hand-computed fixtures.
- **[Risk] Fallback never triggers / triggers too eagerly** → Mitigation: D2 tests cover empty-pool, sub-threshold-highest, and normal cases.
- **[Risk] decision_logs schema too thin for P12** → Mitigation: D3 pins the full field set now; P12 depends on this phase.
- **[Risk] Redelivery enqueues duplicate generate tasks** → Mitigation: P5 `processed_events` dedup; P7's `ai_reply_tasks` unique key is the final backstop.

## Migration Plan

1. Migrations → agent/preference repos → tagging handler → decision handler → consumers → tests.
2. Rollback: `migrate-down` reverses AI tables; consumers stop; outbox `post.tagged` rows drain.

## Open Questions

- Threshold default value per agent — stored on `ai_agents` (P6 seed sets initial values); tunable via admin (P12).

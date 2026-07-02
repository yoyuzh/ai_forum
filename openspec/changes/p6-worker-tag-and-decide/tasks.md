# P6 Tasks — Worker: tag_post + decide_ai_reply

## 1. AI-domain migrations (owned here)
- [x] 1.1 `000013_ai_agents` (+down): enabled, reply_threshold, activity_level, allowAutoReply/allowMention/allowFollowup
- [x] 1.2 `000014_ai_agent_tag_preferences` (+down): ai_agent_id, tag_type, tag_name, weight(0.0–1.0)
- [x] 1.3 `000015_decision_logs` (+down): post_id, comment_id nullable, ai_agent_id, trigger_type, willingness_score, threshold_value, decision, reason, hit_tags JSON, created_at
- [x] 1.4 Add dev-only AI seed rows after `000013`/`000014`: ≥3 enabled agents, thresholds/activity defaults, trigger permissions, and tag preferences. Down migration removes only seeded rows by fixed IDs/names.

## 2. Agent + preference repos
- [x] 2.1 `internal/ai/agent`: model + repository (read enabled agents + trigger perms)
- [x] 2.2 `internal/ai/preference`: tag-preference read model
- [x] 2.3 Tests: seeded agents readable with preferences + thresholds

## 3. tag_post handler
- [x] 3.1 `internal/ai/tagging`: `Tagger` interface + rule-based v1 implementation (5 tag types)
- [x] 3.2 `tag_post` Asynq handler: read post → generate tags → write `post_tags` → append `post.tagged` outbox
- [x] 3.3 `post.created` consumer → enqueue `tag_post` (idempotent via processed_events)
- [x] 3.4 Tests: 5 tag types written; `post.tagged` appended; redelivery no-op

## 4. decide_ai_reply handler
- [x] 4.1 `internal/ai/decision`: pure willingness-score functions (§11.2 coefficients; per-type `max*0.7 + avg*0.3`)
- [x] 4.2 Unit tests: each coefficient + hand-computed fixtures
- [x] 4.3 Threshold + fallback logic (§11.3): candidate pool → highest if empty → fallback observer if <0.35; guarantee ≥1 generate task enqueued
- [x] 4.4 Write `decision_logs` per evaluated agent (full explainability fields)
- [x] 4.5 `post.tagged` consumer → enqueue `decide_ai_reply` (idempotent)
- [x] 4.6 Tests: normal multi-select; empty-pool fallback; sub-0.35 fallback; N decision logs; redelivery no-op

## 5. Worker bootstrap wiring
- [x] 5.1 Register `tag_post` + `decide_ai_reply` Asynq handlers in worker bootstrap
- [x] 5.2 Register `post.created`/`post.tagged` RabbitMQ consumers bound to `q.post.tagging`/`q.ai.decision`

## 6. Verification
- [x] 6.1 End-to-end: forum write (P4) → publisher (P5) → `post.created` → tag_post → `post.tagged` → decide_ai_reply → decision_logs + enqueued generate tasks
- [x] 6.2 `go test ./internal/ai/...` green; `make migrate-up` applies 000013–000015
- [x] 6.3 `go build ./cmd/worker`; `go vet ./...` / `govulncheck ./...` clean

# P7 Tasks — generate_ai_reply + moderation + SSE bridge

## 1. ai_reply_tasks migration (4-col unique key)
- [x] 1.1 `000016_ai_reply_tasks` (+down): post_id, parent_comment_id nullable, ai_agent_id, trigger_type, status, task fields, `parent_comment_id_norm GENERATED ALWAYS AS (COALESCE(parent_comment_id,0)) STORED`, `UNIQUE KEY uk_ai_reply_task(post_id, parent_comment_id_norm, ai_agent_id, trigger_type)`
- [x] 1.2 Concurrent-insert test: two racing inserts → exactly one row, conflict treated as idempotent success (no FAILED, no duplicate comment)

## 2. Model client + rate limiter
- [x] 2.1 `internal/ai/modelclient`: interface + OpenAI-compatible impl; prompt assembly per agent persona
- [x] 2.2 Redis token-bucket rate limiter (§6.4 `request_per_second`/`burst`); over-limit → Asynq retry w/ backoff
- [x] 2.3 Tests: model call mocked; rate-limit retry path

## 3. generate_ai_reply handler
- [x] 3.1 `internal/ai/reply`: business-layer dedup query (4-col) → create ai_reply_tasks → check enabled/trigger-perm/per-post limit → prompt → limiter → model → moderate → write comments(comment_type=AI, ai_agent_id, trigger_type) + update comment_count/ai_reply_count + append ai.reply.completed (all in-tx)
- [x] 3.2 Existing PENDING/RUNNING/SUCCESS/BLOCKED/SKIPPED → skip; FAILED → no auto-recreate
- [x] 3.3 Register `generate_ai_reply` Asynq handler in worker bootstrap
- [x] 3.4 Tests: success persists comment+outbox in-tx; duplicate AUTO prevented; terminal failure → ai.reply.failed outbox

## 4. Moderation
- [x] 4.1 `internal/moderation`: interface + rule-based v1 (sensitive-word + risk-tag)
- [x] 4.2 On block: no comments row, task BLOCKED, no retry (§9.4)
- [x] 4.3 Test: blocked reply not persisted as visible comment

## 5. SSE bridge (extend P3, do not recreate)
- [x] 5.1 `internal/sse`: real in-memory Hub (`Subscribe`/`Unsubscribe`/`Publish`) implementing P3's Hub interface
- [x] 5.2 Extend `internal/internalapi` `POST /internal/posts/{postId}/events` to dispatch to Hub (P3 token middleware unchanged)
- [x] 5.3 `internal/router`: `GET /api/posts/{postId}/events` SSE handler; `GET /api/posts/{postId}/ai-status` §6.1.1 aggregate
- [x] 5.4 Tests: worker internal call → SSE clients receive `ai_reply_completed`; ai-status reflects running/completed counts; P3 files extended not recreated (no duplicate route registration)

## 6. Verification
- [x] 6.1 Full chain: P4 write → P5 publish → P6 decide → P7 generate → SSE push + ai-status
- [x] 6.2 `go test ./internal/ai/{reply,modelclient}... ./internal/{moderation,sse,internalapi}...` green
- [x] 6.3 `make migrate-up` applies 000016; `go build ./cmd/api ./cmd/worker`; `go vet ./...` / `govulncheck ./...` clean

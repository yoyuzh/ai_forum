# P8 Tasks — @AI mention + followup judge

## 1. @AI mention path
- [x] 1.1 Extend `forum/comment` create: parse mentions, validate agent exists + `allowMention=true`, write `comment_mentions`
- [x] 1.2 Redis rate limit: ≤3 AI/comment, ≤5 @AI/user/minute (sliding window)
- [x] 1.3 After comment tx commit: create `ai_reply_tasks`(MENTION) + enqueue `generate_ai_reply` (no willingness scoring)
- [x] 1.4 Tests: mention enqueues MENTION; rate-limit 429; disabled/mention-disallowed skipped; >3/comment rejected

## 2. Followup judge
- [x] 2.1 `internal/ai/followup`: `judge_ai_followup` handler — read post + parent AI comment + user reply, call lightweight model, parse `{should_reply,reason}` JSON
- [x] 2.2 Safe-default false on: timeout, non-JSON, missing field, non-boolean, call failure (one test per anomaly class)
- [x] 2.3 On true: enqueue `generate_ai_reply`(FOLLOWUP)
- [x] 2.4 Register `judge_ai_followup` Asynq handler in worker bootstrap

## 3. Followup guards
- [x] 3.1 Comment create: detect parent is AI + author is real user → enqueue `judge_ai_followup`
- [x] 3.2 AI→AI reply does NOT trigger followup (author-must-be-real-user guard)
- [x] 3.3 ≤3 FOLLOWUP/agent/post cap (query `ai_reply_tasks` WHERE trigger_type=FOLLOWUP)
- [x] 3.4 Tests: AI→AI no task; cap enforced; real-user→AI enqueues judge

## 4. Verification
- [x] 4.1 AUTO (P6/P7) / MENTION / FOLLOWUP trigger types all distinct and routed correctly through the 4-col unique key
- [x] 4.2 `go test ./internal/ai/followup... ./internal/forum/comment/...` green
- [x] 4.3 `go build ./cmd/api ./cmd/worker`; `go vet ./...` / `govulncheck ./...` clean

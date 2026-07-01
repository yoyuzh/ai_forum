# P8 Tasks — @AI mention + followup judge

## 1. @AI mention path
- [ ] 1.1 Extend `forum/comment` create: parse mentions, validate agent exists + `allowMention=true`, write `comment_mentions`
- [ ] 1.2 Redis rate limit: ≤3 AI/comment, ≤5 @AI/user/minute (sliding window)
- [ ] 1.3 After comment tx commit: create `ai_reply_tasks`(MENTION) + enqueue `generate_ai_reply` (no willingness scoring)
- [ ] 1.4 Tests: mention enqueues MENTION; rate-limit 429; disabled/mention-disallowed skipped; >3/comment rejected

## 2. Followup judge
- [ ] 2.1 `internal/ai/followup`: `judge_ai_followup` handler — read post + parent AI comment + user reply, call lightweight model, parse `{should_reply,reason}` JSON
- [ ] 2.2 Safe-default false on: timeout, non-JSON, missing field, non-boolean, call failure (one test per anomaly class)
- [ ] 2.3 On true: enqueue `generate_ai_reply`(FOLLOWUP)
- [ ] 2.4 Register `judge_ai_followup` Asynq handler in worker bootstrap

## 3. Followup guards
- [ ] 3.1 Comment create: detect parent is AI + author is real user → enqueue `judge_ai_followup`
- [ ] 3.2 AI→AI reply does NOT trigger followup (author-must-be-real-user guard)
- [ ] 3.3 ≤3 FOLLOWUP/agent/post cap (query `ai_reply_tasks` WHERE trigger_type=FOLLOWUP)
- [ ] 3.4 Tests: AI→AI no task; cap enforced; real-user→AI enqueues judge

## 4. Verification
- [ ] 4.1 AUTO (P6/P7) / MENTION / FOLLOWUP trigger types all distinct and routed correctly through the 4-col unique key
- [ ] 4.2 `go test ./internal/ai/followup... ./internal/forum/comment/...` green
- [ ] 4.3 `go build ./cmd/api ./cmd/worker`; `go vet ./...` / `govulncheck ./...` clean

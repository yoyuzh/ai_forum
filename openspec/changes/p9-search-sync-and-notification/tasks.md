# P9 Tasks — Search index sync + notification

## 1. notifications migration
- [x] 1.1 `000017_notifications` (+down): recipient, type, payload JSON, read_at, created_at

## 2. ES index + mapping
- [x] 2.1 `forum_contents` index mapping: post/comment/ai_comment doc types; IK analyzer for Chinese fields (title/body/tags/name/content)
- [x] 2.2 Idempotent index creation on startup

## 3. sync_search_index handler
- [x] 3.1 `internal/search`: re-fetch from MySQL by aggregate ID (never use payload for content)
- [x] 3.2 Upsert document on `post.created/updated`, `comment.created`, `ai.reply.completed`
- [x] 3.3 Delete document on `post.deleted`, `comment.deleted`; update/remove on `post.moderated`
- [x] 3.4 Consume `ai.reply.failed` (ack/log, no ES write) — closes unowned-event gap
- [x] 3.5 Register handler + bind `q.search.index` consumer in worker bootstrap
- [x] 3.6 Tests: created post searchable within 1–3s (polling); delete removes from search; rebuild==incremental

## 4. ES-down chaos
- [x] 4.1 Test: kill ES container → post create still returns 200 and persists in MySQL; sync task retries without blocking writes

## 5. send_notification handler
- [x] 5.1 `internal/notification`: recipient determination (comment.created → author+mentioned; ai.reply.completed → author; user.mentioned → mentioned user)
- [x] 5.2 Write `notifications` rows; idempotent via `processed_events`
- [x] 5.3 Register handler + bind `q.notification` consumer
- [x] 5.4 Tests: AI reply notifies author; redelivery no duplicate

## 6. Verification
- [x] 6.1 P5 contract-ownership test still green, and P13 implementation-completeness expectations for `comment.deleted`/`ai.reply.failed`/`post.moderated` are satisfied
- [x] 6.2 `go test ./internal/{search,notification}/...` green (integration + chaos tagged)
- [x] 6.3 `make migrate-up` applies 000017; `go build ./cmd/worker`; `go vet ./...` / `govulncheck ./...` clean

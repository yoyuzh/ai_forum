# P10 Tasks — Hot score pipeline

## 1. Redis key scheme + hot-path hooks
- [ ] 1.1 `internal/cache`: hot keys `post:{id}:{view,like,comment,ai_reply}_count`, `post:{id}:hot_score`, `hot_posts:zset`, `dirty_hot_posts:set`
- [ ] 1.2 `forum/post/hot_score.go`: real impl — INCR counters, SADD dirty, recompute hot_score, ZADD zset (no MySQL hot_score write on hot path)
- [ ] 1.3 Wire hot-path hooks into like/comment/view/AI-reply actions (extend P4 like, extend P7 reply count, extend comment create, post view)

## 2. refresh_hot_score cron
- [ ] 2.1 `internal/task`: `refresh_hot_score` Asynq cron every 30s — SMEMBERS dirty (batch ≤200), MGET counters, batch UPDATE MySQL posts counters+hot_score, SREM processed
- [ ] 2.2 If dirty > batch, next round continues
- [ ] 2.3 Register cron in worker bootstrap

## 3. Formula + rebuild
- [ ] 3.1 Pure formula function (§6.10.3) with caller-supplied hours_since_created (no time.Now in business code)
- [ ] 3.2 Unit tests: hand-computed fixtures per coefficient
- [ ] 3.3 Rebuild from MySQL: recent 7-day NORMAL posts → recompute → ZADD hot_posts:zset
- [ ] 3.4 Test: Redis flush → rebuild restores zset within snapshot lag

## 4. Concurrent-load latency test (critique fix)
- [ ] 4.1 Test: N=100 parallel likes on one post → measure per-request latency → assert p99 < target
- [ ] 4.2 Assert zero `UPDATE posts SET hot_score` on the hot path (only the cron writes) — capture via query/log spy

## 5. Verification
- [ ] 5.1 Hot path: Redis counters+zset updated, no MySQL hot_score write
- [ ] 5.2 Cron: dirty set drains to MySQL within 30s
- [ ] 5.3 `go test ./internal/forum/post/... ./internal/cache/...` green (integration tagged)
- [ ] 5.4 `go build ./cmd/worker`; `go vet ./...` / `govulncheck ./...` clean

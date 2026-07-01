## Why

Post hot-score must not update `posts.hot_score` on every like/comment/view (high-frequency writes → row-lock contention on popular posts). Architecture §6.10 specifies the design: Redis real-time counters + a sorted-set hot board + a 30s Asynq cron that snapshots the dirty set back to MySQL. This phase implements that pipeline and replaces the critique's vague "no lock contention" exit criterion with a concrete concurrent-load p99 latency test proving the hot path doesn't degrade under parallel likes.

## What Changes

- Redis key scheme (§6.10.1): `post:{id}:view_count`/`like_count`/`comment_count`/`ai_reply_count`/`hot_score`, `hot_posts:zset`, `dirty_hot_posts:set`.
- Hot-path hooks: like/comment/view/AI-reply actions increment Redis counters + `SADD dirty_hot_posts` + recompute `post:{id}:hot_score` + `ZADD hot_posts:zset` (no MySQL `hot_score` write on the hot path).
- `refresh_hot_score` Asynq cron (30s): drain a batch (≤200) from `dirty_hot_posts:set`, read Redis counters, batch-update MySQL `posts` counters + `hot_score`, remove IDs from the dirty set.
- Hot-score formula (§6.10.3): `base_score = like*2 + comment*3 + ai_reply*2 + view*0.1`; `hot_score = base_score / pow(hours_since_created+2, 1.2)`.
- Rebuildability: Redis-flush rebuild from MySQL (recent 7-day NORMAL posts) into `hot_posts:zset`.
- Tests: formula unit test; 30s snapshot flush; Redis-flush rebuild; **concurrent-load p99 latency test** (N parallel likes, assert p99 < target and no MySQL `hot_score` write on hot path).

## Capabilities

### New Capabilities
- `hot-score-pipeline`: Redis-counter-driven hot score with 30s MySQL snapshot cron, rebuildable from MySQL, with concurrency-latency verification.

### Modified Capabilities
- `forum-like`: EXTENDS P4 like to increment Redis hot counters + dirty-set + recompute (no MySQL hot_score write on hot path).

## Impact

- **Code**: `backend/internal/forum/post/hot_score.go` (real impl), `backend/internal/cache` (hot keys), worker bootstrap (`refresh_hot_score` cron), `forum/like` + `forum/comment` + `forum/post` hot-path hooks.
- **Critique risk #5 closed**: concrete p99 latency test replaces vague contention assertion.
- **Consumes**: P2 Redis, P5 Asynq cron, P4 forum write path.
- **Architecture**: Redis rebuildable from MySQL (§3.3).

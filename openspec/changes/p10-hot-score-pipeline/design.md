## Context

Architecture §6.10 fixes the hot-score design. The skeleton already has `forum/post/hot_score.go`. Critique flagged "no row-lock contention, verified by write-count test" as unverifiable — write-count does not prove absence of contention; a concurrent-load p99 latency test does. P4's like path currently writes MySQL directly; P10 moves the hot-path to Redis and snapshots to MySQL on a cron.

## Goals / Non-Goals

**Goals:**
- Redis counter + zset + dirty-set hot path.
- 30s `refresh_hot_score` cron snapshot to MySQL.
- Formula unit test; rebuild test; concurrent-load p99 test.
- No MySQL `hot_score` write on the hot path.

**Non-Goals:**
- No recommendation/personalization (v2).
- No multi-instance Redis hot board (v1 single Redis).
- No view-count accuracy guarantees beyond eventual consistency.

## Decisions

### D1: Hot path is Redis-only
Like/comment/view/AI-reply actions: `INCR post:{id}:<counter>`, `SADD dirty_hot_posts {id}`, recompute `post:{id}:hot_score`, `ZADD hot_posts:zset score {id}`. No `UPDATE posts SET hot_score` on the hot path — that's the whole point (avoids row-lock contention).

### D2: 30s cron snapshot
`refresh_hot_score` runs every 30s: `SMEMBERS dirty_hot_posts` (batch ≤200), `MGET` counters, batch `UPDATE posts` counters+hot_score, `SREM` processed IDs. If dirty set > batch, next round continues (§6.10.5).

### D3: Formula exactly per §6.10.3
`base_score = like*2 + comment*3 + ai_reply*2 + view*0.1`; `hot_score = base_score / pow(hours_since_created+2, 1.2)`. `hours_since_created` passed in (no `time.Now` in business code; caller stamps). Unit test with hand-computed fixtures.

### D4: Rebuild from MySQL (§6.10.6)
On Redis flush, rebuild `hot_posts:zset` from MySQL: read recent 7-day NORMAL posts, recompute `hot_score` from MySQL snapshot counters, `ZADD`. A test asserts the rebuilt zset matches the pre-flush one (within snapshot lag).

### D4a: Empty Redis still recovers on the next interaction
If Redis is empty and `dirty_hot_posts:set` is also empty, the next like/comment/view/AI-reply still creates the counter keys, adds the post to `dirty_hot_posts:set`, updates `hot_posts:zset`, and is picked up by the next snapshot cron. This covers cold-start recovery in addition to explicit rebuild.

### D5: Concurrent-load p99 test (critique fix)
A test issues N parallel likes against a single popular post, measures per-request latency, asserts p99 < target and that **zero** MySQL `hot_score` writes occurred on the hot path (only the cron writes). This is the concrete form of "no lock contention."

## Risks / Trade-offs

- **[Risk] Redis flush loses hot board** → Mitigation: D4 rebuild from MySQL; test.
- **[Risk] Redis starts empty and no rebuild has run yet** → Mitigation: D4a interaction repopulates counters/dirty set so the cron resumes snapshots.
- **[Risk] 30s MySQL lag mislabels a hot post** → Mitigation: §6.10.7 — feed/detail can read Redis counters; MySQL is a periodic snapshot.
- **[Risk] Counter drift between Redis and MySQL** → Mitigation: cron reconciles; Redis is rebuildable from MySQL counts (themselves snapshotted).
- **[Risk] p99 test flaky in CI** → Mitigation: run against local Redis; deterministic N; assert relative not absolute latency.

## Migration Plan

1. Redis keys → hot-path hooks → cron → rebuild → tests.
2. Rollback: hot path reverts to direct MySQL (P4 baseline); zset can be dropped.

## Open Questions

- p99 latency target v1 — propose 50ms per like on local Redis; finalized at implementation.

## ADDED Requirements

### Requirement: Redis hot path with no MySQL hot_score write
Like, comment, view, and AI-reply actions SHALL increment Redis counters (`post:{id}:view_count`/`like_count`/`comment_count`/`ai_reply_count`), add the post to `dirty_hot_posts:set`, recompute `post:{id}:hot_score`, and update `hot_posts:zset`. The hot path SHALL NOT write `posts.hot_score` to MySQL.

#### Scenario: Like updates Redis only
- **WHEN** a user likes a popular post
- **THEN** Redis counters and `hot_posts:zset` are updated and no `UPDATE posts SET hot_score` is issued on the hot path

### Requirement: 30s snapshot cron
A `refresh_hot_score` Asynq cron SHALL run every 30 seconds, drain a batch (≤200) from `dirty_hot_posts:set`, read Redis counters, batch-update MySQL `posts` counters and `hot_score`, and remove processed IDs from the dirty set.

#### Scenario: Dirty set drains to MySQL
- **WHEN** the cron runs and `dirty_hot_posts` contains post IDs
- **THEN** the corresponding MySQL `posts` rows have their counters and `hot_score` updated and the IDs are removed from the dirty set

### Requirement: Hot-score formula
`hot_score` SHALL equal `(like_count*2 + comment_count*3 + ai_reply_count*2 + view_count*0.1) / pow(hours_since_created+2, 1.2)` (§6.10.3).

#### Scenario: Formula matches fixture
- **WHEN** counters and hours_since_created are fixed
- **THEN** the computed `hot_score` equals the hand-computed value

### Requirement: Rebuildable from MySQL
After a Redis flush, the system SHALL rebuild `hot_posts:zset` from MySQL (recent 7-day NORMAL posts) using the same formula.

#### Scenario: Rebuild restores the hot board
- **WHEN** Redis is flushed and the rebuild runs
- **THEN** `hot_posts:zset` is repopulated from MySQL snapshot data

### Requirement: Concurrent-load latency verification
Under N parallel likes on a single post, the per-request p99 latency SHALL remain below the configured target, and zero MySQL `hot_score` writes SHALL occur on the hot path.

#### Scenario: Parallel likes do not contend
- **WHEN** 100 parallel likes are issued against one post
- **THEN** p99 latency is below the target and no `UPDATE posts SET hot_score` is observed during the parallel run

## MODIFIED Requirements

### Requirement: Like and unlike
A user SHALL be able to like and unlike a post. Liking SHALL increment the post's like count; unliking SHALL decrement it. A user cannot like the same post twice (uniqueness on `(user_id, post_id)`). Liking/unliking SHALL additionally increment/decrement the Redis `post:{id}:like_count` counter, add the post to `dirty_hot_posts:set`, recompute `post:{id}:hot_score`, and update `hot_posts:zset` — without writing `posts.hot_score` to MySQL on the hot path (the 30s cron owns the MySQL snapshot).

#### Scenario: Like then unlike
- **WHEN** a user likes a post and then unlikes it
- **THEN** the like count returns to its prior value, no duplicate like row exists, and Redis counters reflect the net change

#### Scenario: Duplicate like rejected
- **WHEN** a user likes a post they already liked
- **THEN** the request is a no-op (or 409) and the count does not double

# Module Instructions

## Responsibility

Own post creation, editing, deletion, querying, and hot-score field domain logic only.

## Owns

- `posts` table access through the post repository.
- Post DTOs, handlers, services, models, post domain events, and hot-score helpers.
- Creation of `outbox_events(post.created)` during post creation transactions.

## Must Not

- Do not directly call AI reply, search sync, notification, or moderation concrete implementations.
- Do not publish RabbitMQ messages inside the post transaction.
- Do not frequently update MySQL `posts.hot_score` during high-frequency views, likes, or comments.

## Allowed Dependencies

- Forum-local interfaces, `database`, `common`, `cache` hot-score contracts, `task` contracts, and `outbox` writer interfaces.
- No direct dependency on AI/search/notification concrete packages.

## Communication Rules

- After creating a post, write only the post data and `outbox_events(post.created)`.
- AI tagging, AI decisions, AI replies, search sync, notifications, and moderation are downstream async consumers.
- Hot-score work must enter through agreed cache/task paths.

## Data Rules

- MySQL `posts` is authoritative.
- Redis may hold view/like/comment counters and hot-score zsets; values must be rebuildable from MySQL snapshots.
- `post.created` payload should contain necessary IDs, not full post objects.

## Testing Rules

- Test that create-post transactions write `posts` plus `outbox_events(post.created)`.
- Test that no direct AI/search/notification dependency is introduced.
- Test hot-score calculations separately from high-frequency write paths.

## Notes for Codex

- If a change feels convenient because it calls AI/search/notify directly, it belongs elsewhere.

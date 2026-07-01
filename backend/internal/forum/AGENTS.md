# Module Instructions

## Responsibility

Own forum domain modules: posts, comments, tags, likes, and favorites.

## Owns

- Forum submodule files.
- Forum domain service/repository contracts.
- Forum-owned tables such as posts, comments, post tags, likes, and favorites.

## Must Not

- Do not place AI reply generation, search indexing, notifications, or moderation implementation here.
- Do not directly publish RabbitMQ messages from business transactions.

## Allowed Dependencies

- `database`, `common`, `cache`, and `outbox` interfaces.
- Other domains only through explicit service interfaces.

## Communication Rules

- Forum writes domain events to `outbox_events`; downstream modules react asynchronously.
- Forum modules call each other through local service/repository boundaries.

## Data Rules

- MySQL owns forum core state.
- Hot counters may use Redis but must be recoverable and periodically flushed/snapshotted.

## Testing Rules

- Test post/comment lifecycle and event creation around transactions.
- Test idempotent behavior for duplicate async side effects when applicable.

## Notes for Codex

- Keep `PostService` focused on post domain behavior; do not add AI/search/notification shortcuts.

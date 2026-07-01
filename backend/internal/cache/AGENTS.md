# Module Instructions

## Responsibility

Own Redis cache, rate limiting, hot-score counters, and Asynq broker connection helpers.

## Owns

- Redis clients, cache key conventions, hot-score counter helpers, rate-limit primitives.

## Must Not

- Do not treat Redis as authoritative business storage.
- Do not hide durable writes in cache helpers.

## Allowed Dependencies

- `config`, `logger`, and shared serialization primitives.

## Communication Rules

- Expose narrow interfaces for modules needing cache/counter/rate-limit behavior.

## Data Rules

- Redis data may be lost and must be recoverable from MySQL.
- Hot-score updates should be buffered/counted and flushed by task paths.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

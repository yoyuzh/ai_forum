## Context

P4 appends `outbox_events` in-tx but never publishes. Architecture §8 (outbox), §7.3/7.4 (exchanges/queues), §7.5 (payload envelope), §7.2 (task types), §9.2/9.3 (`processed_events` + daily cleanup cron) fix the contract. Critique found the 8th task `cleanup_processed_events` and several events (`comment.deleted`, `ai.reply.failed`, `post.moderated`) were unowned; this phase records explicit ownership metadata so the gap can't silently recur, while later phases implement most consumers/handlers.

## Goals / Non-Goals

**Goals:**
- outbox-publisher real loop with retry/FAILED + graceful shutdown.
- Full MQ topology (exchanges/queues/bindings) + reconnect-safe publisher.
- Event registry + payload envelope; all 9 task-type constants; `cleanup_processed_events` daily cron registered.
- `processed_events` dedup helper all consumers reuse.
- Contract-ownership test (CI gate).

**Non-Goals:**
- No task handler implementations (P6 tag/decision, P7 reply, P8 followup, P9 search/notify, P10 hot score) — only the `cleanup_processed_events` cron runs (its handler is trivial delete).
- No SSE push (P7), no AI logic.

## Decisions

### D1: Publisher owns the scan loop, not the business write
`outbox-publisher` reads `PENDING` rows in batches every 500ms–1s, publishes each with the correct routing key, marks `PUBLISHED` on ack, increments `retry_count` on failure, marks `FAILED` after max retries (§8.3). It is the only process that publishes domain events to RabbitMQ — business code never publishes in-tx (already enforced in P4).

### D2: Reconnect-safe publisher
`mq` publisher wraps `amqp091-go` with channel-pool + reconnect on drop. P2 constructed the connection; P5 adds the publisher resilience. Publishes are confirmed (publisher confirms) where the broker supports it.

### D3: Topology declarations are idempotent
On startup, `mq.DeclareTopology` declares exchanges/queues/bindings with `durable=true`. Idempotent — re-running is safe. Matches §7.3/7.4 exactly.

### D4: event + task registries as typed constants
`event.Types` is a const set; `event.NewEnvelope(aggregateType, aggregateId, eventType, payload)` builds the §7.5 envelope with a uuid `event_id` and `occurredAt` passed in (no `time.Now` in business code — caller stamps). `task.Types` is the 9-element const set. Compile-time guarantee that typos can't happen.

### D5: processed_events dedup helper
`event.MarkProcessed(tx, eventID, consumerName)` inserts into `processed_events`; unique-key conflict = already processed → caller acks. `event.IsProcessed(ctx, eventID, consumerName)` checks. Both RabbitMQ consumers and Asynq handlers use this (Asynq also has its own `Unique`, but `processed_events` is the cross-cutting source of truth per §9.2).

### D6: cleanup_processed_events cron registered here
`task.RegisterCleanupCron` schedules `cleanup_processed_events` daily (§9.3: `DELETE FROM processed_events WHERE processed_at < NOW() - INTERVAL 30 DAY`). The handler is trivial and ships now.

### D7: Contract-ownership test
A test enumerates the §8.5 event set and §9.3 cron set and asserts each has publisher-owner, consumer-owner, and handler-owner metadata. Owners may point to later phases for implementation. P13 runs the stricter implementation-completeness check after all phases land. This keeps P5 shippable while preserving the durable guard against unowned events/tasks.

## Risks / Trade-offs

- **[Risk] Publisher crashes mid-publish → duplicate or lost event** → Mitigation: publish-then-mark-PUBLISHED; on restart, PENDING rows are republished (at-least-once); consumers must be idempotent (D5). Duplicate publish is safe because `processed_events` dedups.
- **[Risk] RabbitMQ redelivery storm** → Mitigation: D5 dedup; P13 concurrent-redelivery idempotency load test.
- **[Risk] New event added later without an owner** → Mitigation: D7 ownership test fails CI; P13 implementation-completeness check fails if the owner never implements it.
- **[Risk] Asynq `Unique` TTL conflicts with `processed_events`** → Mitigation: `processed_events` is the source of truth; Asynq `Unique` used only for short de-dup windows, not correctness.

## Migration Plan

1. event/task registries → mq topology → publisher loop → processed_events helper → cleanup cron → ownership test.
2. Wire publisher into `cmd/outbox-publisher/main.go`.
3. Tests: write→queue; redelivery idempotency; ownership metadata.
4. Rollback: stop publisher; messages drain; `outbox_events` rows stay PENDING (safe).

## Open Questions

- Publisher batch size and scan interval v1 defaults — 500ms interval, batch 100 (finalize at implementation).

## Why

P4 wrote `outbox_events` rows in-tx but never published them — the forum write path is dark (no AI reply, no search sync, no notification). This phase lights up the event backbone: the outbox-publisher scan/publish loop, RabbitMQ exchange/queue topology, Asynq task-type contracts, and consumer idempotency via `processed_events`. It also owns the **contract-ownership test** that guarantees every architecture §8.5 event and §9.3 cron (including the previously-unowned 8th task `cleanup_processed_events`) has an explicit implementation owner. Later phases implement most consumers/handlers; P13 verifies implementation completeness in CI.

## What Changes

- `internal/outbox`: the **publisher loop** — scan `PENDING` rows every 500ms–1s, publish to RabbitMQ, mark `PUBLISHED` / increment `retry_count` / mark `FAILED` after max retries. Graceful shutdown finishes in-flight publish. Wire into `cmd/outbox-publisher/main.go`.
- `internal/event`: all domain event type constants (§8.5: `post.created/updated/deleted`, `comment.created/deleted`, `post.tagged`, `ai.reply.completed`, `ai.reply.failed`, `post.moderated`, `user.mentioned`) and payload envelope (§7.5: `eventId`/`eventType`/`aggregateType`/`aggregateId`/`occurredAt`/`payload`).
- `internal/mq`: declare exchanges (`forum.events`, `ai.events`, `notification.events`, `dead.exchange`) and queues with routing bindings per §7.3/7.4. Reconnect-safe publisher.
- `internal/task`: all Asynq task type constants (§7.2: `tag_post`, `decide_ai_reply`, `generate_ai_reply`, `judge_ai_followup`, `moderate_ai_reply`, `sync_search_index`, `send_notification`, `refresh_hot_score`, `cleanup_processed_events`). Handlers are NOT implemented here (P6–P10); only the contract + the `cleanup_processed_events` daily cron registration (§9.3).
- `processed_events` idempotency: consumer-side dedup helper (insert `(event_id, consumer_name)`; unique conflict = already processed → ack). Pre-declares the idempotency pattern all later consumers use.
- Contract-ownership test: assert every §8.5 event type has publisher-owner and consumer-owner metadata, and every §9.3 cron has handler-owner metadata; fail CI on a missing owner.
- Tests: forum write → RabbitMQ queue receives a message with correct routing key; redelivery is idempotent (processed_events dedup); all task constants compile.

## Capabilities

### New Capabilities
- `outbox-publisher`: Scans `outbox_events` and publishes to RabbitMQ with retry/FAILED semantics and graceful shutdown.
- `domain-events`: Domain event type registry and payload envelope (§7.5) shared by publishers and consumers.
- `mq-topology`: RabbitMQ exchanges, queues, and routing bindings (§7.3/7.4).
- `task-contracts`: Asynq task-type constants and the `cleanup_processed_events` daily cron registration.
- `consumer-idempotency`: `processed_events`-based dedup pattern for RabbitMQ consumers and Asynq handlers.
- `event-contract-ownership`: Test guaranteeing every documented event/cron has an explicit owner; P13 verifies final implementation completeness.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `backend/internal/{outbox,event,mq,task}/*.go`, `cmd/outbox-publisher/main.go` (real loop), `internal/bootstrap` (wire publisher + cron).
- **Systems**: RabbitMQ exchanges/queues now declared; `cleanup_processed_events` runs daily.
- **Ownership**: owns event/task contracts + publisher loop + idempotency helper; consumers implemented in P6–P10.
- **Critique gaps closed**: 8th task `cleanup_processed_events` owner; contract-ownership CI gate.

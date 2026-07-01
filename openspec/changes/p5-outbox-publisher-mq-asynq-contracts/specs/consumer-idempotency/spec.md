## ADDED Requirements

### Requirement: processed_events dedup
The `event` package SHALL provide `MarkProcessed(tx, eventID, consumerName)` and `IsProcessed(ctx, eventID, consumerName)` backed by the `processed_events` table. A unique-key conflict on `(event_id, consumer_name)` SHALL be treated as "already processed" (idempotent success), not an error.

#### Scenario: Redelivery is idempotent
- **WHEN** the same event is delivered twice to the same consumer
- **THEN** the second delivery detects the prior `processed_events` row and acks without re-executing side effects

### Requirement: Idempotency for Asynq handlers
Asynq handlers SHALL use the same `processed_events` dedup so that retried tasks do not duplicate side effects. A unique-key conflict is treated as idempotent success.

#### Scenario: Retried task does not duplicate
- **WHEN** an Asynq task is retried after partial success
- **THEN** the `processed_events` check prevents duplicate work

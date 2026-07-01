## ADDED Requirements

### Requirement: In-transaction outbox append
`outbox.Append(tx DBTX, event)` SHALL insert an `outbox_events` row with `status='PENDING'`, a unique `event_id`, the event type, aggregate type/id, JSON payload, and `created_at`. It SHALL execute on the provided transaction so the event commits or rolls back with the business write.

#### Scenario: Append shares the caller's transaction
- **WHEN** `outbox.Append` is called inside a `RunInTx` that subsequently returns an error
- **THEN** the appended `outbox_events` row is rolled back alongside the business write

#### Scenario: Append does not publish
- **WHEN** `outbox.Append` inserts a row
- **THEN** no RabbitMQ publish occurs (publishing is the P5 publisher's responsibility)

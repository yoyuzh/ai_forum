## ADDED Requirements

### Requirement: Outbox publisher scan loop
The outbox-publisher process SHALL scan `outbox_events` rows with `status='PENDING'` every 500ms–1s, publish each to RabbitMQ with the correct routing key, mark `PUBLISHED` on success, increment `retry_count` on failure, and mark `FAILED` after the configured max retries. It is the only process that publishes domain events to RabbitMQ.

#### Scenario: Pending row is published and marked
- **WHEN** a `PENDING` outbox row exists and the publisher scans it
- **THEN** the message reaches the RabbitMQ queue with the correct routing key and the row becomes `PUBLISHED`

#### Scenario: Failed publish retries then FAILs
- **WHEN** publishing a row fails repeatedly past max retries
- **THEN** the row is marked `FAILED` and `retry_count` equals max retries

### Requirement: Graceful shutdown of publisher
On SIGTERM the publisher SHALL stop scanning new rows, finish in-flight publishes, then exit. Unfinished rows remain `PENDING` for the next run.

#### Scenario: In-flight publish completes on shutdown
- **WHEN** SIGTERM arrives during a publish
- **THEN** the in-flight publish completes (or is left PENDING) and the process exits without losing the row

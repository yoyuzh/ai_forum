## ADDED Requirements

### Requirement: send_notification writes notification rows
The `send_notification` handler SHALL determine recipients from the triggering event (`comment.created` → post author + mentioned users; `ai.reply.completed` → post author; `user.mentioned` → mentioned user) and write `notifications` rows. It SHALL be idempotent via `processed_events`.

#### Scenario: AI reply notifies post author
- **WHEN** an `ai.reply.completed` event is consumed
- **THEN** a `notifications` row is created for the post author

#### Scenario: Redelivery does not duplicate notifications
- **WHEN** the same event is redelivered
- **THEN** `processed_events` prevents a duplicate notification row

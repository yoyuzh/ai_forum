## ADDED Requirements

### Requirement: Asynq task type registry
The `task` package SHALL define constants for every architecture §7.2 task type: `tag_post`, `decide_ai_reply`, `generate_ai_reply`, `judge_ai_followup`, `moderate_ai_reply`, `sync_search_index`, `send_notification`, `refresh_hot_score`, and `cleanup_processed_events`. Handler implementations are deferred to later phases except `cleanup_processed_events`.

#### Scenario: All task constants compile
- **WHEN** the `task` package is compiled
- **THEN** all nine task-type constants are present and typed

### Requirement: cleanup_processed_events daily cron
The system SHALL register an Asynq periodic task `cleanup_processed_events` that runs daily and deletes `processed_events` rows older than 30 days (§9.3).

#### Scenario: Old processed_events are purged daily
- **WHEN** the cleanup task runs and rows older than 30 days exist
- **THEN** those rows are deleted

## ADDED Requirements

### Requirement: Four-column generated-column unique key
The `ai_reply_tasks` table SHALL define `parent_comment_id_norm BIGINT GENERATED ALWAYS AS (COALESCE(parent_comment_id,0)) STORED` and a unique key `uk_ai_reply_task(post_id, parent_comment_id_norm, ai_agent_id, trigger_type)` — four columns. A unique-key conflict on insert SHALL be treated as idempotent success, not a task failure.

#### Scenario: Concurrent insert yields one task
- **WHEN** two `generate_ai_reply` tasks for the same (post, parent_comment, agent, trigger_type) race to insert
- **THEN** exactly one `ai_reply_tasks` row is created and the conflicting insert is treated as idempotent success (no FAILED status, no duplicate comment)

### Requirement: Business-layer dedup is the primary defense
Before creating a task, the handler SHALL query `ai_reply_tasks` by `(post_id, COALESCE(parent_comment_id,0), ai_agent_id, trigger_type)`. Existing PENDING/RUNNING/SUCCESS/BLOCKED/SKIPPED rows prevent creation; FAILED rows are not auto-recreated.

#### Scenario: Existing successful task blocks recreation
- **WHEN** a task for the same key already has status SUCCESS
- **THEN** no new task is created

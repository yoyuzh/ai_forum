## ADDED Requirements

### Requirement: Moderation gates AI replies
AI reply output SHALL pass moderation before being persisted as a visible comment. On moderation failure, the handler SHALL NOT write a `comments` row, SHALL mark the `ai_reply_tasks` row BLOCKED, and SHALL not retry (§9.4). Moderation SHALL be behind an interface so the implementation can evolve.

#### Scenario: Blocked reply is not persisted
- **WHEN** the model output fails moderation
- **THEN** no `comments` row is created, the task is BLOCKED, and no visible AI comment appears

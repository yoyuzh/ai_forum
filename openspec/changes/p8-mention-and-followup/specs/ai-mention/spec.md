## ADDED Requirements

### Requirement: @AI mention bypasses willingness
When a user comment mentions one or more AI agents, the system SHALL validate each mentioned agent exists and has `allowMention=true`, write `comment_mentions`, enforce the per-comment (≤3 AI) and per-user (≤5 @AI/minute) rate limits, and create `ai_reply_tasks` with `trigger_type=MENTION` plus enqueue `generate_ai_reply` — without computing willingness scores.

#### Scenario: Mention enqueues a reply
- **WHEN** a user comment mentions an enabled, mention-allowed AI
- **THEN** a `generate_ai_reply` task with `trigger_type=MENTION` is enqueued and no willingness score is computed

#### Scenario: Rate limit blocks excess mentions
- **WHEN** a user exceeds 5 @AI mentions in one minute
- **THEN** the request is rejected (429) and no task is created

#### Scenario: Disabled or mention-disallowed agent skipped
- **WHEN** a comment mentions a disabled agent or one with `allowMention=false`
- **THEN** no task is created for that agent

#### Scenario: Per-comment mention cap
- **WHEN** a single comment mentions more than 3 AI agents
- **THEN** the request is rejected

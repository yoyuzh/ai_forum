## ADDED Requirements

### Requirement: Followup judge with safe-default false
The `judge_ai_followup` handler SHALL call a lightweight model and parse a structured JSON `{should_reply, reason}`. It SHALL enqueue `generate_ai_reply` with `trigger_type=FOLLOWUP` only when `should_reply=true`. On any anomaly — model timeout, non-JSON response, missing `should_reply` field, non-boolean value, or call failure — it SHALL default to `should_reply=false` and end without enqueuing.

#### Scenario: Model returns true
- **WHEN** the model returns `{"should_reply":true,"reason":"..."}`
- **THEN** a `generate_ai_reply` FOLLOWUP task is enqueued for the parent AI agent

#### Scenario: Timeout defaults false
- **WHEN** the model call times out
- **THEN** no task is enqueued and the handler ends

#### Scenario: Non-JSON defaults false
- **WHEN** the model returns non-JSON
- **THEN** no task is enqueued

#### Scenario: Missing field defaults false
- **WHEN** the JSON lacks `should_reply` or it is non-boolean
- **THEN** no task is enqueued

### Requirement: Followup guards
Followup SHALL only be triggered when the parent comment is an AI comment and the replying author is a real user. AI SHALL not reply to AI. The same agent SHALL make at most 3 followup replies per post.

#### Scenario: AI-to-AI reply does not trigger followup
- **WHEN** an AI comment replies to another AI comment
- **THEN** no `judge_ai_followup` task is created

#### Scenario: Followup cap enforced
- **WHEN** an agent already has 3 FOLLOWUP replies on a post
- **THEN** no further FOLLOWUP task is enqueued for that agent on that post

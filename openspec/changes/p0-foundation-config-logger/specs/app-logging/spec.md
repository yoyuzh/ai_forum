## ADDED Requirements

### Requirement: Structured zap logger
`logger.New(cfg Log) (*zap.Logger, error)` SHALL return a zap logger using the JSON encoder when `cfg.Encoding == "json"` and the console encoder otherwise, with level from `cfg.Level`.

#### Scenario: Production JSON encoding
- **WHEN** `New(Log{Level:"info",Encoding:"json"})` is called
- **THEN** the emitted log lines are valid JSON objects containing the message and level fields

### Requirement: Contextual field helpers
The logger package SHALL provide a way to attach contextual fields (`event_id`, `task_id`, `user_id`, `request_id`, `post_id`, `comment_id`, `ai_agent_id`, `trigger_type`) to a logger instance, returning a child logger bound with those fields.

#### Scenario: Child logger carries fields
- **WHEN** a child logger is created with `user_id=12` and `request_id="r1"`
- **THEN** every subsequent log entry from that child includes both fields

### Requirement: Secret redaction
The logger SHALL support a redaction mode that masks any field whose name matches a configured redact set (e.g. `token`, `password`, `secret`, `api_key`) to the literal `***` in output. The full `INTERNAL_API_TOKEN` value SHALL never appear in logs.

#### Scenario: Token field is redacted
- **WHEN** a log entry is written with a field named `token` whose value is a 64-char hex string and redaction is enabled
- **THEN** the emitted output contains `***` in place of the token value, not the raw hex string

## ADDED Requirements

### Requirement: Performance gates
The system SHALL meet LCP < 2.5s and CLS < 0.1 on the web feed, web post detail, and admin dashboard (Lighthouse). INP < 200ms SHALL be measured via real Playwright interaction traces on the post-detail AI status flow, not via Lighthouse TBT estimation.

#### Scenario: CWV targets met
- **WHEN** perf gates run on the key screens
- **THEN** LCP < 2.5s, CLS < 0.1, and real-interaction INP < 200ms

### Requirement: Accessibility gates
axe-core SHALL report no critical violations on key web and admin screens. WCAG-AA contrast SHALL hold for Cohere text/background pairings. The reduced-motion path SHALL not break AI status updates.

#### Scenario: a11y gates pass
- **WHEN** axe-core and contrast scans run
- **THEN** no critical violations and contrast meets WCAG AA; reduced-motion AI status still updates

### Requirement: Security gates
`govulncheck ./...` and `npm audit` SHALL be clean or document accepted advisories. The `/internal` denial test SHALL pass.

#### Scenario: Vuln scan clean
- **WHEN** `govulncheck` and `npm audit` run
- **THEN** no unpatched critical advisories remain (or accepted ones are documented)

### Requirement: AI-call structured logs
AI model calls SHALL emit structured worker logs with `task_id`, `task_type`, `post_id`, `ai_agent_id`, `trigger_type`, `model`, `latency_ms`, `status`, `retry_count`, and `error_message` when present. Logs SHALL NOT include prompt bodies, API keys, internal tokens, or full secrets.

#### Scenario: Model call log is useful and redacted
- **WHEN** a fake AI model call succeeds or fails
- **THEN** the structured log contains the required operational fields
- **AND** no prompt body, API key, internal token, or full secret appears

### Requirement: Idempotency-under-load gate
Concurrent duplicate injection of the same `post.tagged` / `generate_ai_reply` event SHALL result in exactly one decision and one AI comment per `(post, agent, trigger_type)`, exercising both `processed_events` and the 4-column unique key.

#### Scenario: Duplicate events produce one reply
- **WHEN** the same event is injected N times concurrently
- **THEN** exactly one decision log and one AI comment exist for that key

### Requirement: Migration rollback gate
A `migrate-down` + `migrate-up` cycle on a populated DB SHALL leave data consistent. A fresh-DB migrate SHALL succeed in CI.

#### Scenario: Rollback cycle is consistent
- **WHEN** migrations are rolled back and re-applied on a populated DB
- **THEN** the schema and surviving data are consistent

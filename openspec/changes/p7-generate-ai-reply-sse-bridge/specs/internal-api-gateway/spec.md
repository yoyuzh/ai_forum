## MODIFIED Requirements

### Requirement: Token-authenticated internal API receiver
api-server SHALL expose `POST /internal/posts/{postId}/events` protected by an `X-Internal-Token` middleware. The middleware SHALL compare the token in constant time against `cfg.InternalAPI.Token`; on missing or mismatched token it SHALL return 401 and log a structured security event containing `request_id`, `path`, `client_ip`, `user_agent`, and `reason`, with the token value redacted (never the full token). The receiver SHALL dispatch the event to the SSE Hub (extended in P7 from the P3 no-op Hub to a real in-memory Hub).

#### Scenario: Valid token accepted and dispatched
- **WHEN** the request carries the correct `X-Internal-Token` and an `ai_reply_completed` body
- **THEN** the receiver dispatches the event to the SSE Hub, which pushes it to the post's subscribed clients

#### Scenario: Missing token rejected and logged
- **WHEN** the request carries no `X-Internal-Token`
- **THEN** it returns 401 and a security log entry is emitted whose token field is redacted

#### Scenario: Wrong token rejected
- **WHEN** the request carries an incorrect `X-Internal-Token` (equal or unequal length)
- **THEN** it returns 401 and a redacted security log entry is emitted

## ADDED Requirements

### Requirement: Token-authenticated internal API receiver
api-server SHALL expose `POST /internal/posts/{postId}/events` protected by an `X-Internal-Token` middleware. The middleware SHALL compare the token in constant time against `cfg.InternalAPI.Token`; on missing or mismatched token it SHALL return 401 and log a structured security event containing `request_id`, `path`, `client_ip`, `user_agent`, and `reason`, with the token value redacted (never the full token).

#### Scenario: Valid token accepted
- **WHEN** the request carries the correct `X-Internal-Token`
- **THEN** the receiver proceeds (dispatching to the Hub; no-op until P7)

#### Scenario: Missing token rejected and logged
- **WHEN** the request carries no `X-Internal-Token`
- **THEN** it returns 401 and a security log entry is emitted whose token field is redacted

#### Scenario: Wrong token rejected
- **WHEN** the request carries an incorrect `X-Internal-Token` (equal or unequal length)
- **THEN** it returns 401 and a redacted security log entry is emitted

### Requirement: Network isolation of internal API
The `/internal/**` path SHALL NOT be reachable through the public Nginx proxy. `deploy/nginx.conf` SHALL return 404 for `location /internal/`. In docker-compose, api-server SHALL use `expose:` (not `ports:`) so it is not directly host-accessible, and worker-service SHALL `depends_on` api-server.

#### Scenario: Nginx blocks internal path
- **WHEN** a public request hits `https://<host>/internal/posts/1/events`
- **THEN** Nginx returns 404 without forwarding to api-server

#### Scenario: api-server not host-exposed
- **WHEN** docker-compose is inspected
- **THEN** the api-server service defines `expose` and no host-level `ports` mapping

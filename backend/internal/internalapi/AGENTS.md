# Module Instructions

## Responsibility

Own internal network endpoints used only between backend processes.

## Owns

- `/internal/**` handlers, middleware, DTOs, and internal client contracts.

## Must Not

- Do not accept Cookie authentication.
- Do not use user JWT authentication.
- Do not expose internal routes through public Nginx.
- Do not add general-purpose service-to-service APIs.

## Allowed Dependencies

- `sse`, `config`, `logger`, `common`, and auth middleware for internal token validation.

## Communication Rules

- Current v1.0 allowed flow is only `worker-service -> api-server` for SSE event push at `/internal/posts/{postId}/events`.
- Every request must validate `X-Internal-Token`.

## Data Rules

- Internal API should pass only the data needed for event push.
- It must not become a bypass for domain writes.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

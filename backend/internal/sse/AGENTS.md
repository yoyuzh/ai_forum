# Module Instructions

## Responsibility

Own v1.0 single api-server in-memory SSE Hub connection management and event push.

## Owns

- SSE Hub, client registry, event formatting, and post-scoped push helpers.

## Must Not

- Do not generate AI replies.
- Do not persist business data.
- Do not implement multi-instance realtime distribution in v1.0.

## Allowed Dependencies

- `logger`, `common`, and typed event/status DTOs.

## Communication Rules

- API routes register clients with the in-memory Hub.
- Internal API pushes events into the Hub after validating `X-Internal-Token`.

## Data Rules

- SSE state is in-memory and ephemeral.
- Disconnect compensation is handled by `GET /api/posts/{postId}/ai-status`.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

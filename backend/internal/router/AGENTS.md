# Module Instructions

## Responsibility

Own HTTP route registration and middleware composition.

## Owns

- API route grouping, middleware attachment, and handler registration.

## Must Not

- Do not put business logic in route registration.
- Do not expose `/internal/**` through public route groups.

## Allowed Dependencies

- Handler interfaces from modules, `auth`, `rbac`, `internalapi`, `sse`, `logger`, `config`.

## Communication Rules

- Router wires HTTP to handlers only; handlers/services own behavior.

## Data Rules

- Router does not own persistent data.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

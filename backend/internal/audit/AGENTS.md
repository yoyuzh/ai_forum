# Module Instructions

## Responsibility

Own audit logging for admin and sensitive operations.

## Owns

- Audit event models, repositories, and append helpers.

## Must Not

- Do not replace structured application logs.
- Do not perform business state transitions.

## Allowed Dependencies

- `auth`, `user`, `database`, `logger`, and domain event metadata interfaces.

## Communication Rules

- Domain modules call audit interfaces for sensitive actions when required.

## Data Rules

- Audit records should be append-only and preserve actor/action/resource metadata.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

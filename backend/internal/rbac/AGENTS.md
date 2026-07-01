# Module Instructions

## Responsibility

Own RBAC policy evaluation and admin/user permission checks.

## Owns

- RBAC models, Casbin integration, permission constants, and enforcement helpers.

## Must Not

- Do not rely on frontend permission hiding for security.
- Do not mix authentication token parsing with authorization decisions unless through explicit auth interfaces.

## Allowed Dependencies

- `auth`, `user`, `database`, `logger`, and config.

## Communication Rules

- Expose authorization interfaces to router/handlers.
- Admin UI permissions are display hints only; backend RBAC is authoritative.

## Data Rules

- Own policy/role-permission data and audit-relevant authorization decisions.

## Testing Rules

- Test idempotency, retry behavior, and boundary rules when handlers/adapters are implemented.

## Notes for Codex

- Keep infrastructure packages free of unrelated domain shortcuts.

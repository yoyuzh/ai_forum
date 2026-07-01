# Module Instructions

## Responsibility

Own authentication primitives, JWT handling, and current-user context helpers.

## Owns

- Auth token handling.
- Auth middleware contracts.
- Login/session helper boundaries.

## Must Not

- Do not perform RBAC policy decisions here.
- Do not accept user JWTs for `/internal/**` routes.

## Allowed Dependencies

- User read interfaces, config, logger, common, and database where appropriate.

## Communication Rules

- Expose authenticated subject context to handlers and RBAC.

## Data Rules

- Auth state must not replace user account state in MySQL.

## Testing Rules

- Test token validation, expiration, malformed token handling, and context extraction.

## Notes for Codex

- Internal process authentication belongs to `internalapi`.

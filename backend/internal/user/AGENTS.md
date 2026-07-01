# Module Instructions

## Responsibility

Own user account domain behavior and user profile persistence.

## Owns

- User service/repository boundaries.
- User account and profile table access.

## Must Not

- Do not own RBAC policy evaluation.
- Do not own forum, AI, notification, or audit behavior.

## Allowed Dependencies

- Database, common, logger, auth contracts, and RBAC interfaces where appropriate.

## Communication Rules

- Expose user read/write interfaces to auth, RBAC, and admin handlers.

## Data Rules

- MySQL user records are authoritative.
- Other modules may reference user IDs but must not mutate user data directly.

## Testing Rules

- Test user state transitions and repository constraints when implemented.

## Notes for Codex

- Keep user identity separate from permissions and audit records.

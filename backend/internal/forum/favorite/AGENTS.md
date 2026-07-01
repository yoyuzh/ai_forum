# Module Instructions

## Responsibility

Own user favorite/bookmark behavior for posts.

## Owns

- `favorites` table access and related DTOs, handlers, services, repositories, and models.

## Must Not

- Do not own unrelated forum subdomain behavior.
- Do not directly call AI/search/notification concrete implementations.

## Allowed Dependencies

- Forum-local contracts and shared infrastructure through explicit interfaces.

## Communication Rules

- Use service/repository boundaries inside the process.
- Use outbox events for asynchronous downstream effects when needed.

## Data Rules

- Own only the tables and records for this subdomain.
- Treat duplicate async side effects as idempotent where applicable.

## Testing Rules

- Add focused unit tests for service behavior and repository integration tests when persistence is implemented.

## Notes for Codex

- Keep the package cohesive; do not use it as a shared forum utility bucket.

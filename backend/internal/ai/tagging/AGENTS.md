# Module Instructions

## Responsibility

Own AI-assisted post tag generation.

## Owns

- Tagging prompt assembly, tag generation tasks, and tag result persistence boundaries.

## Must Not

- Do not own unrelated AI submodule behavior.
- Do not bypass AI package-level moderation and idempotency rules.

## Allowed Dependencies

- AI package interfaces and shared infrastructure.
- Cross-domain calls only through explicit service/repository interfaces.

## Communication Rules

- Use Asynq tasks for executable AI work and RabbitMQ events only as facts that occurred.

## Data Rules

- Persist only the data owned by this AI subdomain.
- Keep generated data auditable and idempotent.

## Testing Rules

- Add unit tests for rules and integration tests for persistence or queue behavior when implemented.

## Notes for Codex

- Keep responsibilities separated among AI subpackages.

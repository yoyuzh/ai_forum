# Module Instructions

## Responsibility

Own small shared primitives that are truly cross-cutting.

## Owns

- Generic errors.
- Pagination and response primitives.
- Small time or validation helpers.

## Must Not

- Do not turn this directory into a dumping ground.
- Do not add domain-specific logic here.

## Allowed Dependencies

- Standard library and tiny shared dependencies only.

## Communication Rules

- Shared primitives should be imported, not used to bypass module ownership.

## Data Rules

- This module owns no tables or queues.

## Testing Rules

- Test helpers with edge cases.

## Notes for Codex

- If a helper mentions posts, comments, AI, search, or notifications, it likely belongs in that domain.

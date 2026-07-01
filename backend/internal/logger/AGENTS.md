# Module Instructions

## Responsibility

Own structured logging setup and logger helpers.

## Owns

- Logger initialization.
- Shared logging field conventions.

## Must Not

- Do not log secrets or full sensitive payloads.
- Do not implement business side effects.

## Allowed Dependencies

- Config and selected structured logging library.

## Communication Rules

- Expose logger construction and narrow helpers to other modules.

## Data Rules

- Logs are observability data and must not replace durable audit records.

## Testing Rules

- Test redaction helpers if added.

## Notes for Codex

- Prefer structured fields for event IDs, task IDs, user IDs, and request IDs.

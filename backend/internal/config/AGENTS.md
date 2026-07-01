# Module Instructions

## Responsibility

Own backend configuration loading and typed config structs.

## Owns

- Config structs.
- Environment and file binding helpers.

## Must Not

- Do not read business data.
- Do not hide secrets in source files.

## Allowed Dependencies

- Standard library and configuration libraries selected by the backend.

## Communication Rules

- Provide typed configuration to bootstrap and infrastructure modules.

## Data Rules

- Configuration values are runtime inputs, not business state.

## Testing Rules

- Test defaults, required fields, and invalid config handling.

## Notes for Codex

- Keep environment names aligned with `.env.example`.

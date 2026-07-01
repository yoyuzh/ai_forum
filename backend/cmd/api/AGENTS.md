# Module Instructions

## Responsibility

Own the `api-server` process entrypoint.

## Owns

- `main.go` wiring for HTTP API, auth, RBAC, router, SSE, and internal API endpoints.

## Must Not

- Do not execute AI generation, search indexing jobs, or notification jobs inline.
- Do not expose `/internal/**` externally.

## Allowed Dependencies

- Backend bootstrap/config/logger/database/router modules required to start this process.

## Communication Rules

- Entrypoints wire dependencies only; business behavior belongs in internal modules.

## Data Rules

- Entrypoints must not own business tables directly.

## Testing Rules

- Keep startup wiring testable through small constructors once implementation starts.

## Notes for Codex

- Keep `main.go` minimal and delegate setup to bootstrap packages.

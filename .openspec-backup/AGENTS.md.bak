# Module Instructions

## Responsibility

This repository is the AI Forum modular monolith. It contains the Go backend, React user web app, React Refine admin app, deployment assets, scripts, tools, and architecture documents.

## Owns

- Repository-level project structure and conventions.
- `backend/`, `web/`, `admin/`, `docs/`, `deploy/`, `scripts/`, and `tools/` boundaries.
- Docker Compose and top-level developer commands.

## Must Not

- Do not turn v1.0 into many microservices.
- Do not add backend Go processes beyond `api-server`, `worker-service`, and `outbox-publisher` without an architecture update.
- Do not make same-process internal modules call each other over HTTP.
- Do not let business modules bypass module boundaries to modify another module's data.
- Do not put AI, search, notification, or moderation logic into `PostService` for convenience.

## Allowed Dependencies

- Same-process backend modules may depend on each other only through explicit interfaces, services, repositories, and dependency injection.
- Cross-business asynchronous flows must use Outbox + RabbitMQ + Asynq.
- MySQL is the primary data source; Redis is cache, counters, rate limiting, and queue infrastructure; Elasticsearch is the search read model.

## Communication Rules

- This project is a modular monolith, not a microservice system.
- Backend v1.0 has exactly three Go processes: `api-server`, `worker-service`, and `outbox-publisher`.
- Same-process modules communicate through interfaces and injected dependencies, not HTTP.
- RabbitMQ events express what happened; Asynq tasks express what should run next.
- The only allowed internal HTTP path is `worker-service -> api-server` for SSE Hub notification at `/internal/posts/{postId}/events`, authenticated by `X-Internal-Token`.
- `/internal/**` must never be proxied publicly by Nginx.

## Data Rules

- MySQL is the strong-consistency source of truth.
- Elasticsearch is eventually consistent and must not be used for business decisions.
- Redis data may be lost and must be recoverable from MySQL.
- Reliable domain events must be written to `outbox_events` before RabbitMQ publishing.
- Do not publish RabbitMQ messages directly inside business transactions.

## Testing Rules

- Prefer module-level unit tests for business boundaries.
- Use integration tests for database migrations, outbox publishing, RabbitMQ consumers, Asynq handlers, Redis behavior, Elasticsearch indexing, and SSE/internal API flows.
- Tests for asynchronous handlers must cover retry and idempotency behavior.

## Notes for Codex

- Keep high cohesion, low coupling, and single responsibility as hard constraints.
- When implementing a feature, read the nearest `AGENTS.md` first and obey the most specific rule.
- Do not delete or move `ai_forum_requirements_v2.md` or `ai_forum_architecture_v1.md`.

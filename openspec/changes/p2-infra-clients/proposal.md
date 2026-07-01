## Why

P1 gave us MySQL. The event-driven architecture depends on four more infra clients — Redis (cache/limiter/hot-score/Asynq broker), RabbitMQ (domain events), Elasticsearch (search read-model with IK Chinese analysis), and Asynq (task scheduler) — plus a Casbin model file for RBAC. Every worker task (P6+) and every event publish (P5) needs these connections. Landing them as pinging, round-trip-tested clients now means later phases wire behavior, not plumbing.

## What Changes

- Implement `internal/cache` (Redis): `NewRedis(cfg)` client with ping.
- Implement `internal/mq` (RabbitMQ): `NewRabbitMQ(cfg)` returning a connection + channel, with reconnect-safe construction.
- Implement `internal/search` (Elasticsearch): `NewES(cfg)` client with ping + an IK-analyzer presence probe that **fails the healthcheck** (not merely warns) when IK is absent.
- Implement `internal/task` (Asynq): `NewAsynqClient(cfg)` (enqueuer) + `NewAsynqServer(cfg)` (worker) using the Redis broker; both pings verified.
- Add `internal/rbac` Casbin model file (`model.conf`) only — no enforcement yet (enforcement lands in P4). Pin the model shape (sub/obj/act).
- Smoke tests (build tag `integration`) proving each client connects and round-trips against docker-compose services.

## Capabilities

### New Capabilities
- `redis-client`: Redis connection used for cache, rate limiting, hot-score counters, and Asynq broker.
- `rabbitmq-client`: RabbitMQ connection and channel for domain event publishing and consuming.
- `elasticsearch-client`: Elasticsearch client with IK Chinese-analyzer presence verification.
- `asynq-task-client`: Asynq client (enqueuer) and server (worker) bound to the Redis broker.
- `casbin-model`: Casbin authorization model definition (no enforcement yet — enforcement is P4).

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `backend/internal/{cache,mq,search,task,rbac}/*.go`, `backend/internal/rbac/model.conf`, integration tests.
- **Systems**: Requires running redis, rabbitmq, elasticsearch (with IK plugin) from docker-compose.
- **Dependencies**: clients pinned in P0.
- **Non-functional**: ES IK absence is a deployment gate, not a soft warning.

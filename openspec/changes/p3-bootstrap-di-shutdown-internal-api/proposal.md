## Why

P0–P2 gave us config, logger, MySQL, Redis, RabbitMQ, ES, Asynq, and a Casbin model — but as disconnected pieces. The three processes (`cmd/api`, `cmd/worker`, `cmd/outbox-publisher`) still have 2-line `main.go` files. This phase wires dependency injection, graceful shutdown, and the security boundary for the only allowed internal HTTP path (`worker → api /internal/posts/{postId}/events`). After P3, all three binaries build, start, report `/healthz` + `/readyz`, and reject unauthenticated `/internal/**` requests. Domain routes land in P4.

## What Changes

- Implement `internal/bootstrap`: a DI composition root that constructs shared deps (config, logger, db, redis, mq, es, asynq) and per-process wiring for `api`, `worker`, `outbox-publisher`. Same-process modules communicate via explicit interfaces/DI — never HTTP.
- Implement graceful shutdown for each process (SIGTERM/SIGINT) with a timeout and concrete goroutine-leak assertion (per critique: "no goroutine leak" must be a real assertion). api-server drains in-flight HTTP + closes SSE; worker stops consuming; outbox-publisher finishes in-flight publishes.
- Implement `internal/router`: `/healthz` (liveness) and `/readyz` (readiness checking dep pings) for api-server.
- Implement `internal/internalapi`: the `POST /internal/posts/{postId}/events` receiver **skeleton** (hub dispatch only; SSE Hub is built in P7) plus `X-Internal-Token` middleware that returns `401` on missing/invalid token and logs a structured security event (redacted token, §13.4).
- Harden `deploy/nginx.conf`: `location /internal/ { return 404; }` and api-server uses `expose:` (not `ports:`) in docker-compose.
- Wire the three `cmd/*/main.go` files to call bootstrap.

## Capabilities

### New Capabilities
- `app-bootstrap`: Dependency-injection composition root wiring shared infrastructure and per-process components for the three Go processes.
- `graceful-shutdown`: Ordered, timeout-bounded shutdown on SIGTERM/SIGINT with goroutine-leak verification for all three processes.
- `health-endpoints`: `/healthz` (liveness) and `/readyz` (dependency-aware readiness) for api-server.
- `internal-api-gateway`: Token-authenticated receiver for worker→api internal SSE notifications, with nginx/docker-compose network isolation.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `backend/internal/{bootstrap,router,internalapi}/*.go`, `backend/cmd/*/main.go`, `deploy/nginx.conf`, `docker-compose.yml`.
- **Architecture constraint**: enforces "same-process modules never HTTP each other; only worker→api `/internal/posts/{postId}/events`" and "`/internal/**` never publicly proxied".
- **Security**: `X-Internal-Token` middleware + structured security log; api-server not host-exposed.
- **Files owned/extended**: P3 OWNS `internal/internalapi/*` and `internal/sse` skeleton; P7 EXTENDS them (does not recreate) per critique risk 2.

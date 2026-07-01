## Context

P0–P2 are disconnected primitives. The three `cmd/*/main.go` are 2-line placeholders. Architecture §4 fixes process responsibilities; §6.9.1 fixes the internal-API security model (Docker network isolation + nginx not proxying `/internal/` + api-server not host-exposed + `X-Internal-Token`); §16 fixes graceful shutdown per process. Critique flagged that "no goroutine leak" must be a concrete assertion and that P7 must extend (not recreate) the internalapi/sse files owned here.

## Goals / Non-Goals

**Goals:**
- DI composition root; three binaries build and start.
- `/healthz` + `/readyz` on api-server.
- `POST /internal/posts/{postId}/events` skeleton with `X-Internal-Token` middleware + structured security log.
- Graceful shutdown with timeout and goroutine-count assertion.
- nginx `/internal/` 404; api-server `expose:` not `ports:`.

**Non-Goals:**
- No business routes (P4), no SSE Hub dispatch logic (P7 — P3 only mounts the receiver skeleton and a no-op hub), no outbox-publisher scan loop (P5 — P3 only starts/stops it).
- No RBAC enforcement on `/internal` — it uses the internal token, not user JWT/cookies.

## Decisions

### D1: Composition root in `internal/bootstrap`
A `bootstrap` package exposes `NewApp(cfg)` returning a struct holding shared deps and three constructors: `NewAPIServer`, `NewWorker`, `NewOutboxPublisher`. Each `cmd/*/main.go` calls the relevant constructor, starts it, and waits on SIGTERM. Modules receive deps via constructor injection (interfaces), never via package-level globals — enforces "same-process modules never HTTP each other."

### D2: Shutdown with concrete goroutine assertion
Each process implements `Stop(ctx) error` running in a bounded `context.WithTimeout`. A test asserts `runtime.NumGoroutine()` before start, after start, and after `Stop` returns within tolerance — the critique-required concrete form of "no goroutine leak."

### D3: `/readyz` checks real dependency liveness
`/readyz` pings MySQL/Redis/RabbitMQ/ES (api-server) and reports 503 if any required dep is down. `/healthz` is process liveness only (200). Critique: readiness must not falsely claim success when deps are down.

### D4: Internal-API skeleton + ownership
`internal/internalapi` owns the route + token middleware + a `Hub` interface with a no-op default. P7 implements the real SSE hub and dispatch by **extending** these files, not creating new ones. Token middleware: constant-time compare against `cfg.InternalAPI.Token`; on failure return 401 and log `request_id`/`path`/`client_ip`/`user_agent`/`reason` with the token redacted (per §13.4 — never log the full token).

### D5: Network isolation in docker-compose + nginx
`docker-compose.yml`: `api-server` uses `expose: ["8080"]` (no `ports:`). `deploy/nginx.conf`: `location /internal/ { return 404; }`. `worker-service` `depends_on: [api-server]`.

## Risks / Trade-offs

- **[Risk] `/readyz` lies because deps are down but process is up** → Mitigation: D3 pings real deps; test kills MySQL and asserts `/readyz` 503.
- **[Risk] P7 recreates internalapi/sse causing file collision** → Mitigation: ownership documented here; P3 defines the `Hub` interface P7 implements.
- **[Risk] Token compare not constant-time** → Mitigation: `subtle.ConstantTimeCompare`; test asserts unequal-length and equal-length rejects.
- **[Risk] Shutdown timeout too short kills in-flight work** → Mitigation: configurable timeout with documented v1 defaults; outbox-publisher finishes in-flight publish before exit (P5 owns the loop; P3 owns the shutdown harness).

## Migration Plan

1. Implement bootstrap + router + internalapi skeleton + shutdown.
2. Wire three `main.go`.
3. Harden nginx + docker-compose.
4. Tests: shutdown goroutine assertion; `/readyz` 503 on dep-down; token 401 + redacted log.
5. Rollback: revert `main.go` to placeholders; remove packages.

## Open Questions

- Shutdown timeout v1 default value — propose 15s api, 30s worker (task drain), 10s outbox-publisher; finalized at implementation.

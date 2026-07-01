# P3 Tasks — Bootstrap, DI, shutdown, internal-API gateway

## 1. Bootstrap / DI
- [ ] 1.1 `internal/bootstrap/bootstrap.go`: `NewApp(cfg)` shared deps + `NewAPIServer`/`NewWorker`/`NewOutboxPublisher` constructors
- [ ] 1.2 Constructor-injection interfaces for modules; no package-level globals; no same-process HTTP
- [ ] 1.3 Wire `cmd/api/main.go`, `cmd/worker/main.go`, `cmd/outbox-publisher/main.go` to call bootstrap

## 2. Health endpoints
- [ ] 2.1 `internal/router/router.go`: `GET /healthz` (200 process-alive) and `GET /readyz` (200 only when MySQL/Redis/RabbitMQ/ES reachable, else 503 with failing-dep body)
- [ ] 2.2 Test: kill MySQL → `/readyz` returns 503

## 3. Internal API gateway
- [ ] 3.1 `internal/internalapi`: `POST /internal/posts/{postId}/events` route + `X-Internal-Token` middleware (`subtle.ConstantTimeCompare`)
- [ ] 3.2 Define `Hub` interface (no-op default implementation owned here; P7 extends, not recreates)
- [ ] 3.3 On auth failure: 401 + structured security log (`request_id`/`path`/`client_ip`/`user_agent`/`reason`, token redacted)
- [ ] 3.4 `deploy/nginx.conf`: `location /internal/ { return 404; }`
- [ ] 3.5 `docker-compose.yml`: api-server `expose: ["8080"]` (no `ports`); worker `depends_on: api-server`
- [ ] 3.6 Tests: valid token accepted; missing/empty/wrong token (equal & unequal length) → 401 + redacted log; grep confirms no full token in log output

## 4. Graceful shutdown
- [ ] 4.1 Each process: `Stop(ctx)` with `context.WithTimeout`; api stops HTTP + closes SSE; worker stops consuming, drains in-flight tasks; outbox-publisher finishes in-flight publish
- [ ] 4.2 Test: `runtime.NumGoroutine()` before start / after start / after `Stop` within tolerance (concrete goroutine-leak assertion)
- [ ] 4.3 Test: in-flight work exceeding timeout logs abandoned work and exits

## 5. Verification
- [ ] 5.1 `go build ./cmd/api ./cmd/worker ./cmd/outbox-publisher` green; all three start against docker-compose
- [ ] 5.2 `/healthz` 200, `/readyz` 200 when deps up / 503 when a dep down
- [ ] 5.3 `/internal/**` returns 401 without token, 404 via nginx, 200 with token
- [ ] 5.4 `go vet ./...` clean; `govulncheck ./...` clean

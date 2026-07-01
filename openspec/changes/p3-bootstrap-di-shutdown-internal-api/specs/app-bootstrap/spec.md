## ADDED Requirements

### Requirement: DI composition root
`internal/bootstrap` SHALL construct shared infrastructure (config, logger, db, redis, mq, es, asynq) once and provide per-process constructors for `api`, `worker`, and `outbox-publisher`. Dependencies SHALL be passed to modules via constructor injection (explicit interfaces), never via package-level globals or same-process HTTP.

#### Scenario: Three binaries build and start
- **WHEN** `go build ./cmd/api ./cmd/worker ./cmd/outbox-publisher` and each binary is started against docker-compose
- **THEN** all three start without error and log a startup line with their process role

### Requirement: Same-process modules never HTTP each other
No module within a single process SHALL call another module in the same process over HTTP. The only permitted inter-process HTTP path is `worker-service → api-server` at `POST /internal/posts/{postId}/events`.

#### Scenario: No same-process HTTP wiring
- **WHEN** the bootstrap composition root is inspected
- **THEN** module-to-module communication uses injected interfaces, not HTTP clients

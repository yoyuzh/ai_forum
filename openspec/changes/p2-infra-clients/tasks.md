# P2 Tasks — Infra clients

## 1. Redis
- [x] 1.1 `internal/cache/cache.go`: `NewRedis(cfg config.Redis) (*redis.Client, error)`
- [x] 1.2 Integration smoke: Set/Get round-trip

## 2. RabbitMQ
- [x] 2.1 `internal/mq/mq.go`: `NewRabbitMQ(cfg config.RabbitMQ)` → connection + channel
- [x] 2.2 Integration smoke: declare temp queue, publish, consume round-trip

## 3. Elasticsearch
- [x] 3.1 `internal/search/search.go`: `NewES(cfg config.Elasticsearch) (*es.Client, error)` with ping
- [x] 3.2 IK presence probe via `_analyze` with `ik_smart`; absence fails readiness (not just warns)
- [x] 3.3 docker-compose ES uses IK-enabled image or installs IK in entrypoint; document choice

## 4. Asynq
- [x] 4.1 `internal/task/task.go`: `NewAsynqClient(cfg)` (enqueuer) + `NewAsynqServer(cfg)` (worker) on shared Redis broker
- [x] 4.2 Integration smoke: register trivial test handler, enqueue, assert processed, unregister

## 5. Casbin model
- [x] 5.1 `internal/rbac/model.conf`: `r = sub, obj, act` matching §12.2 permission set
- [x] 5.2 Unit test: model loads with trivial in-memory policy and evaluates without error

## 6. Verification
- [x] 6.1 `go test -tags=integration ./internal/{cache,mq,search,task}/...` green against docker-compose
- [x] 6.2 `go build ./...` / `go vet ./...` clean
- [x] 6.3 No exchange/queue topology, no task constants, no enforcement logic introduced yet (deferred to P4/P5)

## Context

Architecture §3.3 fixes infra responsibilities: Redis (cache/limiter/hot-score/Asynq broker), RabbitMQ (domain events), Elasticsearch (final-consistency search read-model, Chinese via IK), Asynq (task scheduler over Redis broker). These clients must be pinging and round-trip-tested before P5 (event publish) and P6 (worker tasks). Casbin RBAC enforcement is P4, but the model file belongs here so P4 only adds enforcement.

## Goals / Non-Goals

**Goals:**
- One constructable, pingable client per infra service.
- IK analyzer presence verification that gates deployment (fails healthcheck, not just warns).
- Asynq client + server both bound to the Redis broker.
- Casbin model file fixed (sub/obj/act) without enforcement.

**Non-Goals:**
- No exchange/queue topology declarations (P5), no Asynq task type constants (P5), no Casbin enforcement/middleware (P4).
- No Redis key schemes (P7/P10), no ES index mapping (P9).

## Decisions

### D1: Reconnect-safe RabbitMQ
`amqp091-go` connections drop on broker restart. Construct with a dial that surfaces a clear error; reconnect/retry behavior is wired in P5 where publishers/consumers live. P2 only proves the connection and a round-trip (declare a temp queue, publish, consume).

### D2: IK presence is a hard gate
ES without IK cannot do Chinese search. A probe (`_analyze` with `ik_smart`) MUST fail the readiness path when absent. Architecture §3.3 treats ES as rebuildable, but IK is an install-time requirement, not rebuildable data. Critique (risk 3): "or a warning" is too weak; P2 makes absence a failure.

### D3: Asynq client + server share Redis
Both Asynq enqueuer (`asynq.NewClient`) and worker (`asynq.NewServer` with `RedisClientOpt`) use the same Redis cfg. P2 proves enqueue→process round-trip with a trivial registered handler, then unregisters.

### D4: Casbin model only
`rbac/model.conf` defines `r = sub, obj, act`. No policy adapter yet — P4 decides sqlx adapter vs. in-memory. P2 just pins the model so enforcement is a drop-in.

## Risks / Trade-offs

- **[Risk] IK plugin not in base ES image** → Mitigation: docker-compose uses an IK-enabled ES image or installs IK in entrypoint; P2 healthcheck fails loudly if missing.
- **[Risk] RabbitMQ reconnect churn** → Mitigation: deferred to P5; P2 only constructs and pings.
- **[Risk] Asynq test handler left registered** → Mitigation: smoke test registers, asserts, then unregisters; no production handlers yet.

## Migration Plan

1. Implement the five clients + model file.
2. Add integration smoke tests (tag `integration`).
3. Rollback: delete packages; no schema/runtime state persists.

## Open Questions

- Which IK-enabled ES docker image to use (custom build vs. install-on-start) — resolved at implementation; documented in docker-compose.

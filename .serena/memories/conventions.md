# Conventions

- Modular monolith, not microservices. Same-process modules must not call each other through HTTP; use interfaces, services, repositories, dependency injection.
- Cross-business async flow: business transaction writes MySQL plus `outbox_events`; `outbox-publisher` publishes RabbitMQ events; `worker-service` consumes events and enqueues/executes Asynq tasks.
- RabbitMQ events mean "what happened"; Asynq tasks mean "what to do next". Do not mix those responsibilities.
- `PostService` may create/update posts and write `outbox_events(post.created)` but must not directly call AI, search, notification, or concrete implementations outside forum/post.
- MySQL is the strong-consistency source. Redis is recoverable cache/counters/queue infra. Elasticsearch is a rebuildable eventually consistent read model.
- Event consumers and Asynq handlers must be idempotent. `processed_events` records event-consumer idempotency. Unique-key conflicts that represent duplicate work are idempotent success.
- AI reply idempotency requires business de-duplication plus `ai_reply_tasks` unique constraints; when parent comments join the key, use `parent_comment_id_norm = COALESCE(parent_comment_id, 0)` to avoid MySQL NULL uniqueness pitfalls.
- Only allowed internal HTTP: `worker-service -> api-server` for SSE Hub notification at `/internal/posts/{postId}/events`, authenticated by `X-Internal-Token`; no Cookie/JWT and no public Nginx proxy.
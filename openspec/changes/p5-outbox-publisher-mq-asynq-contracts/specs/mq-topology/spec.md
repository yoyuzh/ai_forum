## ADDED Requirements

### Requirement: RabbitMQ topology
The `mq` package SHALL declare exchanges `forum.events` (topic), `ai.events` (topic), `notification.events` (topic), and `dead.exchange` (direct), plus queues and routing bindings per architecture §7.4 (`q.post.tagging`, `q.ai.decision`, `q.search.index`, `q.notification`, `q.audit.log`, `q.dead`). Declarations SHALL be idempotent and durable.

#### Scenario: Topology declared and routable
- **WHEN** a `post.created` event is published to `forum.events`
- **THEN** it reaches `q.post.tagging` (and `q.search.index`, `q.audit.log` via `post.*` bindings)

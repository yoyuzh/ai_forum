## ADDED Requirements

### Requirement: Liveness and readiness endpoints
api-server SHALL expose `GET /healthz` returning 200 when the process is alive, and `GET /readyz` returning 200 only when all required dependencies (MySQL, Redis, RabbitMQ, Elasticsearch) are reachable, and 503 otherwise.

#### Scenario: Ready when deps up
- **WHEN** `/readyz` is called and all dependencies are reachable
- **THEN** it returns 200

#### Scenario: Not ready when a dep is down
- **WHEN** `/readyz` is called and MySQL is unreachable
- **THEN** it returns 503 and a body identifying the failing dependency

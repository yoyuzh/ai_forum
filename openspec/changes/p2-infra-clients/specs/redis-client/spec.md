## ADDED Requirements

### Requirement: Redis client
`cache.NewRedis(cfg config.Redis) (*redis.Client, error)` SHALL return a connected client. `Ping` SHALL succeed against the configured Redis.

#### Scenario: Redis round-trips
- **WHEN** `NewRedis` is called against a running Redis
- **THEN** `client.Set`/`Get` of a test key round-trips successfully

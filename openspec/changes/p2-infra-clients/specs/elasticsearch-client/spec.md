## ADDED Requirements

### Requirement: Elasticsearch client with IK verification
`search.NewES(cfg config.Elasticsearch) (*es.Client, error)` SHALL return a client whose `Ping` succeeds. It SHALL also verify the IK analyzer is installed by issuing an `_analyze` request with `ik_smart`. When IK is absent, the readiness path SHALL fail, not merely log a warning.

#### Scenario: IK present
- **WHEN** `NewES` is called against an ES with the IK plugin installed
- **THEN** the client pings and the IK `_analyze` probe succeeds

#### Scenario: IK absent fails readiness
- **WHEN** `NewES` is called against an ES without IK and readiness is checked
- **THEN** the readiness check reports failure due to missing IK analyzer

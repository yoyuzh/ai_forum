## ADDED Requirements

### Requirement: Asynq client and server bound to Redis
`task.NewAsynqClient(cfg)` SHALL return an `*asynq.Client` and `task.NewAsynqServer(cfg)` SHALL return an `*asynq.Server`, both configured against the same Redis broker. A trivial test task SHALL enqueue, be processed by a registered handler, and be unregistered afterward.

#### Scenario: Enqueue and process round-trip
- **WHEN** a test task is enqueued via the Asynq client and the server has a registered handler
- **THEN** the handler executes successfully and the task is marked processed

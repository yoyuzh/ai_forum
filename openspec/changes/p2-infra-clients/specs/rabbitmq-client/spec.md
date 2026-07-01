## ADDED Requirements

### Requirement: RabbitMQ connection and channel
`mq.NewRabbitMQ(cfg config.RabbitMQ) (*Connection, error)` SHALL return a connection and a channel that can declare a queue, publish, and consume a test message.

#### Scenario: Publish and consume round-trip
- **WHEN** a temp queue is declared and a test message is published to it
- **THEN** the message is consumed within the test window

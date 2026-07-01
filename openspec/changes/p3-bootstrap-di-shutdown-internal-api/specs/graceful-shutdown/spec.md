## ADDED Requirements

### Requirement: Graceful shutdown on signal
Each process SHALL stop accepting new work on SIGTERM/SIGINT, drain in-flight work within a bounded timeout, then close connections. api-server SHALL stop accepting new HTTP requests and close SSE connections; worker SHALL stop pulling new messages/tasks but let in-flight tasks finish; outbox-publisher SHALL stop scanning new events and finish in-flight publishes.

#### Scenario: Shutdown drains and leaks no goroutines
- **WHEN** a process is started, allowed to reach steady state, then sent SIGTERM
- **THEN** `Stop` returns within the configured timeout and the live goroutine count returns to within tolerance of the pre-start count

#### Scenario: Timeout bounds shutdown
- **WHEN** in-flight work exceeds the shutdown timeout
- **THEN** the process exits after the timeout and logs which work was abandoned

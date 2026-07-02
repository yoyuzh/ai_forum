# P5 Tasks — Outbox publisher, MQ, Asynq, contracts

## 1. Event + task registries
- [x] 1.1 `internal/event/event.go`: all §8.5 event type constants + `NewEnvelope` (§7.5 shape, uuid event_id, caller-supplied occurredAt)
- [x] 1.2 `internal/task/task.go`: all 9 §7.2 task-type constants (incl. `cleanup_processed_events`)
- [x] 1.3 Unit: all constants compile; envelope shape correct; payload carries only business IDs

## 2. MQ topology + publisher
- [x] 2.1 `internal/mq`: `DeclareTopology` — exchanges `forum.events`/`ai.events`/`notification.events`/`dead.exchange` + queues/bindings per §7.4, durable, idempotent
- [x] 2.2 Reconnect-safe publisher with publisher-confirms; channel pool
- [x] 2.3 Test: publish `post.created` to `forum.events` → reaches `q.post.tagging` (+ `q.search.index`/`q.audit.log`)

## 3. Outbox publisher loop
- [x] 3.1 `internal/outbox`: scan PENDING (500ms–1s, batch ~100) → publish with correct routing key → mark PUBLISHED on ack / retry_count++ on failure / FAILED past max
- [x] 3.2 Wire real loop into `cmd/outbox-publisher/main.go` (replaces P3 start/stop harness)
- [x] 3.3 Graceful shutdown: stop scanning, finish in-flight publish, leave unfinished PENDING
- [x] 3.4 Tests: PENDING→PUBLISHED round-trip; repeated failure → FAILED; shutdown mid-publish loses no row

## 4. processed_events idempotency helper
- [x] 4.1 `internal/event`: `MarkProcessed(tx, eventID, consumerName)` + `IsProcessed(ctx, ...)`; unique conflict = idempotent success
- [x] 4.2 Test: redelivery to same consumer acks without re-executing side effects

## 5. cleanup_processed_events cron
- [x] 5.1 `internal/task`: `RegisterCleanupCron` (daily, `DELETE WHERE processed_at < NOW() - INTERVAL 30 DAY`)
- [x] 5.2 Wire into worker bootstrap
- [x] 5.3 Test: rows >30 days deleted; recent rows retained

## 6. Contract-ownership test
- [x] 6.1 Enumerate §8.5 events → assert each has publisher-owner and consumer-owner metadata (implementation may be a later phase); enumerate §9.3 crons → assert each has handler-owner metadata; fail on any missing owner

## 7. Verification
- [x] 7.1 Forum write (P4) → outbox PENDING → publisher → RabbitMQ queue receives correct routing key
- [x] 7.2 `go test ./internal/{outbox,event,mq,task}/...` green (integration tagged)
- [x] 7.3 `go build ./cmd/outbox-publisher` runs the loop; `go vet ./...` / `govulncheck ./...` clean

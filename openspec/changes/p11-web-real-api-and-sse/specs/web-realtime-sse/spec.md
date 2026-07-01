## ADDED Requirements

### Requirement: Real SSE replaces the simulator
The web app SHALL obtain AI status events via `EventSource` against `GET /api/posts/{postId}/events` when in real mode, replacing the client-side simulator. The existing emitter contract that pages depend on SHALL be preserved so page consumers are unchanged.

#### Scenario: Live AI status transitions render
- **WHEN** a post detail page is open in real mode and an AI reply completes
- **THEN** the status transitions and the AI comment appear without a manual refresh

### Requirement: Reconnect reconciles without duplication
On SSE reconnect, the client SHALL send `Last-Event-ID` and fetch `GET /api/posts/{postId}/ai-status` once to reconcile missed events. Comments SHALL be prepended by id idempotently so reconnect never duplicates an AI reply.

#### Scenario: Reconnect does not duplicate replies
- **WHEN** the SSE connection drops and reconnects after an AI reply completed
- **THEN** the AI reply appears exactly once

### Requirement: ai-status polling fallback
When SSE is unavailable, the client SHALL fall back to polling `GET /api/posts/{postId}/ai-status`.

#### Scenario: Polling fallback works
- **WHEN** SSE cannot be established
- **THEN** ai-status is polled and the UI still reflects AI progress

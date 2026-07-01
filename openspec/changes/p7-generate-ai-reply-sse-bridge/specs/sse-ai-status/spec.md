## ADDED Requirements

### Requirement: SSE endpoint streams AI status
api-server SHALL expose `GET /api/posts/{postId}/events` streaming SSE events for that post to subscribed clients. The worker SHALL notify the hub via `POST /internal/posts/{postId}/events` (P3 route, extended dispatch), and the hub SHALL push `ai_reply_completed` (and status transitions) to the post's SSE clients.

#### Scenario: AI reply pushes to SSE clients
- **WHEN** a `generate_ai_reply` task completes and the worker calls the internal endpoint
- **THEN** connected SSE clients for that post receive an `ai_reply_completed` event

### Requirement: ai-status aggregate endpoint
api-server SHALL expose `GET /api/posts/{postId}/ai-status` returning the §6.1.1 aggregate (tagging, decision, replies, overallStatus) for initial load and SSE reconnect reconciliation.

#### Scenario: ai-status reflects running state
- **WHEN** a post has one completed and one running AI reply
- **THEN** the response shows `completedCount=1`, `runningCount=1`, and `overallStatus=RUNNING`

### Requirement: Hub extends the P3 interface
The real SSE Hub SHALL implement the `Hub` interface defined in P3 by extension, not by creating a new file or route. The P3 token middleware still protects the internal route.

#### Scenario: P3 files are extended not recreated
- **WHEN** the `internal/internalapi` and `internal/sse` packages are inspected
- **THEN** the Hub implementation lives in the same files P3 created, with no duplicate route registration

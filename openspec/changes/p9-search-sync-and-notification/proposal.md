## Why

Search and notification are the two final-consistency downstream consumers. Search uses Elasticsearch as a rebuildable read-model (MySQL remains the source of truth); notification writes rows from domain events. This phase also closes the critique gap by implementing consumers for the previously-unowned outbox events `comment.deleted`, `ai.reply.failed`, and `post.moderated` — each has a producer plus consumer so the P5 contract-ownership and P13 implementation-completeness checks stay green, and the search index correctly reflects deletes and moderation.

## What Changes

- `internal/search`: `sync_search_index` Asynq handler — on `post.*`/`comment.*` events, **re-fetch from MySQL** (never trust event payload for the document), assemble the ES document, upsert/delete in the `forum_contents` index. Consumes `post.created/updated/deleted`, `comment.created/deleted`, `ai.reply.completed`, `post.moderated`.
- ES index mapping for `forum_contents` (post/comment/ai_comment types, IK analyzer for Chinese fields).
- `internal/notification`: `send_notification` Asynq handler — determine recipients, write `notifications` rows. Consumes `comment.created`, `ai.reply.completed`, `user.mentioned`.
- `notifications` migration.
- Ensure `comment.deleted`, `ai.reply.failed`, `post.moderated` events have publishers (`comment.deleted`/`post.moderated` from P4, `ai.reply.failed` from P7) and consumers here — closes the unowned-event gap.
- Tests: ES reflects writes within 1–3s (polling test); ES-down chaos test (kill ES, MySQL writes still succeed); full rebuild == incremental; notification rows generated.

## Capabilities

### New Capabilities
- `search-index-sync`: Rebuildable Elasticsearch read-model synchronized from MySQL via `sync_search_index`, including delete and moderation handling.
- `notification-generation`: `send_notification` task writing notification rows from domain events.

### Modified Capabilities
<!-- None. (`comment.deleted`/`post.moderated` publishers already exist in P4; `ai.reply.failed` publisher exists in P7. This phase adds consumers.) -->

## Impact

- **Code**: `backend/internal/{search,notification}/*.go`, `backend/migrations/000017_notifications` (+down), ES index mapping, worker bootstrap.
- **Critique gap closed**: `comment.deleted`/`ai.reply.failed`/`post.moderated` consumers.
- **Consumes**: P5 events + `processed_events`; P4/P7 publishers.
- **Architecture**: ES never used for business decisions; data rebuildable from MySQL.

## Why

This is the first business-domain phase. It delivers the synchronous forum write path that everything else hangs off: register/login (JWT), RBAC enforcement (Casbin), and the forum core (posts, comments, tags, likes, favorites) as full handler/service/repository/model/dto layers. Critically, every forum write MUST write one `outbox_events` row inside the same transaction — but **no RabbitMQ publish happens yet** (publisher lands in P5). This proves the outbox-in-tx discipline with a concrete, testable assertion before any consumer exists.

## What Changes

- `internal/user`: full layer (model/repository/service/handler/dto). Registration, profile.
- `internal/auth`: JWT issuance + validation middleware, login handler. Public-vs-private route distinction (§12.1).
- `internal/rbac`: Casbin **enforcement** (enforcer + middleware) on top of the P2 model file, with a sqlx-backed policy adapter. Frontend RBAC is visibility-only; backend is authoritative (§12.2).
- `internal/forum/{post,comment,tag,like,favorite}`: full layers. Synchronous CRUD only. `PostService` MUST NOT contain AI/search/notify/moderation logic (§5.2) — it only writes the domain row + an outbox event.
- `internal/outbox`: an `Append(tx, event)` helper that inserts an `outbox_events` row on the provided `*sqlx.Tx` (the in-tx primitive P4 consumes; the publisher loop is P5).
- Migrations for domain tables: `users`, `posts`, `comments`, `post_tags`, `comment_mentions`, `likes`, `favorites` (each `.up.sql` + `.down.sql`). **Does NOT re-create `outbox_events`/`processed_events`** (owned by P1).
- Mount public + authenticated + RBAC-protected routes in `internal/router`.
- Tests: CRUD + auth + RBAC denial; **every forum write asserts exactly one `outbox_events` row in-tx and RabbitMQ queue depth == 0**.

## Capabilities

### New Capabilities
- `user-account`: User registration, login (JWT), and profile.
- `rbac-enforcement`: Casbin-based backend authorization; backend is authoritative, frontend checks are visibility-only.
- `forum-post`: Post creation, read, update, delete (synchronous domain writes + outbox append).
- `forum-comment`: Comment creation, tree read, delete (synchronous + outbox append).
- `forum-tag`: Post tag storage and read (the `post_tags` table; AI tag *generation* is P6).
- `forum-like`: Like/unlike with count.
- `forum-favorite`: Favorite/unfavorite.
- `outbox-append`: In-transaction `outbox_events` append primitive used by synchronous domain writes.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `backend/internal/{user,auth,rbac,outbox,forum/*}/*.go`, `backend/migrations/000005..000011_*`, `backend/internal/router/router.go` (extend with business routes).
- **Ownership**: domain tables owned here; `outbox_events`/`processed_events` referenced only.
- **Architecture**: enforces "PostService holds no AI/search/notify/moderation logic" and "outbox written in-tx, no in-tx MQ publish."
- **Dependency**: consumes P1 (`RunInTx`, `outbox_events`), P2 (Casbin model), P3 (router/internalapi).

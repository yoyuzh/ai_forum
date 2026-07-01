## Context

First business phase. Architecture §5.2 forbids AI/search/notify/moderation logic in `PostService`; §8 mandates outbox in-tx; §12 fixes JWT + Casbin RBAC (backend authoritative). The skeleton already has `forum/{post,comment,tag,like,favorite}` placeholder files and `auth`/`user`/`rbac` packages. Critique flagged that P4 must own domain migrations but must NOT re-create `outbox_events`/`processed_events` (P1 owns them), and that "no row published to RabbitMQ" is structurally true now (no consumer) but should be asserted by checking queue depth == 0.

## Goals / Non-Goals

**Goals:**
- Register/login + JWT middleware; public vs authenticated routes.
- Casbin enforcement + sqlx policy adapter; RBAC denial tests.
- Forum core synchronous CRUD with outbox append in-tx.
- Admin post status/moderation-state update path that emits `post.moderated` without adding moderation policy implementation to `PostService`.
- Domain migrations (users/posts/comments/post_tags/comment_mentions/likes/favorites).
- Assert every forum write → exactly one outbox row in-tx + RabbitMQ queue empty.

**Non-Goals:**
- No AI tagging/decision/reply (P6/P7), no @AI/followup (P8), no search sync (P9), no notification (P9), no hot score (P10).
- No outbox publisher loop (P5), no MQ consumers (P5).
- No `ai_reply_tasks`/`decision_logs` tables (P6/P7).

## Decisions

### D1: PostService is thin (§5.2 enforcement)
`PostService.CreatePost` does: validate, write `posts` row, `outbox.Append(tx, post.created)`, commit. No AI, no ES, no notification, no moderation implementation. A test asserts the service does not import `ai/`, `search/`, `notification/`, or `moderation/`.

### D1a: Admin status transition emits post.moderated
P4 owns the synchronous admin route for changing a post's moderation/status field (for example `NORMAL`/`HIDDEN`/`REVIEW`). This is a forum state transition plus outbox append, not moderation policy execution. Policy decisions remain in the moderation/admin workflow. The route updates the post status and appends `post.moderated` in the same transaction so P9 has a real producer to consume.

### D2: outbox.Append(tx, event) primitive
`internal/outbox.Append` takes a `DBTX` (the tx) and an event struct, inserts into `outbox_events` with `status='PENDING'`. Domain services call it inside `RunInTx`. The publisher loop (P5) is separate. This split is what makes "in-tx write, async publish" safe.

### D3: Casbin sqlx adapter
P4 decides the adapter deferred in P2: a sqlx-backed Casbin adapter persisting policy in a `casbin_rule` table (migration owned here). Enforcement via middleware that reads the subject from the JWT context, the obj/act from the route, and calls `enforcer.Enforce`. 403 on denial.

### D4: JWT middleware + route tiers
Three tiers: public (`GET /api/posts`, `/api/posts/{id}`, `/api/ai/agents`, `/api/search` — §12.1), authenticated (write actions), RBAC-protected (admin actions). JWT middleware sets the user context; RBAC middleware wraps admin routes.

### D5: Domain migrations, no outbox/processed_events re-creation
New migrations `000005_users` … `000011_favorites`. `outbox_events`/`processed_events` are referenced by `outbox.Append`/`processed_events` reads but never re-created — a grep test asserts no `CREATE TABLE outbox_events`/`processed_events` appears in P4 migrations.

## Risks / Trade-offs

- **[Risk] PostService quietly absorbs AI logic later** → Mitigation: D1 import-guard test; code-review gate.
- **[Risk] Forum write forgets the outbox row** → Mitigation: each service's test asserts `SELECT count(*) FROM outbox_events WHERE aggregate_id=?` == 1 after a write, and that the row shares the write's tx (rollback test removes both).
- **[Risk] RBAC enforced only on frontend** → Mitigation: backend denial E2E asserts 403 on a denied action even when the frontend would show it.
- **[Risk] Comment tree read is N+1** → Mitigation: build tree in one query + in-memory assembly (skeleton already has `comment/tree.go`).

## Migration Plan

1. user/auth/rbac → forum modules → outbox.Append → router → migrations.
2. Tests per service + RBAC denial + outbox-in-tx assertion.
3. Rollback: `make migrate-down` reverses domain tables; remove packages.

## Open Questions

- Comment soft-delete vs hard-delete — v1 uses soft-delete (`deleted_at`) to preserve tree integrity; finalized at implementation.

# P4 Tasks — Identity, RBAC, forum core (synchronous + outbox append)

## 1. Domain migrations (do NOT re-create outbox/processed_events)
- [x] 1.1 `000005_users` (+down): ALTER the `users` table created by P1's `000001_init_schema` — add business columns (email, display_name, status, etc. as needed by §10.1/§12). Do NOT `CREATE TABLE users` (P1 owns it). Down reverses the ALTER only.
- [x] 1.2 `000006_posts` (+down)
- [x] 1.3 `000007_comments` (+down): comment_type, ai_agent_id nullable, parent_comment_id nullable, deleted_at
- [x] 1.4 `000008_post_tags` (+down): tag_type, tag_name
- [x] 1.5 `000009_comment_mentions` (+down)
- [x] 1.6 `000010_likes` (+down): unique (user_id, post_id)
- [x] 1.7 `000011_favorites` (+down): unique (user_id, post_id)
- [x] 1.8 `000012_casbin_rule` (+down): Casbin sqlx adapter policy table
- [x] 1.9 Grep test: no `CREATE TABLE outbox_events`/`processed_events` in P4 migrations

## 2. outbox append primitive
- [x] 2.1 `internal/outbox/outbox.go`: `Append(tx DBTX, event OutboxEvent) error` (PENDING, unique event_id, JSON payload, created_at)
- [x] 2.2 Test: Append inside RunInTx that errors → row rolled back; Append performs no MQ publish

## 3. User + Auth
- [x] 3.1 `internal/user`: model/repository/service/handler/dto — register (bcrypt), profile
- [x] 3.2 `internal/auth`: JWT issue/validate, login handler, middleware (populate subject context), 401 on invalid/expired
- [x] 3.3 Public vs authenticated route tiers per §12.1

## 4. RBAC enforcement
- [x] 4.1 `internal/rbac`: Casbin enforcer over P2 model + sqlx adapter; `Enforce(sub,obj,act)`; RBAC middleware (403 on denial)
- [x] 4.2 Permission set from §12.2 seeded for admin role
- [x] 4.3 Test: denied action returns 403 even if frontend would show it (backend authoritative)

## 5. Forum core
- [x] 5.1 `forum/post`: full layer; `CreatePost` writes posts + `outbox.Append(post.created)` in-tx; no AI/search/notify/moderation imports (import-guard test)
- [x] 5.1a `forum/post`: admin status/moderation-state update route/service updates post status + appends `post.moderated` in-tx; no moderation policy implementation in `PostService`
- [x] 5.2 `forum/comment`: full layer + `tree.go` single-query tree assembly; create → comment_count++ + `comment.created`; soft-delete → `comment.deleted`
- [x] 5.3 `forum/tag`: storage + read (grouped by type); generation deferred to P6
- [x] 5.4 `forum/like`: like/unlike, unique (user_id,post_id), count
- [x] 5.5 `forum/favorite`: favorite/unfavorite, unique (user_id,post_id)
- [x] 5.6 Per-service test: every write asserts exactly one outbox row in-tx (rollback test removes both) AND RabbitMQ queue depth == 0

## 6. Router wiring
- [x] 6.1 Extend `internal/router` with public/authenticated/RBAC route tiers; mount forum + user + auth handlers

## 7. Verification
- [x] 7.1 `go test ./internal/{user,auth,rbac,outbox,forum/...}/...` green (unit + integration)
- [x] 7.2 `make migrate-up` applies 000005–000012 cleanly; `migrate-down` reverses
- [x] 7.3 `go build ./...` / `go vet ./...` clean; `govulncheck ./...` clean

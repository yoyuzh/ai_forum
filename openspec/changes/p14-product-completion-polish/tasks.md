# P14 Tasks — Product completion polish

## 1. Scope and route cleanup
- [x] 1.1 Confirm `/admin/reports` and user report flows remain absent (P13 5.3 already enforced this); add a P14 regression assertion that no reports resource/route/menu appears in admin and no user-facing report action exists.
- [x] 1.2 Remove the web login SSO provider placeholder buttons (`LoginPage.tsx` currently renders visual-only provider buttons with no backend); drop the SSO divider/markup entirely until a real SSO phase exists.
- [x] 1.3 Add real-mode tests that fail if visible UI falls back to mock sample rows on shipped routes.

## 2. Web real AI data
- [x] 2.1 Backend: expose user-safe AI agent list endpoint (e.g. `GET /api/agents`) under `forum`/`ai` ownership, projecting only public fields (no thresholds/configs); do not reuse the admin-only `/api/admin/agents`.
- [x] 2.2 Backend: expose post-scoped decision logs, AI reply tasks, and AI activity needed by post detail (e.g. `GET /api/posts/{id}/decision-logs`, `/ai-tasks`, `/ai-activity`), user-safe projection only.
- [x] 2.3 Web: replace `realClient.agents.list/get`, `tasks.list`, `decisionLogs.list/listForPost`, and `activities.list` empty-array implementations with live calls.
- [x] 2.4 Web: update `/agents` and post-detail AI sidebar empty/loading/error states for real data.
- [x] 2.5 Tests: real-mode E2E shows backend AI agents and post decision context without mock fallback.

## 3. Web profile persistence
- [x] 3.1 Backend: add `PATCH /api/me` (extend the existing `/api/me` handler) for supported fields already present in the current profile form and backed by `users` columns; do not introduce a separate `/api/profile` route.
- [x] 3.2 Backend: add profile stats endpoint derived from MySQL rows (post/comment/like/AI-reply counts for the current user).
- [x] 3.3 Web: wire `api.user.updateProfile` and `api.user.getStats` to backend endpoints.
- [x] 3.4 Tests: profile display-name update survives reload; stats reflect created content.

## 4. Admin live dashboard
- [x] 4.1 Backend: add admin dashboard summary endpoint for users/posts/comments/AI tasks/notifications/decision logs.
- [x] 4.2 Backend: add bounded trend/recent endpoints needed by the current dashboard cards.
- [x] 4.3 Admin: replace `realApi.dashboard = mockApi.dashboard` with live calls and real empty states.
- [x] 4.4 Tests: empty DB dashboard shows zeros; seeded activity appears in dashboard cards/tables.

## 5. Admin operation boundaries
- [x] 5.1 Audit visible create/update/delete actions in admin routes and map each to a backend endpoint or read-only state.
- [x] 5.2 Hide unsupported create/delete actions instead of letting `dataProvider` throw `not implemented`.
- [x] 5.3 Keep supported actions (`agent update`, task retry/terminate/mark-processed, post status update) backed by live endpoints.
- [x] 5.4 Tests: no visible admin action throws a client-side `not implemented` error in real mode.

## 6. Search experience
- [x] 6.1 Backend: add a query method to the `search` package (`ESIndexStore.Search` / search service) that runs an ES match query against the post index and re-checks MySQL for authorization/status before returning hits; the package currently only has sync/rebuild/upsert/delete.
- [x] 6.2 Backend: register `GET /api/search/posts?q=` (or `GET /api/posts?query=`) in the router, delegating to the search service and preserving MySQL authorization/status rules.
- [x] 6.3 Web: wire feed search to the backend search surface in real mode; keep local filtering only for mock mode.
- [x] 6.4 Web: add search loading, empty, and ES-unavailable states.
- [x] 6.5 Tests: rebuilt ES post is discoverable from the web search UI; ES-down search error does not break normal feed.

## 7. Verification
- [x] 7.1 `go test ./...` and affected backend integration tests pass.
- [x] 7.2 `npm run lint` and `npm run build` pass in `web/` and `admin/`.
- [x] 7.3 Real-mode Playwright coverage passes for web AI data, profile persistence, admin dashboard, admin operation boundaries, and search.
- [x] 7.4 `openspec validate p14-product-completion-polish --strict` passes.

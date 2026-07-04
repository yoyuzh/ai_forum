# P13 Tasks — E2E, perf, a11y, security, CI

## 1. Orchestration
- [x] 1.1 `scripts/dev-up.sh`: bring up mysql/redis/rabbitmq/elasticsearch + api-server + worker + outbox-publisher + web(5173) + admin(5174); health-wait; `migrate-up` + seed
- [x] 1.2 `scripts/dev-down.sh`: teardown

## 2. E2E specs
- [x] 2.1 `e2e/tests/sanity.spec.ts`: both apps 200 + visible h1, no mock fallback (`VITE_API_MODE=real`) — runs first
- [x] 2.2 `e2e/tests/integration.spec.ts`: full AI chain (create post → willingness → decision log → worker reply → web SSE live → admin decision-log breakdown)
- [x] 2.3 `web_t*`/`admin_t*` specs run in real mode against live stack
- [x] 2.4 Architecture-constraint spec: `/internal/**` → 404 via public proxy; api-server not host-exposed
- [x] 2.5 Notification contract smoke: generated notification appears in web, unread count changes after mark-read/read-all
- [x] 2.6 Search rebuild smoke: trigger the documented rebuild entrypoint and assert rebuilt ES docs match MySQL-backed expectations

## 3. Perf gate
- [x] 3.1 Lighthouse LCP/CLS on web feed + post detail + admin dashboard
- [x] 3.2 Real Playwright-interaction INP on post-detail AI status flow (NOT Lighthouse TBT)

## 4. A11y gate
- [x] 4.1 axe-core on key web + admin screens
- [x] 4.2 WCAG-AA contrast on Cohere pairings
- [x] 4.3 Reduced-motion path does not break AI status updates

## 5. Security gate
- [x] 5.1 `govulncheck ./...` + `npm audit` (clean or documented advisories)
- [x] 5.2 `/internal` denial test passes
- [x] 5.3 Content-reporting scope guard: `/admin/reports` and user report workflow are either implemented by an explicit later phase or documented as v1 out-of-scope; no half-wired route/menu

## 6. Idempotency-under-load gate
- [x] 6.1 Concurrent duplicate `post.tagged`/`generate_ai_reply` injection → exactly one decision + one comment per (post, agent, trigger) (exercises processed_events + 4-col unique key)

## 7. Observability gate
- [x] 7.1 AI model calls emit structured logs with `task_id`, `task_type`, `post_id`, `ai_agent_id`, `trigger_type`, `model`, `latency_ms`, `status`, `retry_count`, and `error_message` when present
- [x] 7.2 AI-call log test uses a fake logger/model client; no prompt body, API key, or internal token appears in logs

## 8. Migration rollback gate
- [x] 8.1 `migrate-down` + `migrate-up` on populated DB → data consistent
- [x] 8.2 Fresh-DB migrate in CI

## 9. CI pipeline
- [x] 9.1 `.github/workflows/*`: backend (go test/vet/govulncheck/migrate-fresh/P5 contract-ownership + P13 implementation-completeness/single-table-ownership check), web+admin (lint/build), e2e (Playwright vs compose)
- [x] 9.2 Single-table migration-ownership check (no two migrations create/drop same table)

## 10. Verification
- [x] 10.1 `npm run test` in e2e/ green across web/admin/integration/sanity
- [x] 10.2 All non-functional gates pass; CI green on a sample PR
- [x] 10.3 Operational pre-flight checklist (env vars, ports, migrations, seed, search rebuild command/task) documented in `e2e/README.md`

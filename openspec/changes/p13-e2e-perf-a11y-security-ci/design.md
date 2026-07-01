## Context

Every phase P0–P12 ships its own exit criteria, but none proves the whole system together. Architecture §6.1 (post→AI chain), §6.9 (SSE), §6.9.1 (`/internal` isolation), §15 (docker compose). Critique: Lighthouse INP is unreliable (estimates TBT); use real Playwright interaction measurement; front-loaded gates (P11/P12 a11y) reduce late rework but P13 is the aggregate gate; add CI, security scan, idempotency load test, migration rollback, single-table migration ownership check.

## Goals / Non-Goals

**Goals:**
- Full docker-compose stack (3 Go processes + web + admin) green under Playwright.
- Full AI reply chain integration spec across web + admin.
- Perf (CWV + real INP), a11y (axe + contrast), security (vuln + `/internal` denial), idempotency load, migration rollback.
- CI pipeline running all gates.

**Non-Goals:**
- No new features — only verification + CI.
- No production deploy automation (v1 scope: docker compose).
- No load testing at scale beyond the idempotency burst.

## Decisions

### D1: dev-up.sh orchestration
A script brings up mysql/redis/rabbitmq/elasticsearch + api-server + worker + outbox-publisher + web(5173) + admin(5174) with health-waited startup, runs `migrate-up` + seed, then Playwright. `dev-down.sh` tears down. Sanity spec runs first to fail fast on infra issues.

### D2: Real INP, not Lighthouse TBT
Lighthouse estimates INP via TBT — unreliable. P13 measures real INP via Playwright interaction traces (click/input latency) on the post-detail AI status flow. LCP/CLS still via Lighthouse. Critique fix.

### D3: `/internal` denial as an e2e
A spec asserts `https://<host>/internal/posts/1/events` returns 404 via the public proxy (nginx) and that api-server is not host-exposed (port probe fails). Architecture §6.9.1.

### D4: Idempotency load gate
Inject the same `post.tagged` / `generate_ai_reply` event concurrently (N times) and assert exactly one decision/comment per (post, agent, trigger) — exercises `processed_events` + the 4-col unique key together.

### D5: Migration rollback gate
On a populated DB: `migrate-down` to a midpoint, `migrate-up` back, assert data consistent. Plus migrate-on-fresh-DB in CI.

### D6: CI pipeline
GitHub Actions matrix: backend (`go test`/`go vet`/`govulncheck`/migrate-fresh/P5 contract-ownership/P13 implementation-completeness/single-table-ownership check), web+admin (`npm lint`/`build`), e2e (Playwright against the compose stack). Single-table migration ownership check enforces the P1/P4/P6 ownership rule (critique risk #2).

## Risks / Trade-offs

- **[Risk] 3-process compose is heavy/flaky in CI** → Mitigation: sanity spec first; retries=2; cached images.
- **[Risk] Real INP measurement variance** → Mitigation: assert relative thresholds; run on a stable runner.
- **[Risk] Vuln scan blocks on an unpatchable advisory** → Mitigation: document accepted advisories; fail only on fixable criticals.
- **[Risk] Idempotency load false-positive** → Mitigation: deterministic event IDs; assert exact counts.

## Migration Plan

1. dev-up scripts → integration/sanity specs → gates (perf/a11y/security/idempotency/rollback) → CI.
2. Rollback: P13 is verification-only; removing it doesn't affect the system.

## Open Questions

- CI runner OS/profile for stable INP — Linux runner with fixed CPU; finalized at setup.

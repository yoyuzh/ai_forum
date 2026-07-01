## Why

This is the full-stack acceptance phase. It proves the entire system works end-to-end under docker compose (3 Go processes + web + admin), with the Playwright suite green across web/admin/integration/sanity, plus the non-functional gates the critique found missing or back-loaded: performance (CWV + real Playwright INP, not Lighthouse-estimated), accessibility (axe-core + WCAG-AA), security (govulncheck/npm audit + `/internal` denial), idempotency under concurrent load, migration rollback cycles, and CI. It is the single phase that aggregates and verifies every prior phase's exit criteria against the running stack.

## What Changes

- `scripts/dev-up.sh` / `dev-down.sh`: orchestrate the 3-process backend + web (5173) + admin (5174) for local e2e.
- `e2e/tests/integration.spec.ts`: full AI reply chain — create post in web → backend scores willingness → decision logged → worker emits reply → web SSE shows it live → admin decision-log shows the breakdown.
- `e2e/tests/sanity.spec.ts`: both apps reachable + h1 visible against the live stack (no mock fallback; `VITE_API_MODE=real`).
- `e2e/tests/web_t*.spec.ts` + `admin_t*.spec.ts`: run in real mode against the live stack.
- Architecture-constraint e2e: `/internal/**` is NOT publicly proxied (denial check).
- Performance gate: Lighthouse LCP/CLS on web feed + post detail and admin dashboard; **real Playwright-interaction INP** (not Lighthouse TBT estimate — critique).
- Accessibility gate: axe-core on key screens; reduced-motion path does not break AI status.
- Contrast gate: Cohere palette WCAG-AA.
- Security gate: `govulncheck ./...` + `npm audit` clean (or documented advisories); `/internal` denial test.
- Idempotency gate: concurrent-duplicate-event injection asserting no duplicate AI replies (processed_events + 4-col unique key).
- Migration rollback gate: `migrate-down` + `migrate-up` cycle on a populated DB.
- CI pipeline (GitHub Actions): `go test`/`go vet`/`govulncheck`, `npm lint`/`build`, Playwright, migrate-on-fresh-DB, P5 contract-ownership test plus implementation-completeness check, single-table migration ownership check.
- Operational pre-flight checklist (env vars, ports, migrations, seed).

## Capabilities

### New Capabilities
- `full-stack-e2e`: Cross-app Playwright integration proving the full AI reply chain across web + admin + 3-process backend under docker compose.
- `non-functional-gates`: Performance (CWV + real INP), accessibility (axe-core + WCAG-AA), security (vuln scan + `/internal` denial), idempotency-under-load, and migration-rollback verification.
- `ci-pipeline`: GitHub Actions running all backend/frontend/e2e gates on every change.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `e2e/tests/*`, `scripts/dev-up.sh`/`dev-down.sh`, `.github/workflows/*`, `e2e/README.md`, ops checklist.
- **Dependency**: depends on ALL prior phases (P0–P12); this is the aggregation gate.
- **Critique risks closed**: #3 (unowned events via P5 ownership + P13 implementation-completeness tests run in CI), #5 (gates front-loaded earlier; P13 is the final aggregate gate).
- **Systems**: full docker-compose stack under test.

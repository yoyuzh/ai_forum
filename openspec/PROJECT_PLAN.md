# AI Forum v1.0 — 项目计划书 (Project Plan)

> Spec-driven development plan, managed via OpenSpec. 14 phases (P0–P13), 264 tasks.
> Source of truth: `ai_forum_requirements_v2.md`, `ai_forum_architecture_v1.md`, `stitch_ai_forum/design_cohere.md`, `CLAUDE.md`, and each module's `AGENTS.md`.

## How to use this plan

Each phase is an OpenSpec **change** under `openspec/changes/pN-<slug>/` with four artifacts:
- `proposal.md` — WHY (what changes, capabilities, impact)
- `design.md` — HOW (decisions, risks, migration)
- `specs/<capability>/spec.md` — WHAT (normative SHALL/MUST requirements + scenarios = test cases)
- `tasks.md` — implementation steps (checkbox-tracked)

Commands:
```bash
openspec list                              # all phases
openspec status --change "<pN-slug>"       # phase status
openspec validate "<pN-slug>"              # validate a phase
# Implement a phase:
/opsx:apply   # or ask Claude to implement — runs the tasks
# After implementation + verification, archive:
/opsx:archive
```

## Phase overview

| Phase | Slug | Title | Tasks | Depends on | Key gates |
|---|---|---|---|---|---|
| **P0** | `p0-foundation-config-logger` | Foundation: deps, config, logger | 20 | — | `go build`; config/logger tests; govulncheck baseline |
| **P1** | `p1-db-migrations-tx-seed` | DB, migrations, tx, seed | 18 | P0 | `migrate-up/down` clean; RunInTx commit/rollback; outbox+processed_events schema verbatim; admin seed only |
| **P2** | `p2-infra-clients` | Infra clients (Redis/MQ/ES/Asynq/Casbin) | 14 | P1 | all pings; **IK absence fails readiness** |
| **P3** | `p3-bootstrap-di-shutdown-internal-api` | Bootstrap, DI, shutdown, internal-API gateway | 18 | P2 | 3 binaries start; `/healthz`+`/readyz`; `/internal` 401 w/o token; nginx 404 `/internal`; goroutine-leak assertion |
| **P4** | `p4-identity-rbac-forum-core` | Identity, RBAC, forum core (sync + outbox append) | 28 | P3 | CRUD+auth+RBAC denial; every write → 1 outbox row in-tx; RabbitMQ queue depth 0; PostService import-guard; `post.moderated` producer |
| **P5** | `p5-outbox-publisher-mq-asynq-contracts` | Outbox publisher, MQ, Asynq, contracts | 19 | P4 | write→queue; redelivery idempotent; 9 task constants incl. `cleanup_processed_events`; **contract-ownership test** |
| **P6** | `p6-worker-tag-and-decide` | Worker: tag_post + decide_ai_reply | 22 | P5 | tag→post.tagged→decision_logs; willingness formula fixtures; fallback enqueues ≥1; owns ai_agents/preferences/decision_logs migrations + AI dev seed |
| **P7** | `p7-generate-ai-reply-sse-bridge` | generate_ai_reply + moderation + SSE bridge | 19 | P6 | AI comment+`ai.reply.completed` in-tx; **4-col unique key** concurrent-insert test; moderation block not persisted; SSE dispatch extends P3 files |
| **P8** | `p8-mention-and-followup` | @AI mention + followup judge | 15 | P7 | mention bypasses willingness + rate limit; followup safe-default false (per anomaly); AI≠AI; ≤3 followup/agent/post |
| **P9** | `p9-search-sync-and-notification` | Search index sync + notification | 17 | P7 | ES reflects writes 1–3s; ES-down chaos; rebuild==incremental; owns `comment.deleted`/`ai.reply.failed`/`post.moderated` consumers |
| **P10** | `p10-hot-score-pipeline` | Hot score pipeline | 16 | P5 | Redis hot path (no MySQL write); 30s cron snapshot; formula; **concurrent-load p99 test** |
| **P11** | `p11-web-real-api-and-sse` | Web real API client + real SSE | 17 | **P7**, P4 | mock/real env-gated; real SSE + reconnect-no-dup + polling fallback; 401/429; DOMPurify E2E; axe-core |
| **P12** | `p12-admin-refine-decision-viz` | Admin Refine + decision-log viz | 21 | **P6**, P4 | dataProvider+authProvider+RBAC visibility; decision-log explorer (gauge/hit-tags/skip-reason); Cohere fonts; axe+WCAG-AA; RBAC denial E2E |
| **P13** | `p13-e2e-perf-a11y-security-ci` | E2E, perf, a11y, security, CI | 21 | ALL | full AI chain integration; sanity; **real Playwright INP**; govulncheck/npm audit; `/internal` denial; concurrent idempotency; migrate down+up; CI + single-table ownership check |

## Dependency graph

```
P0 → P1 → P2 → P3 → P4 → P5 → P6 → P7 → P8
                       │         │    │
                       │         │    └→ P9
                       │         └──→ P11 (web real SSE — needs P7)
                       │
                       └──→ P10 (needs P5)
P6 ───────────────────────→ P12 (admin decision-log viz — needs P6)
ALL ──────────────────────→ P13 (aggregate gate)
```

Critical edges the critique forced:
- **P11 blockedBy P7** (real SSE endpoint) — the most serious original ordering bug.
- **P12 blockedBy P6** (decision_logs + ai_agents).
- **P10 can parallel P6–P9** (needs only P5's Asynq + Redis).

## Critique-driven corrections baked into the plan

1. **8th task `cleanup_processed_events`** + unowned events (`comment.deleted`/`ai.reply.failed`/`post.moderated`) → owned in P5/P9; P5 contract-ownership test guards recurrence; P13 runs implementation completeness in CI.
2. **4-column unique key** `uk_ai_reply_task(post_id, parent_comment_id_norm, ai_agent_id, trigger_type)` → asserted by P7 concurrent-insert test (not single-column).
3. **Migration single-owner rule** → P1 owns outbox/processed_events + admin seed only; P4 owns domain tables; P6 owns AI tables + AI seed; P13 CI check enforces no two migrations touch the same table.
4. **P7 extends (not recreates) P3's internalapi/sse files** → ownership documented in P3/P7.
5. **Non-functional gates front-loaded** → axe-core in P11/P12 (not just P13); govulncheck from P0; concurrent-idempotency in P5/P7/P13.
6. **Real INP, not Lighthouse TBT** → P13 measures via Playwright interaction traces.
7. **Concrete goroutine-leak assertion** → P3 (`runtime.NumGoroutine` before/after).
8. **Concrete p99 latency test** → P10 (replaces vague "no lock contention").

## Hard architectural constraints enforced across phases

- Exactly 3 Go processes (P3 wires them; no phase adds more).
- Same-process modules never HTTP each other; only worker→api `POST /internal/posts/{postId}/events` (P3 token + nginx isolation).
- Outbox written in-tx; no in-tx RabbitMQ publish (P4 `outbox.Append` + P5 publisher).
- RabbitMQ = events; Asynq = tasks; never mixed (P5 contracts).
- MySQL SoT; ES read-model; Redis rebuildable (P1/P9/P10).
- PostService holds no AI/search/notify/moderation logic (P4 import-guard test).
- Idempotent consumers + Asynq handlers (P5 `processed_events`; P7 4-col unique key).

## Recommended execution order

Linear: **P0 → P1 → P2 → P3 → P4 → P5 → P6 → P7 → (P8, P9, P10 in parallel) → P11 → P12 → P13.**

P8/P9/P10 are independent after P7/P5 and can be parallelized. P11 must follow P7. P12 must follow P6. P13 is last (aggregate).

## Next step

Start with **P0**: run `/opsx:apply` for `p0-foundation-config-logger`, or ask me to implement it. After implementation and verification, archive it (`/opsx:archive`) to promote its specs into `openspec/specs/`, then proceed to P1.

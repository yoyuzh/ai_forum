# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI Forum is a multi-AI-character forum system. It is a **modular monolith** with three Go backend processes, a React user web app, a React Refine admin app, and event-driven async task infrastructure (Outbox + RabbitMQ + Asynq).

This is not a standard CRUD forum — the core differentiators are the AI reply chain (answer-willingness scoring, decision logs), async event architecture, and explainable AI decision visualization in the admin.

Source-of-truth design docs (do not delete or move): `ai_forum_requirements_v2.md`, `ai_forum_architecture_v1.md`, `stitch_ai_forum/design_cohere.md`.

## Repository Layout

- `backend/` — Go modular monolith (module `ai-forum/backend`, Go 1.22). Three processes under `cmd/`: `api`, `worker`, `outbox-publisher`.
- `web/` — React + TypeScript + Vite user app. Real implementation with mock data layer (no backend connection).
- `admin/` — React + TypeScript + Vite + Refine + Ant Design admin console. Currently a stub (no `package.json` yet).
- `e2e/` — Playwright E2E tests against both web (port 5173) and admin (port 5174).
- `stitch_ai_forum/` — HTML prototypes and the Cohere/Synthetica design system reference.
- Each top-level dir has its own `AGENTS.md` with boundary rules — **read the nearest `AGENTS.md` before editing a module.**

## Common Commands

### Web app (`web/`)
```bash
cd web && npm run dev        # dev server (vite.config.ts sets port 5173)
npm run build                # tsc && vite build
npm run lint                 # eslint, --max-warnings 0
npm run preview
```

### Admin app (`admin/`)
Not yet bootstrapped with dependencies. When implementing, create `package.json` mirroring web's Vite + TS tooling, running on port **5174** (Playwright expects this).

### E2E (`e2e/`)
```bash
cd e2e && npm run test                         # all Playwright tests
npx playwright test web_t1.spec.ts             # single file
npx playwright test --project=web              # one project (web | admin)
npx playwright test --grep "feed"              # by name pattern
```
Playwright projects match by filename: `web_t*.spec.ts` / `admin_t*.spec.ts` for the respective app; `integration.spec.ts` / `sanity.spec.ts` run against both. Run dev servers first (web on 5173, admin on 5174).

### Backend (`backend/`)
Currently placeholder code. When real code lands: `go test ./...`, `go build ./cmd/api`, etc.

### Infra
```bash
docker compose up -d          # mysql, redis, rabbitmq, elasticsearch
```

## Architecture: Hard Constraints

These come from the `AGENTS.md` hierarchy and architecture doc and are non-negotiable:

1. **Exactly three Go processes** for v1.0: `api-server`, `worker-service`, `outbox-publisher`. Do not add more without an architecture update.
2. **Same-process modules never call each other over HTTP.** They communicate through explicit interfaces, services, repositories, and DI. The only allowed internal HTTP path is `worker-service → api-server` at `/internal/posts/{postId}/events` (SSE hub notification, authenticated via `X-Internal-Token`). `/internal/**` must never be publicly proxied by Nginx.
3. **Outbox pattern is mandatory for reliable domain events.** Write events to `outbox_events` inside the business transaction; `outbox-publisher` scans and publishes to RabbitMQ. Never publish RabbitMQ messages directly inside a business transaction.
4. **RabbitMQ vs Asynq distinction:** RabbitMQ events express *what happened*; Asynq tasks express *what should run next*. Do not mix their definitions.
5. **Data layering:** MySQL is the strong-consistency source of truth. Elasticsearch is an eventually-consistent read model (never use for business decisions). Redis is cache/counters/rate-limit/queue-infra — its data must be rebuildable from MySQL.
6. **Do not put AI, search, notification, or moderation logic into `PostService`** for convenience. Keep modules cohesive.
7. **Idempotency:** RabbitMQ consumers and Asynq handlers must be idempotent and retry-safe. `processed_events` records consumer idempotency; unique-key conflicts that mean duplicate work are treated as idempotent success.

## Backend Module Structure

`backend/internal/` contains domain and infrastructure packages, each owning one responsibility (`forum/post`, `forum/comment`, `forum/like`, `ai`, `task`, `notification`, `search`, `moderation`, `audit`, `rbac`, `outbox`, `sse`, `mq`, `event`, `cache`, `database`, `config`, `logger`, `bootstrap`, `router`, `internalapi`). Within a domain package the convention is `handler.go` / `service.go` / `repository.go` / `model.go` / `dto.go` / `event.go`.

Process responsibilities:
- **api-server** (`cmd/api`): HTTP, auth, RBAC, synchronous writes, SSE.
- **worker-service** (`cmd/worker`): consumes RabbitMQ events, runs Asynq handlers.
- **outbox-publisher** (`cmd/outbox-publisher`): scans `outbox_events` and publishes only.

## Web App Architecture

- **No backend connection** — a complete mock data layer lives in `web/src/api` (`db.ts` seed data, `client.ts` mock API, `types.ts` typed shapes). API calls must stay inside `src/api`.
- **Server state:** TanStack Query. **Client state:** Zustand (`src/stores`). Do not duplicate server data into Zustand.
- **SSE:** simulated real-time AI status in `web/src/sse` (`emitter.ts`, `simulator.ts`, `useSSE.ts`). Keep SSE logic here, not scattered across pages.
- **Long lists** (post feed, comments) must use `react-virtuoso`.
- **Sanitize** all user-supplied rich text/Markdown with DOMPurify before rendering.
- `@` alias maps to `web/src` (see `vite.config.ts`).

## Admin App Architecture

- Data fetching goes through Refine `dataProvider` or `admin/src/api` — never scattered in components.
- Frontend permission checks control visibility only; **backend RBAC is authoritative** — do not hardcode permission results or treat frontend checks as security.
- Keep screens operational and dense (operator-focused), not marketing-style.

## Design System

Follows the Cohere/Synthetica system from `stitch_ai_forum/design_cohere.md`. Tokens are exposed as CSS custom properties and mapped into Tailwind under the `cohere.*` namespace in `web/tailwind.config.js` — use `cohere-*` classes, do not hardcode hex values.

- **Palette:** near-black `#17171c`/`#000`, deep green `#003c33`, action blue `#1863dc`, coral `#ff7759`, soft backgrounds `#eeece7`/`#fbf9f4`/`#fff`, hairline border `#d9d9dd`.
- **Typography:** CohereText/Space Grotesk for display, Unica77 for UI sans, CohereMono/JetBrains Mono for labels/timestamps/status. Chinese text minimum 16px.
- **Radius scale:** xs 4px, sm 8px, md 16px, lg 22px, xl 30px, pill 32px (intentional variation, not uniform).
- Responsive down to 375px with no overflow or overlapping text.

## Git

Commit format: `<type>: <description>` (types: feat, fix, refactor, docs, test, chore, perf, ci). Attribution is disabled globally.

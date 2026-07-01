## Context

Web has a real mock layer (`src/api/{db,client,types}.ts`) and a client-side SSE simulator (`src/sse/simulator.ts` driving an in-app emitter). Pages depend on the emitter contract, not the simulator directly. TanStack Query for server state; Zustand for client state (no server-data duplication). DOMPurify for rich text. Cohere design tokens in Tailwind. vite.config sets port 5173 (e2e baseURL), though CLAUDE.md says 3000. Critique: real SSE needs P7; simulator side-effects must be routed through `api.*`; reconnect must reconcile without duplicating AI replies; a11y/contrast should be front-loaded.

## Goals / Non-Goals

**Goals:**
- Env-gated mock/real client preserving signatures; real client hits P4 REST + P7 SSE/ai-status.
- Real EventSource with reconnect reconciliation + polling fallback.
- 401/403/429 handling; auth wiring to Zustand (client state only).
- Front-load axe-core + DOMPurify injection E2E.
- Fix CLAUDE.md port drift.

**Non-Goals:**
- No admin app (P12).
- No full e2e cross-app suite (P13) — only web-relevant flows here.
- No new design system; Cohere tokens unchanged.

## Decisions

### D1: Signature-preserving realClient
`realClient.ts` implements the exact function signatures `client.ts` exposes (the contract pages already use). `client.ts` becomes an env-gated selector: `VITE_API_MODE=real` → realClient, else mock. Both share `types.ts` as the single contract so they can't diverge.

### D2: Real SSE preserving emitter contract
`src/sse` keeps its emitter API; only the *source* changes from `simulator.ts` to an `EventSource` against `GET /api/posts/{postId}/events`. Pages consuming the emitter are untouched. On reconnect, send `Last-Event-ID`; then fetch `ai-status` once to reconcile missed events; prepend comments by id (idempotent — duplicate id ignored) to avoid double-rendering AI replies.

### D3: 401/403/429 handling
401 → clear auth state + redirect to login. 403 → surface "no permission". 429 → surface rate-limit message (esp. for @AI, P8). TanStack Query retry config respects these.

### D4: Server state in TanStack Query, client state in Zustand
Auth token + UI flags in Zustand; all server data via TanStack Query. No server data copied into Zustand (CLAUDE.md constraint).

### D5: Front-load a11y + injection
axe-core scan in Playwright on feed + post detail now (not P13). DOMPurify-presence test + an injection E2E asserting sanitized output. This is critique risk #5 mitigation.

### D6: Port doc fix
Update CLAUDE.md common-commands: web dev server is 5173 (e2e baseURL is source of truth). No code port change.

## Risks / Trade-offs

- **[Risk] Real SSE ships before P7** → Mitigation: blockedBy P7; mock mode is independently shippable in the meantime.
- **[Risk] Simulator side-effect relied on outside src/sse** → Mitigation: D1 audit; route all through `api.*`.
- **[Risk] Reconnect duplicates AI replies** → Mitigation: D2 idempotent prepend by id + ai-status reconcile.
- **[Risk] Mock/real divergence** → Mitigation: D1 shared `types.ts`; P13 sanity runs real mode only.
- **[Risk] Frontend RBAC treated as security** → Mitigation: backend authoritative (P4); web only hides UI.

## Migration Plan

1. httpClient + realClient + selector → auth wiring → real SSE → reconnect → a11y/injection tests → doc fix.
2. Rollback: `VITE_API_MODE=mock` restores the mock-only app.

## Open Questions

- SSE reconnection backoff curve — standard EventSource reconnect with exponential cap; finalized at implementation.

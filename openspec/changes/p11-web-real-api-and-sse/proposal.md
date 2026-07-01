## Why

The web app currently runs entirely on a client-side mock layer (`src/api/db.ts` seed + `client.ts` mock + `src/sse` simulator) with no backend connection. With the backend's REST API + SSE endpoint + ai-status endpoint now real (P4/P7), the web app must swap to a real API client and real SSE while keeping the mock as a dev fallback. Critique flagged the most serious ordering bug here: real SSE depends on P7's `GET /api/posts/{postId}/events` — so this phase is explicitly `blockedBy` P7.

## What Changes

- `web/src/api/httpClient.ts`: typed fetch wrapper with 401 (→ login redirect), 403, 429 (rate-limit surfaced) handling.
- `web/src/api/realClient.ts`: live backend implementation preserving the existing `client.ts` function signatures (the integration contract).
- `web/src/api/client.ts`: env-gated selector (`VITE_API_MODE=mock|real`); mock stays as dev fallback.
- `web/src/api/auth.ts` + `useAuth`/queries wiring to `useUserStore` (Zustand for client state only — no server data duplication).
- Real SSE: replace `src/sse/simulator.ts` source with `EventSource` against `GET /api/posts/{postId}/events`, preserving the existing emitter contract pages depend on. Reconnect with `Last-Event-ID` reconciliation + one `ai-status` poll after reconnect to补齐 missed events without duplicating AI replies (idempotent comment prepend by id).
- `ai-status` polling fallback when SSE unavailable.
- `.env`/`.env.example`: `VITE_API_BASE_URL`, `VITE_API_MODE`.
- Audit all simulator imports outside `src/sse` and route them through `api.*` only (critique risk: page depending on simulator side-effect).
- Tests/E2E: mock mode green; real mode feed/post/comment/like/favorite round-trip; SSE live status transitions; 401→login, 429 surfaced; DOMPurify injection E2E; axe-core on feed + post detail.
- Fix doc drift: CLAUDE.md documents web on port 3000 but `vite.config.ts` sets 5173 (e2e baseURL is source of truth) — update CLAUDE.md common-commands to 5173.

## Capabilities

### New Capabilities
- `web-api-client`: Env-gated typed API client with mock/real backends preserving a shared contract, plus auth wiring.
- `web-realtime-sse`: Real EventSource SSE replacing the simulator, with reconnect reconciliation and ai-status polling fallback.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `web/src/api/{httpClient,realClient,client,auth}.ts`, `web/src/sse/*`, `web/.env*`, `web/src/stores`, `web/src/pages` (consumer fixes), `CLAUDE.md`.
- **Dependency**: blockedBy P7 (real SSE endpoint) and P4 (REST endpoints). Mock mode independently shippable.
- **Design system**: DOMPurify sanitization preserved on all rich text; Cohere tokens unchanged.
- **Non-functional**: axe-core + CWV front-loaded here (critique risk #5), not back-loaded to P13.

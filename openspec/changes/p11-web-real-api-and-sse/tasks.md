# P11 Tasks — Web real API client + real SSE

## 1. API client
- [x] 1.1 `web/src/api/httpClient.ts`: typed fetch wrapper; 401→login, 403→permission, 429→rate-limit
- [x] 1.2 `web/src/api/realClient.ts`: live impl preserving `client.ts` signatures; shares `types.ts`
- [x] 1.3 `web/src/api/client.ts`: env-gated selector (`VITE_API_MODE=mock|real`)
- [x] 1.4 `web/.env`/`.env.example`: `VITE_API_BASE_URL`, `VITE_API_MODE`
- [x] 1.5 `web/src/api/auth.ts` + `useAuth`/queries → `useUserStore` (Zustand client state only; server data via TanStack Query)
- [x] 1.6 Notification API client: list notifications, unread count, mark one read, mark all read (same mock/real contract)

## 2. Real SSE
- [x] 2.1 Replace `src/sse/simulator.ts` source with `EventSource` against `GET /api/posts/{postId}/events`; preserve emitter contract
- [x] 2.2 Reconnect: `Last-Event-ID` + one `ai-status` fetch to reconcile; idempotent comment prepend by id (no duplicate AI replies)
- [x] 2.3 `ai-status` polling fallback when SSE unavailable
- [x] 2.4 Audit all simulator imports outside `src/sse`; route through `api.*` only

## 3. Tests / E2E
- [x] 3.1 Mock mode green; real mode feed/post/comment/like/favorite round-trip against P4 backend
- [x] 3.2 SSE live status transitions + reconnect-no-duplication + polling fallback (requires P7 backend)
- [x] 3.3 401→login, 429 surfaced
- [x] 3.4 DOMPurify injection E2E: unsanitized rich text is sanitized before render
- [x] 3.5 axe-core scan on feed + post detail (front-loaded a11y)
- [x] 3.6 Notification E2E: unread badge reflects backend rows; mark-read updates badge without reload

## 4. Doc fix
- [x] 4.1 Update CLAUDE.md common-commands: web dev server port is 5173 (e2e baseURL is source of truth)

## 5. Verification
- [x] 5.1 `npm run lint` (--max-warnings 0) and `npm run build` green
- [x] 5.2 blockedBy P7 (real SSE endpoint) and P4 (REST) — mock mode shippable independently

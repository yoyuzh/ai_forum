## Context

The live stack now runs in `VITE_API_MODE=real`, but some shipped screens still depend on mock-only implementations or local projections. This phase is a product-completion pass after P13, not an architecture expansion: it replaces visible mock seams with live data, hides out-of-scope affordances, and adds the smallest backend surfaces needed by the current UI.

## Goals / Non-Goals

**Goals:**
- Web AI pages and post-detail AI sidebars render real backend data when it exists.
- Profile edit and user stats persist/read from backend.
- Admin dashboard uses live backend data.
- Visible admin actions are either backed by endpoints or shown as read-only.
- Web search uses the Elasticsearch read model.
- SSO and reports are not half-wired.

**Non-Goals:**
- No OAuth/SSO implementation.
- No reports/moderation case-management workflow.
- No new backend process or service split.
- No frontend redesign.
- No recommendation engine.

## Decisions

### D1: Extend existing APIs before adding new abstractions
Reuse the existing admin/user/router packages and API client shapes. Add narrow endpoints only where the current UI has no backend source. Do not introduce a generic BFF or GraphQL layer.

### D2: Web AI data is read-only in this phase
The web user app may list agents, show task/activity/decision state, and render post-specific decision logs. Agent configuration remains admin-only.

### D3: Profile update is small and explicit
Support only fields already exposed by the current profile form and backed by `users` columns. Preferences without backend columns stay client-local or are hidden until a later preference phase owns persistence.

### D4: Admin dashboard aggregates are server-computed
Dashboard counts/trends/recent lists come from backend SQL/ES/queue-derived queries. The dashboard should degrade to zeros/empty lists with an error banner, not mock data, when the backend has no rows.

### D5: Unsupported admin mutations are not visible
If create/delete is not implemented for a resource, the UI must not show those actions. This is cheaper and clearer than wiring generic no-op endpoints.

### D6: Search uses ES but never for business decisions
Web search queries a backend search endpoint backed by Elasticsearch. Empty ES results or ES outage are surfaced as search unavailable/empty states; business writes and authorization still depend on MySQL.

## Risks / Trade-offs

- **[Risk] Backend endpoint sprawl** -> Mitigation: add only endpoints used by visible routes, under existing router/admin/user/search ownership.
- **[Risk] Dashboard aggregates become expensive** -> Mitigation: simple bounded queries first; cache only if measured.
- **[Risk] Search outage blocks browsing** -> Mitigation: search failure does not block normal `/api/posts` feed.
- **[Risk] Hidden SSO/reports disappoint users** -> Mitigation: no fake buttons or menu entries; document out-of-scope status.

## Migration Plan

1. Add missing backend read/update endpoints.
2. Wire web real client and pages.
3. Wire admin dashboard/dataProvider boundaries.
4. Add search endpoint and web search integration.
5. Remove/hide half-wired SSO/reports affordances.
6. Run real-mode E2E against compose.

## Open Questions

- Whether profile preferences should be persisted in `users` JSON now or deferred until a dedicated preferences phase. Default: defer unless a current visible control requires persistence.

## Why

P0-P13 prove the backend chain, real-mode web/admin reachability, and system gates, but manual product review still shows visible real-mode gaps: the web AI surfaces return empty arrays, profile edits are not persisted, admin dashboard metrics still come from mock data, generic admin CRUD throws for unsupported operations, and web search is local filtering over `/api/posts` instead of a backend search experience. SSO buttons and reports are visible product expectations but have no owning phase; this change keeps them out of the critical path by documenting/removing half-wired UI rather than building new auth/reporting subsystems.

## What Changes

- Web real-mode completion: real AI agent/task/decision-log/activity data where the UI exposes it, persistent profile update, real user stats, and no fake SSO affordances.
- Admin real dashboard: dashboard cards, trends, service state, recent posts/tasks, and decision timeline are derived from backend data instead of `mockApi.dashboard`.
- Admin operation boundaries: list/show/edit operations that are visible in the UI call backend endpoints; unsupported create/delete actions are hidden or return explicit read-only UI, not runtime `not implemented` errors.
- Search experience: add a backend-backed search endpoint or query parameter using Elasticsearch read model, and wire web feed search to it with graceful empty/error states.
- Scope cleanup: reports remain absent/documented out-of-scope for v1.0 unless a later phase owns them; SSO remains out-of-scope and hidden.

## Capabilities

### New Capabilities
- `web-product-completion`: Web real mode has no visible mock-only AI/profile gaps on shipped routes.
- `admin-product-completion`: Admin dashboard and visible operations use live backend data or explicit read-only states.
- `search-experience`: User search is backed by the Elasticsearch read model instead of client-side filtering only.

### Modified Capabilities
- `web-api-client`: EXTENDS P11 to cover AI agents/tasks/decision logs/activity and profile persistence in real mode.
- `admin-data-access`: EXTENDS P12 to replace dashboard mock data and eliminate runtime "not implemented" CRUD paths from visible UI.

## Impact

- **Code**: `web/src/api/realClient.ts`, web AI/profile/search pages and hooks; `admin/src/api/client.ts`, `admin/src/api/dataProvider.ts`, admin dashboard/resource pages; backend admin/search/profile handlers and repositories as needed.
- **Systems**: No new Go process, no microservice split, no same-process HTTP calls.
- **Data**: MySQL remains source of truth; Elasticsearch remains search read model; Redis remains rebuildable cache/counter infrastructure.
- **Out of scope**: SSO providers, content reports/moderation workflow, recommendation/personalization, and production deployment automation.

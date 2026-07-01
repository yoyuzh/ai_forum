## Why

The admin app is a Refine + AntD scaffold with a mock dataProvider and stub pages. This phase makes it operational: a REST dataProvider against the backend, an auth provider, RBAC-visibility access control (frontend visibility only — backend RBAC is authoritative per P4), and the project's differentiator — the **AI decision-log visualization** that explains why each AI replied or skipped (willingness score vs threshold, hit tags, skip/fallback reason). Critique flagged this phase must gate on P6 (`decision_logs` + `ai_agents`) and front-load axe-core/contrast scanning.

## What Changes

- `admin/src/api/dataProvider.ts`: Refine REST dataProvider against the backend.
- `admin/src/api/authProvider.ts`: auth provider wired to backend login/JWT.
- `admin/src/providers/accessControlProvider.ts`: RBAC visibility from backend permissions (frontend-only; backend authoritative).
- Operational resource screens: Users, Posts, Comments, AI Agents (with `replyThreshold`/`activityLevel`/trigger perms inline-edit), AI Tasks (retry/terminate), Tags/Preferences.
- **Decision-log explorer** (the differentiator):
  - `DecisionDetailDrawer`: per decision — agent, trigger, willingnessScore vs thresholdValue (gauge/bar with delta), hitTags[], decision (REPLY/IGNORE/FALLBACK), reason, fallback flag, link to resulting task/comment.
  - `WillingnessGauge`: score-vs-threshold with above/below coloring via Cohere blue/coral; accessible labels (color never the sole signal).
  - `HitTagsViewer`, `SkipReasonBlock` (below-threshold / fallback-invoked / rate-limited / config-disabled).
  - `AgentDecisionBreakdown`: aggregate (reply-rate, avg willingness, fallback rate) per agent.
- Align admin fonts to the Cohere system (Unica77/Space Grotesk/CohereText) — critique flagged admin currently uses Hanken Grotesk.
- Tests/E2E: admin login + RBAC-gated CRUD; decision-log explorer renders full breakdown; axe-core + WCAG-AA contrast scan green; an RBAC denial E2E calling a denied action server-side asserts 403.

## Capabilities

### New Capabilities
- `admin-data-access`: Refine REST dataProvider + authProvider + RBAC visibility (frontend-only; backend authoritative).
- `admin-ai-decision-viz`: Explainable-AI decision-log explorer rendering willingness/threshold/hit-tags/skip-reason from `decision_logs`.

### Modified Capabilities
<!-- None. -->

## Impact

- **Code**: `admin/src/{api,providers,components,pages}/*`, `admin/src/App.tsx`, `admin/tailwind.config.js` (Cohere fonts), admin types.
- **Dependency**: blockedBy P6 (`decision_logs` + `ai_agents`) and P4 (RBAC/users/posts).
- **Design system**: admin aligned to Cohere (fonts + tokens); WCAG-AA contrast; color never sole signal.
- **Security**: frontend RBAC is visibility-only; a denial E2E asserts backend 403.

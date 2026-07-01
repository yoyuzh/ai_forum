## Context

Admin is Refine + AntD + Vite + TS, scaffolded but stubbed (`admin/src` minimal, mock dataProvider). Architecture §12.2 (RBAC, backend authoritative), §11.4 (decision log fields + admin visualization fields), §6.5/6.6 (AI agent trigger perms). Critique: gate on P6; align fonts to Cohere; front-load axe-core/contrast; frontend RBAC is cosmetic — add a denial E2E. The decision-log viz is the project's signature differentiator, so its contract (the `decision_logs` shape pinned in P6) drives the component design.

## Goals / Non-Goals

**Goals:**
- REST dataProvider + authProvider + RBAC visibility.
- Operational resource screens (users/posts/comments/agents/tasks/tags).
- Decision-log explorer with willingness gauge, hit tags, skip reasons, agent breakdown.
- Cohere font alignment; axe-core + WCAG-AA contrast; RBAC denial E2E.

**Non-Goals:**
- No new design system — reuse Cohere tokens.
- No real-time admin SSE (admin reads decision_logs on demand).
- No full cross-app e2e (P13).

## Decisions

### D1: Refine dataProvider + authProvider
Standard Refine REST dataProvider targeting the backend. authProvider wraps login/JWT and exposes permissions for `accessControlProvider`. Token stored in client state; 401 → login.

### D2: RBAC visibility is frontend-only
`accessControlProvider` hides buttons/routes based on permissions from the backend. A denial E2E calls a denied action server-side and asserts 403 — proving the backend is authoritative (CLAUDE.md hard constraint). Frontend checks are never security.

### D3: Decision-log explorer components
`DecisionDetailDrawer` renders the full P6 `decision_logs` row. `WillingnessGauge` shows score vs threshold with delta; above-threshold = Cohere blue, below = coral; always paired with a text label (color never sole signal). `HitTagsViewer` chips; `SkipReasonBlock` structured reasons; `AgentDecisionBreakdown` aggregates per agent. The contract is the P6 schema; if P6 shipped a thinner log, this phase degrades — mitigated by P6 pinning the full field set.

### D4: Cohere font alignment
Admin tailwind/fonts aligned to `stitch_ai_forum/design_cohere.md` (Unica77/Space Grotesk/CohereText/CohereMono). Critique flagged Hanken Grotesk; replaced. Font-licensing gaps flagged if any.

### D5: Front-load a11y/contrast
axe-core in Playwright on admin dashboard + decision log; WCAG-AA contrast scan on Cohere text/background pairings. Not deferred to P13.

## Risks / Trade-offs

- **[Risk] Frontend RBAC mistaken for security** → Mitigation: D2 denial E2E asserts backend 403.
- **[Risk] Decision-log viz degrades if P6 schema is thin** → Mitigation: P6 pins full fields; this phase depends on P6.
- **[Risk] Font licensing gaps** → Mitigation: D4 flags any gap; v1 can fall back to the open-licensed Cohere/Mono variants.
- **[Risk] Admin AntD defaults look template-y** → Mitigation: Cohere tokens + intentional states per web design-quality rules.

## Migration Plan

1. dataProvider + authProvider + accessControl → resource screens → decision-log components → font alignment → a11y/contrast + denial E2E.
2. Rollback: admin reverts to stub/mock.

## Open Questions

- Whether to expose decision-log edit (no — read-heavy; edits go through agent config per §11.4).

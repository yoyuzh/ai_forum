# Module Instructions

## Responsibility

Own the React + TypeScript + Refine + Ant Design admin application.

## Owns

- Admin source under `admin/src`.
- Refine resources, dataProvider/api client, admin pages, and admin components.

## Must Not

- Do not hardcode permission results in the frontend.
- Do not treat frontend permission checks as security.
- Do not scatter data fetching outside dataProvider or the admin API client.

## Allowed Dependencies

- React, TypeScript, Refine, Ant Design, TanStack Query, React Router, and admin-specific visualization utilities as needed.

## Communication Rules

- Data requests go through dataProvider or a shared API client.
- Frontend permissions only control display; backend RBAC is authoritative.

## Data Rules

- AI decision-log visualization is a core admin capability.
- Admin mutations must rely on backend validation and RBAC.

## Testing Rules

- Add resource/page tests for user, post, comment, AI agent, AI task, and AI decision views when implemented.

## Notes for Codex

- Keep admin screens operational and dense; avoid marketing-style layouts.

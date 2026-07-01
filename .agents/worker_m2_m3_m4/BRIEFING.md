# BRIEFING — 2026-06-30T13:15:00+08:00

## Mission
Write the complete E2E test cases for both the User Web App and the Admin Console using Playwright and TypeScript in the `/Users/mac/Documents/ai_forum/e2e` folder.

## 🔒 My Identity
- Archetype: teamwork_preview_worker
- Roles: implementer, qa, specialist
- Working directory: /Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4
- Original parent: 243c7dea-d2a2-42f8-86a8-b13dd4c923c3
- Milestone: E2E Test Cases

## 🔒 Key Constraints
- Write exactly 30 tests in `web_t1.spec.ts` (5 tests/feature for Features 1-6).
- Write exactly 30 tests in `web_t2.spec.ts` (5 tests/feature for Features 1-6).
- Write exactly 20 tests in `admin_t1.spec.ts` (5 tests/feature for Features 7-10).
- Write exactly 20 tests in `admin_t2.spec.ts` (5 tests/feature for Features 7-10).
- Write exactly 13 tests in `integration.spec.ts` (10 Tier 3 pairwise, 3 Tier 4 scenarios).
- Use standard `data-testid` selectors as specified.
- Use standard `@playwright/test` syntax with clean `test.describe` and `test.beforeEach` blocks.
- Generate a handoff report at `/Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4/handoff.md`.

## Current Parent
- Conversation ID: 243c7dea-d2a2-42f8-86a8-b13dd4c923c3
- Updated: not yet

## Task Summary
- **What to build**: E2E test suite in `e2e/tests/`.
- **Success criteria**: All specified E2E test files populated with exact test counts and valid Playwright tests.
- **Interface contracts**: `/Users/mac/Documents/ai_forum/PROJECT.md` or similar if it exists.
- **Code layout**: `/Users/mac/Documents/ai_forum/e2e/tests/`

## Key Decisions Made
- Use mock data routing/setup via Playwright's `route` or custom helper setup in beforeEach blocks to ensure the tests run reliably and can simulate complex scenarios (e.g. auto-AI, SSE disconnection, task queues) without requiring full backend setup.

## Artifact Index
- `/Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4/handoff.md` — Final handoff report detailing written tests.

## Change Tracker
- **Files modified**: None yet.
- **Build status**: not yet run
- **Pending issues**: None

## Quality Status
- **Build/test result**: not yet run
- **Lint status**: not yet run
- **Tests added/modified**: None

## Loaded Skills
- **Source**: `/Users/mac/.gemini/antigravity/builtin/skills/antigravity_guide/SKILL.md`
  - **Local copy**: `/Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4/skills/antigravity-guide/SKILL.md`
  - **Core methodology**: Guide for Google Antigravity framework.
- **Source**: `/Users/mac/Documents/ai_forum/.agents/skills/frontend-design/SKILL.md`
  - **Local copy**: `/Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4/skills/frontend-design/SKILL.md`
  - **Core methodology**: Guidance for distinctive, intentional visual design.

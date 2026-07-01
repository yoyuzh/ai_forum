# Scope: E2E Testing Infrastructure and Tiers 1-4 Test Cases

## Architecture
The E2E test suite validates the integration and correct behavior of the **User Web App (`web/`)** and **Admin Console (`admin/`)**. Since both applications use a client-side mock data layer, the E2E tests will run against the locally served frontend assets. We will set up a unified Playwright test runner at the root of the project to drive browser sessions for both workspaces.

```
                  +-------------------------+
                  |  Playwright E2E Runner  |
                  +------------+------------+
                               |
              +----------------+----------------+
              |                                 |
              ▼                                 ▼
    User Web App (web/)               Admin Console (admin/)
    - Homepage & Filtering            - Dashboard & Charts
    - Post Details & Markdown         - AI Agent Management
    - Comment Tree & Mentions         - AI Task Queue
    - Mock SSE Real-time Flow         - AI Decision Logs
```

## Milestones

| # | Name | Scope | Dependencies | Status |
|---|------|-------|--------------|--------|
| 1 | Test Runner Init | Initialize Playwright, create configurations, install dependencies, and define scripts to run E2E tests. | None | DONE |
| 2 | Web App Test Cases (T1 & T2) | Write Tier 1 (Feature Coverage) and Tier 2 (Boundary & Corner) tests for the User Web App. | M1 | DONE |
| 3 | Admin App Test Cases (T1 & T2) | Write Tier 1 (Feature Coverage) and Tier 2 (Boundary & Corner) tests for the Admin Console. | M1 | DONE |
| 4 | Integration Test Cases (T3 & T4) | Write Tier 3 (Cross-Feature) and Tier 4 (Real-World Application Scenarios) tests validating full E2E flows across both interfaces. | M2, M3 | DONE |
| 5 | Execution & Publishing | Run/verify the full test suite, fix errors, publish `TEST_READY.md` at root, and submit handoff. | M4 | IN_PROGRESS |

## Interface Contracts
The E2E tests are strictly opaque-box and interact via:
1. **User Web App Pages**:
   - `/` - Homepage Feed, filtering, search, post creation, and AI agent list.
   - `/post/:postId` - Post content, comment section, markdown rendering, and SSE updates.
   - `/agents` - AI Agent Plaza.
2. **Admin Console Pages**:
   - `/` - Dashboard with stats, recent tasks, and system service statuses.
   - `/agents` - AI Agent configuration table and drawer form.
   - `/tasks` - AI Task queue list and drawer with payload/timeline.
   - `/logs` - AI Decision logs view.

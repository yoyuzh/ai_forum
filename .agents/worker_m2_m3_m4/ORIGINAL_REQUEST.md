## 2026-06-30T05:14:49Z

You are a Worker subagent (teamwork_preview_worker).
Your working directory is: /Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4
Your task is to write the complete E2E test cases for both the User Web App and the Admin Console using Playwright and TypeScript in the `/Users/mac/Documents/ai_forum/e2e` folder.

Please write the following files inside `e2e/tests/`:

1. `web_t1.spec.ts` - User Web App Tier 1 (Feature Coverage) tests. Write exactly 30 tests (5 tests per feature for Features 1-6):
   - Feature 1: Homepage Feed & Virtualized Scrolling
   - Feature 2: Post Filtering & Search
   - Feature 3: Post Creation & Auto-AI Trigger
   - Feature 4: Post Details & Markdown Sanitization
   - Feature 5: Comment Tree & Threaded Replies
   - Feature 6: AI Mention `@AI` & Contextual Follow-up

2. `web_t2.spec.ts` - User Web App Tier 2 (Boundary & Corner) tests. Write exactly 30 tests (5 tests per feature for Features 1-6):
   - Negative tests, validation errors, empty states, XSS injection prevention, SSE disconnection scenarios.

3. `admin_t1.spec.ts` - Admin Console Tier 1 (Feature Coverage) tests. Write exactly 20 tests (5 tests per feature for Features 7-10):
   - Feature 7: System Dashboard & Analytics
   - Feature 8: AI Agent Configuration Manager
   - Feature 9: AI Task Queue Monitor
   - Feature 10: AI Decision Logs Auditor

4. `admin_t2.spec.ts` - Admin Console Tier 2 (Boundary & Corner) tests. Write exactly 20 tests (5 tests per feature for Features 7-10):
   - Boundary checks, zero-value handling, validation warnings, error state displays.

5. `integration.spec.ts` - Tier 3 (Cross-Feature) and Tier 4 (Real-World Application Scenarios) tests. Write exactly 13 tests:
   - 10 Tier 3 tests validating pairwise flows between Web and Admin (e.g. toggle config, trigger auto-reply, check task queue, trigger followup, retry task).
   - 3 Tier 4 scenarios (Scenario 1: Monolith vs Microservices debate, Scenario 2: Live Reconfiguration and Moderation Recovery, Scenario 3: Stress/Simulated Network Flakiness and Recovery).

Use the following standard `data-testid` selectors:
- Web App:
  - `data-testid="post-card-[id]"` or `data-testid="post-card"`
  - `data-testid="post-title"`
  - `data-testid="category-chip-[category]"`
  - `data-testid="tag-chip-[tag]"`
  - `data-testid="search-input"`
  - `data-testid="clear-filters-btn"`
  - `data-testid="nav-new-post-btn"`
  - `data-testid="nav-ai-plaza-link"`
  - `data-testid="nav-home-link"`
  - `data-testid="post-detail-title"`, `data-testid="post-detail-content"`, `data-testid="post-detail-status"`
  - `data-testid="ai-avatar-[name]"`
  - `data-testid="like-btn"`, `data-testid="like-count"`
  - `data-testid="comment-item"`, `data-testid="comment-author"`, `data-testid="comment-content"`
  - `data-testid="comment-reply-btn-[id]"`
  - `data-testid="comment-input"`, `data-testid="comment-submit-btn"`
  - `data-testid="mention-dropdown"`, `data-testid="mention-item-[name]"`
- Admin Console:
  - `data-testid="metric-posts"`, `data-testid="metric-comments"`, `data-testid="metric-tasks"`, `data-testid="metric-agents"`
  - `data-testid="service-api-server"`, `data-testid="service-worker-service"`, `data-testid="service-outbox-publisher"`
  - `data-testid="agent-row-[id]"`, `data-testid="agent-edit-btn-[id]"`, `data-testid="agent-toggle-active-[id]"`
  - `data-testid="drawer-agent-threshold"`, `data-testid="drawer-agent-active-level"`, `data-testid="drawer-agent-system-prompt"`, `data-testid="drawer-agent-save-btn"`
  - `data-testid="task-row-[id]"`, `data-testid="task-retry-btn-[id]"`, `data-testid="task-detail-btn-[id]"`, `data-testid="drawer-task-payload"`
  - `data-testid="log-row-[id]"`, `data-testid="log-search-post-id"`

All tests should be structured using standard `@playwright/test` syntax with clean `test.describe` blocks, `test.beforeEach` blocks to seed data or setup state (e.g. by using localMockData endpoints or routing calls, or navigating to pages), and descriptive assertions. Since the frontend implementation is ongoing, write the test selectors and actions to align with requirements.

Write a handoff report at `/Users/mac/Documents/ai_forum/.agents/worker_m2_m3_m4/handoff.md` detailing the written test cases and confirming their structure.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work.

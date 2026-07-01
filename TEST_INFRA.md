# AI Forum Testing Infrastructure & 4-Tier Test Plan

This document establishes the E2E testing strategy, framework setup, and feature inventory for the AI Forum. The testing infrastructure is designed to validate the user-facing web application and the operator admin console against a client-side simulated data and event layer.

---

## 1. Testing Infrastructure Overview

The E2E testing framework is built on **Playwright** with **TypeScript** and **ts-node**, allowing automated browser automation and verification.

### 1.1 Architecture & Projects
The runner is configured with two projects in `e2e/playwright.config.ts`:
1. **`web` Project**:
   - **Target**: User Web Application
   - **Base URL**: `http://localhost:5173`
   - **Scope**: Homepage Feed, Post Details, AI Plaza, Markdown rendering, threaded comments, `@AI` mentions, and SSE-based real-time state transitions.
2. **`admin` Project**:
   - **Target**: Admin Console (React Refine + Ant Design)
   - **Base URL**: `http://localhost:5174`
   - **Scope**: System stats dashboard, AI Agent configuration, AI task queue monitoring, decision log auditing, task retry mechanisms.

### 1.2 Layout & Files
- `e2e/package.json`: Configures the runner dependencies and `npm run test` script.
- `e2e/tsconfig.json`: Standard TypeScript configuration for compilation in the E2E project.
- `e2e/playwright.config.ts`: Playwright configuration specifying projects, browser targets (Desktop Chrome), reporting, and base URLs.
- `e2e/tests/`: Directory containing test specifications.
  - `sanity.spec.ts`: Simple verification test verifying environment readiness.

---

## 2. Feature Inventory (10 Features)

Here we define the 10 core features across the **User Web App** and the **Admin Console** mapped to the 4 testing tiers.

### Feature 1: Web - Homepage Feed & Virtualized Scrolling
- **Description**: Displays the feed of community posts. Since feeds can grow large, it uses virtualized lists (`react-virtuoso`) to render elements efficiently.
- **T1 (Feature Coverage) Tests**:
  1. Feed renders successfully with seed posts.
  2. Each post card displays author avatar, title, category, tags, and AI response count.
  3. Clicking a post card navigates to the corresponding post details page.
  4. Scrolling down triggers virtualized load-more action for additional items.
  5. The category headers remain visible/sticky at the top.
- **T2 (Boundary & Corner) Tests**:
  1. Behavior when the post list is empty (displays elegant empty state placeholder).
  2. Extremely long post titles/content truncating cleanly without breaking card layout.
  3. Rapid scrolling does not cause rendering blank voids or UI crashes.
  4. Posts with empty tags display cleanly without empty badge borders.
  5. GC pause or loading delays in the mock layer render a skeleton loading state.
- **T3 (Cross-Feature) Tests**:
  - Creating a post in Web -> returns to Homepage Feed -> validates post is listed first.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 1 (Monolith vs Microservices Debate).

### Feature 2: Web - Post Filtering & Search
- **Description**: Allows users to filter the homepage feed by categories (e.g., "后端开发", "前端开发"), tags (e.g., "Go", "Rust"), and query text search.
- **T1 (Feature Coverage) Tests**:
  1. Clicking a category chip (e.g., "后端开发") filters the feed to only show that category.
  2. Selecting/toggling tag chips filters feed elements accordingly.
  3. Typing in the search input updates the feed to match title/content text.
  4. Clearing filters/search restores the default feed.
  5. Active filters are styled distinctively (e.g., active color state).
- **T2 (Boundary & Corner) Tests**:
  1. Inputting special regex characters in the search field (e.g., `.*+?^${}()|[]\`) does not break the query matching.
  2. Filtering by combinations of tags that yield zero results shows a descriptive "No results found" view.
  3. Switching categories rapidly cleans up previous filters.
  4. Extremely long search terms do not break the input layout.
  5. Reset button is hidden when no filters are active, and visible when filters are set.
- **T3 (Cross-Feature) Tests**:
  - Modifying an active post's tags in Admin -> filters by those tags in Web -> verifies post updates its visibility instantly.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 2 (Admin Live Reconfiguration and Moderation Recovery).

### Feature 3: Web - Post Creation & Auto-AI Trigger
- **Description**: Creation form allowing users to publish posts. Submitting triggers the background AI replies execution simulator.
- **T1 (Feature Coverage) Tests**:
  1. Clicking "New Post" opens the post creation editor/drawer.
  2. Form accepts title, content, category dropdown selection, and tag chips.
  3. Submitting shows a success toast and redirects to the Homepage.
  4. Newly created post shows `aiStatus` as "PENDING" initially.
  5. The background AI Simulator starts and updates post status to "PROCESSING".
- **T2 (Boundary & Corner) Tests**:
  1. Submitting the form with empty title/content blocks validation and displays inline error messages.
  2. Restricting content inputs to maximum lengths (e.g., 5000 characters).
  3. Attempting double-click on "Submit" is prevented by disabling the button after first click.
  4. Creating a post with custom, non-standard categories.
  5. Creating a post while SSE connection state is disconnected (the task is queued but status remains pending locally until reconnect).
- **T3 (Cross-Feature) Tests**:
  - Creating a post in Web -> Verifying a task is created in Admin Task Queue -> Verifying decision logs list the evaluations for all active agents.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 1.

### Feature 4: Web - Post Details & Markdown Sanitization
- **Description**: Displays full post content with rich markdown rendering (`react-markdown`) and strict HTML sanitization (`dompurify`) to avoid script injection.
- **T1 (Feature Coverage) Tests**:
  1. Post details display post content matching markdown styles (headers, tables, bold text, lists).
  2. Displays author username, avatar, and publication timestamp.
  3. Lists the avatars of all AI agents that have replied.
  4. Status indicator displays current AI progress status ("PENDING" / "PROCESSING" / "COMPLETED").
  5. Renders code blocks with proper monospace formatting.
- **T2 (Boundary & Corner) Tests**:
  1. Safe rendering: Injecting `<script>alert('xss')</script>` or `onload` actions in post content shows plain text or strips tags entirely.
  2. Loading a non-existent `postId` shows a 404 page or redirects to the homepage.
  3. Markdown with broken syntax (e.g. unclosed code block brackets) renders gracefully without page crashes.
  4. Render post details containing extremely long code snippets.
  5. Verification that user profile details do not render unescaped text.
- **T3 (Cross-Feature) Tests**:
  - Admin changes an Agent's name/avatar -> Web Post Details updates the avatars and tags of the agent in the replies section.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 1.

### Feature 5: Web - Comment Tree & Threaded Replies
- **Description**: Standard hierarchical comment tree that lists user and AI responses under a post, allowing users to reply to any specific comment.
- **T1 (Feature Coverage) Tests**:
  1. Comments section displays nested replies with indentation matching hierarchy.
  2. Shows distinct badges for AI agents ("AI" label next to username).
  3. Clicking "Reply" on a comment opens a text input nested under that comment.
  4. Submitting a comment reply successfully updates the comment list.
  5. Timestamps are formatted relative to current time.
- **T2 (Boundary & Corner) Tests**:
  1. Comment content validation (prevent empty comments).
  2. Deep nesting limit: Comment trees indentation levels flat-out after a specific depth (e.g., 5 levels) to prevent horizontal overflow on mobile screens.
  3. Submitting duplicate comments triggers warning or rate limits.
  4. Deleting or modifying a comment with children keeps the children intact under a "Deleted Comment" placeholder.
  5. Render comments containing long strings without spaces (asserts wrap-word formatting).
- **T3 (Cross-Feature) Tests**:
  - Creating a comment under a post -> triggers followup simulation -> Admin Task Queue registers a new FOLLOWUP task for that post.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 3 (Stress/Simulated Network Flakiness and Recovery).

### Feature 6: Web - AI Mention `@AI` & Contextual Follow-up
- **Description**: Rich text or mention capability letting users summon a specific AI Agent by typing `@AgentName`. This forces the selected agent to reply, bypassing willingness score calculations.
- **T1 (Feature Coverage) Tests**:
  1. Typing `@` in comment input field displays a dropdown list of active AI agents.
  2. Selecting an agent inserts the mention component (e.g. `@ArchTechLead`).
  3. Submitting the comment triggers the simulation specifically targeting the mentioned agent.
  4. The mentioned agent responds to the thread, contextually aware of the parent comment.
  5. Verify that the agent replies even if its `replyThreshold` is set to maximum (1.0).
- **T2 (Boundary & Corner) Tests**:
  1. Mentioning an inactive AI Agent (does not display in the dropdown, or if manually typed, ignores execution).
  2. Mentioning multiple agents in a single comment (only the first active agent is triggered or both are queued sequentially).
  3. Backspacing the mention text removes the mention metadata tag.
  4. Mentioning an agent in a thread where the agent has already reached its `maxFollowupRepliesPerPost` (triggers task, but fails or log shows threshold/count exceeded).
  5. Submitting comments containing multiple mock tags resembling mentions (e.g. `@FakeAgent`).
- **T3 (Cross-Feature) Tests**:
  - User submits a mention -> Admin Decision Logs shows triggerType as "MENTION" and indicates score calculations were bypassed.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 1.

### Feature 7: Admin - System Dashboard & Analytics
- **Description**: Operator workspace displaying real-time metrics, system health, service statuses, task execution summary, and decision log updates.
- **T1 (Feature Coverage) Tests**:
  1. Dashboard cards render counts of Total Posts, Total Comments, Total Tasks, and Active AI Agents.
  2. Displays status list of backend simulated services (`api-server`, `worker-service`, `outbox-publisher`) as active.
  3. Renders charts showing task execution success/failure ratios.
  4. Lists recent AI reply tasks and recent decision logs.
  5. Clicking recent logs/tasks redirects to the detailed logs/tasks tab.
- **T2 (Boundary & Corner) Tests**:
  1. Metrics displaying correctly when the database values are zero (counters show 0 instead of NaN or empty spaces).
  2. Service statuses changing to "offline" if simulated connection fails.
  3. Stats refresh interval: Dashboard periodically re-fetches or subscribes to SSE to sync counts.
  4. Large numbers are formatted cleanly (e.g., 15000 -> 15.0k).
  5. The charts render smoothly in dark mode / custom themes.
- **T3 (Cross-Feature) Tests**:
  - Adding multiple posts and comments in Web -> Admin Dashboard counters increment instantly.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 2.

### Feature 8: Admin - AI Agent Configuration Manager
- **Description**: Interface to view active AI agents, update their parameters (thresholds, prompts), and toggle their availability status.
- **T1 (Feature Coverage) Tests**:
  1. Lists all AI agents in a structured table showing personality traits and toggle states.
  2. Clicking "Edit" opens a drawer containing configurations (reply threshold, activity level, prompts).
  3. Changing parameters and clicking "Save" updates the agent properties in the mock database.
  4. Toggling the "Active" status switches the agent inline.
  5. Displays validation warnings on invalid inputs (e.g. threshold > 1.0 or < 0.0).
- **T2 (Boundary & Corner) Tests**:
  1. Setting `replyThreshold` to exactly `0.0` or `1.0` (valid boundary values).
  2. System prompt fields containing extreme amounts of text (drawer handles scrolling cleanly).
  3. Setting `maxAutoRepliesPerPost` to a negative value is blocked by input validation.
  4. Disabling all agents (system displays fallback agents warning or defaults to fallback logic).
  5. Toggling an agent off during an active processing queue (does not disrupt tasks already running).
- **T3 (Cross-Feature) Tests**:
  - Toggle Agent "ArchTechLead" off in Admin -> Create post in Web -> Verify "ArchTechLead" does not evaluate or reply.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 2.

### Feature 9: Admin - AI Task Queue Monitor
- **Description**: Displays the historical queue of task executions, allowing operators to monitor statuses (`PENDING`, `PROCESSING`, `COMPLETED`, `FAILED`) and retry failures.
- **T1 (Feature Coverage) Tests**:
  1. Renders table of AI tasks with filters for status, agent, and trigger type.
  2. Displays task payloads (prompts, results, timestamps) in a drawer detail view.
  3. Displays retry count and error messages for failed tasks.
  4. Completed tasks show start/end times and total duration calculations.
  5. Shows a manual "Retry" action link/button for failed tasks.
- **T2 (Boundary & Corner) Tests**:
  1. Tasks table pagination: Correctly navigates page pages.
  2. Retrying a task that is already `PROCESSING` or `COMPLETED` is disabled.
  3. Displays realistic error strings when task status is `FAILED`.
  4. Task creation date filtering handles timezone conversions cleanly.
  5. Renders prompts containing complex characters and escape sequences.
- **T3 (Cross-Feature) Tests**:
  - Triggering a retry on a failed task in Admin -> changes status to PENDING -> triggers simulated background execution -> comment is appended to Web Post Details page.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 3.

### Feature 10: Admin - AI Decision Logs Auditor
- **Description**: Timeline showing decision evaluations for every agent on every post, documenting why an agent decided to respond or pass.
- **T1 (Feature Coverage) Tests**:
  1. Lists decision logs displaying Agent Name, Trigger Type, Willingness Score, Threshold, and Decision result.
  2. Shows structural "Reason" description (e.g., "Willingness score did not satisfy threshold").
  3. Integrates quick search filter by Post ID or Agent ID.
  4. Highlights "REPLY" decisions with green indicators and "IGNORE" decisions with neutral indicators.
  5. Refreshes instantly on new logs creation.
- **T2 (Boundary & Corner) Tests**:
  1. Display logic when willingness score is equal to the threshold (decides REPLY).
  2. Performance under heavy logs volume: Verify table pagination or virtualized rows load smoothly.
  3. Filter by non-existent Post ID returns empty state gracefully.
  4. Rendering logs where the target agent was deleted (displays "Deleted Agent" placeholder).
  5. Truncating long reasoning text and providing tooltips/popovers.
- **T3 (Cross-Feature) Tests**:
  - User summons AI via `@ArchTechLead` -> Decision Logs record created -> score evaluation bypassed indicator checked.
- **T4 (Real-World Journey) Tests**:
  - Covered in Scenario 2.

---

## 3. 4-Tier Testing Roadmap

The E2E suite will be rolled out progressively following these specifications:

### 3.1 Tier 1: Feature Coverage (Opaque-Box Happy Path)
- **Objective**: Ensure all features perform as expected under normal circumstances.
- **Volume**: At least 5 independent, distinct test cases per feature.
- **Interaction**: UI selectors, form inputs, button clicks, page transitions.

### 3.2 Tier 2: Boundary & Corner (Negative & Edge Case Validation)
- **Objective**: Ensure the system does not crash or corrupt state during anomalies.
- **Volume**: At least 5 edge cases per feature.
- **Focus Areas**: Validation failures, network loss simulation (SSE disconnected state), invalid inputs, empty states, XSS attempt sanitization.

### 3.3 Tier 3: Cross-Feature Combinations (Pairwise Workspace Flows)
- **Objective**: Validate the full closed loop between the User Web App and the Admin Console.
- **Focal Points**:
  - **Auto-Reply Loop**: Post created (Web) -> Task queued (Admin) -> Simulation updates (Admin) -> Comment appears (Web).
  - **Configuration Lock**: Parameter modified (Admin) -> Post created (Web) -> Decision score reflects parameter changes (Admin).
  - **Retry Lifecycle**: Task fails (Admin) -> Retry clicked (Admin) -> State progresses -> Reply renders (Web).

### 3.4 Tier 4: Real-World Application Scenarios
This tier contains high-level integration flows matching common community usage.

#### Scenario 1: The Monolith vs Microservices Debate
- **Sequence**:
  1. User `alex_dev` posts a question: "Is it time to rewrite our Go monolithic API in Rust?"
  2. The Auto-AI trigger processes the post.
  3. `DevilsAdvocate` evaluates: willingness score exceeds threshold, post status becomes `PROCESSING`.
  4. `DevilsAdvocate` publishes reply: Skeptical critique of Rust rewrite complexity.
  5. User reads comment, types reply: "What about GC pauses?" and mentions `@ArchTechLead`.
  6. `ArchTechLead` bypasses willingness checks, processes followup, and publishes comment proposing system telemetry and `pprof` profile steps.
  7. User reads comment, updates thread.
  8. Operator reviews the Admin logs to audit the task lifecycle.

#### Scenario 2: Live Reconfiguration and Moderation Recovery
- **Sequence**:
  1. Operator notices AI agent `GrowthProductManager` is replying too frequently with promotional content.
  2. Operator opens Admin Console, navigates to AI Agent configurations, sets `replyThreshold` to `0.95` and disables `allowAutoReply`.
  3. Operator navigates to Web App and publishes a post.
  4. System logs indicate `GrowthProductManager` skipped ("IGNORE") due to `allowAutoReply` being disabled.
  5. Operator edits agent again, turns `active` status off.
  6. Operator verifies that the agent is completely excluded from the active plaza list in the Web App.

#### Scenario 3: Stress/Simulated Network Flakiness and Recovery
- **Sequence**:
  1. User creates a post.
  2. While the tasks are in `PROCESSING` state, the simulated SSE connection is toggled to `disconnected` (using the connection status control component).
  3. The Web interface correctly displays the network reconnection loading spinner.
  4. The background task finishes in the mock database state.
  5. User toggles the connection back to `connected`.
  6. The client retrieves the catch-up updates, updates the post status to `COMPLETED`, and appends the AI reply comment to the UI without duplications.

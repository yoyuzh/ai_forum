# Handoff Report — Web Pages Explorer 3 (Gen 2)

This report details the architectural analysis and design/implementation strategy for Milestone 2 (Web App Pages), along with a patch proposal to fix the background simulation concurrency bug.

## 1. Observation

During our investigation of the `ai_forum` repository in `CODE_ONLY` mode, we observed the following:

1. **HTML Prototypes**: We found high-fidelity HTML templates located in the `stitch_ai_forum/` directory:
   * **Homepage / Post Feed**: `stitch_ai_forum/ai_forum_5/code.html` (comprises navigation bar, left feed list, right sidebar with popular tags, active AI agents, and a dotted-line timeline for recent AI activity).
   * **Post Details**: `stitch_ai_forum/_2/code.html` (includes detailed post canvas, dynamic AI processing stepper/loader, thread-nested comments with a distinct background for AI comments, and a comment input form with styling buttons and `@AI` trigger selector).
   * **AI Agents Plaza**: `stitch_ai_forum/ai_ai_forum_2/code.html` (features a 3-column roster of AI agent cards containing display names, active status badges, behavior parameters, thresholds, and checkbox rules).
   * **Create Post**: `stitch_ai_forum/ai_forum_4/code.html` (consists of title, category, tags inputs, markdown editor, 4-mode AI participation radio buttons, and a live post card preview sidebar).

2. **React Pages**: The existing react pages are located in `web/src/pages/`:
   * `HomePage.tsx` (using `Virtuoso` for post listing scroll performance and mapping categories/tags).
   * `PostDetailPage.tsx` (rendering markdown via `SafeMarkdown` and loading comments).
   * `AIAgentsPage.tsx` (rendering a grid of `AIAgentCard` components).
   * `PostsListPage.tsx` (handling feed search and post creation logic).

3. **Design Guidelines**: We found styling references:
   * `stitch_ai_forum/design_cohere.md` (defining near-black `#17171c` primary, deep green `#003c33` bands, action blue `#1863dc`, coral `#ff7759` for tags, soft stone `#eeece7` backgrounds, and a Hanken Grotesk / JetBrains Mono font split).
   * `web/tailwind.config.js` (exposing these custom styles under `cohere.*` theme classes, e.g. `bg-cohere-canvas`, `font-display`, `font-mono`, and custom radii).

4. **SSE Simulator Concurrency Bug**: In `web/src/sse/simulator.ts`:
   * The simulator updates the post state to `COMPLETED` inside two delayed execution blocks:
     * Line 151: `db.updatePost(postId, { aiStatus: "COMPLETED" });` inside a 1000ms timeout when `replyQueue.length === 0`.
     * Line 243: `db.updatePost(postId, { aiStatus: "COMPLETED" });` inside a timeout of `totalDuration = (replyQueue.length * 1000) + 4000`.
   * These update calls do not check whether other concurrent background tasks for the same `postId` are still actively running.

5. **Previous Explorer Work**: We read the first-generation explorer files located at:
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/analysis.md`
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/proposed_FeedPage.tsx`
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/proposed_PostDetailPage.tsx`
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/proposed_AgentPlazaPage.tsx`
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/proposed_CreatePostPage.tsx`
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/proposed_App.tsx`
   * `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/simulator_concurrency_fix.patch`

---

## 2. Logic Chain

1. **Routing & Component Consolidation**:
   * The user request specifies consolidating the page structures and routing in `web/src/App.tsx` into `FeedPage.tsx`, `PostDetailPage.tsx`, and `AgentPlazaPage.tsx`.
   * By combining `HomePage.tsx` and `PostsListPage.tsx` into a singular `FeedPage.tsx` at `/` (and `/posts`), we gather feed listings, tags, active agents list, and timeline.
   * By redirecting post creation from `FeedPage` to `/create-post` mapping to `CreatePostPage.tsx` (or embedding the composer), we maintain clarity.
   * Modifying `AIAgentsPage.tsx` to `AgentPlazaPage.tsx` at `/ai-agents` matches the prototype's interactive settings toggles (active status switch, willingness/activity sliders).

2. **Design Tokens**:
   * Aligning with `design_cohere.md` and `DESIGN.md` requires applying the dual-font strategy (Hanken Grotesk for copy and display titles, JetBrains Mono for system metrics, timestamps, and log labels).
   * It also requires utilizing flat layering (white blocks on soft stone backgrounds) and thin borders (`border-cohere-hairline`) instead of card drop shadows.
   * The custom tailwind configuration file in `web/tailwind.config.js` maps these tokens to `cohere-*` CSS properties which must be referenced directly.

3. **Concurrency Fix**:
   * If two simulations run concurrently (e.g. Simulation A starts at t=0 and Simulation B starts at t=1s), Simulation A's scheduled timeout (executing at t=5s) will unconditionally update `aiStatus` to `COMPLETED` even if Simulation B's tasks are still actively `PENDING` or `PROCESSING` at that time.
   * By querying the database `db.getTasks()` and filtering by matching `postId` and status (`PENDING` or `PROCESSING`), we can ensure that we only update `aiStatus` to `COMPLETED` when no other active tasks for that post exist.

---

## 3. Caveats

* **Network Restrictions**: Since we are in `CODE_ONLY` network mode, we did not verify any third-party external CDNs or remote dependencies.
* **Database Mock Limitations**: The database layer is simulated in `web/src/api/db.ts` utilizing `localStorage`. Actual production systems with real SQL backends will require implementing corresponding transactional queries to guarantee isolation levels.

---

## 4. Conclusion

1. **Design Proposal**: The React page components should be restructured to follow the drafts located in `/Users/mac/Documents/ai_forum/.agents/explorer_m2_3/`. Specifically:
   * `web/src/pages/FeedPage.tsx` -> Infinite feed list with `Virtuoso` scroll, categories, popular tags, and dynamic AI status timelines.
   * `web/src/pages/PostDetailPage.tsx` -> Stepper status banner, Markdown renderer via `SafeMarkdown` + `DOMPurify`, and AI comments nested layout with followup controls.
   * `web/src/pages/AgentPlazaPage.tsx` -> Roster cards with toggle active switch and sliders linked to `useAgents()` mutation methods.
   * `web/src/pages/CreatePostPage.tsx` -> 4-mode radio selections and sidebar live cards preview.
2. **Concurrency Fix**: The bug in `web/src/sse/simulator.ts` can be resolved by applying the proposed patch file `simulator_concurrency_fix.patch` generated in our working directory.

---

## 5. Verification Method

To verify the proposed implementation and styling layout:
1. **Linter & Compilability**:
   Run the following validation commands inside the `web/` directory to ensure no TypeScript or CSS styling compile errors exist:
   ```bash
   cd web && npm run lint && npm run build
   ```
2. **Patch Application**:
   Validate that the patch compiles cleanly with no rejects by running:
   ```bash
   git apply --check .agents/explorer_m2_3_gen2/simulator_concurrency_fix.patch
   ```
3. **Interactive Simulation test**:
   Trigger multiple comments or posts quickly in the browser to ensure the AI status indicator transitions to `COMPLETED` only after all spawned AI responses have fully compiled and rendered.

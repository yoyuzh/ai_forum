# Handoff Report — Milestone 2: Web App Pages

## 1. Observation
*   **Prototypes examined**: 
    *   `stitch_ai_forum/ai_forum_5/code.html` (Homepage/Feed)
    *   `stitch_ai_forum/_2/code.html` (Post Details)
    *   `stitch_ai_forum/ai_ai_forum_2/code.html` (AI Plaza)
    *   `stitch_ai_forum/ai_forum_4/code.html` (Create Post)
*   **Existing codebase inspected**:
    *   `web/package.json` contains dependencies for virtualized rendering, markdown rendering, and sanitization:
        ```json
        "dependencies": {
          "@tanstack/react-query": "^5.56.2",
          "dompurify": "^3.1.6",
          "lucide-react": "^0.439.0",
          "react": "^18.3.1",
          "react-dom": "^18.3.1",
          "react-markdown": "^9.0.1",
          "react-router-dom": "^6.26.2",
          "react-virtuoso": "^4.7.11",
          "zustand": "^4.5.5"
        }
        ```
    *   `web/src/sse/simulator.ts` has the following timeout configuration for post completed transitions (lines 238–246):
        ```typescript
        // Schedule final COMPLETED status transition
        const totalDuration = (replyQueue.length * 1000) + 4000;
        setTimeout(() => {
          const latestPost = db.getPost(postId);
          if (latestPost) {
            db.updatePost(postId, { aiStatus: "COMPLETED" });
            sseEmitter.emit("post.updated", db.getPost(postId));
          }
        }, totalDuration);
        ```

## 2. Logic Chain
1.  **Observation 1**: The HTML prototypes define distinct pages for post listing, detailed view, AI plaza settings, and post creation, using precise tailwind design configurations.
2.  **Observation 2**: `web/package.json` contains `react-virtuoso` for virtualized list rendering, `react-markdown` for parsing rich text, and `dompurify` for HTML sanitization.
3.  **Deduction 1**: We can structure four modular pages under `web/src/pages/` (FeedPage, PostDetailPage, AgentPlazaPage, CreatePostPage) and orchestrate routing through `react-router-dom` in `web/src/App.tsx` matching these prototypes.
4.  **Observation 3**: `simulator.ts` schedules an unconditional `aiStatus = "COMPLETED"` update after a hardcoded `totalDuration`.
5.  **Deduction 2**: If multiple client triggers (comments/follow-ups) launch background simulations concurrently, an earlier simulation's timeout will update the post to `COMPLETED` while subsequent simulated reply tasks remain active (`PENDING` or `PROCESSING`), leading to data inconsistency.
6.  **Resolution**: We must check `db.getTasks()` and ensure no task matching `postId` is in `PENDING` or `PROCESSING` state before updating the status to `COMPLETED`.

## 3. Caveats
*   We assumed the user is running React 18 and Vite as indicated in `web/package.json`.
*   We did not investigate the production Go backend services, as Milestone 2 focuses exclusively on client-side React page layouts and mock SSE simulation.

## 4. Conclusion
We have formulated a robust design strategy and written the proposed source files representing:
*   `proposed_App.tsx` (routing & layout)
*   `proposed_FeedPage.tsx` (virtualized list feed & sidebar)
*   `proposed_PostDetailPage.tsx` (sanitized markdown display & virtual comments tree)
*   `proposed_AgentPlazaPage.tsx` (agent toggle & configuration sliders)
*   `proposed_CreatePostPage.tsx` (publishing workflow & mode selectors)
*   `concurrency_bug.patch` (precise bug fix targeting simulator concurrency race conditions)

## 5. Verification Method
To verify the implementation:
1.  Copy the proposed pages into the active workspace:
    ```bash
    cp proposed_FeedPage.tsx web/src/pages/FeedPage.tsx
    cp proposed_PostDetailPage.tsx web/src/pages/PostDetailPage.tsx
    cp proposed_AgentPlazaPage.tsx web/src/pages/AgentPlazaPage.tsx
    cp proposed_CreatePostPage.tsx web/src/pages/CreatePostPage.tsx
    cp proposed_App.tsx web/src/App.tsx
    ```
2.  Apply the simulator bug fix patch:
    ```bash
    git apply concurrency_bug.patch
    ```
3.  Verify the project compiles without errors and check the linter output:
    ```bash
    cd web
    npm run build
    npm run lint
    ```

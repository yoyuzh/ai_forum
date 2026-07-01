# Handoff Report: Milestone 1: Web App Init & Mock Layer Design

## 1. Observation
- `web/` contains only `AGENTS.md` and empty subdirectories with `README.md` placeholders (e.g. `web/src/api/README.md`, `web/src/sse/README.md`). No `package.json`, `vite.config.ts`, `tailwind.config.js`, or source logic exists.
- `design_cohere.md` and `DESIGN.md` outline specific typography (Hanken Grotesk, Unica77, JetBrains Mono, CohereMono), colors (Action Blue, Coral, Deep Green, etc.), border radii (xs: 4px to pill: 32px), and spacing grids.
- `PROJECT.md` and `SCOPE.md` define exact interface schemas for `Post`, `Comment`, `AIAgent`, `AIReplyTask`, and `AIDecisionLog`.

## 2. Logic Chain
- To enable developers to compile and build the frontend, we must initialize `web/package.json` with correct build/dev scripts and dependencies (`react-virtuoso`, `react-markdown`, `dompurify`, `@tanstack/react-query`, `zustand`, `tailwindcss`, `vite`).
- To implement the design variables faithfully, we map custom theme tokens to CSS custom variables in `web/src/styles/index.css` and extend Tailwind colors/radii in `web/tailwind.config.js`.
- To establish a client-side mock backend, we map schemas to TypeScript interfaces in `web/src/api/types.ts` and write a persistence manager `web/src/api/mockDb.ts` backed by `localStorage` (`__ai_forum_db__`).
- To simulate the asynchronous outbox-worker flow of the backend, we construct a Pub/Sub client `MockSSEHub` and an async simulator loop `runAgentSimulation(postId)` that evaluates agent reply thresholds, updates status transitions, appends mock AI comments, and pushes events.
- To reflect updates instantly in the UI, the custom SSE hook `usePostSSE` intercepts events and updates TanStack Query's cache (`['comments', postId]` and `['post', postId]`) in real-time.

## 3. Caveats
- Since the user application (`web/`) and admin console (`admin/`) are separate React applications, they will only share the `localStorage` database state if they are hosted on the same origin (e.g., via Nginx or a Vite dev server proxy).
- Simulating LLM text completion is done through static template matching based on keywords in the post content (such as "Go", "Zustand").

## 4. Conclusion
- The designed implementation plan for Milestone 1 is robust, fully typed, persistent, and integrates simulated backend asynchronous processes (Outbox/Asynq) into a client-side real-time experience using a Pub/Sub hub and query-cache mutations.

## 5. Verification Method
- **Vite Compilation**: Change directory to `web/` and execute `npm run build` once the files are written. The compiler should output 0 errors.
- **Asynchronous Execution**: Invoke a mutation for a new post or comment, check that `localStorage` has registered the `__ai_forum_db__` structure, and verify that the SSE hook receives updates and updates the query cache.

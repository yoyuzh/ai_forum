# Handoff Report: Web App Init & Mock Layer Analysis

This handoff outlines the analysis, structural designs, and configuration definitions prepared for the initialization and mock database/SSE setup of the User Web Application (`web/`).

## 1. Observation
- **Root Directory Layout**: Searched files in `/Users/mac/Documents/ai_forum/`. Confirmed no existing `package.json` file is present at the workspace root or nested subdirectories (found 0 matches).
- **Web Project Placeholder State**:
  - `web/AGENTS.md` (lines 12-20) defines the allowed dependencies: `"React, TypeScript, Vite, Ant Design, TanStack Query, Zustand, React Router, React Virtuoso, Tiptap or mentions input, react-markdown, DOMPurify, and small UI utilities"`.
  - `web/src/App.tsx` contains only: `// User web root component placeholder.`.
  - `web/src/main.tsx` contains only: `// User web application entrypoint placeholder.`.
- **Design Guidelines**:
  - `/Users/mac/Documents/ai_forum/stitch_ai_forum/design_cohere.md` defines the Cohere brand palette, including `primary` (`#17171c`), `cohere-black` (`#000000`), `deep-green` (`#003c33`), `soft-stone` (`#eeece7`), and `coral` (`#ff7759`).
  - `/Users/mac/Documents/ai_forum/stitch_ai_forum/synthetica_ai_forum/DESIGN.md` defines custom styling layers, where functional elements (buttons/inputs) use `xs/sm` (4px-8px) corners, and posts/agent cards use `md/lg` (16px-22px) corners. It details the font fallback split: CohereText (display) and Unica77/Hanken Grotesk (body/UI).

## 2. Logic Chain
- Since no `package.json` files are initialized, a complete boilerplate workspace must be specified under `web/` using Vite + TypeScript, declaring the dependencies approved in `web/AGENTS.md` (React 18, React Router, TanStack Query, Zustand, React Virtuoso, DOMPurify, Tailwind, etc.).
- Cohere colors and spacing metrics must be mapped to CSS custom variables in `web/src/styles/index.css` and linked dynamically into `web/tailwind.config.js` to avoid hardcoding theme colors in component files.
- The browser mock data layer requires an in-memory database (`db.ts`) configured with `localStorage` persistence, ensuring that user posts/comments update state synchronously and trigger simulated background event pipelines.
- SSE simulation can be resolved client-side via a unified custom event emitter on `window` or a dedicated listener class (`emitter.ts`). When a user mutates data, a background generator (`simulator.ts`) staggers agent triggers, writes decision logs, handles task status changes (`PENDING` -> `PROCESSING` -> `COMPLETED`), and writes the corresponding reply comment.
- TanStack Query hooks should wrap the mock client endpoints and subscribe to simulated SSE triggers to automatically invalidate caches on `post.updated`, `comment.created`, and `task.updated` events.

## 3. Caveats
- **Cross-Origin Storage**: LocalStorage is naturally partitioned by origin. If the User Web App runs on `http://localhost:3000` and the Admin Console runs on `http://localhost:3001`, their local storage stores will be isolated. The implementation agent must document that during test runs, they should be served on the same origin (e.g. proxying `/admin` requests) or synchronization code (using frames or postMessage) should be considered if cross-origin testing is necessary. We assume here they will share the same origin or run on the same server route `/admin` for testing.
- **AI Response Content**: The AI reply generator uses pre-configured template responses rather than real LLM API connections to ensure offline consistency and zero API-key dependencies.

## 4. Conclusion
The initialization configurations and mock layer designs detailed in `analysis.md` provide a complete, low-coupling blueprint that implements the mock backend and real-time event simulation in-browser. This will allow pages to fetch, filter, and render simulated AI threads in high fidelity.

## 5. Verification Method
1. **Config Verification**: Confirm that `package.json` successfully installs packages under node, and `npm run dev` boots the dev server at `localhost:3000`.
2. **Data Integrity Test**: Create a post via `api.posts.create()`. Verify that:
   - A new post appears in `db.getPosts()`.
   - The status is immediately updated in local storage.
3. **Simulated Event Loop Trace**: Subscribe to `sseEmitter` using a script. Verify that posting a comment yields `task.created` -> `task.updated` (PROCESSING) -> `comment.created` -> `task.updated` (COMPLETED) in order over 3-5 seconds.
4. **Markdown Security**: Verify that rendering custom HTML or script elements in markdown comments fails to execute due to `dompurify` sanitization.

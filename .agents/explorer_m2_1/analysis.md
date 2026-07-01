# Milestone 2: Web App Pages — Analysis and Implementation Strategy

This document provides a comprehensive analysis and implementation strategy for the user-facing web pages and components of the AI Forum monolith, aligning HTML prototypes with existing client-side stores, query hooks, and the design systems.

---

## 1. Analysis of HTML Prototypes

We examined the static HTML prototypes located in `stitch_ai_forum/` to map layout zones, styling variables, and interaction components to React.

### A. Homepage / Post Feed (`stitch_ai_forum/ai_forum_5/code.html`)
*   **Grid Structure**: Asymmetrical 12-column layout (Main Feed: `lg:col-span-8`, Sidebar: `lg:col-span-4`).
*   **Feed Design**: Thin horizontal separators (`border-b border-hairline`), clean typography hierarchy. Active tags are indicated by colored badges.
*   **Sidebar Content**:
    *   *Active Agents*: A roster with circular avatars showing agent names and brief descriptions.
    *   *Recent Activity log*: A dotted-line timeline layout displaying recent comments written by AI agents.
*   **Key CSS Styling**: Background uses `#fbf9f4` (`bg-background`), content text uses `#1b1c19` (`text-on-surface`), card frames use `#ffffff` (`bg-surface-container-lowest`).

### B. Post Details (`stitch_ai_forum/_2/code.html`)
*   **Article Header**: Displays title with a massive headline scale (`font-headline-xl`), author's avatar, timestamp, and numeric telemetry (views, comments counts).
*   **AI Processing Status Bar**: A timeline tracker located in the sidebar showing the progression of AI responses:
    1.  *Extract Context Tags* (Complete)
    2.  *Compute AI Willingness Scores* (Active / complete)
    3.  *LLM Generation* (Running)
    4.  *Post to Thread* (Pending)
*   **Comments Layout**: Alternates standard user replies (rounded cards) with AI responses (styled in a pale green-wash panel `#f5f7f6` with a `psychology` brain icon). A dotted left thread guide line represents nesting depth.
*   **Interaction Controls**: Includes "Ask Followup" buttons next to AI comments to request elaboration.

### C. AI Agents Plaza (`stitch_ai_forum/ai_ai_forum_2/code.html`)
*   **Grid layout**: 3-column roster cards.
*   **Agent Profile Card**: Displays avatar, specialty designation (e.g. "System Architect"), age setting, and trait tags (e.g. "Pragmatic", "Analytical").
*   **Controls**: Includes toggles showing state ("Auto Reply" or "On/Off") and configuration triggers.

### D. Create Post (`stitch_ai_forum/ai_forum_4/code.html`)
*   **Form controls**: Standard text inputs, category dropdown, tags input field, and a detailed Markdown textarea.
*   **AI Mode Panel**: Interactive radio selection for AI interaction modes:
    *   *Humans Only*: Suppresses AI.
    *   *Low AI*: High willingness thresholds required.
    *   *Standard AI*: Default activity levels.
    *   *Spirited Mode*: Maximized response frequency.
*   **Live Preview**: Real-time card mockup rendering.

---

## 2. Page Component Structure & Routing Design

To integrate the prototypes into `web/src/`, we design four pages under `web/src/pages/` and configure routes in `web/src/App.tsx`.

### A. Component Mapping
| Page Component | Path | Key Functionalities | Key Dependencies / Hooks |
|---|---|---|---|
| `FeedPage.tsx` | `/` or `/home` | Virtualized post list, search input, category tab switcher, active agents list, tag filter. | `usePosts`, `useAgents`, `useFilterStore`, `react-virtuoso` |
| `PostDetailPage.tsx` | `/post/:id` | Full article view, Markdown parser, virtualized flat comment tree, AI processing checklist, decision logs. | `usePostDetail`, `useComments`, `react-virtuoso`, `react-markdown`, `dompurify` |
| `AgentPlazaPage.tsx` | `/agents` | Agent cards display, settings drawer (system prompts, reply threshold sliders, toggle switches). | `useAgents`, `updateAgent` mutation |
| `CreatePostPage.tsx` | `/create-post` | Create topic form, AI mode selector cards, live sidebar mockup card rendering. | `usePosts` (`createPost`) |

### B. Routing Integration (`web/src/App.tsx`)
We introduce `react-router-dom` to manage clean routing.
*   A global `Layout` wraps the pages, rendering an **Announcement Bar** at the top showing the current SSE connection state (`sseStatus` from `useConnectionStore`), a global sticky `Header` (TopNavBar), and a global `Footer`.

*Detailed implementation structures are stored as proposed files in this agent folder.*

---

## 3. Design System Alignment (`design_cohere.md` & `DESIGN.md`)

Our components strictly match the visual tokens of the Cohere design:

### A. Color Palette Alignment
*   **Canvas & Surface**: Background set to `#ffffff` (`bg-cohere-canvas`), sidebar zones and cards use `#eeece7` (`bg-cohere-soft-stone`) or `#ffffff` with a `#d9d9dd` (`border-cohere-hairline`) border.
*   **Primary Text**: Near-black ink `#212121` (`text-cohere-ink`) and `#17171c` (`text-cohere-primary`).
*   **Accents**: Action links use `#1863dc` (`text-cohere-action-blue`), and taxonomy markers use `#ff7759` (`text-cohere-coral`).
*   **Status Indicators**: AI processing tracks use `#003c33` (`text-cohere-deep-green`) and `#edfce9` (`bg-cohere-pale-green`).

### B. Typography Split
*   **UI/Headlines**: Rendered in Unica77/Hanken Grotesk (`font-sans`), styled with tight line-heights and negative letter-spacing for displays.
*   **Technical Labels**: Rendered in JetBrains Mono (`font-mono`), used for metadata tags, AI scoring, timestamps, and log records.

### C. Radii & Spacing
*   **Borders**: Cards use a precise `16px` radius (`rounded-md`). Inputs use `8px` (`rounded-sm`), and primary action buttons use a `32px` pill radius (`rounded-pill`).
*   **Negative Space**: Spacing utilizes the 8px-grid values. Sections maintain wide `80px` gaps (`py-section` / `gap-section`) to enforce editorial structure.

---

## 4. SSE Simulation Concurrency Bug Fix

### A. The Bug Analysis (`web/src/sse/simulator.ts`)
In the original simulator implementation, when a background simulation is triggered for a post (on creation or followup reply), the code schedules an unconditional transition of the post state to `COMPLETED` using a simple timer:
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
**Why this fails (Concurrency Race Condition)**:
If multiple simulations are triggered concurrently on the same post (e.g. a user writes a new comment while a previous AI reply task is still executing, or multiple users comment at similar times), a timer scheduled by an earlier simulation will fire and force the post's state to `COMPLETED` prematurely, even though newer tasks are still in `PENDING` or `PROCESSING` state in the database.

### B. The Solution Strategy
We introduce a helper function:
```typescript
function checkAndTransitionPostToCompleted(postId: number) {
  const tasks = db.getTasks().filter(t => t.postId === postId);
  const activeTasks = tasks.filter(t => t.status === "PENDING" || t.status === "PROCESSING");
  if (activeTasks.length === 0) {
    const latestPost = db.getPost(postId);
    if (latestPost && latestPost.aiStatus !== "COMPLETED") {
      db.updatePost(postId, { aiStatus: "COMPLETED" });
      sseEmitter.emit("post.updated", db.getPost(postId));
    }
  }
}
```
We apply this validation check before making any transition to `COMPLETED`:
1.  **Simulation Timeout Check**: In the final scheduled `setTimeout`, we call `checkAndTransitionPostToCompleted(postId)` instead of unconditionally updating the state.
2.  **Task Life-Cycle Check**: At the end of each task execution loop (when a task's status changes to `COMPLETED`), we call `checkAndTransitionPostToCompleted(postId)`. This acts as an immediate check as soon as simulated generation ends.
3.  **No-Reply Short Circuit**: If `replyQueue.length === 0`, we check active tasks before updating status to ensure we do not conflict with previous ongoing simulations.

---

## 5. Directory Index of Proposed Artifacts

All proposed components and the bug fix patch are generated in this agent directory:
*   `proposed_App.tsx` — Routing and root layout setup.
*   `proposed_FeedPage.tsx` — Feed listing, Virtuoso rendering, sidebar active agents and tag filters.
*   `proposed_PostDetailPage.tsx` — Post reading canvas, Markdown/Purify viewer, Virtuoso comment tree, dynamic AI timeline.
*   `proposed_AgentPlazaPage.tsx` — AI Roster grid, settings toggles and threshold customization sliders.
*   `proposed_CreatePostPage.tsx` — Post write layout with AI mode selector.
*   `concurrency_bug.patch` — Unified diff patch fixing the concurrency bug.

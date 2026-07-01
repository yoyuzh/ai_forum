# Milestone 2: Web App Pages — Analysis and Strategy Report

This report presents a thorough analysis and design/implementation strategy for **Milestone 2: Web App Pages**, including HTML prototype extraction, page component hierarchies, routing structure in `web/src/App.tsx`, CSS design system compliance, and the resolution of the asynchronous simulation concurrency bug in `web/src/sse/simulator.ts`.

---

## 1. Prototype Structure & Asset Analysis

We examined the high-fidelity HTML prototypes inside `stitch_ai_forum/` and extracted their styling conventions, page layout skeletons, and interactive mechanisms.

### A. Homepage / Post Feed (`stitch_ai_forum/ai_forum_5/code.html`)
*   **Grid Structure**: Standard 12-column responsive layout. On desktop, it splits into an asymmetric **8-column left workspace** and a **4-column right sidebar** with `gap-lg` (24px) gutters.
*   **Header & Footer**: High-contrast, sticky header (64px / `h-16`) containing logo, search bar, main menu, notifications button, and profile entry. The footer occupies a large segment (`py-section`, 80px) at the bottom, containing legal and API links.
*   **Hero Card**: Headline set in `font-display-lg` (72px) with tight tracking and a single-line explanation, accompanied by two key CTAs ("发布帖子" as a solid dark pill and "查看 AI 角色" as a border-only pill).
*   **Navigation & Tabs**: Category filter tabs ("最新", "最热", "待回复", "AI 参与最多") with bottom border emphasis indicating the active filtering status.
*   **Post Cards**: Rendered with `#d9d9dd` hairline borders and `rounded-[16px]` corners. Uses `line-clamp-2` for content summaries and a right-aligned indicator displaying active AI replies or processing loaders.

### B. Post Details (`stitch_ai_forum/_2/code.html`)
*   **Grid Structure**: Shares the same responsive 8:4 grid layout.
*   **Left Column**:
    *   **Post Content Canvas**: Displays the post category, localized date/time, and title (`font-headline-xl`, 48px). Author information (avatar, name, role) is placed adjacent to view count/reply count parameters.
    *   **Comment Form**: Includes Markdown formatting helper buttons (Bold, Code) and a clickable `@AI` label button to quickly target agents.
    *   **User/AI Comments Feed**: Shows a clear vertical nested thread tree using a dotted line to indicate response paths. AI-authored comments stand out with a distinct green-tinted background (`bg-[#f5f7f6]`, border `#e0e5e3`), featuring willingness scores (e.g., `意愿分: 92/100`) and a "继续追问" trigger.
*   **Right Column**:
    *   **AI Processing Status**: Stepper component illustrating a 4-step processing timeline:
        1.  `分析与标签提取` (Semantic Parsing)
        2.  `评估回答意愿` (Willingness Scoring)
        3.  `生成 AI 回复` (LLM Generation)
        4.  `写入评论区` (DB Commit)
    *   **Participating AI**: Vertical list showing active decision parameters (willingness scores, decisions like `REPLY`/`IGNORE`) for all matched agents.

### C. AI Agents Plaza (`stitch_ai_forum/ai_ai_forum_2/code.html`)
*   **Grid Structure**: 3-column responsive card grid (`grid-cols-1 md:grid-cols-2 lg:grid-cols-3`) with generous spacing.
*   **Roster Card Layout**:
    *   AI Avatar: Utilizes a signature **22px border radius** with an adjacent active/inactive status light.
    *   Configuration options: Custom threshold sliders for `replyThreshold` (Willingness) and `activityLevel` (Activity), checkboxes to toggle trigger permissions (`allowAutoReply`, `allowMentionReply`, `allowFollowupReply`), and limit inputs.
    *   Prompt inspect section: Displays raw prompt text boxes (System and Speaking Style) styled inside monospace font containers.

### D. Create Post (`stitch_ai_forum/ai_forum_4/code.html`)
*   **Form Area (8 cols)**: Single input fields for Title, select list for Category, tag field, and a dedicated Markdown-supporting text editor with editing shortcuts.
*   **AI Mode Selection**: 4 radio card triggers implementing AI personality modes:
    *   `humans-only` (仅人类): Blocks all AI replies.
    *   `low-ai` (少量 AI): Restricts replies to high-willingness scorers or direct `@` mentions.
    *   `standard-ai` (标准 AI): Default simulation behavior.
    *   `busy-ai` (热闹模式): Lowers thresholds to stimulate high engagement.
*   **Live Preview Sidebar (4 cols)**: A real-time rendering of how the post card will appear on the homepage, alongside a workflow visualizer explaining the publishing process steps.

---

## 2. Page Components & Routing Design

The React web application under `web/` will consolidate its views and routing to reflect the modular pages requested.

### A. Routing Configuration in `web/src/App.tsx`
The primary layout wrapper (`HeaderAndLayout`) contains the announcement status bar, top navigation header, routing outlet, and the shared footer.
```typescript
import React from "react";
import { BrowserRouter, Routes, Route, Link, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { FeedPage } from "./pages/FeedPage";
import { PostDetailPage } from "./pages/PostDetailPage";
import { AgentPlazaPage } from "./pages/AgentPlazaPage";
import { CreatePostPage } from "./pages/CreatePostPage";
import AppLayout from "./components/layout/AppLayout";

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            {/* Primary navigation routes */}
            <Route path="/" element={<FeedPage />} />
            <Route path="/home" element={<Navigate to="/" replace />} />
            <Route path="/posts" element={<FeedPage />} />
            <Route path="/posts/:id" element={<PostDetailPage />} />
            <Route path="/ai-agents" element={<AgentPlazaPage />} />
            <Route path="/create-post" element={<CreatePostPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
```

### B. Consolidated Page Architectures

#### 1. `FeedPage.tsx` (Consolidates Post Feeds & Filters)
*   **Virtualization**: Uses `react-virtuoso`'s `Virtuoso` component with `useWindowScroll={true}` to display the post cards feed. This ensures the DOM footprint remains flat as the feed grows.
*   **Sidebar Controls**:
    *   **Popular Tags**: Extracted dynamically from the posts payload, showing interactive keyword tags.
    *   **Active AI Agents**: Roster showcasing active agents, linking to the plaza configuration page.
    *   **Recent AI Activities**: Subscribes to the decision logs SSE event stream to show real-time comments and updates in a vertical timeline.
*   **Filter States**: Binds category click handlers to the global Zustand `useFilterStore`, triggering query updates.

#### 2. `PostDetailPage.tsx` (Detail View & AI Processing Timeline)
*   **Rich Text rendering**: Incorporates the custom `SafeMarkdown` component which sanitizes the raw content string via `DOMPurify.sanitize` and parses it into JSX nodes using `ReactMarkdown`.
*   **Comments Feed**: Differentiates between `HumanComment` and `AIComment`. AI comments are styled with unique colors, displaying willingness badges and an inline trigger button calling `handleAskFollowup` (appends `@Agent` and sets parent comment ID).
*   **AI Stepper**: Uses Tailwind classes mapping current task parameters:
    *   *Pending*: Dotted gray circles.
    *   *Processing*: Rotating hourglasses/spinners.
    *   *Completed/Failed*: Checkmarks and status-colored labels.

#### 3. `AgentPlazaPage.tsx` (AI Configuration Control Center)
*   ** Roster Grid**: Grid containing interactive `AIAgentCard` components.
*   **Config UI**:
    *   **Enable/Disable Toggle**: Uses a toggle switch styled with Tailwind peer modifiers. Triggers `updateAgent({ id, updates: { active: !agent.active } })`.
    *   **Threshold Sliders**: Sliders modifying `replyThreshold` and `activityLevel`.
    *   **Rule Selectors**: Checkbox list to toggle the agent's behavior flags.

#### 4. `CreatePostPage.tsx` (Composer with Real-time Previews)
*   **Creation form**: Text fields for title, tag strings, and category, and a markdown textarea.
*   **Modes Card Grids**: 4 radio card elements to choose the AI participation style, mapping to the backend payload.
*   **Preview Card**: Left-to-right state sync. Updates a mock post card in the sidebar as the user types, including the target AI mode.

---

## 3. Design System & CSS Token Alignment

Our pages will strictly align with the specifications defined in `design_cohere.md` and `DESIGN.md`.

| Design Attribute | Token Reference | CSS Mapping / Tailwind Classes | Applied Context |
| :--- | :--- | :--- | :--- |
| **Theme Canvas** | Stark Neutral | `bg-cohere-canvas` (`#ffffff`) \| `bg-cohere-background` (`#fbf9f4`) | Main background surfaces |
| **Secondary Neutral** | Soft Stone | `bg-cohere-surface-low` (`#f5f3ee`) \| `bg-cohere-soft-stone` (`#eeece7`) | Sidebar cards, secondary layouts |
| **Deep Band** | Brand Green | `bg-cohere-deep-green` (`#003c33`) \| `bg-cohere-primary` (`#17171c`) | Status banners, headers, primary buttons |
| **Interactive Blue** | Action Blue | `text-cohere-action-blue` (`#1863dc`) | Inline hyperlinks, secondary actions |
| **Taxonomy Orange** | Coral Accent | `border-cohere-coral` \| `text-cohere-coral` (`#ff7759`) | Categories, post tags, warning borders |
| **UI Typography** | Sans Type | `font-display` \| `font-sans` (Hanken Grotesk / Inter) | Headlines, body text, buttons |
| **Technical Labels** | Monospace | `font-mono` (JetBrains Mono) | Timestamps, counters, logs, code, metrics |
| **Functional Radius** | Soft Corner | `rounded-xs` (4px) \| `rounded-sm` (8px) | Text inputs, dropdown lists |
| **Standard Card Radius** | Card Corner | `rounded-lg` (16px) | Post cards, agent plaza cards |
| **Signature Radius** | Smooth Corner | `rounded-ai` (22px) | AI avatars, highlight images |
| **Button Radius** | Pill | `rounded-pill` (32px) | Primary solid CTAs, form submissions |
| **Grid Gutters** | Column Gaps | `gap-gutter` (16px) \| `gap-lg` (24px) | Responsive grid spacing |

### Layout Spacing & Depth Rules:
1.  **Whitespace**: Desktop layout maintains massive `80px` section spacing gaps to establish editorial hierarchy. Spacing reduces to `48px` on tablets and `32px` on mobile displays.
2.  **Elevation**: The system is completely flat. **No box-shadows** are permitted. Card boundaries are established using a 1px hairline border (`border-cohere-hairline` / `#d9d9dd`) and contrasting background color block alternation (e.g., white cards sitting on soft-stone backgrounds).

---

## 4. SSE Simulator Concurrency Bug Resolution

### A. Root Cause Analysis
Overlapping simulation lifecycles (e.g., when a user submits comments or posts in quick succession) cause concurrent invocations of `runBackgroundAISimulation`. Each call schedules independent `setTimeout` callbacks:
1.  If no replies are scheduled, a 1000ms timer transitions the post's `aiStatus` to `COMPLETED`.
2.  If replies are scheduled, a staggered queue processes the AI replies, and a final timer `totalDuration = (replyQueue.length * 1000) + 4000` transitions the post's `aiStatus` to `COMPLETED`.

If a fast simulation runs concurrently with a slow simulation, the fast simulation's timeout will execute early, unconditionally updating the post's `aiStatus` to `COMPLETED` and broadcasting it. This occurs even if the slow simulation's tasks are still actively in `PENDING` or `PROCESSING` states in the database, resulting in premature loading state termination in the UI.

### B. Resolution Strategy
Before updating the post's status to `COMPLETED` inside any simulation timeout callback, the simulator must check the database for any other tasks associated with the same `postId` that are currently in `PENDING` or `PROCESSING` states. The post status is updated to `COMPLETED` *only* if no such active tasks remain.

### C. Implementation Patch for `web/src/sse/simulator.ts`
We have formulated the precise patch code to fix this concurrency issue:

```typescript
diff --git a/web/src/sse/simulator.ts b/web/src/sse/simulator.ts
index b0c6ff0..00419d8 100655
--- a/web/src/sse/simulator.ts
+++ b/web/src/sse/simulator.ts
@@ -148,8 +148,11 @@ export function runBackgroundAISimulation(postId: number, commentId: number | nu
     setTimeout(() => {
       const latestPost = db.getPost(postId);
       if (latestPost) {
-        db.updatePost(postId, { aiStatus: "COMPLETED" });
-        sseEmitter.emit("post.updated", db.getPost(postId));
+        const activeTasks = db.getTasks().filter(t => t.postId === postId && (t.status === "PENDING" || t.status === "PROCESSING"));
+        if (activeTasks.length === 0) {
+          db.updatePost(postId, { aiStatus: "COMPLETED" });
+          sseEmitter.emit("post.updated", db.getPost(postId));
+        }
       }
     }, 1000);
     return;
@@ -240,8 +243,11 @@ export function runBackgroundAISimulation(postId: number, commentId: number | nu
   setTimeout(() => {
     const latestPost = db.getPost(postId);
     if (latestPost) {
-      db.updatePost(postId, { aiStatus: "COMPLETED" });
-      sseEmitter.emit("post.updated", db.getPost(postId));
+      const activeTasks = db.getTasks().filter(t => t.postId === postId && (t.status === "PENDING" || t.status === "PROCESSING"));
+      if (activeTasks.length === 0) {
+        db.updatePost(postId, { aiStatus: "COMPLETED" });
+        sseEmitter.emit("post.updated", db.getPost(postId));
+      }
     }
   }, totalDuration);
 }
```

This patch resolves the concurrency bug while keeping the simulator completely robust against multiple fast overlapping requests.

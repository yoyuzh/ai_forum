# Milestone 2 Web App Pages: Analysis & Strategy

This document provides the structural analysis of the HTML prototypes, page component architecture, responsive design guidelines, and the strategy to resolve the background simulation concurrency bug.

---

## 1. Prototype Analysis & Key Extractions

We examined four prototype HTML templates under `stitch_ai_forum/` and extracted their functional structure, typography classes, color variables, and visual layout.

### A. Homepage / Post Feed (`stitch_ai_forum/ai_forum_5/code.html`)
- **Key Sections**:
  - **Top Navigation**: Clean bar with site branding, a search input, links to navigation targets (首页, 帖子, AI 角色), notification button, and user status profile center.
  - **Left Area (8/12 grid columns)**:
    - **Hero Card**: Titled "AI Forum" with explanation text and CTA buttons to create a new post or browse active agents.
    - **Feed Tabs**: Interactive tabs ("最新", "最热", "待回复", "AI 参与最多") indicating sorting and category criteria.
    - **Post Feed List**: Cards showing title, content snippet (line-clamp), category badge (e.g., `#后端开发` with coral style), author profile, tag list, and AI status indicators.
  - **Right Area (4/12 grid columns)**:
    - **Popular Tags**: Horizontal tag list.
    - **Active AI Agents**: Roster displaying agent avatars, names, and taglines.
    - **Recent AI Activities**: Dotted-timeline layout displaying real-time agent decision and response logs.

### B. Post Details (`stitch_ai_forum/_2/code.html`)
- **Key Sections**:
  - **Post Container**: Large layout presenting category badge, post metadata (view count, publish time), title, author profile, and formatted post content.
  - **AI Processing Panel**: Real-time progress bar/stepper illustrating:
    1. Tag Extraction (`分析与标签提取`)
    2. Willingness Evaluation (`评估回答意愿`)
    3. Response Generation (`生成 AI 回复` with active loader)
    4. Comment Writing (`写入评论区` as waiting/done)
  - **Comments Feed**: Vertical lists representing comments:
    - **Human Comments**: Traditional style showing avatar, username, metadata, content body, and action buttons (Reply, Like).
    - **AI Comments**: Styled with a unique tint background (`bg-[#f5f7f6]`, border `#e0e5e3`), displaying specific badges (e.g. `发帖后自动回复`), willingness score metadata (`意愿分: 92/100`), and interactive triggers (e.g. `继续追问` button).
  - **Comments Form**: Large textarea with custom styling, styling shortcuts (Bold, Code), and a quick `@AI` selector button.

### C. AI Agents Plaza (`stitch_ai_forum/ai_ai_forum_2/code.html`)
- **Key Sections**:
  - **Layout Header**: Overview text explaining AI Agent perspectives.
  - **Filters**: Horizontal pills to filter agents by category domains (技术, 生活, 情感, etc.).
  - **Agent Cards Grid**: 3-column card layouts:
    - Profile picture (22px signature border-radius).
    - Details: Personality style, system prompts, default viewpoint tags, activity thresholds, and automatic/mention flags.
    - Status badges (Active/Disabled).

### D. Create Post (`stitch_ai_forum/ai_forum_4/code.html`)
- **Key Sections**:
  - **Form fields**: Input for title, select dropdown for category, input for comma-separated tags, and a rich markdown text editor.
  - **AI Mode radio options**:
    - **Humans Only** (`humans-only`): Avoids all AI replies.
    - **Low AI** (`low-ai`): Standard selective responses.
    - **Standard AI** (`standard-ai`): Regular agent replies.
    - **Busy AI** (`busy-ai`): Extreme agent engagement.
  - **Live Preview Sidebar**: Side panel demonstrating real-time card render.

---

## 2. Page Components and Routing Strategy

The pages will be structured as individual components under `web/src/pages/` and routed in `web/src/App.tsx` using `react-router-dom`. The implementation maps directly to query clients, global states, and SSE channels.

### A. Routing Configuration in `web/src/App.tsx`
The primary layout structures the layout skeleton (Announcement Bar, Top Navigation Bar, Page Outlet, and Footer).
Routes will map:
- `/` & `/home` & `/posts` $\rightarrow$ `FeedPage`
- `/posts/:id` $\rightarrow$ `PostDetailPage`
- `/ai-agents` $\rightarrow$ `AgentPlazaPage`
- `/create-post` $\rightarrow$ `CreatePostPage`

*Detailed mockup file can be found in `proposed_App.tsx`.*

### B. Page Modules Design
1. **`FeedPage.tsx`**
   - **List Virtualization**: Integrate `Virtuoso` from `react-virtuoso` with `useWindowScroll={true}` to support smooth, infinite scroll rendering for heavy discussion threads.
   - **State Integration**: Connects with `useFilterStore` to filter posts using `selectedCategory`, `searchQuery`, and `selectedTags`.
   - **Timeline Sidebars**: Subscribes to `decisionLogs` API queries. Triggers real-time invalidation using `useSSE("post.updated", ...)` and `useSSE("comment.created", ...)`.
   - *Draft file created: `proposed_FeedPage.tsx`.*

2. **`PostDetailPage.tsx`**
   - **Markdown Renderer**: Implements safe HTML sanitization and rich-text parsing:
     ```typescript
     import ReactMarkdown from 'react-markdown';
     import DOMPurify from 'dompurify';
     const cleanHTML = DOMPurify.sanitize(rawMarkdown);
     <ReactMarkdown>{cleanHTML}</ReactMarkdown>
     ```
   - **Comments List**: Renders using `Virtuoso` to accommodate potentially long threads. Differentiates AI author comments with custom backgrounds (`#f5f7f6`), showing automated reply tags and willingness scores.
   - **AI Status Banner**: A dynamic sidebar stepper mapping the post's current `aiStatus` and `tasks` from the DB to visualize active rendering steps in real-time.
   - *Draft file created: `proposed_PostDetailPage.tsx`.*

3. **`AgentPlazaPage.tsx`**
   - **Interactive Config**: Cards featuring sliders to configure `replyThreshold` and `activityLevel`, checkboxes for reply rules (`allowAutoReply`, `allowMentionReply`, `allowFollowupReply`), and inputs for limit metrics.
   - **Active State Toggle**: Includes a dual-state switch connected to `updateAgent` mutations to easily enable/disable agent simulation participation in the database.
   - *Draft file created: `proposed_AgentPlazaPage.tsx`.*

4. **`CreatePostPage.tsx`**
   - **Preview System**: Multi-input binding allowing real-time card previews in the sidebar before hitting publish.
   - **Interactive Modes**: Selecting AI participation modes maps parameters to post payloads when dispatching `createPost` mutations.
   - *Draft file created: `proposed_CreatePostPage.tsx`.*

---

## 3. Style & Layout Guidelines (Cohere & Synthetica Design)

To ensure high-fidelity compliance with `design_cohere.md` and `DESIGN.md` rules:

- **Type Split (UI vs Technical Labels)**:
  - Default layout text and headings use **Hanken Grotesk** (mapping to Unica77) with tight tracking and line-heights ($1.0 \sim 1.2$) for displays.
  - Timestamps, tags, viewcounts, willingness scores, status chips, and code snippets use **JetBrains Mono** (`font-label-mono` / `jetbrainsMono`) to establish a precision tool look.
- **Color Blocks (Flat Tonal Layering)**:
  - The default canvas background is Soft Stone (`#fbf9f4`).
  - Container panels use Canvas White (`#ffffff`) or low-surface containers (`#f5f3ee` / `#eae8e3`) with thin rules (`#d9d9dd` hairline borders) rather than drop shadows.
  - Active buttons and banners use dark theme blocks (Primary `#17171c` or Brand Green `#003c33` backgrounds) to create focal weight.
- **Border Radii**:
  - Buttons and input controls use **Soft (4px - 8px)** corners.
  - Standard discussion cards and panels use **Rounded (16px)** corners.
  - Avatars and highlight hero media use **Signature (22px)** corners to soften the layout.
- **Responsive Collapsing**:
  - Grid structures transition from `grid-cols-12` (with an asymmetric $8:4$ layout) on desktop, to stacked single-columns on mobile.
  - Sidebars reflow below content, and spacing scales down from $80\text{px}$ (desktop) to $32\text{px}$ (mobile).

---

## 4. Concurrency Bug Resolution in `web/src/sse/simulator.ts`

### The Problem
During overlapping simulation lifecycles (e.g., when a user submits a post, or adds consecutive comments in quick succession), multiple separate invocations of `runBackgroundAISimulation` schedule independent `setTimeout` callbacks:
1. When no replies are scheduled, a $1000\text{ms}$ timeout transitions the post to `COMPLETED`.
2. When replies are scheduled, a staggered queue processes replies, and a final timeout `totalDuration = (replyQueue.length * 1000) + 4000` transitions the post to `COMPLETED`.

If simulation $B$ completes early (e.g., because it has no replies and finishes in $1000\text{ms}$), its timeout callback updates the post's `aiStatus` to `COMPLETED` and emits `"post.updated"`, even if simulation $A$'s tasks are still actively `PENDING` or `PROCESSING` in the database. This breaks UI indicators, showing a premature `COMPLETED` state while responses are still pending.

### The Solution
Before modifying the post state to `COMPLETED`, we check the database for any active tasks associated with the `postId` that are still in `PENDING` or `PROCESSING` state. The post's status is updated to `COMPLETED` *only* if no such active tasks exist.

### Implementation Patch
The following logic is integrated into both completion timeouts in `web/src/sse/simulator.ts`:

```typescript
const activeTasks = db.getTasks().filter(t => t.postId === postId && (t.status === "PENDING" || t.status === "PROCESSING"));
if (activeTasks.length === 0) {
  db.updatePost(postId, { aiStatus: "COMPLETED" });
  sseEmitter.emit("post.updated", db.getPost(postId));
}
```

*The complete patch is saved as `simulator_concurrency_fix.patch` in this directory.*

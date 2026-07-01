# Milestone 2 Analysis & Implementation Strategy: Web App Pages

## 1. Executive Summary
This document outlines the visual structure, routing layout, and technical implementation details for **Milestone 2: Web App Pages**. By analyzing the HTML prototypes in `stitch_ai_forum/` and the design specifications in `design_cohere.md` and `DESIGN.md`, we establish a clear blueprint for the pages, styling definitions, and mock backend integration, including a crucial concurrency bug fix in the AI response simulator.

---

## 2. HTML Prototype Analysis & Component Mapping

### A. Homepage / Post Feed (`stitch_ai_forum/ai_forum_5/code.html`)
* **Layout Structure**: 
  - A responsive 12-column grid layout (`max-w-7xl mx-auto px-margin-mobile md:px-margin-desktop py-xl grid grid-cols-1 lg:grid-cols-12 gap-lg`).
  - Left Content Area: 8 columns, housing the Hero section and Main Feed.
  - Right Sidebar: 4 columns, housing Hot Tags (`HotTags`), Active AI Agents roster (`ActiveAIRoles`), and Recent AI activity timeline (`RecentAIActivity`).
* **Visual Elements**:
  - Global Header: Stark logo, search input, notifications button, and user profile button.
  - Feed Tabs: "最新" (Latest), "最热" (Hottest), "待回复" (Unanswered), "AI 参与最多" (AI Most), separated by a hairline border (`border-b border-hairline`).
  - Post Card: `bg-surface-container-lowest border border-hairline rounded-[16px] p-lg flex flex-col gap-sm hover:border-secondary transition-colors`. Features a category badge at the top-right, title, content summary (`line-clamp-2`), author details, tag list, and AI responses indicator (overlapped avatars, response counts, status badges).
* **AI Indicators**:
  - **Processing**: `bg-secondary-container text-on-secondary-container px-xs py-[2px] rounded font-micro text-micro flex items-center gap-xs` with a `psychology` icon.
  - **Completed**: `bg-success-green text-secondary px-xs py-[2px] rounded font-micro text-micro flex items-center gap-xs` with a `check_circle` icon.

### B. Post Details (`stitch_ai_forum/_2/code.html`)
* **Layout Structure**:
  - 12-column grid (`grid-cols-1 md:grid-cols-12 gap-gutter`).
  - Left Content Area (8 columns): Detailed post container, AI Processing Status timeline, and Discuss section.
  - Right Sidebar (4 columns): "查看 AI 决策日志" button, Participating AI details list, Post tags, and Related discussions.
* **Visual Elements**:
  - Detailed Post Canvas: Category badge, relative timestamp, large title (`font-headline-xl text-ink`), author info, counts, and markdown-rendered content.
  - Discussion Area: Textarea input container with a markdown toolbar and `@AI` quick-mention button.
  - Comments Feed:
    - **Human Comment**: Bordered white card (`bg-surface-container-lowest border border-hairline`), avatar on the left, author name, timestamp, and body content.
    - **AI Comment**: Soft green background (`bg-[#f5f7f6] border border-[#e0e5e3] rounded-tr-[22px] rounded-br-[22px] rounded-bl-[22px] p-md`). Headers display the agent identity, trigger description (e.g., "发帖后自动回复"), and a footer containing the calculated willingness score ("意愿分: 92/100") and a "继续追问" (Follow-up query) button.
    - **Thread Connection**: Vertically centered dotted line connecting parent comments to replies.
  - Sidebar Components:
    - AI Processing Status: Displays progress stages (e.g., extracting concepts, calculating willingness, running reasoning engine) with icons (`check`, `hourglass_empty`, `edit_note`).
    - Participating AI list: Roster of active agents showing reply status (已回复, 待定, 忽略) and raw willingness score.

### C. AI Agents Plaza (`stitch_ai_forum/ai_ai_forum_2/code.html`)
* **Layout Structure**:
  - Layout matches the grid composition with a centered plaza header ("AI 角色广场" and description).
  - Filter Bar: Allows segmenting agents by domain/category (All, Technical, Life, Emotion), personality traits, or features.
  - Roster Grid: 3-column layout on desktop containing agent profile cards.
* **Agent Cards**:
  - Bordered card (`bg-surface-container-low rounded-[16px] p-lg border border-hairline`).
  - Avatar (`w-[44px] h-[44px] rounded-full`), verified badge, name, perspective label, description, and list of characteristic tags (e.g. "客观冷静", "实用主义").
  - System prompts, speaking style description, and activation toggles to control automated execution.

### D. Create Post (`stitch_ai_forum/ai_forum_4/code.html`)
* **Layout Structure**:
  - Left Column (8 columns): Title, Category selector, Tags comma-separated input, and Content editor.
  - Right Column (4 columns): Real-time card preview and interactive workflow visualizer.
* **AI Participation Mode Selection**:
  - Implements radio cards for selectable options:
    1. **仅真人** (Humans only, no AI analysis).
    2. **少量 AI** (Low AI, only high-affinity/explicitly mentioned agents).
    3. **标准 AI** (Standard AI interaction).
    4. **热闹模式** (Busy AI, boosted willingness triggers).
  - Highlights options with borders and distinct background colors (`bg-secondary-container border-secondary` or coral overlays) on select.

---

## 3. Page Components Structure & Routing Layout

### A. Routing Configuration (`web/src/App.tsx`)
We consolidate the routing hierarchy using React Router DOM. `HomePage` and `PostsListPage` are refactored into a unified, feature-complete `FeedPage` that includes post lists, filters, and composer toggle states.

Proposed routing table:
* `/` -> `FeedPage.tsx` (Homepage & main feed list, defaults to latest)
* `/posts/:id` -> `PostDetailPage.tsx` (Detailed view, comments, and decision logs)
* `/agents` -> `AgentPlazaPage.tsx` (AI Agents Plaza, including prompt configurations & toggles)
* `*` -> `NotFoundPage.tsx`

```tsx
// App.tsx Router configuration outline
<Routes>
  <Route element={<AppLayout />}>
    <Route path="/" element={<FeedPage />} />
    <Route path="/posts/:id" element={<PostDetailPage />} />
    <Route path="/agents" element={<AgentPlazaPage />} />
    <Route path="*" element={<NotFoundPage />} />
  </Route>
</Routes>
```

### B. Pages Components Structure

#### 1. `FeedPage.tsx` (replaces `HomePage.tsx` and `PostsListPage.tsx`)
* **Path**: `web/src/pages/FeedPage.tsx`
* **Features**:
  - Includes a modern Hero header referencing the forum research lab.
  - Unified grid layout: Main content feed (8-cols), sidebar utilities (4-cols).
  - **Feed Lists**: Powered by `React Virtuoso` for smooth performance during infinite/large scrolls. Reads filtered items based on selected tabs (`latest`, `hottest`, `unanswered`, `ai_most`).
  - **Filters**: Inline Category selection tags (e.g. "技术探讨", "前端开发") and Tag parameters parsed from search queries.
  - **Creation Entry**: Incorporates a prominent "发布新话题" toggle button that displays the detailed creation composer form inside an expandable card or modal, matching the style of the prototype (incorporating title, category, tags, markdown text, and AI Participation Mode).
  - **Sidebar Widgets**:
    - `HotTags`: Hot tags index.
    - `ActiveAIRoles`: Visual list of active AI profiles.
    - `RecentAIActivity`: Timeline of latest comments/analyses.

#### 2. `PostDetailPage.tsx`
* **Path**: `web/src/pages/PostDetailPage.tsx`
* **Features**:
  - Detailed article panel featuring metadata (author, role, views, comments counts).
  - **AI Status Banner**: A top-anchored banner showing whether AI processing is `PENDING`, `PROCESSING`, or `COMPLETED`.
  - **Markdown Viewer**: Employs `SafeMarkdown` (built on `react-markdown` + `dompurify`) to safely parse text content and format technical code blocks.
  - **User/AI Comments Feed**: Powered by `React Virtuoso` for list visualization. Alternates between `HumanComment.tsx` and `AIComment.tsx` components.
  - **Sidebar Modules**:
    - `AIProcessingStatus` (Step visualizer mapping tag extraction, willingness evaluations, execution, and text rendering).
    - `ParticipatingAI` (AI response roster with raw willingness scores and statuses).
    - `PostTags` (Tag pill container).
    - `RelatedDiscussions` (List of related research topics).

#### 3. `AgentPlazaPage.tsx` (replaces `AIAgentsPage.tsx`)
* **Path**: `web/src/pages/AgentPlazaPage.tsx`
* **Features**:
  - Introduces a polished filter header matching `ai_ai_forum_2/code.html` to screen profiles by domains, traits, or prompt features.
  - Renders a grid of extended `AIAgentCard` components.
  - **Card Structure**:
    - Agent description, value orientations, specialties.
    - Explains detailed personality prompts (`systemPrompt`) and speaking style formatting parameters (`stylePrompt`).
    - Reply Willingness Threshold: Illustrated using a customized slider/progress bar displaying the threshold (e.g. `0.60`).
    - Toggle Switch: Inline enable/disable toggle input that hooks into the `updateAgent` mutation to toggle state dynamically in the local mock database.

---

## 4. CSS Style and Responsive Layout Alignment
The styling strategy must use Tailwind CSS utilities mapped on top of CSS variables as defined in `web/src/styles/index.css` and `web/tailwind.config.js`:

### Color Mappings
- **Text & Body**: Use `text-cohere-on-surface` (`--c-on-surface`, `#1b1c19`) for baseline readability. Primary titles use `text-cohere-primary` (`--c-primary`, `#000000`).
- **Accent Lines**: Use `border-cohere-hairline` (`--c-hairline`, `#d9d9dd`) for rule-based borders and dividers.
- **Categorization**: Outlines on category chips use `border-cohere-coral` (`#ff7759`) and matching text classes.
- **AI Processing**: Use `bg-cohere-secondary-container` (`#b8ede0`) with `text-cohere-on-secondary-container` (`#3b6d63`) for warning or pending indicators. Completed states utilize `bg-cohere-success` (`#edfce9`) with `text-cohere-secondary` (`#35675d`).

### Borders and Elevation
- Avoid drop-shadow classes. Visual hierarchies must rely entirely on flat layering, outlines, and surface alternation (e.g., placing white cards `bg-cohere-surface-lowest` inside a slightly darker canvas background `bg-cohere-background`).
- **Corner Radii Rules**:
  - Buttons and inputs: `rounded-sm` (8px) or `rounded-xs` (4px).
  - Main Cards: `rounded-lg` (16px) for post cards and agent cards.
  - Avatars & Special media containers: `rounded-ai` (22px).
  - Action pills: `rounded-pill` (32px).

### Responsive Layout Reflow
- **Desktop (1024px+)**: Unified 12-column grid. Full sidebar layouts are placed side-by-side with 8-column main content streams. Section padding utilizes `py-xl md:py-xxl` to keep generous editorial spacing.
- **Tablet (768px-1024px)**: Grid shifts where sidebars either wrap below the main feed or adapt as stacked containers, keeping layout gutters fluid.
- **Mobile (<768px)**: Stacks elements in 1-column layout. Section vertical spacing reduces to `py-md` or `py-lg`. Standard margins default to `px-margin-mobile` (16px).

---

## 5. Concurrency Bug Fix in Response Simulator

### A. Problem Diagnosis
In `web/src/sse/simulator.ts`, multiple client simulations can trigger concurrently (e.g., if a user quickly posts a new topic and writes comments, or inputs multiple messages in sequence). 

The simulator uses asynchronous staggered timeouts (`setTimeout`) to simulate delayed agent executions. When a simulation ends, it unconditionally transitions the post’s status to `COMPLETED`:
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
If a second simulation starts while the first is running, the first simulation's timeout will trigger and set the post to `COMPLETED` even though the second simulation’s tasks are still actively `PENDING` or `PROCESSING` in the background.

### B. Resolution Strategy
Before updating `aiStatus` to `COMPLETED`, the simulator must query the database to verify if there are any *other* tasks matching the given `postId` that remain in an active state (`PENDING` or `PROCESSING`). If active tasks exist, the status transition is skipped, allowing the final active simulation's timeout to perform the final transition to `COMPLETED`.

### C. Code Modifications (Proposed Diff)
Below are the code snippets mapping the fix in `web/src/sse/simulator.ts`:

#### Snippet 1 (Post Auto Reply Empty Flow)
**Before**:
```typescript
  if (replyQueue.length === 0) {
    // If no agent replies, transition post state to COMPLETED after a brief period
    setTimeout(() => {
      const latestPost = db.getPost(postId);
      if (latestPost) {
        db.updatePost(postId, { aiStatus: "COMPLETED" });
        sseEmitter.emit("post.updated", db.getPost(postId));
      }
    }, 1000);
    return;
  }
```

**After**:
```typescript
  if (replyQueue.length === 0) {
    // If no agent replies, transition post state to COMPLETED after a brief period
    // provided there are no other active simulation tasks running.
    setTimeout(() => {
      const hasActiveTasks = db.getTasks().some(
        (t) => t.postId === postId && (t.status === "PENDING" || t.status === "PROCESSING")
      );
      if (!hasActiveTasks) {
        const latestPost = db.getPost(postId);
        if (latestPost) {
          db.updatePost(postId, { aiStatus: "COMPLETED" });
          sseEmitter.emit("post.updated", db.getPost(postId));
        }
      }
    }, 1000);
    return;
  }
```

#### Snippet 2 (Scheduled Tasks Flow)
**Before**:
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

**After**:
```typescript
  // Schedule final COMPLETED status transition, verifying no active tasks remain
  const totalDuration = (replyQueue.length * 1000) + 4000;
  setTimeout(() => {
    const hasActiveTasks = db.getTasks().some(
      (t) => t.postId === postId && (t.status === "PENDING" || t.status === "PROCESSING")
    );
    if (!hasActiveTasks) {
      const latestPost = db.getPost(postId);
      if (latestPost) {
        db.updatePost(postId, { aiStatus: "COMPLETED" });
        sseEmitter.emit("post.updated", db.getPost(postId));
      }
    }
  }, totalDuration);
```

---

## 6. Implementation Checklist & Verification Method
The implementer agent must execute and verify the changes through the following checklist:

1. **Verify Existing Setup**:
   - Confirm code builds via `npm run build` or `npm run dev` within the `web/` directory.
2. **Implement File Changes**:
   - Refactor `web/src/pages/HomePage.tsx` and `web/src/pages/PostsListPage.tsx` into a singular, clean `FeedPage.tsx`. Ensure `React Virtuoso` is imported and used for post lists, and incorporate the composer with AI participation mode choices.
   - Refactor `web/src/pages/AIAgentsPage.tsx` into `AgentPlazaPage.tsx`, adding detail prompts displays, slider threshold visualizations, and local active state mutations.
   - Refactor `web/src/App.tsx` routing paths to match the updated files.
   - Inject the concurrency guard into `web/src/sse/simulator.ts`.
3. **Verify Refactor Performance**:
   - Check page routing transitions in the browser client.
   - Validate markdown content rendering inside `PostDetailPage.tsx` using `SafeMarkdown`.
   - Test simultaneous/consecutive comment publications on the details page to trigger concurrent simulator run loops. Verify that `aiStatus` remains `PROCESSING` until all staggered tasks complete, transitioning to `COMPLETED` only when the final task is marked as done.

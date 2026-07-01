# Original User Request

## Initial Request — 2026-06-30T05:07:42Z

Implement and integrate the user web application (web/) and administrative panel (admin/) for the AI Forum, based on the HTML prototypes in stitch_ai_forum/ and adhering to the custom design system (DESIGN.md/design_cohere.md).

Working directory: /Users/mac/Documents/ai_forum
Integrity mode: development

## Requirements

### R1. User Web Application (web/)
- **Tech Stack**: React + TypeScript + Vite + React Router + TanStack Query + Zustand.
- **Styling**: Tailwind CSS combined with custom CSS Variables to faithfully reproduce the Cohere/Synthetica design system guidelines.
- **Pages**: 
  - Homepage / Post Feed: Includes post feed, post creation entry, active AI agents list, filter controls (latest, hot, pending reply, AI participation).
  - Post Details: Displays post content, tag list, AI processing state, rich-text markdown viewer with DOMPurify sanitization, user comments, and AI-generated comments.
  - AI Agents Page: Lists AI agents with avatars, capability tags, personality traits, and active status.
- **Data & APIs**:
  - Implement a complete, rich Frontend Mock Data layer; do NOT connect to a real backend.
  - Mock API clients and TanStack query queries must reside inside `src/api` and follow clear typed interface shapes.
  - Server-Sent Events (SSE) mock logic should be isolated in `src/sse` or dedicated hooks (e.g., simulating real-time AI updates/status changes).
  - Server state via TanStack Query; UI state via Zustand.
  - Render lists using `react-virtuoso` to optimize rendering of long post streams and comment sections.

### R2. Admin Console (admin/)
- **Tech Stack**: React + TypeScript + Vite + Refine + Ant Design.
- **Styling**: Customize Ant Design's theme system using the configuration provider (Theme Config) to implement custom Cohere brand colors, typography spacing, and specific corner rounding, while keeping a highly dense, operator-focused control panel aesthetic.
- **Pages**:
  - Dashboard: Metrics cards (posts, comments, AI responses, tasks), 7-day post trend chart, task status distribution, system service statuses, recent tasks list, and AI decision logs summary.
  - AI Agent Management: Tables listing agents, their parameters (temperature, threshold, status), and a drawer form for editing agent configurations.
  - AI Task Queue: Table showing tasks, status filters, drawer with task payload (input/output), execution timeline, and retry buttons.
  - AI Decision Logs: Timeline or detailed view of agent evaluation matrices (answer willingness, thresholds, rationale).
- **Architecture**:
  - Keep all data fetching inside Refine's `dataProvider` or `admin/src/api`.
  - Frontend controls visibility, but does not hardcode auth; backend is authoritative (mocked auth roles/permissions).

### R3. Design System Alignment
- Implement the exact design details and color palettes from `stitch_ai_forum/synthetica_ai_forum/DESIGN.md` and `stitch_ai_forum/design_cohere.md`.
- **Colors**: Near-black `#17171c` / `#000000`, Deep Green `#003c33`, Action Blue `#1863dc`, Coral `#ff7759`, soft backgrounds (`#eeece7`, `#fbf9f4`, `#ffffff`), and thin border `#d9d9dd`.
- **Typography**: Hanken Grotesk/system sans for headings; JetBrains Mono for system labels, timestamps, and status badges; Chinese text minimum 16px.
- **Layout**: Dynamic reflow for mobile viewports, no overflows or overlapping text.

## Acceptance Criteria

### Build and Type Verification
- [ ] Both `web/` and `admin/` successfully compile with TypeScript (`tsc`) and build production assets (`npm run build`) without errors.
- [ ] Custom script configurations exist to easily run dev, build, and lint checks.

### Functional Completeness
- [ ] User web app correctly displays the Post Feed, allows filtering, lets users open a post, displays sanitized markdown comments, and lists active AI characters. All state uses TanStack Query + Zustand based on mock layer.
- [ ] Admin panel includes a functioning Dashboard with charts, an AI Agent config drawer, a Task Queue drawer showing inputs/outputs, and an AI Decision logs page.

### Styling & Responsiveness
- [ ] Layout conforms to the Cohere/Synthetica design guidelines (thin borders, roundings, correct color scheme, typography split).
- [ ] Pages are fully responsive down to 375px mobile resolution.

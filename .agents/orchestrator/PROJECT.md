# Project: AI Forum Frontend Implementation & Integration

## Architecture
The AI Forum Frontend consists of two primary applications:
1. **User Web Application (`web/`)**: A client-side React SPA built with Vite, TypeScript, React Router, Zustand, and TanStack Query. It faithfully reproduces the Cohere/Synthetica design system using custom CSS variables and Tailwind. It uses `react-virtuoso` for virtualized list rendering, `react-markdown` + `dompurify` for markdown post/comment rendering, and simulated Server-Sent Events (SSE) for real-time AI updates.
2. **Admin Console (`admin/`)**: An operator control panel built with React, TypeScript, Vite, Refine, and Ant Design. It customizes Ant Design's theme system using the configuration provider (Theme Config) to adhere to the Cohere brand palette.

Both applications run entirely in-browser, powered by a comprehensive **Frontend Mock Data Layer** representing the schemas defined in the database requirements (posts, comments, agents, tasks, decision logs). No real backend is connected.

```
+-------------------------------------------------------------+
|                       AI Forum                              |
+------------------------------+------------------------------+
|     User Web App (web/)      |     Admin Console (admin/)   |
|   - Feed, Post Details, AI   |   - Dashboard, Agent Config, |
|     Agent Plaza              |     Tasks, Decision Logs     |
|   - Tailwind + Custom CSS    |   - Refine + Ant Design      |
|   - Zustand + Query          |   - Theme Configuration      |
+------------------------------+------------------------------+
|                    Frontend Mock Data Layer                 |
|            (Typed mock API client and SSE hooks)            |
+-------------------------------------------------------------+
```

## Milestones

| # | Name | Scope | Dependencies | Status |
|---|------|-------|--------------|--------|
| 1 | E2E Test Suite | Create opaque-box E2E test infra and Tiers 1-4 test cases; publish `TEST_READY.md`. | None | IN_PROGRESS |
| 2 | Web App Init | Initialize `web/` workspace with Vite, TS, Tailwind, custom design CSS variables, and typed Mock Data Layer (with simulated SSE). | None | IN_PROGRESS |
| 3 | Web App Pages | Implement Web App pages: Homepage/Feed, Post Details (Markdown + DOMPurify + Virtuoso), AI Agent Plaza. | M2 | PLANNED |
| 4 | Admin App Init | Initialize `admin/` workspace with Vite, TS, Refine, and customized Ant Design Theme Config. | None | IN_PROGRESS |
| 5 | Admin App Pages | Implement Admin pages: Dashboard, AI Agent Management, AI Task Queue, AI Decision Logs. | M4 | PLANNED |
| 6 | Integration & Verification | Run all E2E tests, fix failures, perform Tier 5 adversarial hardening, and execute Forensic Audit. | M1, M3, M5 | PLANNED |

## Code Layout
- `web/`:
  - `web/package.json`, `web/vite.config.ts`, `web/tailwind.config.js`
  - `web/src/main.tsx` - App entry point
  - `web/src/App.tsx` - App routes & layout
  - `web/src/api/` - Typed Mock API services and mock database store
  - `web/src/sse/` - Mock SSE hooks / provider
  - `web/src/stores/` - Zustand global stores for UI states
  - `web/src/components/` - Virtualized lists, post card, comment sections
  - `web/src/pages/` - FeedPage, PostDetailPage, AgentPlazaPage
  - `web/src/styles/` - Global Tailwind directive styles with Cohere variables
- `admin/`:
  - `admin/package.json`, `admin/vite.config.ts`, `admin/src/main.tsx`
  - `admin/src/App.tsx` - Refine configuration, custom AntD ThemeConfig, and routing
  - `admin/src/dataProvider/` - Custom Refine data provider mapping to mock database
  - `admin/src/pages/` - DashboardPage, AgentManagementPage, TaskQueuePage, DecisionLogsPage

## Interface Contracts

### Mock Database Schema Shapes

#### Post
```typescript
interface Post {
  id: number;
  title: string;
  content: string;
  category: string; // e.g. "后端开发", "前端开发"
  tags: string[]; // e.g. ["Go", "微服务", "React"]
  author: {
    username: string;
    avatar: string;
  };
  aiStatus: "PENDING" | "PROCESSING" | "COMPLETED";
  aiResponsesCount: number;
  aiAvatars: string[]; // Avatars of AI agents who participated
  createdAt: string;
}
```

#### Comment
```typescript
interface Comment {
  id: number;
  postId: number;
  parentId: number | null; // 0 or null for top-level comments
  content: string;
  author: {
    username: string;
    avatar: string;
    isAi: boolean;
    aiAgentId?: number;
  };
  createdAt: string;
}
```

#### AIAgent
```typescript
interface AIAgent {
  id: number;
  name: string;
  avatar: string;
  description: string;
  ageViewpoint: string;
  personality: string;
  valueOrientation: string;
  speakingStyle: string;
  systemPrompt: string;
  stylePrompt: string;
  replyThreshold: number; // e.g. 0.60
  activityLevel: number; // e.g. 0.50
  allowAutoReply: boolean;
  allowMentionReply: boolean;
  allowFollowupReply: boolean;
  maxAutoRepliesPerPost: number;
  maxFollowupRepliesPerPost: number;
  isFallback: boolean;
  active: boolean;
}
```

#### AIReplyTask
```typescript
interface AIReplyTask {
  id: number;
  postId: number;
  parentCommentId: number | null;
  targetCommentId: number | null;
  aiAgentId: number;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  status: "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED";
  prompt: string;
  result: string;
  errorMessage: string;
  retryCount: number;
  createdAt: string;
  startedAt: string | null;
  finishedAt: string | null;
}
```

#### AIDecisionLog
```typescript
interface AIDecisionLog {
  id: number;
  postId: number;
  commentId: number | null;
  aiAgentId: number;
  aiAgentName: string;
  triggerType: string;
  willingnessScore: number;
  thresholdValue: number;
  decision: "REPLY" | "IGNORE" | "FAILED";
  reason: string;
  createdAt: string;
}
```

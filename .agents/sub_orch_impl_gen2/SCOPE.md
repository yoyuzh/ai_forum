# Scope: AI Forum Frontend Implementation & Integration

## Architecture
The AI Forum Frontend consists of two primary applications:
1. **User Web Application (web/)**: React, TypeScript, Vite, Tailwind CSS with custom Cohere/Synthetica variables, Zustand for local UI state, TanStack Query for server state. Long lists of posts/comments are rendered with `react-virtuoso`. Post/comment markdown content is rendered with `react-markdown` and sanitized with `dompurify`.
2. **Admin Console (admin/)**: React, TypeScript, Vite, Refine, Ant Design. Customize the AntD theme via `ConfigProvider` to reflect Cohere brand colors, typography, and corner roundings. Dense, operator-focused layout.
3. **Frontend Mock Data Layer**: A shared in-browser mock database (in-memory or localStorage) containing posts, comments, AI agents, AI reply tasks, and decision logs. Shared mock API client in `web/src/api` and `admin/src/api`. Mock SSE hook in `web/src/sse` to simulate real-time AI reply triggers and status changes.

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|---|---|---|---|
| 1 | Web App Init & Mock Layer | Initialize `web/` Vite project with TS, Tailwind CSS, custom design variables. Implement the complete Frontend Mock Data Layer (typed API clients, mock database, SSE hooks simulating real-time updates). | None | DONE |
| 2 | Web App Pages | Implement FeedPage (feed, filters, creation entry, agents list), PostDetailPage (sanitized markdown viewer, comments stream via Virtuoso), and AgentPlazaPage (agent list with capability tags, personalities). | M1 | IN_PROGRESS |
| 3 | Admin Console Init & Config | Initialize `admin/` Vite project with TS, Refine, and Ant Design. Configure Refine dataProvider pointing to the shared mock database. Customize Ant Design Theme Config (colors, rounded corners, density). | None | PLANNED |
| 4 | Admin Console Pages | Implement DashboardPage (metrics, charts, recent tasks, decision logs summary), AgentManagementPage (agent table + parameter edit drawer), TaskQueuePage (task table, payload drawer, retry button), and DecisionLogsPage (evaluation matrices timeline). | M3 | PLANNED |
| 5 | Integration & E2E Testing | Verify that both applications run, check integrations, wait for `TEST_READY.md`, run the E2E test suite, and debug/resolve all issues until 100% of tests pass. | M2, M4 | PLANNED |
| 6 | Adversarial Hardening & Audit | Spawn Challengers to perform Tier 5 white-box adversarial coverage hardening, then run Forensic Auditor to verify clean verdict. | M5 | PLANNED |

## Interface Contracts

### Mock Database Schema Shapes
The mock database represents the unified source of truth for both applications, running client-side.

#### Post
- `id: number`
- `title: string`
- `content: string`
- `category: string`
- `tags: string[]`
- `author: { username: string; avatar: string }`
- `aiStatus: "PENDING" | "PROCESSING" | "COMPLETED"`
- `aiResponsesCount: number`
- `aiAvatars: string[]`
- `createdAt: string`

#### Comment
- `id: number`
- `postId: number`
- `parentId: number | null`
- `content: string`
- `author: { username: string; avatar: string; isAi: boolean; aiAgentId?: number }`
- `createdAt: string`

#### AIAgent
- `id: number`
- `name: string`
- `avatar: string`
- `description: string`
- `ageViewpoint: string`
- `personality: string`
- `valueOrientation: string`
- `speakingStyle: string`
- `systemPrompt: string`
- `stylePrompt: string`
- `replyThreshold: number`
- `activityLevel: number`
- `allowAutoReply: boolean`
- `allowMentionReply: boolean`
- `allowFollowupReply: boolean`
- `maxAutoRepliesPerPost: number`
- `maxFollowupRepliesPerPost: number`
- `isFallback: boolean`
- `active: boolean`

#### AIReplyTask
- `id: number`
- `postId: number`
- `parentCommentId: number | null`
- `targetCommentId: number | null`
- `aiAgentId: number`
- `triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP"`
- `status: "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED"`
- `prompt: string`
- `result: string`
- `errorMessage: string`
- `retryCount: number`
- `createdAt: string`
- `startedAt: string | null`
- `finishedAt: string | null`

#### AIDecisionLog
- `id: number`
- `postId: number`
- `commentId: number | null`
- `aiAgentId: number`
- `aiAgentName: string`
- `triggerType: string`
- `willingnessScore: number`
- `thresholdValue: number`
- `decision: "REPLY" | "IGNORE" | "FAILED"`
- `reason: string`
- `createdAt: string`

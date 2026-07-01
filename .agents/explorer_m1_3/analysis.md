# Milestone 1: Web App Init & Mock Layer Analysis and Strategy Report

This report outlines the design and implementation strategy for the initialization and mock data layer of the AI Forum frontend (`web/`). It is structured to guide the implementer to create a fully typed, persistent mock database, simulated real-time SSE stream, query cache integration, and layout style framework based on Cohere and Synthetica design system guidelines.

---

## 1. Workspace Initialization (`web/`)

We configure the React SPA using Vite, TypeScript, and Tailwind CSS. The workspace is structured to align with the rest of the project and allow compilation via `npm run build` and TypeScript checks via `tsc`.

### 1.1 `web/package.json`
This file defines the dependencies required for both UI rendering (`react-virtuoso`, `react-markdown`, `dompurify`), global and server state management (`zustand`, `@tanstack/react-query`), routing (`react-router-dom`), icons (`lucide-react`), and build/dev scripts.

```json
{
  "name": "ai-forum-web",
  "private": true,
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "preview": "vite preview"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.45.0",
    "dompurify": "^3.1.5",
    "lucide-react": "^0.395.0",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-markdown": "^9.0.1",
    "react-router-dom": "^6.23.1",
    "react-virtuoso": "^4.7.11",
    "zustand": "^4.5.2"
  },
  "devDependencies": {
    "@types/dompurify": "^3.0.5",
    "@types/node": "^20.12.12",
    "@types/react": "^18.3.3",
    "@types/react-dom": "^18.3.0",
    "@typescript-eslint/eslint-plugin": "^7.9.0",
    "@typescript-eslint/parser": "^7.9.0",
    "@vitejs/plugin-react": "^4.3.0",
    "autoprefixer": "^10.4.19",
    "eslint": "^8.57.0",
    "eslint-plugin-react-hooks": "^4.6.2",
    "eslint-plugin-react-refresh": "^0.4.7",
    "postcss": "^8.4.38",
    "tailwindcss": "^3.4.4",
    "typescript": "^5.4.5",
    "vite": "^5.2.11"
  }
}
```

### 1.2 `web/vite.config.ts`
Vite is configured to use `@vitejs/plugin-react` and resolve the absolute path alias `@/` pointing to `web/src/`. The server is configured to run on port `3000` to avoid conflict with the admin console (which will run on port `3001`).

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    host: true,
  },
});
```

### 1.3 `web/tailwind.config.js`
The Tailwind configuration exposes design variables (colors, border-radius, fonts, letter-spacings) mapped to CSS custom variables. This ensures styling consistency and supports custom theme overrides if needed.

```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: 'rgb(var(--primary) / <alpha-value>)',
        'cohere-black': 'rgb(var(--cohere-black) / <alpha-value>)',
        ink: 'rgb(var(--ink) / <alpha-value>)',
        'deep-green': 'rgb(var(--deep-green) / <alpha-value>)',
        'dark-navy': 'rgb(var(--dark-navy) / <alpha-value>)',
        canvas: 'rgb(var(--canvas) / <alpha-value>)',
        'soft-stone': 'rgb(var(--soft-stone) / <alpha-value>)',
        'pale-green': 'rgb(var(--pale-green) / <alpha-value>)',
        'pale-blue': 'rgb(var(--pale-blue) / <alpha-value>)',
        hairline: 'rgb(var(--hairline) / <alpha-value>)',
        'border-light': 'rgb(var(--border-light) / <alpha-value>)',
        'card-border': 'rgb(var(--card-border) / <alpha-value>)',
        muted: 'rgb(var(--muted) / <alpha-value>)',
        slate: 'rgb(var(--slate) / <alpha-value>)',
        'body-muted': 'rgb(var(--body-muted) / <alpha-value>)',
        'action-blue': 'rgb(var(--action-blue) / <alpha-value>)',
        'focus-blue': 'rgb(var(--focus-blue) / <alpha-value>)',
        coral: 'rgb(var(--coral) / <alpha-value>)',
        'coral-soft': 'rgb(var(--coral-soft) / <alpha-value>)',
        'form-focus': 'rgb(var(--form-focus) / <alpha-value>)',
        error: 'rgb(var(--error) / <alpha-value>)',
        'success-green': 'rgb(var(--success-green) / <alpha-value>)',
        'surface-dim': 'rgb(var(--surface-dim) / <alpha-value>)',
        'surface-container-low': 'rgb(var(--surface-container-low) / <alpha-value>)',
        'surface-container': 'rgb(var(--surface-container) / <alpha-value>)',
        'surface-container-high': 'rgb(var(--surface-container-high) / <alpha-value>)',
        'surface-container-highest': 'rgb(var(--surface-container-highest) / <alpha-value>)',
      },
      borderRadius: {
        'xs': '4px', // Utility elements, inputs
        'sm': '8px', // Chips, dialogs, smaller cards
        'md': '16px', // Standard cards (posts/agent profiles)
        'lg': '22px', // Signature media containers
        'xl': '30px', // Filter pills
        'pill': '32px', // Primary CTAs
      },
      fontFamily: {
        // Dual-font strategy mapping to Fallbacks
        display: ['CohereText', 'Space Grotesk', 'Inter', 'ui-sans-serif', 'system-ui'],
        body: ['Unica77 Cohere Web', 'Inter', 'Arial', 'ui-sans-serif', 'system-ui'],
        mono: ['CohereMono', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'monospace'],
        technical: ['jetbrainsMono', 'ui-monospace', 'monospace'],
        hanken: ['hankenGrotesk', 'Inter', 'sans-serif'],
      },
      letterSpacing: {
        'tight-display': '-0.02em',
        'tighter-display': '-0.03em',
      },
      lineHeight: {
        'tight-display': '1.0',
        'tighter-display': '1.2',
      }
    },
  },
  plugins: [],
}
```

### 1.4 `web/src/styles/index.css`
Defines CSS Custom Properties for RGB channels. This enables Tailwind's alpha transparency values (e.g. `bg-primary/10`).

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --primary: 23 23 28;             /* #17171c */
    --cohere-black: 0 0 0;           /* #000000 */
    --ink: 33 33 33;                 /* #212121 */
    --deep-green: 0 60 51;           /* #003c33 */
    --dark-navy: 7 24 41;            /* #071829 */
    --canvas: 255 255 255;           /* #ffffff */
    --soft-stone: 238 236 231;       /* #eeece7 */
    --pale-green: 237 252 233;       /* #edfce9 */
    --pale-blue: 241 245 255;        /* #f1f5ff */
    --hairline: 217 217 221;         /* #d9d9dd */
    --border-light: 229 231 235;     /* #e5e7eb */
    --card-border: 242 242 242;      /* #f2f2f2 */
    --muted: 147 147 159;            /* #93939f */
    --slate: 117 117 138;            /* #75758a */
    --body-muted: 97 97 97;          /* #616161 */
    --action-blue: 24 99 220;        /* #1863dc */
    --focus-blue: 76 110 230;        /* #4c6ee6 */
    --coral: 255 119 89;             /* #ff7759 */
    --coral-soft: 255 173 155;       /* #ffad9b */
    --form-focus: 155 96 170;        /* #9b60aa */
    --error: 179 0 0;                /* #b30000 */
    --success-green: 237 252 233;    /* #edfce9 */
    --surface-dim: 220 218 213;      /* #dcdad5 */
    --surface-container-low: 245 243 238;   /* #f5f3ee */
    --surface-container: 240 238 233;        /* #f0eee9 */
    --surface-container-high: 234 232 227;   /* #eae8e3 */
    --surface-container-highest: 228 226 221;/* #e4e2dd */
  }

  body {
    @apply bg-canvas text-ink font-body antialiased;
  }

  /* Cohere Custom Headline Typography styling classes */
  .display-hero {
    @apply font-display text-[96px] leading-[1.0] tracking-[-0.02em] font-normal;
  }
  
  .display-product {
    @apply font-display text-[72px] leading-[1.0] tracking-[-0.02em] font-normal;
  }

  .display-section {
    @apply font-body text-[60px] leading-[1.0] tracking-[-0.02em] font-normal;
  }
}
```

---

## 2. Frontend Mock Data Layer

To achieve seamless operability in-browser and enable shared state between `web/` and `admin/` during local testing, a fully simulated client-side database wrapper is proposed.

### 2.1 Interface Schema Definitions (`web/src/api/types.ts`)
This file defines types for all core entities, mirroring the database structure.

```typescript
export interface Post {
  id: number;
  title: string;
  content: string;
  category: string; // e.g., "后端开发" | "前端开发" | "人工智能"
  tags: string[];
  author: {
    username: string;
    avatar: string;
  };
  aiStatus: "PENDING" | "PROCESSING" | "COMPLETED";
  aiResponsesCount: number;
  aiAvatars: string[]; // Avatars of participating agents
  createdAt: string;
}

export interface Comment {
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

export interface AIAgent {
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

export interface AIReplyTask {
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

export interface AIDecisionLog {
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

### 2.2 Seed Data
The database is pre-seeded with highly descriptive records representing practical technical queries, diverse AI agents, simulated history, and evaluation decisions.

```typescript
export const SEED_AGENTS: AIAgent[] = [
  {
    id: 1,
    name: "ArchitectCommand",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=architect",
    description: "专注系统架构的高并发 Go 语言专家，逻辑严密，追求极限极致优化。",
    ageViewpoint: "偏好微服务集群、领域驱动设计(DDD)与事件驱动，反对冗余代码与盲目引入依赖。",
    personality: "严谨、高冷、直截了当，常用性能基准测试数据说服他人。",
    valueOrientation: "高内聚低耦合，稳定性第一，架构演进需以量化指标为导向。",
    speakingStyle: "简洁紧凑，大量使用代码片段、架构图表或基准指标。",
    systemPrompt: "你是一个资深的 Go 语言高并发系统架构专家。分析问题时，必须评估网络I/O、内存分配、锁竞争以及通道(Channel)设计。推荐高内聚低耦合的 Go monolithic / microservice 最佳实践。",
    stylePrompt: "回答要专业，减少情绪词汇，常用‘基于...的理由，建议...’。附带带行号的 Go 优化代码。",
    replyThreshold: 0.65,
    activityLevel: 0.80,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 2,
    maxFollowupRepliesPerPost: 3,
    isFallback: false,
    active: true
  },
  {
    id: 2,
    name: "CynicDeveloper",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=cynic",
    description: "饱经风霜的资深前端程序员，以嘲讽业界各种繁杂浮躁的技术栈为乐，偏爱精简方案。",
    ageViewpoint: "对过度炒作的AI、花哨的网页动效和每周都在换代的前端框架持批判怀疑态度。",
    personality: "毒舌、冷幽默、喜欢泼冷水。虽然态度悲观，但给出的底层技术原理剖析非常准确。",
    valueOrientation: "大道至简，实用主义。宁愿用原生 DOM 和纯 CSS 也别乱堆臃肿框架。",
    speakingStyle: "语气慵懒，常带有反问，比如‘难道我们真的需要为这个功能写 50MB 的依赖吗？’。",
    systemPrompt: "你是一个经历过无数次前端技术更迭的毒舌程序员。指出新技术中的冗余设计，用尖酸刻薄但理性的语气剖析系统中的问题。",
    stylePrompt: "多用反问句，穿插冷嘲热讽，但最后要落脚于简化技术栈、降低复杂度。",
    replyThreshold: 0.50,
    activityLevel: 0.60,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: false,
    maxAutoRepliesPerPost: 1,
    maxFollowupRepliesPerPost: 1,
    isFallback: false,
    active: true
  },
  {
    id: 3,
    name: "MildModerator",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=moderator",
    description: "温和的技术协调者，擅长在冲突各方中寻求平衡，鼓励建设性的技术讨论。",
    ageViewpoint: "任何技术选择都是折中(Trade-off)，没有绝对的对与错，要看具体业务背景。",
    personality: "温柔包容、鼓励导向、富有同理心，维护讨论秩序。",
    valueOrientation: "和谐沟通，协作创值。保护初学者的积极性，引导资深研发良性交流。",
    speakingStyle: "和蔼可亲，句式多以‘我很理解...’、‘我们可以尝试从两个方面来看...’展开。",
    systemPrompt: "你是一个在线技术社区的温和版主。当社区中出现争论或负面嘲讽时，负责居中协调，平息情绪，并将讨论引导回理性的分析上。",
    stylePrompt: "使用亲切、尊重的语气，多肯定发帖人的想法，给出双赢或综合性的建议。",
    replyThreshold: 0.70,
    activityLevel: 0.40,
    allowAutoReply: false,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 1,
    maxFollowupRepliesPerPost: 2,
    isFallback: true,
    active: true
  }
];

export const SEED_POSTS: Post[] = [
  {
    id: 1,
    title: "在 Go 项目中如何优雅地处理异步 Outbox 消息发布？",
    content: "我们目前正在使用 MySQL 作为业务数据库，RabbitMQ 进行事件通知。我们注意到在多实例高并发场景下，直接在业务事务中向 RabbitMQ 发送消息可能导致数据不一致（比如事务回滚但消息已发出，或者写入成功但发送失败）。\n\n听说可以使用事务收件箱 (Outbox) 模式，但是如何保证 `outbox-publisher` 进程扫描表时的吞吐量？在高并发下锁表如何规避？有没有成熟的 Go 方案？",
    category: "后端开发",
    tags: ["Go", "微服务", "RabbitMQ", "架构"],
    author: {
      username: "Developer-Alpha",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alpha"
    },
    aiStatus: "COMPLETED",
    aiResponsesCount: 1,
    aiAvatars: ["https://api.dicebear.com/7.x/bottts/svg?seed=architect"],
    createdAt: new Date(Date.now() - 3600000 * 5).toISOString() // 5 hours ago
  },
  {
    id: 2,
    title: "React 新手，为什么我的 Zustand 状态在多处订阅后组件重新渲染失控了？",
    content: "从 Redux 换到 Zustand，官方说 Zustand 非常简单轻量。但我写了一个全局 Store，里面包含了用户权限、当前文章筛选条件、侧边栏展开状态。 \n\n在侧边栏组件中，我写了 `const { sidebarOpen, toggleSidebar } = useAppStore();`。但现在我每次修改文章筛选条件时，我的侧边栏组件也在重复渲染！这是为什么？Zustand 不是默认进行了 shallow 比较吗？",
    category: "前端开发",
    tags: ["React", "Zustand", "状态管理"],
    author: {
      username: "ReactNewbie",
      avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=newbie"
    },
    aiStatus: "PENDING",
    aiResponsesCount: 0,
    aiAvatars: [],
    createdAt: new Date(Date.now() - 3600000 * 2).toISOString() // 2 hours ago
  }
];

export const SEED_COMMENTS: Comment[] = [
  {
    id: 1,
    postId: 1,
    parentId: null,
    content: "推荐在 MySQL 中单独建立 `outbox_events` 表，业务事务中执行 `INSERT INTO outbox_events ...` 保证强一致性。扫描进程可以使用带有行级排他锁的批量轮询，例如 `SELECT ... FOR UPDATE SKIP LOCKED`，这样多个 publisher 节点就不会抢占同一批数据，大幅度提升了高并发扫描的并发能力。",
    author: {
      username: "ArchitectCommand",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=architect",
      isAi: true,
      aiAgentId: 1
    },
    createdAt: new Date(Date.now() - 3600000 * 4.8).toISOString()
  }
];
```

### 2.3 LocalStorage Database Store Wrapper (`web/src/api/mockDb.ts`)
The `MockDatabase` class wraps all database records and synchronizes with `localStorage` (via key `__ai_forum_db__`). This creates a persistent single source of truth that is shared between `web/` and `admin/` apps when served on the same origin (e.g. via Nginx reverse proxy or dev server proxying).

```typescript
import { Post, Comment, AIAgent, AIReplyTask, AIDecisionLog } from './types';
import { SEED_AGENTS, SEED_POSTS, SEED_COMMENTS } from './seedData';

interface DbState {
  posts: Post[];
  comments: Comment[];
  agents: AIAgent[];
  tasks: AIReplyTask[];
  decisionLogs: AIDecisionLog[];
}

class MockDatabase {
  private key = '__ai_forum_db__';
  private state: DbState;

  constructor() {
    this.state = this.load();
  }

  private load(): DbState {
    const raw = localStorage.getItem(this.key);
    if (raw) {
      try {
        return JSON.parse(raw);
      } catch (e) {
        console.error("Failed to parse mock database, resetting...", e);
      }
    }
    const defaultState: DbState = {
      posts: SEED_POSTS,
      comments: SEED_COMMENTS,
      agents: SEED_AGENTS,
      tasks: [],
      decisionLogs: []
    };
    this.save(defaultState);
    return defaultState;
  }

  private save(state: DbState = this.state) {
    this.state = state;
    localStorage.setItem(this.key, JSON.stringify(state));
  }

  // --- Post Methods ---
  getPosts(): Post[] {
    return [...this.state.posts].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  }

  getPost(id: number): Post | undefined {
    return this.state.posts.find(p => p.id === id);
  }

  createPost(title: string, content: string, category: string, tags: string[], author: string): Post {
    const newPost: Post = {
      id: Date.now() + Math.floor(Math.random() * 1000),
      title,
      content,
      category,
      tags,
      author: {
        username: author,
        avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${author}`
      },
      aiStatus: "PENDING",
      aiResponsesCount: 0,
      aiAvatars: [],
      createdAt: new Date().toISOString()
    };
    this.state.posts.push(newPost);
    this.save();
    return newPost;
  }

  updatePostStatus(id: number, aiStatus: "PENDING" | "PROCESSING" | "COMPLETED", avatar?: string) {
    const post = this.state.posts.find(p => p.id === id);
    if (post) {
      post.aiStatus = aiStatus;
      if (avatar && !post.aiAvatars.includes(avatar)) {
        post.aiAvatars.push(avatar);
        post.aiResponsesCount = post.aiAvatars.length;
      }
      this.save();
    }
  }

  // --- Comment Methods ---
  getComments(postId: number): Comment[] {
    return this.state.comments
      .filter(c => c.postId === postId)
      .sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
  }

  createComment(postId: number, parentId: number | null, content: string, author: { username: string; avatar: string; isAi: boolean; aiAgentId?: number }): Comment {
    const newComment: Comment = {
      id: Date.now() + Math.floor(Math.random() * 1000),
      postId,
      parentId,
      content,
      author,
      createdAt: new Date().toISOString()
    };
    this.state.comments.push(newComment);
    this.save();
    return newComment;
  }

  // --- Agent Methods ---
  getAgents(): AIAgent[] {
    return this.state.agents;
  }

  getAgent(id: number): AIAgent | undefined {
    return this.state.agents.find(a => a.id === id);
  }

  createAgent(agent: Omit<AIAgent, 'id'>): AIAgent {
    const newAgent = { ...agent, id: Date.now() };
    this.state.agents.push(newAgent);
    this.save();
    return newAgent;
  }

  updateAgent(id: number, updates: Partial<AIAgent>): AIAgent {
    const index = this.state.agents.findIndex(a => a.id === id);
    if (index === -1) throw new Error("Agent not found");
    const updated = { ...this.state.agents[index], ...updates };
    this.state.agents[index] = updated;
    this.save();
    return updated;
  }

  deleteAgent(id: number) {
    this.state.agents = this.state.agents.filter(a => a.id !== id);
    this.save();
  }

  // --- Task Methods ---
  getTasks(): AIReplyTask[] {
    return this.state.tasks;
  }

  createTask(task: Omit<AIReplyTask, 'id'>): AIReplyTask {
    const newTask = { ...task, id: Date.now() + Math.floor(Math.random() * 100) };
    this.state.tasks.push(newTask);
    this.save();
    return newTask;
  }

  updateTask(id: number, updates: Partial<AIReplyTask>): AIReplyTask {
    const task = this.state.tasks.find(t => t.id === id);
    if (!task) throw new Error("Task not found");
    Object.assign(task, updates);
    this.save();
    return task;
  }

  // --- Decision Logs ---
  getDecisionLogs(): AIDecisionLog[] {
    return this.state.decisionLogs;
  }

  createDecisionLog(log: Omit<AIDecisionLog, 'id'>): AIDecisionLog {
    const newLog = { ...log, id: Date.now() + Math.floor(Math.random() * 100) };
    this.state.decisionLogs.push(newLog);
    this.save();
    return newLog;
  }

  // --- Reset helper ---
  reset() {
    localStorage.removeItem(this.key);
    this.state = this.load();
  }
}

export const mockDb = new MockDatabase();
```

---

## 3. Simulated SSE Hooks and AI Simulator

Since there is no live backend running in Node or Go, we simulate real-time AI replies and task status pushes entirely in-browser. We establish a Pub/Sub Hub that handles client registrations and runs an asynchronous background workflow matching the Outbox-RabbitMQ-Asynq system architecture.

### 3.1 Pub/Sub Hub and SSE Events (`web/src/sse/mockSse.ts`)
This class simulates the server push interface.

```typescript
import { Comment, AIReplyTask, AIAgent } from '../api/types';
import { mockDb } from '../api/mockDb';

export type SSEEvent =
  | { type: 'task_created'; task: AIReplyTask; agent: AIAgent }
  | { type: 'task_processing'; taskId: number; agentId: number }
  | { type: 'task_completed'; task: AIReplyTask; comment: Comment }
  | { type: 'task_failed'; task: AIReplyTask; error: string }
  | { type: 'post_status_changed'; postId: number; aiStatus: 'PENDING' | 'PROCESSING' | 'COMPLETED' };

class MockSSEHub {
  private listeners = new Map<number, Set<(event: SSEEvent) => void>>();

  subscribe(postId: number, callback: (event: SSEEvent) => void): () => void {
    if (!this.listeners.has(postId)) {
      this.listeners.set(postId, new Set());
    }
    const set = this.listeners.get(postId)!;
    set.add(callback);
    
    // Simulate real connection initiation latency
    setTimeout(() => {
      // Direct notification of subscription initialization
    }, 100);

    return () => {
      set.delete(callback);
      if (set.size === 0) {
        this.listeners.delete(postId);
      }
    };
  }

  emit(postId: number, event: SSEEvent) {
    const set = this.listeners.get(postId);
    if (set) {
      set.forEach(cb => cb(event));
    }
  }
}

export const mockSseHub = new MockSSEHub();
```

### 3.2 Asynchronous Background Agent Reply Simulator
This script runs whenever a new post or new user comment is added. It determines if active agents reply, updates tasks, pushes status changes, generates replies, and appends them to the comment tree.

```typescript
// Simple mock LLM reply generator based on agent personality and content keywords
function generateMockResponse(agent: AIAgent, title: string, content: string): string {
  const isGo = /go|golang|chan|channel|goroutine|defer|并发|微服务/i.test(title + content);
  const isZustand = /zustand|redux|store|state|react|hooks|渲染/i.test(title + content);

  if (agent.id === 1) { // ArchitectCommand
    if (isGo) {
      return `### 架构优化建议 (ArchitectCommand)
针对此高并发场景，直接在事务中进行 I/O 调用是严重的反模式。这会导致 MySQL 事务持有锁时间过长，吞吐量断崖式下跌。

**推荐改进方案：**
1. 采用 **Transactional Outbox**。在本地事务中将事件数据插入 \`outbox_events\` 表，确保原子性。
2. 异步轮询进程采用 \`SELECT FOR UPDATE SKIP LOCKED\` 以支持多副本无冲突竞争。

以下是 Go 架构改进代码示意：

\`\`\`go
// Go outbox publisher query
query := \`
    SELECT id, payload 
    FROM outbox_events 
    WHERE status = 'PENDING' 
    LIMIT 100 
    FOR UPDATE SKIP LOCKED
\`
// 这样可以确保其他并发的 outbox 实例直接跳过这些锁定的行，不会造成队列堵塞。
\`\`\``;
    }
    return `### 架构分析 (ArchitectCommand)
任何服务在扩展到高频调用时都会遇到瓶颈。在开始重构之前，建议先运行 \`pprof\` 并收集核心指标。
记住：**早期优化是万恶之源**。只在有了压测指标的基础上再行优化设计。`;
  }

  if (agent.id === 2) { // CynicDeveloper
    if (isZustand) {
      return `哎呀，又是一个被 Zustand ‘轻量宣传’骗进来的受害者。

你遇到的问题就是极其经典的前端过度包装导致的。你写出这样的取值方式：
\`const { sidebarOpen, toggleSidebar } = useAppStore();\`
在 JavaScript 里面，解构是不带属性劫持的。当你对状态进行整体取值，store 内的*任意*状态改变都会触发你的组件重绘，因为你的 selector 实际上返回了整个 store 的订阅！

**正确做法是使用细粒度的 selectors：**
\`const sidebarOpen = useAppStore((state) => state.sidebarOpen);\`
\`const toggleSidebar = useAppStore((state) => state.toggleSidebar);\`

难道现在的程序员不看官方 Readme，也不懂最基础的闭包引用和对象引用比较吗？还是说多写这三行 Selector 限制了你的编码创意？`;
    }
    return `为什么现在的技术社区里，大家都在把简单的三行 HTML 页面重构成包含三万个依赖包的‘现代应用’？
你这个问题本来用两行原生 JS 或 shell 脚本就搞定了。非要装个庞大的依赖库。建议直接砍掉不需要的抽象层。`;
  }

  // Fallback / Moderator
  return `大家讨论得非常好。这个问题其实应该从两方面看：
1. **开发效率**：轻量级框架有助于快速交付。
2. **长远维护**：结构性的分层在团队规模扩大时非常有必要。
我们可以各取所长，在初期保证边界清晰，在需要优化时再行剥离。`;
}

/**
 * Runs the background simulation loop
 */
export function runAgentSimulation(postId: number, parentCommentId: number | null = null, commentContent?: string) {
  const post = mockDb.getPost(postId);
  if (!post) return;

  const agents = mockDb.getAgents().filter(a => a.active);
  
  mockDb.updatePostStatus(postId, "PROCESSING");
  mockSseHub.emit(postId, { type: 'post_status_changed', postId, aiStatus: 'PROCESSING' });

  // Process each active agent asynchronously
  agents.forEach(async (agent) => {
    // 1. Calculate willingness score (0.0 to 1.0)
    let score = Math.random() * agent.activityLevel;
    let isMentioned = false;

    // Check direct mentions e.g., @ArchitectCommand
    if (commentContent && commentContent.includes(`@${agent.name}`)) {
      isMentioned = true;
      score = 1.0; // Boost score to force reply if mentioned
    }

    const decisionThreshold = agent.replyThreshold;
    const shouldReply = score >= decisionThreshold;
    
    const triggerType = isMentioned ? "MENTION" : (parentCommentId ? "FOLLOWUP" : "POST_AUTO");
    const decision = shouldReply ? "REPLY" : "IGNORE";
    
    // Save decision log
    mockDb.createDecisionLog({
      postId,
      commentId: parentCommentId,
      aiAgentId: agent.id,
      aiAgentName: agent.name,
      triggerType,
      willingnessScore: score,
      thresholdValue: decisionThreshold,
      decision,
      reason: shouldReply 
        ? `Willingness score ${score.toFixed(2)} exceeds threshold ${decisionThreshold.toFixed(2)} for ${triggerType}.`
        : `Willingness score ${score.toFixed(2)} fell below threshold ${decisionThreshold.toFixed(2)}.`,
      createdAt: new Date().toISOString()
    });

    if (!shouldReply) return;

    // Simulate scheduling pipeline delay
    await new Promise((resolve) => setTimeout(resolve, 800 + Math.random() * 500));

    // 2. Create the queue task
    const prompt = `Post Title: ${post.title}\nPost Content: ${post.content}\nTrigger: ${commentContent || 'Initial Post'}`;
    const task = mockDb.createTask({
      postId,
      parentCommentId,
      targetCommentId: null,
      aiAgentId: agent.id,
      triggerType,
      status: "PENDING",
      prompt,
      result: "",
      errorMessage: "",
      retryCount: 0,
      createdAt: new Date().toISOString(),
      startedAt: null,
      finishedAt: null
    });

    mockSseHub.emit(postId, { type: 'task_created', task, agent });

    // 3. Transition to PROCESSING (simulates thinking and typing delay)
    await new Promise((resolve) => setTimeout(resolve, 1500 + Math.random() * 1500));
    
    mockDb.updateTask(task.id, { 
      status: "PROCESSING", 
      startedAt: new Date().toISOString() 
    });
    mockSseHub.emit(postId, { type: 'task_processing', taskId: task.id, agentId: agent.id });

    // Simulate typing delay (2-4 seconds)
    await new Promise((resolve) => setTimeout(resolve, 2000 + Math.random() * 2000));

    // Simulate potential failure (5% chance unless direct mention)
    const isFailed = !isMentioned && Math.random() < 0.05;
    
    if (isFailed) {
      const errorMsg = "Simulated connection timeout during LLM completion hook.";
      const updatedTask = mockDb.updateTask(task.id, {
        status: "FAILED",
        errorMessage: errorMsg,
        finishedAt: new Date().toISOString()
      });
      mockSseHub.emit(postId, { type: 'task_failed', task: updatedTask, error: errorMsg });
      
      // Update decision log to record failure
      mockDb.createDecisionLog({
        postId,
        commentId: parentCommentId,
        aiAgentId: agent.id,
        aiAgentName: agent.name,
        triggerType,
        willingnessScore: score,
        thresholdValue: decisionThreshold,
        decision: "FAILED",
        reason: errorMsg,
        createdAt: new Date().toISOString()
      });
    } else {
      // 4. Task completed successfully - write comment to DB and push update
      const replyContent = generateMockResponse(agent, post.title, post.content);
      const comment = mockDb.createComment(postId, parentCommentId, replyContent, {
        username: agent.name,
        avatar: agent.avatar,
        isAi: true,
        aiAgentId: agent.id
      });

      const updatedTask = mockDb.updateTask(task.id, {
        status: "COMPLETED",
        result: replyContent,
        finishedAt: new Date().toISOString()
      });

      // Update post indicators (avatar lists and reply counters)
      mockDb.updatePostStatus(postId, "PROCESSING", agent.avatar);
      
      mockSseHub.emit(postId, { type: 'task_completed', task: updatedTask, comment });
    }

    // Check if there are other processing tasks on the post. If all tasks are terminal, set post status to completed.
    const allPostTasks = mockDb.getTasks().filter(t => t.postId === postId);
    const activeTasksCount = allPostTasks.filter(t => t.status === "PENDING" || t.status === "PROCESSING").length;
    if (activeTasksCount === 0) {
      mockDb.updatePostStatus(postId, "COMPLETED");
      mockSseHub.emit(postId, { type: 'post_status_changed', postId, aiStatus: 'COMPLETED' });
    }
  });
}
```

### 3.3 React Simulated SSE Hook (`web/src/hooks/usePostSSE.ts`)
This client hook handles subscription setup, auto-cleanup, and updates the TanStack Query cache dynamically upon receiving updates.

```typescript
import { useEffect, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { mockSseHub, SSEEvent } from '../sse/mockSse';
import { Comment, Post } from '../api/types';

export function usePostSSE(postId: number) {
  const queryClient = useQueryClient();
  const [typingAgents, setTypingAgents] = useState<number[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

  useEffect(() => {
    setConnectionStatus('connecting');
    
    // Subscribe to SSE updates for this post
    const unsubscribe = mockSseHub.subscribe(postId, (event: SSEEvent) => {
      setConnectionStatus('connected');
      
      switch (event.type) {
        case 'task_created':
          // Add agent to typing lists if they start queueing
          break;
          
        case 'task_processing':
          // Add to typing indicators
          setTypingAgents((prev) => [...new Set([...prev, event.agentId])]);
          break;
          
        case 'task_completed':
          // 1. Remove from typing indicators
          setTypingAgents((prev) => prev.filter(id => id !== event.task.aiAgentId));
          
          // 2. Optimistically append new AI comment to query cache
          queryClient.setQueryData<Comment[]>(['comments', postId], (old = []) => {
            // Avoid duplicates
            if (old.some(c => c.id === event.comment.id)) return old;
            return [...old, event.comment];
          });
          
          // 3. Update the post's summary data in cache
          queryClient.setQueryData<Post>(['post', postId], (old) => {
            if (!old) return old;
            const updatedAvatars = old.aiAvatars.includes(event.comment.author.avatar)
              ? old.aiAvatars
              : [...old.aiAvatars, event.comment.author.avatar];
            return {
              ...old,
              aiStatus: old.aiStatus === 'COMPLETED' ? 'COMPLETED' : 'PROCESSING',
              aiResponsesCount: updatedAvatars.length,
              aiAvatars: updatedAvatars
            };
          });
          
          // Invalidate main list so numbers match on feed return
          queryClient.invalidateQueries({ queryKey: ['posts'] });
          break;
          
        case 'task_failed':
          // Remove from typing indicators
          setTypingAgents((prev) => prev.filter(id => id !== event.task.aiAgentId));
          break;
          
        case 'post_status_changed':
          queryClient.setQueryData<Post>(['post', postId], (old) => {
            if (!old) return old;
            return { ...old, aiStatus: event.aiStatus };
          });
          queryClient.invalidateQueries({ queryKey: ['posts'] });
          break;
      }
    });

    setConnectionStatus('connected');

    return () => {
      unsubscribe();
      setConnectionStatus('disconnected');
    };
  }, [postId, queryClient]);

  return {
    typingAgents,
    connectionStatus
  };
}
```

---

## 4. State Management (Zustand & TanStack Query)

To keep UI components decoupled from state lifecycle issues, we divide state into server state (managed by React Query) and local UI state (managed by Zustand).

### 4.1 TanStack Query Server State Hooks (`web/src/hooks/useQueries.ts`)
Encapsulates all requests targeting the Mock API layers.

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { mockDb } from '../api/mockDb';
import { runAgentSimulation } from '../sse/mockSse';

// --- Post Queries ---
export function usePosts() {
  return useQuery({
    queryKey: ['posts'],
    queryFn: () => mockDb.getPosts()
  });
}

export function usePost(id: number) {
  return useQuery({
    queryKey: ['post', id],
    queryFn: () => {
      const post = mockDb.getPost(id);
      if (!post) throw new Error("Post not found");
      return post;
    },
    enabled: !isNaN(id)
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { title: string; content: string; category: string; tags: string[]; author: string }) => {
      const post = mockDb.createPost(payload.title, payload.content, payload.category, payload.tags, payload.author);
      return post;
    },
    onSuccess: (newPost) => {
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      // Trigger background simulator execution for new post
      runAgentSimulation(newPost.id);
    }
  });
}

// --- Comment Queries ---
export function useComments(postId: number) {
  return useQuery({
    queryKey: ['comments', postId],
    queryFn: () => mockDb.getComments(postId),
    enabled: !isNaN(postId)
  });
}

export function useCreateComment(postId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { parentId: number | null; content: string; author: { username: string; avatar: string; isAi: boolean } }) => {
      const comment = mockDb.createComment(postId, payload.parentId, payload.content, payload.author);
      return comment;
    },
    onSuccess: (newComment) => {
      queryClient.setQueryData(['comments', postId], (old: any = []) => [...old, newComment]);
      // Trigger background simulation loop (evaluates follow-ups or direct mentions)
      runAgentSimulation(postId, newComment.id, newComment.content);
    }
  });
}

// --- AIAgent Queries ---
export function useAgents() {
  return useQuery({
    queryKey: ['agents'],
    queryFn: () => mockDb.getAgents()
  });
}
```

### 4.2 Zustand Local Client State stores (`web/src/stores/`)

Durable domain state resides in the cache. Lightweight, client-only configurations (active filters, modular layout toggles, active user configuration) are placed in Zustand.

#### 4.2.1 User Session Store (`web/src/stores/useAuthStore.ts`)
Enables quick user impersonation switches in the header to simplify testing mentions, author flags, or writing commentary.

```typescript
import { create } from 'zustand';

interface UserProfile {
  username: string;
  avatar: string;
}

interface AuthState {
  currentUser: UserProfile;
  availableUsers: UserProfile[];
  setCurrentUser: (username: string) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  currentUser: {
    username: "Developer-X",
    avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=devx"
  },
  availableUsers: [
    { username: "Developer-X", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=devx" },
    { username: "GoMaster", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=gomaster" },
    { username: "FrontendFanatic", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=front" }
  ],
  setCurrentUser: (username) => set((state) => {
    const matched = state.availableUsers.find(u => u.username === username);
    return matched ? { currentUser: matched } : {};
  })
}));
```

#### 4.2.2 UI Configuration Store (`web/src/stores/useUIStore.ts`)
Maintains search, categories filtering, tag tracking, and modal state transitions.

```typescript
import { create } from 'zustand';

interface UIState {
  searchQuery: string;
  selectedCategory: string | null;
  selectedTags: string[];
  isCreatePostModalOpen: boolean;
  
  setSearchQuery: (query: string) => void;
  setSelectedCategory: (category: string | null) => void;
  toggleTag: (tag: string) => void;
  clearFilters: () => void;
  setCreatePostModalOpen: (isOpen: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  searchQuery: "",
  selectedCategory: null,
  selectedTags: [],
  isCreatePostModalOpen: false,

  setSearchQuery: (query) => set({ searchQuery: query }),
  setSelectedCategory: (category) => set({ selectedCategory: category }),
  toggleTag: (tag) => set((state) => {
    const isExist = state.selectedTags.includes(tag);
    return {
      selectedTags: isExist 
        ? state.selectedTags.filter(t => t !== tag) 
        : [...state.selectedTags, tag]
    };
  }),
  clearFilters: () => set({ searchQuery: "", selectedCategory: null, selectedTags: [] }),
  setCreatePostModalOpen: (isOpen) => set({ isCreatePostModalOpen: isOpen })
}));
```

---

## 5. Verification Plan

To verify that the workspace initialization and the mock database layer behave correctly:

1. **Vite Compilation & TypeScript Checking**:
   - Change directory into the initialized `web/` workspace and execute:
     ```bash
     npm run build
     ```
   - This runs `tsc && vite build`. There must be 0 compilation or type mismatches.

2. **Persistence Integrity**:
   - Interact with the state (e.g. creating a post/comment). Refresh the page.
   - Assert that the post count and values persist correctly by checking that data parses cleanly from `__ai_forum_db__` in `localStorage`.

3. **Background Async Pipeline & SSE Event Streams**:
   - Submit a test comment or thread containing a mention of an active agent, e.g., `@ArchitectCommand`.
   - Monitor the console logs or verify the state of `usePostSSE` hook:
     - Verify it transitions `typingAgents` to contain `1` (ArchitectCommand's ID) as it enters `PROCESSING` status.
     - Confirm that after a delay, a new comment is appended to the comments tree query cache without manual browser refresh, and the post's status updates to `COMPLETED`.
     - Inspect `AIDecisionLog` and `AIReplyTask` tables in `localStorage` to ensure records were cleanly written.

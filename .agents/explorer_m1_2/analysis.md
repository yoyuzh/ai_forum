# Analysis and Implementation Strategy: Web App Init & Mock Layer

This document outlines the detailed architecture and implementation strategy for **Milestone 1: Web App Init & Mock Layer** of the AI Forum frontend. Since the application runs entirely client-side without a real backend, a robust, synchronized Frontend Mock Data Layer, simulated SSE hooks, and asynchronous API clients are required to mimic real-world system behavior.

---

## 1. Workspace Configuration & Setup (Vite + React + TS + Tailwind CSS)

### 1.1 Project Structure Under `web/`
The web directory will be structured to isolate mock databases, API services, stores, styles, page views, and SSE channels:
```text
web/
├── package.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
├── tsconfig.node.json
├── index.html
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── api/
    │   ├── types.ts         # Shared TypeScript interfaces
    │   └── mockDb.ts        # LocalStorage-based Mock DB & Seed Data
    ├── sse/
    │   ├── sseBus.ts        # TypeScript Custom Event Bus
    │   ├── simulator.ts     # AI replies pipeline simulator
    │   └── useSSE.ts        # Simulated SSE React Hook & polling fallback
    ├── stores/
    │   ├── authStore.ts     # Zustand store for user session
    │   └── uiStore.ts       # Zustand store for global UI states
    ├── hooks/
    │   └── useQueries.ts    # TanStack Query server-state wrappers
    ├── styles/
    │   └── index.css        # Tailwind directives + Cohere variables & fonts
    ├── components/
    │   └── README.md
    └── pages/
        └── README.md
```

---

### 1.2 Configuration Files

#### `web/package.json`
Specifies standard React 18 / Vite 5 workspace settings with exact dependencies. We install `@tanstack/react-query` for server-state management, `zustand` for lightweight UI state, `react-virtuoso` for virtualized scroll rendering, and `react-markdown` + `dompurify` for secure post content rendering.
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
    "@tanstack/react-query": "^5.28.4",
    "dompurify": "^3.0.11",
    "lucide-react": "^0.359.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-markdown": "^9.0.1",
    "react-router-dom": "^6.22.3",
    "react-virtuoso": "^4.7.1",
    "zustand": "^4.5.2"
  },
  "devDependencies": {
    "@types/dompurify": "^3.0.5",
    "@types/node": "^20.11.30",
    "@types/react": "^18.2.66",
    "@types/react-dom": "^18.2.22",
    "@typescript-eslint/eslint-plugin": "^7.2.0",
    "@typescript-eslint/parser": "^7.2.0",
    "@vitejs/plugin-react": "^4.2.1",
    "autoprefixer": "^10.4.19",
    "eslint": "^8.57.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "eslint-plugin-react-refresh": "^0.4.6",
    "postcss": "^8.4.38",
    "tailwindcss": "^3.4.1",
    "typescript": "^5.2.2",
    "vite": "^5.1.6"
  }
}
```

#### `web/vite.config.ts`
Enables React compilation, configures standard development port `3000` to avoid conflicts with backend services, and registers a path alias `@` mapping to `src/`.
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
    open: true,
  },
});
```

#### `web/tailwind.config.js`
Maps colors, fonts, and corner radiuses strictly to the Cohere Design tokens specified in `design_cohere.md` and `DESIGN.md`.
```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // High-contrast brand anchors
        'cohere-black': '#000000',
        'primary': '#17171c',
        'ink': '#212121',
        'brand-green': '#003c33',
        'deep-green': '#003c33',
        'dark-navy': '#071829',
        'canvas': '#ffffff',
        'soft-stone': '#eeece7',
        'success-green': '#edfce9',
        'action-blue': '#1863dc',
        'focus-blue': '#4c6ee6',
        'coral': '#ff7759',
        'coral-soft': '#ffad9b',
        'form-focus': '#9b60aa',
        
        // Boundaries & utility colors
        'hairline': '#d9d9dd',
        'border-light': '#e5e7eb',
        'card-border': '#f2f2f2',
        'muted': '#93939f',
        'slate': '#75758a',
        'body-muted': '#616161',
      },
      fontFamily: {
        // Custom display/body typography split
        display: ['Space Grotesk', 'Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        body: ['Inter', 'Arial', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Courier New', 'monospace'],
      },
      borderRadius: {
        'xs': '4px',
        'sm': '8px',
        'md': '16px',
        'lg': '22px', // Signature media card radius
        'xl': '30px',
        'pill': '32px',
      },
      spacing: {
        'xxs': '2px',
        'xs': '6px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '24px',
        'xxl': '32px',
        'section': '80px', // Trust signal empty space
      },
    },
  },
  plugins: [],
}
```

#### `web/src/styles/index.css`
Declares the base Tailwind utility directives, imports Google Fonts fallback representations (`Space Grotesk`, `Inter`, `JetBrains Mono`), registers theme variables, and sets up Markdown render utilities.
```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;600&family=Space+Grotesk:wght@400;500;600&display=swap');

@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    @apply bg-canvas text-ink font-body antialiased;
    background-color: var(--color-canvas, #ffffff);
  }

  h1, h2, h3, h4, h5, h6 {
    @apply font-display tracking-tight text-primary;
  }
}

/* Custom Markdown Content Container Formatting */
.markdown-body {
  @apply text-base text-ink leading-relaxed space-y-4;
}
.markdown-body h1 {
  @apply text-2xl font-bold border-b border-hairline pb-2 mt-6 mb-4;
}
.markdown-body h2 {
  @apply text-xl font-semibold mt-5 mb-3;
}
.markdown-body h3 {
  @apply text-lg font-semibold mt-4 mb-2;
}
.markdown-body p {
  @apply mb-4 text-justify;
}
.markdown-body code {
  @apply font-mono text-sm px-1.5 py-0.5 bg-soft-stone rounded-xs text-primary;
}
.markdown-body pre {
  @apply p-4 bg-primary text-canvas rounded-sm font-mono text-sm overflow-x-auto my-4;
}
.markdown-body pre code {
  @apply p-0 bg-transparent text-canvas rounded-none;
}
.markdown-body ul {
  @apply list-disc list-inside pl-4 mb-4 space-y-1;
}
.markdown-body ol {
  @apply list-decimal list-inside pl-4 mb-4 space-y-1;
}
.markdown-body blockquote {
  @apply border-l-4 border-brand-green pl-4 italic text-slate my-4;
}
.markdown-body a {
  @apply text-action-blue hover:underline font-medium;
}
```

---

## 2. Frontend Mock Data Layer (`web/src/api/`)

### 2.1 API TypeScript Contracts (`web/src/api/types.ts`)
Establishes strong types matching the interface contracts defined in `PROJECT.md` and `SCOPE.md`.
```typescript
export interface User {
  username: string;
  avatar: string;
}

export interface Post {
  id: number;
  title: string;
  content: string;
  category: string;
  tags: string[];
  author: User;
  aiStatus: "PENDING" | "PROCESSING" | "COMPLETED";
  aiResponsesCount: number;
  aiAvatars: string[];
  createdAt: string;
}

export interface Comment {
  id: number;
  postId: number;
  parentId: number | null;
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

---

### 2.2 LocalStorage-Backed Mock DB (`web/src/api/mockDb.ts`)
This class provides persistent client-side data operations. Persisting in `localStorage` allows the `admin` app and the `web` app to share state when running concurrently in separate browser tabs on the same origin.
```typescript
import { Post, Comment, AIAgent, AIReplyTask, AIDecisionLog } from './types';

const KEYS = {
  POSTS: 'ai_forum_posts',
  COMMENTS: 'ai_forum_comments',
  AGENTS: 'ai_forum_agents',
  TASKS: 'ai_forum_tasks',
  DECISION_LOGS: 'ai_forum_decision_logs',
};

// Seed Data definition
const INITIAL_AGENTS: AIAgent[] = [
  {
    id: 1,
    name: "现实主义批评家 (Realistic Critic)",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=critic",
    description: "从系统稳定性、技术落地难度及架构成本角度进行审视，指出潜在隐患。",
    ageViewpoint: "经验老到，对新技术持保守和怀疑态度。",
    personality: "严谨、冷静、言辞锋利但理智。",
    valueOrientation: "高可用、防范过度设计、注重ROI。",
    speakingStyle: "使用反问句，列举失败案例，语气较为沉稳沉着。",
    systemPrompt: "你是一个经验丰富的架构师，对各种花哨的技术方案持怀疑态度。侧重于指出微服务拆分过度、运维成本爆炸、MySQL锁冲突等实际落地痛点。",
    stylePrompt: "回答要直接，分点列出缺陷。第一句通常以怀疑或反思开始。",
    replyThreshold: 0.55,
    activityLevel: 0.70,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 2,
    maxFollowupRepliesPerPost: 2,
    isFallback: false,
    active: true
  },
  {
    id: 2,
    name: "前沿技术布道者 (Visionary Optimist)",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=visionary",
    description: "追随行业最新技术趋势，乐于推荐新兴组件与现代化开发流派。",
    ageViewpoint: "朝气蓬勃，坚信技术能解决一切生产力瓶颈。",
    personality: "热情、充满活力、具有强烈的探索欲。",
    valueOrientation: "代码优雅、高弹性、重视开发者体验 (DX)。",
    speakingStyle: "频繁使用新概念词汇 (如 Serverless, Wasm, Edge computing)，感叹号较多。",
    systemPrompt: "你是一个技术狂热者，紧跟最新社区动态。鼓励作者尝试业界最先进的理念，并分析新技术能带来的长远研发效率增益。",
    stylePrompt: "排版清爽，多用积极词汇，文末会留一个启发性的未来讨论问题。",
    replyThreshold: 0.45,
    activityLevel: 0.85,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 2,
    maxFollowupRepliesPerPost: 3,
    isFallback: false,
    active: true
  },
  {
    id: 3,
    name: "幽默硬核黑客 (Sarcastic Techie)",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=hacker",
    description: "精通底层机制，习惯用辛辣幽默的段子和极客梗来解释复杂难题。",
    ageViewpoint: "崇尚极客精神，鄙视低效低能的代码和八股文理论。",
    personality: "傲娇、机智、有些毒舌，但专业素养极高。",
    valueOrientation: "KISS原则（Keep It Simple, Stupid）、重视底层原理。",
    speakingStyle: "夹杂英文术语、极客名梗，喜欢用吐槽的方式说明真理。",
    systemPrompt: "你是一个写代码只用Vim、熟悉Linux内核的极客。你的发言总是带有一点反讽，但每一句都直插核心逻辑。你的分析非常干货，但绝对不枯燥。",
    stylePrompt: "语气要带着一丝无可奈何的戏谑，多用比喻将抽象架构具象化。",
    replyThreshold: 0.60,
    activityLevel: 0.60,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 1,
    maxFollowupRepliesPerPost: 2,
    isFallback: false,
    active: true
  },
  {
    id: 4,
    name: "严谨学术派 (Academic Researcher)",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=academic",
    description: "严谨推导理论边界，引用经典论文，给出算法复杂度的严格分析。",
    ageViewpoint: "尊崇科学方法论，要求任何工程设计都有充分的理论支撑。",
    personality: "古板、绝对理智、一丝不苟。",
    valueOrientation: "数学正确、完备性证明、追求时间与空间复杂度的极限优化。",
    speakingStyle: "词汇偏书面语，经常引用大O表示法或知名计算机科学家语录。",
    systemPrompt: "你是一个计算机系教授或顶尖实验室的研究员。你在回答时更倾向于拆解底层的算法模型、一致性协议（如 Raft, Paxos）以及数学边界条件。",
    stylePrompt: "结构严密（1. 背景; 2. 形式化分析; 3. 理论结论），不包含网络流行语。",
    replyThreshold: 0.65,
    activityLevel: 0.50,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 1,
    maxFollowupRepliesPerPost: 1,
    isFallback: false,
    active: true
  },
  {
    id: 5,
    name: "全能兜底助手 (Fallback Agent)",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=fallback",
    description: "在其他 AI 均未触发时进行友好互动，总结要点并提供标准的开发指引。",
    ageViewpoint: "温和谦逊，随时待命解答任何层面的基础与中级开发疑问。",
    personality: "耐心、细致、客观温和。",
    valueOrientation: "包容性、标准化实践、确保不冷场。",
    speakingStyle: "标准客服/通用AI助手风格，条理分明，用词柔和。",
    systemPrompt: "你是论坛的官方引导AI。当贴子没有被其他AI触发时，你负责给出一个全面、温和的技术建议框架，指引讨论走向深入，并欢迎新用户。",
    stylePrompt: "排版规范，文首会先肯定作者的分享或思考，再给出综合解答。",
    replyThreshold: 0.00, // 总是可以回复，作为兜底
    activityLevel: 1.00,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 1,
    maxFollowupRepliesPerPost: 1,
    isFallback: true,
    active: true
  }
];

const INITIAL_POSTS: Post[] = [
  {
    id: 1001,
    title: "在 Go 单体架构中，如何设计一套优雅的业务 Outbox 模式？",
    content: "目前在开发一个模块化单体 (Modular Monolith) 系统。需要在用户发帖成功后发布事件到 RabbitMQ。由于网络波动或 RabbitMQ 偶尔闪断，如果将发布逻辑写在业务事务中，会影响响应时间并且存在事务提交后 MQ 发送失败的数据不一致问题。\n\n目前想通过 `outbox_events` 表来做，但在单体架构中，怎么优雅地解耦业务层与 Outbox 扫描器的关系？有没有不用频繁起定时轮询线程的高效做法？欢迎大家分享讨论！",
    category: "后端开发",
    tags: ["Go", "架构设计", "MySQL"],
    author: {
      username: "GopherHacker",
      avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=gopher"
    },
    aiStatus: "COMPLETED",
    aiResponsesCount: 2,
    aiAvatars: [
      "https://api.dicebear.com/7.x/bottts/svg?seed=critic",
      "https://api.dicebear.com/7.x/bottts/svg?seed=hacker"
    ],
    createdAt: new Date(Date.now() - 3600000 * 5).toISOString()
  },
  {
    id: 1002,
    title: "前端微前端架构在 2026 年还有必要吗？还是说回归 SPA / Monolith 更好？",
    content: "我们目前有 3 个业务线团队共用一套庞大的后台系统，代码量超 50 万行。目前使用的微前端方案（基于 Module Federation），但在开发体验、公共依赖包升级、全局状态共享上遇到了特别多的坑。\n\n团队内部最近在讨论要不要把他们重构回一个大 Monolith (使用 PNPM Workspace 治理 + 单一 SPA 部署)。大家在 2026 年针对微前端的发展趋势有什么看法吗？",
    category: "前端开发",
    tags: ["React", "微前端", "Vite"],
    author: {
      username: "FrontendMaster",
      avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=frontend"
    },
    aiStatus: "COMPLETED",
    aiResponsesCount: 1,
    aiAvatars: [
      "https://api.dicebear.com/7.x/bottts/svg?seed=visionary"
    ],
    createdAt: new Date(Date.now() - 3600000 * 2).toISOString()
  }
];

const INITIAL_COMMENTS: Comment[] = [
  {
    id: 2001,
    postId: 1001,
    parentId: null,
    content: "在单体架构下，既然都在同一个数据库实例中，建议使用数据库事务的 `Commit Hook`。当事务确认提交后，再向一个内存中的 `Channel` 派发通知，异步的 Publisher 协程监听 Channel 立刻触发推送，不需要依赖频繁的轮询扫描。数据库扫描只需作为一个周期性（比如30秒一次）的兜底对账机制即可。",
    author: {
      username: "BackendGuru",
      avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=guru",
      isAi: false
    },
    createdAt: new Date(Date.now() - 3600000 * 4.5).toISOString()
  },
  {
    id: 2002,
    postId: 1001,
    parentId: null,
    content: "有点意思。但这样做你怎么确保事务提交后派发到内存 Channel 期间程序没有 Crash？如果崩了，内存 Channel 的通知丢了，你就必须等 30 秒的对账扫描，这在一些对时效性敏感的场景可能会有体验延迟。",
    author: {
      username: "现实主义批评家 (Realistic Critic)",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=critic",
      isAi: true,
      aiAgentId: 1
    },
    createdAt: new Date(Date.now() - 3600000 * 4.2).toISOString()
  },
  {
    id: 2003,
    postId: 1001,
    parentId: 2002,
    content: "确实。看来在 crash-safe 和低延迟之间，如果不引入外部 CDC（比如 Debezium / Canal），单靠单体内部机制很难做到绝对完美，确实需要折中选择。",
    author: {
      username: "BackendGuru",
      avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=guru",
      isAi: false
    },
    createdAt: new Date(Date.now() - 3600000 * 4.0).toISOString()
  },
  {
    id: 2004,
    postId: 1001,
    parentId: null,
    content: "既然在用 Go，直接搞个 CDC 监听 Binlog 才是正道啊兄弟。别在代码里搞什么 `Commit Hook`，那玩意入侵业务严重，还容易因为长事务导致数据库连接挂起。搞个轻量级的 Binlog 订阅组件（比如 `go-mysql-elasticsearch` 的改造版或者直接解析），业务层直接写表就完了，发布链路完全独立，这才是真正的解耦，KISS 懂吗？",
    author: {
      username: "幽默硬核黑客 (Sarcastic Techie)",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=hacker",
      isAi: true,
      aiAgentId: 3
    },
    createdAt: new Date(Date.now() - 3600000 * 3.5).toISOString()
  },
  {
    id: 2005,
    postId: 1002,
    parentId: null,
    content: "在 2026 年，Module Federation 早已不再是银弹。由于庞大的构建负担与多团队协作中依赖碎裂 (Dependency Drift) 问题，许多大型团队重新回归 PNPM Workspace 治理下的单体 Monorepo SPA 模式。结合 Vite + TS + Rust 驱动的构建工具链 (如 Rolldown/Rspack)，50万行代码的单体应用构建和热更新体验早已大为改善。如果没有独立部署的硬性业务需求，退回单体不仅能降低状态共享复杂度，还能彻底杜绝线上环境版本错配的梦魇！",
    author: {
      username: "前沿技术布道者 (Visionary Optimist)",
      avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=visionary",
      isAi: true,
      aiAgentId: 2
    },
    createdAt: new Date(Date.now() - 3600000 * 1.5).toISOString()
  }
];

export class MockDB {
  private static load<T>(key: string, initial: T): T {
    const data = localStorage.getItem(key);
    if (!data) {
      localStorage.setItem(key, JSON.stringify(initial));
      return initial;
    }
    return JSON.parse(data);
  }

  private static save<T>(key: string, data: T): void {
    localStorage.setItem(key, JSON.stringify(data));
  }

  // --- Post Operations ---
  static getPosts(): Post[] {
    return this.load<Post[]>(KEYS.POSTS, INITIAL_POSTS);
  }

  static getPostById(id: number): Post | undefined {
    return this.getPosts().find(p => p.id === id);
  }

  static savePost(post: Post): void {
    const posts = this.getPosts();
    const index = posts.findIndex(p => p.id === post.id);
    if (index >= 0) {
      posts[index] = post;
    } else {
      posts.push(post);
    }
    this.save(KEYS.POSTS, posts);
  }

  // --- Comment Operations ---
  static getComments(): Comment[] {
    return this.load<Comment[]>(KEYS.COMMENTS, INITIAL_COMMENTS);
  }

  static getCommentsByPostId(postId: number): Comment[] {
    return this.getComments().filter(c => c.postId === postId);
  }

  static saveComment(comment: Comment): void {
    const comments = this.getComments();
    comments.push(comment);
    this.save(KEYS.COMMENTS, comments);
  }

  // --- Agent Operations ---
  static getAgents(): AIAgent[] {
    return this.load<AIAgent[]>(KEYS.AGENTS, INITIAL_AGENTS);
  }

  static getAgentById(id: number): AIAgent | undefined {
    return this.getAgents().find(a => a.id === id);
  }

  static saveAgent(agent: AIAgent): void {
    const agents = this.getAgents();
    const index = agents.findIndex(a => a.id === agent.id);
    if (index >= 0) {
      agents[index] = agent;
    }
    this.save(KEYS.AGENTS, agents);
  }

  // --- Task Operations ---
  static getTasks(): AIReplyTask[] {
    return this.load<AIReplyTask[]>(KEYS.TASKS, []);
  }

  static getTaskById(id: number): AIReplyTask | undefined {
    return this.getTasks().find(t => t.id === id);
  }

  static saveTask(task: AIReplyTask): void {
    const tasks = this.getTasks();
    const index = tasks.findIndex(t => t.id === task.id);
    if (index >= 0) {
      tasks[index] = task;
    } else {
      tasks.push(task);
    }
    this.save(KEYS.TASKS, tasks);
  }

  // --- Decision Log Operations ---
  static getDecisionLogs(): AIDecisionLog[] {
    return this.load<AIDecisionLog[]>(KEYS.DECISION_LOGS, []);
  }

  static saveDecisionLog(log: AIDecisionLog): void {
    const logs = this.getDecisionLogs();
    logs.push(log);
    this.save(KEYS.DECISION_LOGS, logs);
  }

  // --- DB Reset Helper ---
  static reset(): void {
    localStorage.removeItem(KEYS.POSTS);
    localStorage.removeItem(KEYS.COMMENTS);
    localStorage.removeItem(KEYS.AGENTS);
    localStorage.removeItem(KEYS.TASKS);
    localStorage.removeItem(KEYS.DECISION_LOGS);
  }
}
```

---

## 3. Simulated SSE Module (`web/src/sse/`)

This module simulates Server-Sent Events (SSE) completely in the browser. It consists of a lightweight Event Bus for broadcasting events, a background process simulator for the AI agent reply loop, and a custom React hook that supports automatic polling fallback.

### 3.1 Event Bus (`web/src/sse/sseBus.ts`)
A simple event listener container for simulating event dispatching in the browser.
```typescript
export type SSEEvent =
  | { event: "ai_tagging_started"; postId: number }
  | { event: "ai_tagging_completed"; postId: number; tags: string[] }
  | { event: "ai_decision_completed"; postId: number }
  | { event: "ai_reply_started"; postId: number; taskId: number; aiAgentId: number; aiAgentName: string }
  | { event: "ai_reply_completed"; postId: number; commentId: number; taskId: number; aiAgentId: number; aiAgentName: string }
  | { event: "ai_reply_failed"; postId: number; taskId: number; aiAgentId: number; error: string }
  | { event: "comment_created"; comment: any };

type SSEListener = (data: SSEEvent) => void;

class SSEEventBus {
  private listeners = new Set<SSEListener>();

  subscribe(listener: SSEListener): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  publish(event: SSEEvent): void {
    // Dispatch async to mimic network event microtask queues
    setTimeout(() => {
      this.listeners.forEach(listener => {
        try {
          listener(event);
        } catch (err) {
          console.error("SSE Listener error:", err);
        }
      });
    }, 0);
  }
}

export const sseBus = new SSEEventBus();
```

---

### 3.2 AI Process Pipeline Simulator (`web/src/sse/simulator.ts`)
When a user creates a post or replies to an active topic, this simulation module intercepts the transaction and triggers a timed execution pipeline. The pipeline runs the AI Agent Decision engine, logs their rationale, creates asynchronous background worker tasks, and writes reply comments in the DB, while triggering real-time SSE broadcasts.

```typescript
import { MockDB } from '../api/mockDb';
import { Post, Comment, AIReplyTask, AIDecisionLog, AIAgent } from '../api/types';
import { sseBus } from './sseBus';

// Generates an AI response based on the post context and agent characteristics
function generateFakeAIContent(agent: AIAgent, title: string, content: string): string {
  const templates = [
    `关于你的议题 **"${title}"**，我深入分析了其架构痛点。在${agent.speakingStyle}的影响下，我们需要考虑：\n\n1. **${agent.valueOrientation}** 的第一性原理：当前设计可能在可用性和扩展性上面临考验。\n2. 推荐采取的工程范式是根据其底层状态机进行优化，减少长链接与进程锁定。\n\n针对你提到的 “${content.substring(0, 40)}...” 这个问题，我的具体建议是：直接精简调用栈，将状态同步从同步写入解耦为持久化的事件流。`,
    `我有一些不同看法。从计算机科学经典理论来看，这属于分布式一致性边界控制的问题。按照 **${agent.systemPrompt.substring(0, 30)}** 的规范要求：\n\n- 时间复杂度分析：目前在最坏情况下退化到了 O(N^2) 的资源开销。\n- 推荐策略：采用两阶段退避重试或直接切换到支持 CAS (Compare-And-Swap) 的乐观锁机制。\n\n${agent.stylePrompt}`,
    `技术在2026年的最佳实践通常倡导：极简化开发 (KISS)。面对这个问题，我的建议是：\n\n\`\`\`go\n// 推荐的简化模式示例\npackage main\n\nfunc ProcessOutbox() {\n    // 核心重试策略：解耦事务提交，使用 Binlog 或是轻量轮询\n    println("保持底层实现简单可靠。")\n}\n\`\`\`\n\n避免引入过度复杂的微服务网关，那只是增加运维复杂度而已。`
  ];
  return templates[Math.floor(Math.random() * templates.length)];
}

export function triggerAIPipeline(postId: number, parentCommentId: number | null = null, targetCommentId: number | null = null) {
  const post = MockDB.getPostById(postId);
  if (!post) return;

  // 1. Trigger ai_tagging_started
  post.aiStatus = "PENDING";
  MockDB.savePost(post);
  sseBus.publish({ event: "ai_tagging_started", postId });

  // 2. Perform Tag Analysis (Delay: 1.5 seconds)
  setTimeout(() => {
    const updatedPost = MockDB.getPostById(postId);
    if (!updatedPost) return;

    // Simulate appending a custom tag if missing
    if (!updatedPost.tags.includes("AI助理")) {
      updatedPost.tags = [...updatedPost.tags, "AI助理"];
    }
    MockDB.savePost(updatedPost);
    sseBus.publish({ event: "ai_tagging_completed", postId, tags: updatedPost.tags });

    // 3. AI Agent Dispatch & Decisioning (Delay: 1 second)
    setTimeout(() => {
      const activeAgents = MockDB.getAgents().filter(a => a.active);
      const chosenTasks: { agent: AIAgent; task: AIReplyTask }[] = [];
      const decisionLogs: AIDecisionLog[] = [];

      let hasReplied = false;

      for (const agent of activeAgents) {
        // Simple willingness formula: base activityLevel modified by randomness
        const willingnessScore = Math.min(1.0, Math.max(0.0, agent.activityLevel * 0.7 + Math.random() * 0.4));
        const threshold = agent.replyThreshold;
        const willReply = willingnessScore >= threshold && (!agent.isFallback || !hasReplied);

        if (willReply) {
          hasReplied = true;
          const decision: "REPLY" = "REPLY";
          
          // Write Decision Log
          const log: AIDecisionLog = {
            id: Date.now() + Math.floor(Math.random() * 1000),
            postId,
            commentId: parentCommentId,
            aiAgentId: agent.id,
            aiAgentName: agent.name,
            triggerType: parentCommentId ? "FOLLOWUP" : "POST_AUTO",
            willingnessScore,
            thresholdValue: threshold,
            decision,
            reason: `Willingness score (${willingnessScore.toFixed(2)}) exceeded threshold (${threshold.toFixed(2)}). Agent personality triggered.`,
            createdAt: new Date().toISOString()
          };
          MockDB.saveDecisionLog(log);
          decisionLogs.push(log);

          // Create AIReplyTask
          const task: AIReplyTask = {
            id: Date.now() + Math.floor(Math.random() * 10000),
            postId,
            parentCommentId,
            targetCommentId,
            aiAgentId: agent.id,
            triggerType: parentCommentId ? "FOLLOWUP" : "POST_AUTO",
            status: "PENDING",
            prompt: `Context Title: ${updatedPost.title}. Trigger content snapshot: ${updatedPost.content.substring(0, 100)}`,
            result: "",
            errorMessage: "",
            retryCount: 0,
            createdAt: new Date().toISOString(),
            startedAt: null,
            finishedAt: null
          };
          MockDB.saveTask(task);
          chosenTasks.push({ agent, task });
        } else {
          // Log Ignored Decision
          const log: AIDecisionLog = {
            id: Date.now() + Math.floor(Math.random() * 1000),
            postId,
            commentId: parentCommentId,
            aiAgentId: agent.id,
            aiAgentName: agent.name,
            triggerType: parentCommentId ? "FOLLOWUP" : "POST_AUTO",
            willingnessScore,
            thresholdValue: threshold,
            decision: "IGNORE",
            reason: `Score (${willingnessScore.toFixed(2)}) is below threshold (${threshold.toFixed(2)}). Ignored.`,
            createdAt: new Date().toISOString()
          };
          MockDB.saveDecisionLog(log);
          decisionLogs.push(log);
        }
      }

      // If no agent triggered, run Fallback Agent
      if (!hasReplied) {
        const fallbackAgent = activeAgents.find(a => a.isFallback);
        if (fallbackAgent) {
          const decision: "REPLY" = "REPLY";
          const log: AIDecisionLog = {
            id: Date.now() + Math.floor(Math.random() * 1000),
            postId,
            commentId: parentCommentId,
            aiAgentId: fallbackAgent.id,
            aiAgentName: fallbackAgent.name,
            triggerType: parentCommentId ? "FOLLOWUP" : "POST_AUTO",
            willingnessScore: 1.0,
            thresholdValue: 0.0,
            decision,
            reason: "No active agents triggered. Fallback Agent activated.",
            createdAt: new Date().toISOString()
          };
          MockDB.saveDecisionLog(log);
          decisionLogs.push(log);

          const task: AIReplyTask = {
            id: Date.now() + Math.floor(Math.random() * 10000),
            postId,
            parentCommentId,
            targetCommentId,
            aiAgentId: fallbackAgent.id,
            triggerType: parentCommentId ? "FOLLOWUP" : "POST_AUTO",
            status: "PENDING",
            prompt: `Fallback prompt for: ${updatedPost.title}`,
            result: "",
            errorMessage: "",
            retryCount: 0,
            createdAt: new Date().toISOString(),
            startedAt: null,
            finishedAt: null
          };
          MockDB.saveTask(task);
          chosenTasks.push({ agent: fallbackAgent, task });
        }
      }

      // Publish decision completed
      sseBus.publish({ event: "ai_decision_completed", postId });

      // 4. Run Chosen Tasks sequentially with simulated delay (2-4 seconds per agent)
      executeTasksSequentially(postId, chosenTasks, updatedPost);

    }, 1000);
  }, 1500);
}

function executeTasksSequentially(
  postId: number, 
  tasks: { agent: AIAgent; task: AIReplyTask }[], 
  post: Post
) {
  if (tasks.length === 0) {
    post.aiStatus = "COMPLETED";
    MockDB.savePost(post);
    return;
  }

  post.aiStatus = "PROCESSING";
  MockDB.savePost(post);

  const current = tasks[0];
  const remaining = tasks.slice(1);

  // Mark task processing
  current.task.status = "PROCESSING";
  current.task.startedAt = new Date().toISOString();
  MockDB.saveTask(current.task);

  sseBus.publish({
    event: "ai_reply_started",
    postId,
    taskId: current.task.id,
    aiAgentId: current.agent.id,
    aiAgentName: current.agent.name
  });

  // Latency delay (2.5 seconds) simulating model calculation
  setTimeout(() => {
    try {
      const generatedText = generateFakeAIContent(current.agent, post.title, post.content);

      // Create comment
      const newComment: Comment = {
        id: Date.now() + Math.floor(Math.random() * 1000),
        postId,
        parentId: current.task.parentCommentId,
        content: generatedText,
        author: {
          username: current.agent.name,
          avatar: current.agent.avatar,
          isAi: true,
          aiAgentId: current.agent.id
        },
        createdAt: new Date().toISOString()
      };
      MockDB.saveComment(newComment);

      // Update Task status
      current.task.status = "COMPLETED";
      current.task.result = generatedText;
      current.task.finishedAt = new Date().toISOString();
      MockDB.saveTask(current.task);

      // Update Post counts and avatar stamps
      const updatedPost = MockDB.getPostById(postId);
      if (updatedPost) {
        updatedPost.aiResponsesCount += 1;
        if (!updatedPost.aiAvatars.includes(current.agent.avatar)) {
          updatedPost.aiAvatars = [...updatedPost.aiAvatars, current.agent.avatar];
        }
        MockDB.savePost(updatedPost);
      }

      // Publish reply success
      sseBus.publish({ event: "comment_created", comment: newComment });
      sseBus.publish({
        event: "ai_reply_completed",
        postId,
        commentId: newComment.id,
        taskId: current.task.id,
        aiAgentId: current.agent.id,
        aiAgentName: current.agent.name
      });

    } catch (err: any) {
      current.task.status = "FAILED";
      current.task.errorMessage = err.message || "Simulated generation error";
      current.task.finishedAt = new Date().toISOString();
      MockDB.saveTask(current.task);

      sseBus.publish({
        event: "ai_reply_failed",
        postId,
        taskId: current.task.id,
        aiAgentId: current.agent.id,
        error: current.task.errorMessage
      });
    }

    // Recursively handle the rest
    setTimeout(() => {
      executeTasksSequentially(postId, remaining, post);
    }, 500);

  }, 2500);
}
```

---

### 3.3 React SSE Connection Hook (`web/src/sse/useSSE.ts`)
Standard custom hook that components mount to listen for update payloads. To satisfy robust reliability requirements, if the simulated connection fails or is disconnected, it immediately falls back to a background REST polling cycle (interval: 2 seconds) fetching the post status.

```typescript
import { useEffect, useState, useRef } from 'react';
import { sseBus, SSEEvent } from './sseBus';
import { MockDB } from '../api/mockDb';
import { Post } from '../api/types';

export function useSSE(postId: number, onEvent?: (event: SSEEvent) => void) {
  const [isConnected, setIsConnected] = useState(false);
  const [aiStatus, setAiStatus] = useState<"PENDING" | "PROCESSING" | "COMPLETED">("COMPLETED");
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  // Track polling fallback interval
  const pollTimerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // Initial fetch of current status
    const currentPost = MockDB.getPostById(postId);
    if (currentPost) {
      setAiStatus(currentPost.aiStatus);
    }

    // Simulate connecting to the SSE stream GET /api/posts/{postId}/events
    setIsConnected(true);
    console.log(`[SSE] Connected to /api/posts/${postId}/events`);

    const unsubscribe = sseBus.subscribe((event) => {
      // Direct message routing logic
      if (
        event.event === "ai_tagging_started" ||
        event.event === "ai_tagging_completed" ||
        event.event === "ai_decision_completed" ||
        event.event === "ai_reply_started" ||
        event.event === "ai_reply_completed" ||
        event.event === "ai_reply_failed" ||
        event.event === "comment_created"
      ) {
        // Filter event based on postId matching
        const payloadPostId = (event as any).postId || ((event as any).comment && (event as any).comment.postId);
        
        if (payloadPostId === postId) {
          // Adjust state locally
          const updatedPost = MockDB.getPostById(postId);
          if (updatedPost) {
            setAiStatus(updatedPost.aiStatus);
          }

          // Trigger listener callback
          if (onEventRef.current) {
            onEventRef.current(event);
          }
        }
      }
    });

    return () => {
      unsubscribe();
      setIsConnected(false);
      console.log(`[SSE] Disconnected from /api/posts/${postId}/events`);
      if (pollTimerRef.current) {
        clearInterval(pollTimerRef.current);
      }
    };
  }, [postId]);

  // Fallback Polling Trigger: triggers if simulated connectivity breaks
  const forceTriggerPollingFallback = () => {
    if (pollTimerRef.current) return;
    
    console.warn(`[SSE Fallback] Triggering 2-second REST polling for post ${postId}`);
    pollTimerRef.current = setInterval(() => {
      const updatedPost = MockDB.getPostById(postId);
      if (updatedPost) {
        setAiStatus(updatedPost.aiStatus);
        
        if (updatedPost.aiStatus === "COMPLETED") {
          console.log("[SSE Fallback] AI pipeline complete. Stopping polling.");
          if (pollTimerRef.current) {
            clearInterval(pollTimerRef.current);
            pollTimerRef.current = null;
          }
        }
      }
    }, 2000);
  };

  return {
    isConnected,
    aiStatus,
    forceTriggerPollingFallback
  };
}
```

---

## 4. React Query Hooks & Zustand State Store Designs

To keep code modular and aligned with `web/AGENTS.md` communication rules:
- **Server state** uses TanStack Query hooks, calling async database clients that mimic network latency.
- **Client state** uses Zustand stores for auth context and UI overlays.

### 4.1 Mock Async API Query Service (`web/src/hooks/useQueries.ts`)
Integrates the TanStack query engine, mapping operations to MockDB wrappers with simulated latency of 300ms.
```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { MockDB } from '../api/mockDb';
import { Post, Comment, AIAgent } from '../api/types';
import { triggerAIPipeline } from '../sse/simulator';

// Helper to simulate REST network latency
const delay = <T>(result: T, ms = 300): Promise<T> => {
  return new Promise(resolve => setTimeout(() => resolve(result), ms));
};

export const QUERY_KEYS = {
  POSTS: ['posts'] as const,
  POST: (id: number) => ['posts', id] as const,
  COMMENTS: (postId: number) => ['comments', postId] as const,
  AGENTS: ['agents'] as const,
  TASKS: ['tasks'] as const,
  DECISION_LOGS: ['decision_logs'] as const,
};

// --- Post Hooks ---
export function usePosts(category?: string) {
  return useQuery({
    queryKey: category ? [...QUERY_KEYS.POSTS, category] : QUERY_KEYS.POSTS,
    queryFn: () => {
      let posts = MockDB.getPosts();
      if (category) {
        posts = posts.filter(p => p.category === category);
      }
      // Sort newest first
      posts = [...posts].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
      return delay(posts);
    }
  });
}

export function usePost(id: number) {
  return useQuery({
    queryKey: QUERY_KEYS.POST(id),
    queryFn: () => {
      const post = MockDB.getPostById(id);
      if (!post) throw new Error("Post not found");
      return delay(post);
    }
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (newPost: Omit<Post, 'id' | 'aiStatus' | 'aiResponsesCount' | 'aiAvatars' | 'createdAt'>) => {
      const created: Post = {
        ...newPost,
        id: Date.now(),
        aiStatus: "PENDING",
        aiResponsesCount: 0,
        aiAvatars: [],
        createdAt: new Date().toISOString(),
      };
      MockDB.savePost(created);
      return delay(created);
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.POSTS });
      // Trigger the background AI Pipeline simulator immediately
      triggerAIPipeline(data.id);
    }
  });
}

// --- Comment Hooks ---
export function useComments(postId: number) {
  return useQuery({
    queryKey: QUERY_KEYS.COMMENTS(postId),
    queryFn: () => {
      const comments = MockDB.getCommentsByPostId(postId);
      // Sort chronologically
      const sorted = [...comments].sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
      return delay(sorted);
    }
  });
}

export function useCreateComment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (newComment: Omit<Comment, 'id' | 'createdAt'>) => {
      const comment: Comment = {
        ...newComment,
        id: Date.now(),
        createdAt: new Date().toISOString()
      };
      MockDB.saveComment(comment);
      return delay(comment);
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.COMMENTS(data.postId) });
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.POST(data.postId) });
      
      // If the comment triggers mentions or is a followup, simulate AI checks
      const mentionsAI = data.content.includes("@");
      if (mentionsAI) {
        triggerAIPipeline(data.postId, data.id);
      }
    }
  });
}

// --- Agent Hooks ---
export function useAgents() {
  return useQuery({
    queryKey: QUERY_KEYS.AGENTS,
    queryFn: () => {
      const agents = MockDB.getAgents();
      return delay(agents);
    }
  });
}

export function useUpdateAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (updated: AIAgent) => {
      MockDB.saveAgent(updated);
      return delay(updated);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.AGENTS });
    }
  });
}

// --- Admin-specific Tasks & Logs Hooks ---
export function useTasks() {
  return useQuery({
    queryKey: QUERY_KEYS.TASKS,
    queryFn: () => {
      const tasks = MockDB.getTasks();
      // Sort newest first
      const sorted = [...tasks].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
      return delay(sorted);
    }
  });
}

export function useDecisionLogs() {
  return useQuery({
    queryKey: QUERY_KEYS.DECISION_LOGS,
    queryFn: () => {
      const logs = MockDB.getDecisionLogs();
      const sorted = [...logs].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
      return delay(sorted);
    }
  });
}
```

---

### 4.2 Zustand Client-State Stores

#### Auth Store (`web/src/stores/authStore.ts`)
Manages the user session details entirely in the browser.
```typescript
import { create } from 'zustand';
import { User } from '../api/types';

interface AuthState {
  currentUser: User | null;
  login: (username: string) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  currentUser: (() => {
    const saved = localStorage.getItem('ai_forum_current_user');
    return saved ? JSON.parse(saved) : {
      username: 'DeveloperGuy',
      avatar: 'https://api.dicebear.com/7.x/adventurer/svg?seed=DeveloperGuy'
    };
  })(),
  login: (username: string) => {
    const user: User = {
      username,
      avatar: `https://api.dicebear.com/7.x/adventurer/svg?seed=${username}`
    };
    localStorage.setItem('ai_forum_current_user', JSON.stringify(user));
    set({ currentUser: user });
  },
  logout: () => {
    localStorage.removeItem('ai_forum_current_user');
    set({ currentUser: null });
  },
  isAuthenticated: () => !!get().currentUser,
}));
```

#### UI Store (`web/src/stores/uiStore.ts`)
Handles application-wide lightweight states like sidebars, categories, alert queues, and themes.
```typescript
import { create } from 'zustand';

interface UIState {
  sidebarOpen: boolean;
  selectedCategory: string | null;
  themeMode: 'light' | 'stone' | 'dark';
  setSidebarOpen: (open: boolean) => void;
  setSelectedCategory: (category: string | null) => void;
  setThemeMode: (mode: 'light' | 'stone' | 'dark') => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  selectedCategory: null,
  themeMode: 'light',
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setSelectedCategory: (category) => set({ selectedCategory: category }),
  setThemeMode: (mode) => set({ themeMode: mode }),
}));
```

---

## 5. Verification & Test Plan

To verify this implementation plan:
1. **Initialize Sandbox Validation**: Run a validation check on Vite startup using configuration files in a clean test structure.
2. **Unit Testing queries and SSE hooks**:
   - Write a unit test using `@testing-library/react-hooks` or a lightweight test harness, simulating the unmounting of `useSSE` to ensure the event listener unregisters correctly.
   - Assert the status pipeline of a created post transitions from `PENDING` -> `PROCESSING` -> `COMPLETED` when the AI simulator ticks.
3. **Database Concurrency Check**: Open two browsers/iframes accessing the same local storage to confirm additions synchronize correctly.

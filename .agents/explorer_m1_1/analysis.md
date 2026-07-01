# Analysis & Implementation Strategy: Milestone 1 (Web App Init & Mock Layer)

This document outlines the detailed architecture and implementation strategy for the User Web Application (`web/`) workspace setup, mock database layer, simulated Server-Sent Events (SSE) system, and query/state management hooks.

---

## 1. Workspace Initialization Configuration

The workspace runs on Vite + TypeScript + React + Tailwind CSS. The following files establish the structural foundation under `web/`.

### 1.1 `web/package.json`
Specifies standard, lightweight, stable versions for React 18, TanStack Query v5, Zustand, React Virtuoso, and Markdown parsing.

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
    "@tanstack/react-query": "^5.56.2",
    "dompurify": "^3.1.6",
    "lucide-react": "^0.439.0",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-markdown": "^9.0.1",
    "react-router-dom": "^6.26.2",
    "react-virtuoso": "^4.7.11",
    "zustand": "^4.5.5"
  },
  "devDependencies": {
    "@types/dompurify": "^3.0.5",
    "@types/react": "^18.3.5",
    "@types/react-dom": "^18.3.0",
    "@vitejs/plugin-react": "^4.3.1",
    "autoprefixer": "^10.4.20",
    "postcss": "^8.4.45",
    "tailwindcss": "^3.4.10",
    "typescript": "^5.5.3",
    "vite": "^5.4.2"
  }
}
```

### 1.2 `web/vite.config.ts`
Enables the `@vitejs/plugin-react` compiler, defines port `3000` for the user application, and configures path aliasing (`@/*` pointing to `web/src/*`).

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
Maps CSS variables defined in the stylesheet to Tailwind utility classes. This supports custom Cohere spacing rules, typography choices, and border-radius tokens.

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
        cohere: {
          primary: "var(--cohere-primary)",
          black: "var(--cohere-black)",
          ink: "var(--cohere-ink)",
          "deep-green": "var(--cohere-deep-green)",
          "dark-navy": "var(--cohere-dark-navy)",
          canvas: "var(--cohere-canvas)",
          "soft-stone": "var(--cohere-soft-stone)",
          "pale-green": "var(--cohere-pale-green)",
          "pale-blue": "var(--cohere-pale-blue)",
          hairline: "var(--cohere-hairline)",
          "border-light": "var(--cohere-border-light)",
          "card-border": "var(--cohere-card-border)",
          muted: "var(--cohere-muted)",
          slate: "var(--cohere-slate)",
          "body-muted": "var(--cohere-body-muted)",
          "action-blue": "var(--cohere-action-blue)",
          "focus-blue": "var(--cohere-focus-blue)",
          coral: "var(--cohere-coral)",
          "coral-soft": "var(--cohere-coral-soft)",
          "form-focus": "var(--cohere-form-focus)",
          error: "var(--cohere-error)",
          success: "var(--cohere-success)",
        }
      },
      fontFamily: {
        display: ["CohereText", "Space Grotesk", "Inter", "sans-serif"],
        sans: ["Unica77 Cohere Web", "Inter", "Arial", "sans-serif"],
        mono: ["CohereMono", "Courier New", "monospace"],
      },
      borderRadius: {
        xs: "4px",
        sm: "8px",
        md: "16px",
        lg: "22px",
        xl: "30px",
        pill: "32px",
      },
      spacing: {
        xxs: "2px",
        xs: "6px",
        sm: "8px",
        md: "12px",
        lg: "16px",
        xl: "24px",
        xxl: "32px",
        section: "80px",
      }
    },
  },
  plugins: [],
}
```

### 1.4 `web/src/styles/index.css`
Declares the CSS custom properties and configures core Tailwind layer components to enforce Cohere's editorial design system.

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --cohere-primary: #17171c;
    --cohere-black: #000000;
    --cohere-ink: #212121;
    --cohere-deep-green: #003c33;
    --cohere-dark-navy: #071829;
    --cohere-canvas: #ffffff;
    --cohere-soft-stone: #eeece7;
    --cohere-pale-green: #edfce9;
    --cohere-pale-blue: #f1f5ff;
    --cohere-hairline: #d9d9dd;
    --cohere-border-light: #e5e7eb;
    --cohere-card-border: #f2f2f2;
    --cohere-muted: #93939f;
    --cohere-slate: #75758a;
    --cohere-body-muted: #616161;
    --cohere-action-blue: #1863dc;
    --cohere-focus-blue: #4c6ee6;
    --cohere-coral: #ff7759;
    --cohere-coral-soft: #ffad9b;
    --cohere-form-focus: #9b60aa;
    --cohere-error: #b30000;
    --cohere-success: #edfce9;
  }

  body {
    @apply bg-cohere-canvas text-cohere-ink font-sans antialiased;
    font-size: 16px;
    line-height: 1.5;
  }
}

@layer components {
  /* Typography Classes based on design_cohere.md */
  .font-hero-display {
    @apply font-display text-[96px] font-normal leading-[1] -tracking-[1.92px];
  }
  
  .font-product-display {
    @apply font-display text-[72px] font-normal leading-[1] -tracking-[1.44px];
  }
  
  .font-section-display {
    @apply font-sans text-[60px] font-normal leading-[1] -tracking-[1.2px];
  }
  
  .font-section-heading {
    @apply font-sans text-[48px] font-normal leading-[1.2] -tracking-[0.48px];
  }
  
  .font-card-heading {
    @apply font-sans text-[32px] font-normal leading-[1.2] -tracking-[0.32px];
  }
  
  .font-feature-heading {
    @apply font-sans text-[24px] font-normal leading-[1.3] tracking-normal;
  }
  
  .font-body-large {
    @apply font-sans text-[18px] font-normal leading-[1.4] tracking-normal;
  }
  
  .font-body {
    @apply font-sans text-[16px] font-normal leading-[1.5] tracking-normal;
  }
  
  .font-button {
    @apply font-sans text-[14px] font-medium leading-[1.71] tracking-normal;
  }
  
  .font-caption {
    @apply font-sans text-[14px] font-normal leading-[1.4] tracking-normal;
  }
  
  .font-mono-label {
    @apply font-mono text-[14px] font-normal leading-[1.4] tracking-[0.28px] uppercase;
  }
  
  .font-micro {
    @apply font-sans text-[12px] font-normal leading-[1.4] tracking-normal;
  }

  /* Cohere Signature Component Classes */
  .btn-primary {
    @apply font-button bg-cohere-primary text-white rounded-pill px-6 py-3 transition-colors hover:bg-cohere-black focus:outline-none focus:ring-2 focus:ring-cohere-focus-blue;
  }
  
  .btn-secondary {
    @apply font-body text-cohere-ink underline bg-transparent py-2 border-none transition-opacity hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-cohere-focus-blue;
  }
  
  .btn-pill-outline {
    @apply font-button bg-transparent text-cohere-primary border border-cohere-primary rounded-xl px-3 py-1.5 transition-colors hover:bg-cohere-soft-stone focus:outline-none focus:ring-2 focus:ring-cohere-focus-blue;
  }
  
  .announcement-bar {
    @apply font-micro bg-cohere-black text-white h-[36px] flex items-center justify-center px-4 relative;
  }
  
  .card-base {
    @apply bg-cohere-canvas border border-cohere-hairline rounded-md p-6 transition-shadow hover:shadow-sm;
  }

  .agent-console-card {
    @apply bg-cohere-primary text-white rounded-sm p-6;
  }

  .product-card {
    @apply bg-cohere-soft-stone text-cohere-ink rounded-sm p-8 border border-cohere-card-border;
  }

  .blog-filter-chip {
    @apply font-card-heading text-cohere-coral border border-transparent rounded-sm px-3.5 py-2 transition-colors hover:bg-cohere-soft-stone;
  }
  
  .blog-filter-chip.active {
    @apply bg-cohere-coral text-white border-transparent;
  }
}
```

---

## 2. Frontend Mock Data Layer

A fully-typed in-memory/localStorage database acts as the single source of truth for both `web/` and `admin/` clients.

### 2.1 Schema Definition (`web/src/api/types.ts`)

```typescript
export interface Post {
  id: number;
  title: string;
  content: string;
  category: string;
  tags: string[];
  author: {
    username: string;
    avatar: string;
  };
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
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  willingnessScore: number;
  thresholdValue: number;
  decision: "REPLY" | "IGNORE" | "FAILED";
  reason: string;
  createdAt: string;
}

export interface DatabaseState {
  posts: Post[];
  comments: Comment[];
  agents: AIAgent[];
  tasks: AIReplyTask[];
  decisionLogs: AIDecisionLog[];
}
```

### 2.2 Seed Data (`web/src/api/db.ts` - Seed Section)
Initializes state with a set of diverse agents, sample posts, and initial decision/comment structures.

```typescript
import { DatabaseState, AIAgent } from "./types";

export const DEFAULT_AGENTS: AIAgent[] = [
  {
    id: 1,
    name: "ArchTechLead",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=ArchTechLead",
    description: "Experienced Backend Systems Architect specializing in performance, distributed Go structures, and security.",
    ageViewpoint: "Advocates for simplicity, strict decoupled service design, and exhaustive integration testing.",
    personality: "Pragmatic, rigorous, analytical, but constructive.",
    valueOrientation: "Stability, low technical debt, defensive coding.",
    speakingStyle: "Direct, professional, using precise technical terminology (e.g., decoupled boundaries, high-cohesion).",
    systemPrompt: "You are ArchTechLead. Analyze technical design and point out potential bugs, scale issues, and layout compliance errors.",
    stylePrompt: "Start with a direct structural summary. Use markdown tables and lists to organize critique. Do not use corporate speak.",
    replyThreshold: 0.60,
    activityLevel: 0.80,
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
    name: "GrowthProductManager",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=PM",
    description: "Product Manager focused on usability, clean user acquisition loops, and telemetry-driven design decisions.",
    ageViewpoint: "Prioritizes user experience, developer-friendly layouts, and short time-to-first-interaction metrics.",
    personality: "Encouraging, business-oriented, communicative.",
    valueOrientation: "User retention, quick visual validation, clarity over pure engine efficiency.",
    speakingStyle: "Conversational, enthusiastic, metric-driven, frequently referencing KPIs and user loops.",
    systemPrompt: "You are GrowthProductManager. Assess product designs and UI usability issues, ensuring alignment with user acquisition loops.",
    stylePrompt: "Structure with warm encouragement. Detail UX gaps using bullet points and ask follow-up questions.",
    replyThreshold: 0.50,
    activityLevel: 0.70,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 1,
    maxFollowupRepliesPerPost: 1,
    isFallback: false,
    active: true
  },
  {
    id: 3,
    name: "DevilsAdvocate",
    avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=Devil",
    description: "Skeptical senior developer who challenges trending tech, microservice complexity, and premature optimization.",
    ageViewpoint: "Argues that most systems are over-engineered and should start as a single monolith with minimal dependencies.",
    personality: "Critical, sarcastic, challenging, yet highly knowledgeable.",
    valueOrientation: "Extreme cost-efficiency, sanity checks, minimizing complexity.",
    speakingStyle: "Ironical, provocative, questioning assumptions, using phrases like 'Do we really need X?'",
    systemPrompt: "You are DevilsAdvocate. Critique codebases by identifying unnecessary engineering, over-architected systems, and extra layers.",
    stylePrompt: "Pose direct skeptical questions. Avoid consensus statements. Keep it sharp and provocative.",
    replyThreshold: 0.45,
    activityLevel: 0.85,
    allowAutoReply: true,
    allowMentionReply: true,
    allowFollowupReply: true,
    maxAutoRepliesPerPost: 2,
    maxFollowupRepliesPerPost: 2,
    isFallback: false,
    active: true
  }
];

export const INITIAL_DB_STATE: DatabaseState = {
  posts: [
    {
      id: 1,
      title: "Is it time to rewrite our Go monolithic API in Rust?",
      content: "Our Go API server handles around 15k RPS. CPU usage stays at 40%, but GC pauses occasionally spike to 8ms. Would translating the hot paths to Rust or rewrite the service entirely solve this without introducing extreme complexity?",
      category: "后端开发",
      tags: ["Go", "Rust", "Performance"],
      author: {
        username: "alex_dev",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=alex"
      },
      aiStatus: "COMPLETED",
      aiResponsesCount: 2,
      aiAvatars: [
        "https://api.dicebear.com/7.x/bottts/svg?seed=ArchTechLead",
        "https://api.dicebear.com/7.x/bottts/svg?seed=Devil"
      ],
      createdAt: new Date(Date.now() - 3600000 * 4).toISOString()
    }
  ],
  comments: [
    {
      id: 1,
      postId: 1,
      parentId: null,
      content: "Before writing any Rust, did you profile memory allocations? A GC pause of 8ms suggests large heaps or excessive short-lived objects. Run `pprof` first.",
      author: {
        username: "senior_gopher",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=gopher",
        isAi: false
      },
      createdAt: new Date(Date.now() - 3600000 * 3.5).toISOString()
    },
    {
      id: 2,
      postId: 1,
      parentId: 1,
      content: "Exactly what senior_gopher said. Rewriting 15k RPS monolithic Go code into Rust just to save 8ms is classic premature optimization. You'll inflate maintenance overhead by 3x.",
      author: {
        username: "DevilsAdvocate",
        avatar: "https://api.dicebear.com/7.x/bottts/svg?seed=Devil",
        isAi: true,
        aiAgentId: 3
      },
      createdAt: new Date(Date.now() - 3600000 * 3).toISOString()
    }
  ],
  agents: DEFAULT_AGENTS,
  tasks: [],
  decisionLogs: [
    {
      id: 1,
      postId: 1,
      commentId: 1,
      aiAgentId: 3,
      aiAgentName: "DevilsAdvocate",
      triggerType: "FOLLOWUP",
      willingnessScore: 0.89,
      thresholdValue: 0.45,
      decision: "REPLY",
      reason: "High interest in contesting a premature Rust rewrite of Go code.",
      createdAt: new Date(Date.now() - 3600000 * 3.01).toISOString()
    }
  ]
};
```

### 2.3 Local Storage Database Engine (`web/src/api/db.ts`)
Implements the active read-write mechanism using `localStorage` to ensure operations trigger notifications and save persistent modifications.

```typescript
import { DatabaseState, Post, Comment, AIAgent, AIReplyTask, AIDecisionLog } from "./types";
import { INITIAL_DB_STATE } from "./db";

const DB_KEY = "ai_forum_db_state";

export class MockDatabase {
  private state: DatabaseState;

  constructor() {
    const saved = localStorage.getItem(DB_KEY);
    if (saved) {
      try {
        this.state = JSON.parse(saved);
      } catch {
        this.state = INITIAL_DB_STATE;
        this.save();
      }
    } else {
      this.state = INITIAL_DB_STATE;
      this.save();
    }
  }

  private save() {
    localStorage.setItem(DB_KEY, JSON.stringify(this.state));
  }

  // --- Post Operations ---
  public getPosts(): Post[] {
    return this.state.posts;
  }

  public getPost(id: number): Post | undefined {
    return this.state.posts.find(p => p.id === id);
  }

  public createPost(post: Omit<Post, "id" | "aiStatus" | "aiResponsesCount" | "aiAvatars" | "createdAt">): Post {
    const newPost: Post = {
      ...post,
      id: this.generateId(this.state.posts),
      aiStatus: "PENDING",
      aiResponsesCount: 0,
      aiAvatars: [],
      createdAt: new Date().toISOString()
    };
    this.state.posts.unshift(newPost); // New posts at the top
    this.save();
    return newPost;
  }

  public updatePost(id: number, updates: Partial<Post>): Post {
    const index = this.state.posts.findIndex(p => p.id === id);
    if (index === -1) throw new Error(`Post ${id} not found`);
    this.state.posts[index] = { ...this.state.posts[index], ...updates };
    this.save();
    return this.state.posts[index];
  }

  // --- Comment Operations ---
  public getComments(postId: number): Comment[] {
    return this.state.comments.filter(c => c.postId === postId);
  }

  public createComment(comment: Omit<Comment, "id" | "createdAt">): Comment {
    const newComment: Comment = {
      ...comment,
      id: this.generateId(this.state.comments),
      createdAt: new Date().toISOString()
    };
    this.state.comments.push(newComment);
    this.save();
    return newComment;
  }

  // --- Agent Operations ---
  public getAgents(): AIAgent[] {
    return this.state.agents;
  }

  public updateAgent(id: number, updates: Partial<AIAgent>): AIAgent {
    const index = this.state.agents.findIndex(a => a.id === id);
    if (index === -1) throw new Error(`Agent ${id} not found`);
    this.state.agents[index] = { ...this.state.agents[index], ...updates };
    this.save();
    return this.state.agents[index];
  }

  // --- Task Operations ---
  public getTasks(): AIReplyTask[] {
    return this.state.tasks;
  }

  public createTask(task: Omit<AIReplyTask, "id" | "createdAt" | "retryCount">): AIReplyTask {
    const newTask: AIReplyTask = {
      ...task,
      id: this.generateId(this.state.tasks),
      retryCount: 0,
      createdAt: new Date().toISOString()
    };
    this.state.tasks.unshift(newTask);
    this.save();
    return newTask;
  }

  public updateTask(id: number, updates: Partial<AIReplyTask>): AIReplyTask {
    const index = this.state.tasks.findIndex(t => t.id === id);
    if (index === -1) throw new Error(`Task ${id} not found`);
    this.state.tasks[index] = { ...this.state.tasks[index], ...updates };
    this.save();
    return this.state.tasks[index];
  }

  // --- Decision Log Operations ---
  public getDecisionLogs(): AIDecisionLog[] {
    return this.state.decisionLogs;
  }

  public createDecisionLog(log: Omit<AIDecisionLog, "id" | "createdAt">): AIDecisionLog {
    const newLog: AIDecisionLog = {
      ...log,
      id: this.generateId(this.state.decisionLogs),
      createdAt: new Date().toISOString()
    };
    this.state.decisionLogs.unshift(newLog);
    this.save();
    return newLog;
  }

  private generateId(arr: { id: number }[]): number {
    return arr.reduce((max, item) => item.id > max ? item.id : max, 0) + 1;
  }
}

export const db = new MockDatabase();
```

---

## 3. Simulated SSE Hooks and Background Execution Engine

To model agent background activity without a backend running WebSockets or SSE channels, we implement a client-side notification and background execution system.

### 3.1 Global Event Emitter (`web/src/sse/emitter.ts`)
Facilitates event broadcasting within the client browser window.

```typescript
type SSECallback = (data: any) => void;

class ClientEventEmitter {
  private listeners: Map<string, Set<SSECallback>> = new Map();

  public subscribe(eventType: string, callback: SSECallback): () => void {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set());
    }
    this.listeners.get(eventType)!.add(callback);

    return () => {
      const set = this.listeners.get(eventType);
      if (set) {
        set.delete(callback);
        if (set.size === 0) this.listeners.delete(eventType);
      }
    };
  }

  public emit(eventType: string, data: any) {
    const set = this.listeners.get(eventType);
    if (set) {
      set.forEach(callback => callback(data));
    }
    // Also emit to wildcards or generic log channel
    const generic = this.listeners.get("*");
    if (generic) {
      generic.forEach(callback => callback({ type: eventType, data }));
    }
  }
}

export const sseEmitter = new ClientEventEmitter();
```

### 3.2 Background Simulation Engine (`web/src/sse/simulator.ts`)
Fires automatically whenever a user publishes a Post or Comment. It simulates agent decision calculations (willingness against threshold), writes to decision logs, queues reply tasks, updates task steps, and eventually creates the automated response comment.

```typescript
import { db } from "../api/db";
import { sseEmitter } from "./emitter";
import { Post, Comment, AIAgent } from "../api/types";

// Generates an LLM-styled response template based on agent personality rules.
function generateAgentText(agentName: string, title: string, category: string): string {
  const time = new Date().toLocaleTimeString();
  switch (agentName) {
    case "ArchTechLead":
      return `### Design Critique: "${title}"
Reviewing the current proposal, here are key architectural observations at ${time}:
1. **Decoupled Boundaries**: If we implement this in "${category}", make sure we establish clean interfaces to prevent high coupling.
2. **Resource Metrics**: Monolithic structures are often easier to instrument. Profile your active database connections before scaling.
3. **Recommendation**: Write integration tests covering state corruption on network loss first.`;

    case "GrowthProductManager":
      return `Wow! The proposal around "${title}" touches some vital user retention paths! 🚀
- Have we defined core metrics for evaluating this change?
- From a UX standpoint, developer interfaces benefit from compact information grids.
- Let's schedule a brief sync to refine the telemetry dashboard!`;

    case "DevilsAdvocate":
      return `Let's pause and do a simple sanity check on this "${title}" proposal.
- **Why are we adding this complexity?** It feels like premature optimization.
- Monolithic setups don't need additional microservice dependencies. What happens if the network layer breaks down?
- Keep it simple, test the basic path, and don't introduce fancy solutions for simple problems.`;

    default:
      return `Interesting post on "${title}". From my perspective as an AI collaborator, maintaining clean boundaries is key.`;
  }
}

export function runBackgroundAISimulation(postId: number, commentId: number | null) {
  const post = db.getPost(postId);
  if (!post) return;

  const targetComment = commentId ? db.getComments(postId).find(c => c.id === commentId) : null;
  const agents = db.getAgents().filter(a => a.active);

  // Mark post status in database
  db.updatePost(postId, { aiStatus: "PROCESSING" });
  sseEmitter.emit("post.updated", db.getPost(postId));

  // Run simulation sequence per agent with offset timeouts
  agents.forEach((agent, index) => {
    setTimeout(() => {
      // 1. Calculate decision
      const triggerType = commentId ? "FOLLOWUP" : "POST_AUTO";
      const randomWillingness = Math.random();
      const decision = randomWillingness >= agent.replyThreshold ? "REPLY" : "IGNORE";
      const reason = decision === "REPLY" 
        ? `Willingness score (${randomWillingness.toFixed(2)}) exceeded threshold (${agent.replyThreshold}).`
        : `Willingness score (${randomWillingness.toFixed(2)}) did not satisfy threshold (${agent.replyThreshold}).`;

      // 2. Log decision
      const decisionLog = db.createDecisionLog({
        postId,
        commentId,
        aiAgentId: agent.id,
        aiAgentName: agent.name,
        triggerType,
        willingnessScore: randomWillingness,
        thresholdValue: agent.replyThreshold,
        decision,
        reason
      });
      sseEmitter.emit("decision_log.created", decisionLog);

      if (decision === "REPLY") {
        // 3. Queue reply task
        const task = db.createTask({
          postId,
          parentCommentId: targetComment ? targetComment.id : null,
          targetCommentId: targetComment ? targetComment.id : null,
          aiAgentId: agent.id,
          triggerType,
          status: "PENDING",
          prompt: `System: ${agent.systemPrompt}\nStyle: ${agent.stylePrompt}\nContext: ${post.title}`,
          result: "",
          errorMessage: "",
          startedAt: null,
          finishedAt: null
        });
        sseEmitter.emit("task.created", task);

        // 4. Transition to PROCESSING (simulating network latency of LLM call)
        setTimeout(() => {
          db.updateTask(task.id, {
            status: "PROCESSING",
            startedAt: new Date().toISOString()
          });
          sseEmitter.emit("task.updated", db.getTasks().find(t => t.id === task.id));

          // 5. Complete Task & Create Comment (simulating text generation latency)
          setTimeout(() => {
            const replyText = generateAgentText(agent.name, post.title, post.category);
            
            // Create comment
            const createdComment = db.createComment({
              postId,
              parentId: targetComment ? targetComment.id : null,
              content: replyText,
              author: {
                username: agent.name,
                avatar: agent.avatar,
                isAi: true,
                aiAgentId: agent.id
              }
            });
            sseEmitter.emit("comment.created", createdComment);

            // Update task to COMPLETED
            db.updateTask(task.id, {
              status: "COMPLETED",
              result: replyText,
              finishedAt: new Date().toISOString()
            });
            sseEmitter.emit("task.updated", db.getTasks().find(t => t.id === task.id));

            // Update post statistics
            const latestPost = db.getPost(postId);
            if (latestPost) {
              const updatedAvatars = Array.from(new Set([...latestPost.aiAvatars, agent.avatar]));
              db.updatePost(postId, {
                aiResponsesCount: latestPost.aiResponsesCount + 1,
                aiAvatars: updatedAvatars
              });
              sseEmitter.emit("post.updated", db.getPost(postId));
            }
          }, 2000);

        }, 1500);
      }
    }, index * 1000); // Stagger agent execution to prevent thread blocking
  });

  // Final check to mark execution finished on the post object
  const totalDuration = (agents.length * 1000) + 4000;
  setTimeout(() => {
    const latestPost = db.getPost(postId);
    if (latestPost) {
      db.updatePost(postId, { aiStatus: "COMPLETED" });
      sseEmitter.emit("post.updated", db.getPost(postId));
    }
  }, totalDuration);
}
```

### 3.3 React hook subscribing to mock SSE (`web/src/sse/useSSE.ts`)
Exposes state and hook subscriptions to real-time database modifications, triggering UI updates smoothly.

```typescript
import { useEffect } from "react";
import { sseEmitter } from "./emitter";

export function useSSE(eventType: string, callback: (data: any) => void) {
  useEffect(() => {
    const unsubscribe = sseEmitter.subscribe(eventType, callback);
    return () => unsubscribe();
  }, [eventType, callback]);
}
```

---

## 4. TanStack Query Hooks & Zustand Store Design

Enforces clean boundaries: client states remain light inside Zustand, while server-authoritative mock API data remains queryable using TanStack Query cache.

### 4.1 Mock API Client Layer (`web/src/api/client.ts`)
Applies a simulated latency offset of `250ms` to mimic network roundtrips.

```typescript
import { db } from "./db";
import { Post, Comment, AIAgent, AIReplyTask, AIDecisionLog } from "./types";
import { runBackgroundAISimulation } from "../sse/simulator";

const delay = <T>(value: T): Promise<T> => {
  return new Promise(resolve => setTimeout(() => resolve(value), 250));
};

export const api = {
  posts: {
    list: async (): Promise<Post[]> => delay(db.getPosts()),
    get: async (id: number): Promise<Post> => {
      const p = db.getPost(id);
      if (!p) throw new Error("Post not found");
      return delay(p);
    },
    create: async (post: Omit<Post, "id" | "aiStatus" | "aiResponsesCount" | "aiAvatars" | "createdAt">): Promise<Post> => {
      const created = db.createPost(post);
      // Trigger async simulation pipeline
      runBackgroundAISimulation(created.id, null);
      return delay(created);
    }
  },
  comments: {
    list: async (postId: number): Promise<Comment[]> => delay(db.getComments(postId)),
    create: async (comment: Omit<Comment, "id" | "createdAt">): Promise<Comment> => {
      const created = db.createComment(comment);
      // Trigger followup reply flow
      runBackgroundAISimulation(created.postId, created.id);
      return delay(created);
    }
  },
  agents: {
    list: async (): Promise<AIAgent[]> => delay(db.getAgents()),
    update: async (id: number, updates: Partial<AIAgent>): Promise<AIAgent> => {
      return delay(db.updateAgent(id, updates));
    }
  },
  tasks: {
    list: async (): Promise<AIReplyTask[]> => delay(db.getTasks())
  },
  decisionLogs: {
    list: async (): Promise<AIDecisionLog[]> => delay(db.getDecisionLogs())
  }
};
```

### 4.2 TanStack Query Hooks (`web/src/hooks/`)
Provides React Query configurations for automated query cache validation, invalidation triggers, and mutation hooks.

#### `web/src/hooks/usePosts.ts`
```typescript
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { useSSE } from "../sse/useSSE";

export function usePosts() {
  const queryClient = useQueryClient();

  // Real-time cache refresh on post creation/updates
  useSSE("post.updated", () => {
    queryClient.invalidateQueries({ queryKey: ["posts"] });
  });

  const postsQuery = useQuery({
    queryKey: ["posts"],
    queryFn: api.posts.list
  });

  const createPostMutation = useMutation({
    mutationFn: api.posts.create,
    onSuccess: (newPost) => {
      // Optimistic cache update or invalidation
      queryClient.invalidateQueries({ queryKey: ["posts"] });
    }
  });

  return {
    posts: postsQuery.data || [],
    isLoading: postsQuery.isLoading,
    createPost: createPostMutation.mutateAsync,
    isCreating: createPostMutation.isPending
  };
}

export function usePostDetail(id: number) {
  const queryClient = useQueryClient();

  useSSE("post.updated", (updatedPost) => {
    if (updatedPost.id === id) {
      queryClient.setQueryData(["post", id], updatedPost);
    }
  });

  return useQuery({
    queryKey: ["post", id],
    queryFn: () => api.posts.get(id),
    enabled: !isNaN(id)
  });
}
```

#### `web/src/hooks/useComments.ts`
```typescript
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { useSSE } from "../sse/useSSE";

export function useComments(postId: number) {
  const queryClient = useQueryClient();

  useSSE("comment.created", (newComment) => {
    if (newComment.postId === postId) {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
    }
  });

  const commentsQuery = useQuery({
    queryKey: ["comments", postId],
    queryFn: () => api.comments.list(postId),
    enabled: !isNaN(postId)
  });

  const createCommentMutation = useMutation({
    mutationFn: api.comments.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
    }
  });

  return {
    comments: commentsQuery.data || [],
    isLoading: commentsQuery.isLoading,
    createComment: createCommentMutation.mutateAsync,
    isSubmitting: createCommentMutation.isPending
  };
}
```

#### `web/src/hooks/useAgents.ts`
```typescript
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";

export function useAgents() {
  const queryClient = useQueryClient();

  const agentsQuery = useQuery({
    queryKey: ["agents"],
    queryFn: api.agents.list
  });

  const updateAgentMutation = useMutation({
    mutationFn: ({ id, updates }: { id: number; updates: Partial<any> }) =>
      api.agents.update(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
    }
  });

  return {
    agents: agentsQuery.data || [],
    isLoading: agentsQuery.isLoading,
    updateAgent: updateAgentMutation.mutateAsync
  };
}
```

### 4.3 Zustand UI Stores (`web/src/stores/`)
Coordinates lightweight, client-only UI configurations: current logged-in identity, navigation filters, and simulated connection states.

#### `web/src/stores/useUserStore.ts`
```typescript
import { create } from "zustand";

interface UserIdentity {
  username: string;
  avatar: string;
}

interface UserStore {
  currentUser: UserIdentity;
  setCurrentUser: (user: UserIdentity) => void;
}

export const useUserStore = create<UserStore>((set) => ({
  currentUser: {
    username: "user_developer_1",
    avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=dev1"
  },
  setCurrentUser: (currentUser) => set({ currentUser })
}));
```

#### `web/src/stores/useFilterStore.ts`
```typescript
import { create } from "zustand";

interface FilterStore {
  selectedCategory: string | null;
  searchQuery: string;
  selectedTags: string[];
  setCategory: (category: string | null) => void;
  setSearchQuery: (query: string) => void;
  toggleTag: (tag: string) => void;
  resetFilters: () => void;
}

export const useFilterStore = create<FilterStore>((set) => ({
  selectedCategory: null,
  searchQuery: "",
  selectedTags: [],
  setCategory: (selectedCategory) => set({ selectedCategory }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),
  toggleTag: (tag) => set((state) => {
    const active = state.selectedTags.includes(tag);
    return {
      selectedTags: active 
        ? state.selectedTags.filter(t => t !== tag)
        : [...state.selectedTags, tag]
    };
  }),
  resetFilters: () => set({ selectedCategory: null, searchQuery: "", selectedTags: [] })
}));
```

#### `web/src/stores/useConnectionStore.ts`
Models the operational state of the mock event stream for testing E2E network-loss behaviors.

```typescript
import { create } from "zustand";

interface ConnectionStore {
  sseStatus: "connected" | "connecting" | "disconnected";
  setSSEStatus: (sseStatus: "connected" | "connecting" | "disconnected") => void;
}

export const useConnectionStore = create<ConnectionStore>((set) => ({
  sseStatus: "connected",
  setSSEStatus: (sseStatus) => set({ sseStatus })
}));
```

---

## 5. Architectural Verification Strategy

To verify this implementation is robust, the implementer must validate the following behaviors:
1. **Mock Storage Consistency**: Ensure data added via the user post feeds is directly queryable inside the console logs view. Both client instances must import from `web/src/api/db.ts` to query the equivalent localStorage key.
2. **Staggered Task Updates**: Verification must capture the state progression `PENDING` -> `PROCESSING` -> `COMPLETED` sequentially, with appropriate latency intervals, in the simulated SSE listener tests.
3. **Safe Markdown Rendering**: Validate that any simulated agent Markdown response (e.g. including headers, inline lists) parses correctly, using DOMPurify filters to wipe scripts or injection tags.

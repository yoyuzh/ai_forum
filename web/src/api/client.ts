import { db } from "./db";
import {
  Post,
  Comment,
  AIAgent,
  AIReplyTask,
  AIDecisionLog,
  AIActivity,
  FeedTab,
  UserProfile,
  UserStats,
  AuthResult,
} from "./types";
import { runBackgroundAISimulation } from "../sse/simulator";

// Simulated network latency so loading states and optimistic UI are exercised.
const delay = <T>(value: T): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(value), 250));

/**
 * All HTTP-shaped calls live here. Swap the bodies for real `fetch` calls
 * against the Go api-server when the backend lands — the function signatures
 * and return shapes are the contract the rest of the app depends on.
 */
export const api = {
  posts: {
    list: async (): Promise<Post[]> => delay(db.getPosts()),

    listByFilter: async (tab: FeedTab, query = "", tag?: string): Promise<Post[]> => {
      const posts = db.getPosts();
      const byTab: Record<FeedTab, () => Post[]> = {
        latest: () => [...posts].sort((a, b) => +new Date(b.createdAt) - +new Date(a.createdAt)),
        hottest: () => [...posts].sort((a, b) => b.viewCount - a.viewCount),
        // 待回复: posts with no AI replies yet.
        unanswered: () => posts.filter((p) => p.aiResponsesCount === 0),
        // AI 参与最多: posts ranked by AI reply count.
        "ai_most": () => [...posts].sort((a, b) => b.aiResponsesCount - a.aiResponsesCount),
      };
      const normalizedQuery = query.trim().toLowerCase();
      const normalizedTag = tag?.trim().toLowerCase();
      const filtered = byTab[tab]().filter((post) => {
        const matchesQuery =
          !normalizedQuery ||
          [post.title, post.content, post.category, post.author.username, ...post.tags]
            .join(" ")
            .toLowerCase()
            .includes(normalizedQuery);
        const matchesTag =
          !normalizedTag || post.tags.some((postTag) => postTag.toLowerCase() === normalizedTag);
        return matchesQuery && matchesTag;
      });
      return delay(filtered);
    },

    get: async (id: number): Promise<Post> => {
      const p = db.getPost(id);
      if (!p) throw new Error("Post not found");
      return delay(p);
    },

    create: async (
      post: Omit<
        Post,
        | "id"
        | "aiStatus"
        | "aiResponsesCount"
        | "aiAvatars"
        | "createdAt"
        | "viewCount"
        | "commentCount"
        | "likeCount"
      >,
    ): Promise<Post> => {
      const created = db.createPost(post);
      // Trigger the async AI simulation pipeline (decision → task → reply).
      runBackgroundAISimulation(created.id, null);
      return delay(created);
    },
  },

  comments: {
    list: async (postId: number): Promise<Comment[]> => delay(db.getComments(postId)),

    create: async (comment: Omit<Comment, "id" | "createdAt" | "likeCount">): Promise<Comment> => {
      const created = db.createComment(comment);
      // Trigger the followup reply flow when a human comments.
      runBackgroundAISimulation(created.postId, created.id);
      return delay(created);
    },
  },

  agents: {
    list: async (): Promise<AIAgent[]> => delay(db.getAgents()),

    get: async (id: number): Promise<AIAgent> => {
      const a = db.getAgent(id);
      if (!a) throw new Error("Agent not found");
      return delay(a);
    },

    update: async (id: number, updates: Partial<AIAgent>): Promise<AIAgent> =>
      delay(db.updateAgent(id, updates)),
  },

  tasks: {
    list: async (): Promise<AIReplyTask[]> => delay(db.getTasks()),
  },

  decisionLogs: {
    list: async (): Promise<AIDecisionLog[]> => delay(db.getDecisionLogs()),

    listForPost: async (postId: number): Promise<AIDecisionLog[]> =>
      delay(db.getDecisionLogsForPost(postId)),
  },

  activities: {
    list: async (): Promise<AIActivity[]> => delay(db.getActivities()),
  },

  // --- Auth & user profile (mock; swap for real fetch calls to api-server later) ---
  auth: {
    /** Mock login: matches identifier + non-empty password against local users. */
    login: async (identifier: string, password: string): Promise<AuthResult> => {
      const trimmedId = identifier.trim();
      if (!trimmedId || !password) {
        throw new Error("请输入账号和密码");
      }
      const user = db.findUserByIdentifier(trimmedId);
      if (!user) {
        throw new Error("账号不存在，请检查或先注册");
      }
      // Mock only — any non-empty password is accepted. Real auth is backend.
      db.setCurrentUser(user.username);
      return delay({ user });
    },

    /** Mock register: creates a local user and marks them as current. */
    register: async (input: {
      username: string;
      nickname: string;
      email: string;
      password: string;
    }): Promise<AuthResult> => {
      const username = input.username.trim();
      const email = input.email.trim();
      if (!username || !email || !input.password) {
        throw new Error("请填写完整的注册信息");
      }
      if (db.findUserByIdentifier(username) || db.findUserByIdentifier(email)) {
        throw new Error("用户名或邮箱已被注册，请直接登录");
      }
      const user = db.createUser({
        username,
        nickname: input.nickname.trim(),
        email,
        password: input.password,
      });
      return delay({ user });
    },

    logout: async (): Promise<void> => {
      db.clearCurrentUser();
      return delay(undefined);
    },
  },

  user: {
    getProfile: async (): Promise<UserProfile> => {
      const user = db.getCurrentUser();
      if (!user) throw new Error("未登录，请先登录");
      return delay(user);
    },

    getStats: async (username: string): Promise<UserStats> =>
      delay(db.getUserStats(username)),

    updateProfile: async (updates: Partial<UserProfile>): Promise<UserProfile> =>
      delay(db.updateUser(updates)),
  },
};

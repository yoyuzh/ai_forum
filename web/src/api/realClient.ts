import { setAuthToken } from "./auth";
import { http } from "./httpClient";
import type {
  AIAgent,
  AIDecisionLog,
  AIActivity,
  AIReplyTask,
  ApiClient,
  Comment,
  FeedTab,
  NotificationItem,
  Post,
  UserProfile,
  UserStats,
} from "./types";

type BackendPost = {
  id: number;
  author_id?: number;
  title: string;
  content: string;
  status?: string;
  category?: string;
  tags?: string[];
  view_count?: number;
  comment_count?: number;
  like_count?: number;
  ai_reply_count?: number;
  created_at?: string;
};

type BackendComment = {
  id: number;
  post_id?: number;
  postId?: number;
  parent_comment_id?: number | null;
  parentId?: number | null;
  content: string;
  comment_type?: string;
  ai_agent_id?: number | null;
  author?: { username?: string; avatar?: string; isAi?: boolean; role?: string };
  created_at?: string;
};

type BackendUser = {
  id?: number;
  username: string;
  display_name?: string;
  role?: string;
  status?: string;
  email?: string;
};

type BackendNotification = {
  id: number;
  type: string;
  payload?: Record<string, unknown>;
  read_at?: string | null;
  readAt?: string | null;
  created_at?: string;
  createdAt?: string;
};

function localAvatar(seed: string): string {
  const text = encodeURIComponent(seed.slice(0, 2) || "AI");
  return `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='32' fill='%23b8ede0'/%3E%3Ctext x='32' y='38' text-anchor='middle' font-size='20' font-family='Arial' fill='%2335675d'%3E${text}%3C/text%3E%3C/svg%3E`;
}

function postFromBackend(p: BackendPost): Post {
  return {
    id: p.id,
    title: p.title,
    content: p.content,
    category: p.category ?? "技术探讨",
    tags: p.tags ?? [],
    author: {
      username: p.author_id ? `user_${p.author_id}` : "backend_user",
      avatar: localAvatar(String(p.author_id ?? p.id)),
      role: "研究员",
    },
    aiStatus: (p.ai_reply_count ?? 0) > 0 ? "COMPLETED" : "PENDING",
    aiResponsesCount: p.ai_reply_count ?? 0,
    aiAvatars: [],
    viewCount: p.view_count ?? 0,
    commentCount: p.comment_count ?? 0,
    likeCount: p.like_count ?? 0,
    createdAt: p.created_at ?? new Date().toISOString(),
  };
}

function commentFromBackend(c: BackendComment, postId: number): Comment {
  const isAi = c.author?.isAi ?? c.comment_type === "AI";
  const username = c.author?.username ?? (isAi ? "ArchTechLead" : "backend_user");
  return {
    id: c.id,
    postId: c.post_id ?? c.postId ?? postId,
    parentId: c.parent_comment_id ?? c.parentId ?? null,
    content: c.content,
    author: {
      username,
      avatar:
        c.author?.avatar ??
        localAvatar(username),
      isAi,
      aiAgentId: c.ai_agent_id ?? undefined,
      role: c.author?.role,
    },
    likeCount: 0,
    createdAt: c.created_at ?? new Date().toISOString(),
  };
}

function userFromBackend(user: BackendUser): UserProfile {
  const username = user.username;
  return {
    username,
    nickname: user.display_name || username,
    email: user.email ?? `${username}@local.invalid`,
    avatar: localAvatar(username),
    bio: "",
    role: user.role ?? "USER",
    uid: String(user.id ?? username),
    joinedAt: new Date().toISOString(),
    emailVerified: true,
    preferences: {
      aiReplyNotifications: true,
      liveActivity: true,
      themePreference: "system",
    },
  };
}

function notificationFromBackend(n: BackendNotification): NotificationItem {
  const title = String(n.payload?.title ?? n.payload?.post_title ?? n.type);
  const body =
    typeof n.payload?.body === "string"
      ? n.payload.body
      : typeof n.payload?.message === "string"
        ? n.payload.message
        : undefined;
  return {
    id: n.id,
    type: n.type,
    title,
    body,
    readAt: n.read_at ?? n.readAt ?? null,
    createdAt: n.created_at ?? n.createdAt ?? new Date().toISOString(),
  };
}

async function listPosts(): Promise<Post[]> {
  const rows = await http<BackendPost[]>("/api/posts");
  return rows.map(postFromBackend);
}

export const realApi: ApiClient = {
  posts: {
    list: listPosts,
    listByFilter: async (tab: FeedTab, query = "", tag?: string) => {
      const normalizedQuery = query.trim().toLowerCase();
      const normalizedTag = tag?.trim().toLowerCase();
      const posts = await listPosts();
      const sorted = [...posts].sort((a, b) => {
        if (tab === "hottest") return b.viewCount - a.viewCount;
        if (tab === "ai_most") return b.aiResponsesCount - a.aiResponsesCount;
        return +new Date(b.createdAt) - +new Date(a.createdAt);
      });
      return sorted.filter((post) => {
        if (tab === "unanswered" && post.aiResponsesCount > 0) return false;
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
    },
    get: async (id: number) => postFromBackend(await http<BackendPost>(`/api/posts/${id}`)),
    create: async (post) =>
      postFromBackend(
        await http<BackendPost>("/api/posts", {
          method: "POST",
          body: JSON.stringify({ title: post.title, content: post.content }),
        }),
      ),
  },

  comments: {
    list: async (postId: number) =>
      (await http<BackendComment[]>(`/api/posts/${postId}/comments`)).map((row) =>
        commentFromBackend(row, postId),
      ),
    create: async (comment) =>
      commentFromBackend(
        await http<BackendComment>(`/api/posts/${comment.postId}/comments`, {
          method: "POST",
          body: JSON.stringify({
            content: comment.content,
            parent_comment_id: comment.parentId,
          }),
        }),
        comment.postId,
      ),
  },

  likes: {
    likePost: (postId: number) => http<void>(`/api/posts/${postId}/like`, { method: "POST" }),
    unlikePost: (postId: number) => http<void>(`/api/posts/${postId}/like`, { method: "DELETE" }),
  },

  favorites: {
    favoritePost: (postId: number) =>
      http<void>(`/api/posts/${postId}/favorite`, { method: "POST" }),
    unfavoritePost: (postId: number) =>
      http<void>(`/api/posts/${postId}/favorite`, { method: "DELETE" }),
  },

  agents: {
    list: async (): Promise<AIAgent[]> => [],
    get: async (id: number): Promise<AIAgent> => {
      throw new Error(`Agent ${id} unavailable in real mode`);
    },
    update: async (): Promise<AIAgent> => {
      throw new Error("Agent update unavailable in web real mode");
    },
  },

  tasks: { list: async (): Promise<AIReplyTask[]> => [] },
  decisionLogs: {
    list: async (): Promise<AIDecisionLog[]> => [],
    listForPost: async (): Promise<AIDecisionLog[]> => [],
  },
  activities: { list: async (): Promise<AIActivity[]> => [] },

  auth: {
    login: async (identifier, password) => {
      const result = await http<{ token: string }>("/api/login", {
        method: "POST",
        body: JSON.stringify({ username: identifier, password }),
      });
      setAuthToken(result.token);
      const user = userFromBackend({ username: identifier });
      return { user, token: result.token };
    },
    register: async (input) => {
      const user = userFromBackend(
        await http<BackendUser>("/api/register", {
          method: "POST",
          body: JSON.stringify({
            username: input.username,
            password: input.password,
            displayName: input.nickname,
          }),
        }),
      );
      return { user };
    },
    logout: async () => setAuthToken(null),
  },

  user: {
    getProfile: async () => userFromBackend(await http<BackendUser>("/api/me")),
    getStats: async (): Promise<UserStats> => ({
      postCount: 0,
      commentCount: 0,
      likeCount: 0,
      aiReplyCount: 0,
    }),
    updateProfile: async (updates: Partial<UserProfile>) => updates as UserProfile,
  },

  notifications: {
    list: async () =>
      (await http<BackendNotification[]>("/api/notifications", { skipAuthRedirect: true })).map(notificationFromBackend),
    unreadCount: async () => (await http<{ count: number }>("/api/notifications/unread-count", { skipAuthRedirect: true })).count,
    markRead: async (id: number) => http<void>(`/api/notifications/${id}/read`, { method: "PUT" }),
    markAllRead: async () => http<void>("/api/notifications/read-all", { method: "PUT" }),
  },

  aiStatus: {
    get: (postId: number) => http(`/api/posts/${postId}/ai-status`),
  },
};

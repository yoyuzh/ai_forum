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
  ApiClient,
  NotificationItem,
  AIChat,
  AIChatMessage,
  AIChatSendResult,
  AIChatSession,
  AIChatSessionSummary,
  AIChatSessionPage,
  AIChatStreamHandler,
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
let notifications: NotificationItem[] = [
  {
    id: 1,
    type: "ai.reply.completed",
    title: "AI 回复了你的讨论",
    body: "ArchTechLead 已参与讨论",
    readAt: null,
    createdAt: new Date().toISOString(),
  },
];

const chatSessions = new Map<number, AIChatSession>();
const chatMessages = new Map<number, AIChatMessage[]>();
let nextChatSessionId = 1;
let nextChatMessageId = 1;

function createMockChat(agentId: number): AIChat {
  const agent = db.getAgent(agentId);
  if (!agent) throw new Error("Agent not found");
  const now = new Date().toISOString();
  const session = {
    id: nextChatSessionId++,
    userId: 1,
    aiAgentId: agentId,
    title: agent.displayName,
    status: "ACTIVE" as const,
    lastMessagePreview: "",
    messageCount: 0,
    createdAt: now,
    updatedAt: now,
  };
  chatSessions.set(session.id, session);
  chatMessages.set(session.id, []);
  return { session, agent, messages: [] };
}

function getMockChat(agentId: number, sessionId?: number): AIChat {
  const agent = db.getAgent(agentId);
  if (!agent) throw new Error("Agent not found");
  let session = sessionId ? chatSessions.get(sessionId) : undefined;
  if (session && session.aiAgentId !== agentId) throw new Error("Session not found");
  if (!session) throw new Error("Session not found");
  return { session, agent, messages: chatMessages.get(session.id) ?? [] };
}

function listMockChats(): AIChatSessionPage {
  const items = Array.from(chatSessions.values())
    .filter((session) => session.status !== "DELETED" && session.messageCount > 0)
    .map((session) => {
      const agent = db.getAgent(session.aiAgentId);
      if (!agent) return null;
      const messages = chatMessages.get(session.id) ?? [];
      return {
        session,
        agent,
        lastMessage: messages.at(-1)?.content ?? "",
        messageCount: messages.length,
      };
    })
    .filter((item): item is AIChatSessionSummary => Boolean(item))
    .sort((a, b) => +new Date(b.session.updatedAt) - +new Date(a.session.updatedAt));
  return { items, page: 1, pageSize: 20, total: items.length };
}

function createMockChatMessage(sessionId: number, role: AIChatMessage["role"], content: string): AIChatMessage {
  const message = {
    id: nextChatMessageId++,
    sessionId,
    role,
    content,
    status: "DONE" as const,
    sequenceNo: (chatMessages.get(sessionId)?.length ?? 0) + 1,
    requestId: null,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };
  chatMessages.set(sessionId, [...(chatMessages.get(sessionId) ?? []), message]);
  for (const [sessionKey, session] of chatSessions) {
    if (session.id === sessionId) {
      const messages = chatMessages.get(sessionId) ?? [];
      const title = messages.length === 1 ? titleFromContent(content) : session.title;
      chatSessions.set(sessionKey, {
        ...session,
        title,
        lastMessagePreview: content,
        messageCount: messages.length,
        updatedAt: message.createdAt,
      });
      break;
    }
  }
  return message;
}

function titleFromContent(content: string): string {
  const title = content.trim().replace(/\s+/g, " ");
  return title.length > 28 ? `${title.slice(0, 28)}...` : title || "新对话";
}

export const mockApi: ApiClient = {
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

  likes: {
    likePost: async (): Promise<void> => delay(undefined),
    unlikePost: async (): Promise<void> => delay(undefined),
  },

  favorites: {
    favoritePost: async (): Promise<void> => delay(undefined),
    unfavoritePost: async (): Promise<void> => delay(undefined),
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

  chat: {
    list: async (): Promise<AIChatSessionPage> => delay(listMockChats()),
    get: async (conversationId: number): Promise<AIChat> => {
      const session = chatSessions.get(conversationId);
      if (!session) throw new Error("Session not found");
      return delay(getMockChat(session.aiAgentId, conversationId));
    },
    sendMessage: async (
      agentId: number,
      content: string,
      conversationId: number | null,
      requestId: string,
      onEvent?: AIChatStreamHandler,
    ): Promise<AIChatSendResult> => {
      const chat = conversationId ? getMockChat(agentId, conversationId) : createMockChat(agentId);
      const trimmed = content.trim();
      if (!trimmed) throw new Error("请输入消息内容");
      if (!conversationId) {
        onEvent?.({ event: "conversation_created", data: { conversationId: chat.session.id, session: chat.session } });
      }
      const userMessage = { ...createMockChatMessage(chat.session.id, "user", trimmed), requestId };
      onEvent?.({ event: "user_message_saved", data: { message: userMessage, messageId: userMessage.id, sequenceNo: userMessage.sequenceNo } });
      const assistantMessage = createMockChatMessage(
        chat.session.id,
        "assistant",
        `${chat.agent.displayName}：我先按自己的视角回应你。${trimmed.length > 24 ? trimmed.slice(0, 24) + "…" : trimmed}`,
      );
      onEvent?.({ event: "ai_message_created", data: { message: { ...assistantMessage, content: "", status: "STREAMING" }, messageId: assistantMessage.id, sequenceNo: assistantMessage.sequenceNo } });
      onEvent?.({ event: "token", data: { content: assistantMessage.content } });
      const session = chatSessions.get(chat.session.id) ?? chat.session;
      onEvent?.({ event: "done", data: { message: assistantMessage, messageId: assistantMessage.id, status: "DONE", session } });
      return delay({ session, userMessage, assistantMessage });
    },
    retryMessage: async (messageId: number, _requestId: string, onEvent?: AIChatStreamHandler): Promise<AIChatMessage> => {
      for (const messages of chatMessages.values()) {
        const message = messages.find((item) => item.id === messageId);
        if (message) {
          message.status = "DONE";
          message.content = "重新生成后的回复。";
          onEvent?.({ event: "done", data: { message, messageId, status: "DONE", session: chatSessions.get(message.sessionId)! } });
          return delay(message);
        }
      }
      throw new Error("Message not found");
    },
    deleteConversation: async (conversationId: number) => {
      const session = chatSessions.get(conversationId);
      if (session) chatSessions.set(conversationId, { ...session, status: "DELETED" });
      return delay({ success: true });
    },
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

  notifications: {
    list: async (): Promise<NotificationItem[]> => delay(notifications),

    unreadCount: async (): Promise<number> =>
      delay(notifications.filter((item) => item.readAt === null).length),

    markRead: async (id: number): Promise<void> => {
      notifications = notifications.map((item) =>
        item.id === id ? { ...item, readAt: new Date().toISOString() } : item,
      );
      return delay(undefined);
    },

    markAllRead: async (): Promise<void> => {
      const now = new Date().toISOString();
      notifications = notifications.map((item) => ({ ...item, readAt: item.readAt ?? now }));
      return delay(undefined);
    },
  },

  aiStatus: {
    get: async (postId: number) => {
      const post = db.getPost(postId);
      return delay({
        completedCount: post?.aiResponsesCount ?? 0,
        runningCount: post?.aiStatus === "PROCESSING" ? 1 : 0,
        failedCount: 0,
        retryableCount: 0,
        overallStatus:
          post?.aiStatus === "PROCESSING"
            ? "RUNNING"
            : post?.aiStatus === "COMPLETED"
              ? "COMPLETED"
              : "IDLE",
      });
    },
    retry: async () => delay({ retried: 0 }),
  },
};

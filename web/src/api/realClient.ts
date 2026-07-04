import { getAuthToken, setAuthToken } from "./auth";
import { aiAgentAvatar } from "./agentAvatars";
import { aiAgentProfile } from "./agentProfiles";
import { defaultUserAvatar } from "../assets/brand";
import { HttpError, apiURL, http } from "./httpClient";
import type {
  AIAgent,
  AIChat,
  AIChatMessage,
  AIChatSendResult,
  AIChatSessionPage,
  AIChatStreamEvent,
  AIChatStreamHandler,
  AIChatSession,
  AIChatSessionSummary,
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
  user_id?: number | null;
  userId?: number | null;
  parent_comment_id?: number | null;
  parentId?: number | null;
  content: string;
  comment_type?: string;
  ai_agent_id?: number | null;
  aiAgentId?: number | null;
  trigger_type?: "AUTO" | "POST_AUTO" | "MENTION" | "FOLLOWUP";
  triggerType?: "AUTO" | "POST_AUTO" | "MENTION" | "FOLLOWUP";
  willingness_score?: number;
  willingnessScore?: number;
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

type BackendAgent = Partial<AIAgent> & {
  id: number;
  name: string;
  displayName?: string;
  display_name?: string;
  active?: boolean;
  enabled?: boolean;
  allow_auto_reply?: boolean;
  allow_mention?: boolean;
  allow_followup?: boolean;
  is_fallback?: boolean;
};

type BackendChatMessage = {
  id: number;
  sessionId?: number;
  session_id?: number;
  role?: AIChatMessage["role"] | "USER" | "AI";
  senderType?: "USER" | "AI";
  sender_type?: "USER" | "AI";
  content: string;
  status?: AIChatMessage["status"];
  sequenceNo?: number;
  sequence_no?: number;
  requestId?: string | null;
  request_id?: string | null;
  errorMessage?: string | null;
  error_message?: string | null;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type BackendChatSession = {
  id: number;
  userId?: number;
  user_id?: number;
  aiAgentId?: number;
  ai_agent_id?: number;
  title: string;
  status?: AIChatSession["status"];
  lastMessagePreview?: string;
  last_message_preview?: string;
  messageCount?: number;
  message_count?: number;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type BackendChat = {
  session: BackendChatSession;
  agent?: BackendAgent;
  messages: BackendChatMessage[];
};

type BackendChatSessionSummary = {
  session: BackendChatSession;
  agent: BackendAgent;
  lastMessage?: string;
  last_message?: string;
  messageCount?: number;
  message_count?: number;
};

type BackendChatSessionPage = {
  items: BackendChatSessionSummary[];
  page: number;
  pageSize?: number;
  page_size?: number;
  total: number;
};

type BackendTask = {
  id: number;
  postId?: number;
  post_id?: number;
  parentCommentId?: number | null;
  parent_comment_id?: number | null;
  targetCommentId?: number | null;
  commentId?: number | null;
  comment_id?: number | null;
  aiAgentId?: number;
  ai_agent_id?: number;
  triggerType?: AIReplyTask["triggerType"];
  trigger_type?: AIReplyTask["triggerType"];
  status: AIReplyTask["status"];
  prompt?: string;
  result?: string;
  errorMessage?: string | null;
  error_message?: string | null;
  retryCount?: number;
  retry_count?: number;
  attempt_count?: number;
  createdAt?: string;
  created_at?: string;
  startedAt?: string | null;
  started_at?: string | null;
  finishedAt?: string | null;
  finished_at?: string | null;
};

type BackendDecisionLog = {
  id: number;
  postId?: number;
  post_id?: number;
  commentId?: number | null;
  comment_id?: number | null;
  aiAgentId?: number;
  ai_agent_id?: number;
  aiAgentName?: string;
  ai_agent_name?: string;
  triggerType?: AIDecisionLog["triggerType"];
  trigger_type?: AIDecisionLog["triggerType"];
  willingnessScore?: number;
  willingness_score?: number;
  thresholdValue?: number;
  threshold_value?: number;
  decision: AIDecisionLog["decision"] | "FALLBACK";
  reason?: string;
  hitTags?: string[];
  hit_tags?: string[];
  createdAt?: string;
  created_at?: string;
};

function localAvatar(seed: string): string {
  return defaultUserAvatar(seed);
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

function commentTriggerType(
  triggerType?: BackendComment["trigger_type"] | BackendComment["triggerType"],
): Comment["triggerType"] | undefined {
  if (triggerType === "AUTO") return "POST_AUTO";
  return triggerType;
}

function commentFromBackend(c: BackendComment, postId: number, fallbackAuthor?: Comment["author"]): Comment {
  const isAi = c.author?.isAi ?? c.comment_type === "AI";
  const aiAgentId = c.ai_agent_id ?? c.aiAgentId ?? undefined;
  const userId = c.user_id ?? c.userId ?? undefined;
  const profile = isAi && aiAgentId ? aiAgentProfile(aiAgentId) : undefined;
  const username = c.author?.username ?? fallbackAuthor?.username ?? profile?.displayName ?? (userId ? `user_${userId}` : "backend_user");
  const role = c.author?.role ?? fallbackAuthor?.role ?? (isAi ? profile?.ageViewpoint : "研究员");
  const willingnessScore = c.willingnessScore ?? c.willingness_score;
  return {
    id: c.id,
    postId: c.post_id ?? c.postId ?? postId,
    parentId: c.parent_comment_id ?? c.parentId ?? null,
    content: c.content,
    author: {
      username,
      avatar:
        c.author?.avatar ??
        fallbackAuthor?.avatar ??
        (aiAgentId ? aiAgentAvatar(aiAgentId) : undefined) ??
        localAvatar(username),
      isAi,
      aiAgentId,
      role,
    },
    likeCount: 0,
    willingnessScore: willingnessScore === undefined ? undefined : willingnessScore > 1 ? willingnessScore : willingnessScore * 100,
    triggerType: commentTriggerType(c.trigger_type ?? c.triggerType),
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

function agentFromBackend(a: BackendAgent): AIAgent {
  const profile = aiAgentProfile(a.id);
  const displayName = profile?.displayName ?? a.displayName ?? a.display_name ?? a.name;
  const backendAvatar = typeof a.avatar === "string" && a.avatar.trim() !== "" ? a.avatar : undefined;
  return {
    id: a.id,
    name: a.name,
    displayName,
    avatar: aiAgentAvatar(a.id) ?? backendAvatar ?? localAvatar(displayName),
    icon: profile?.icon ?? a.icon ?? "smart_toy",
    description: profile?.description ?? a.description ?? "AI reply decision agent",
    ageViewpoint: profile?.ageViewpoint ?? a.ageViewpoint ?? "",
    personality: profile?.personality ?? a.personality ?? "",
    valueOrientation: profile?.valueOrientation ?? a.valueOrientation ?? "",
    speakingStyle: profile?.speakingStyle ?? a.speakingStyle ?? "",
    systemPrompt: "",
    stylePrompt: "",
    traits: profile?.traits ?? a.traits ?? [],
    specialties: profile?.specialties ?? a.specialties ?? [],
    replyThreshold: a.replyThreshold ?? 0,
    activityLevel: a.activityLevel ?? 0,
    temperature: a.temperature ?? 0.6,
    allowAutoReply: a.allowAutoReply ?? a.allow_auto_reply ?? true,
    allowMentionReply: a.allowMentionReply ?? a.allow_mention ?? true,
    allowFollowupReply: a.allowFollowupReply ?? a.allow_followup ?? true,
    maxAutoRepliesPerPost: a.maxAutoRepliesPerPost ?? 0,
    maxFollowupRepliesPerPost: a.maxFollowupRepliesPerPost ?? 0,
    isFallback: a.isFallback ?? a.is_fallback ?? false,
    active: a.active ?? a.enabled ?? true,
  };
}

function chatSessionFromBackend(s: BackendChatSession): AIChatSession {
  return {
    id: s.id,
    userId: s.userId ?? s.user_id ?? 0,
    aiAgentId: s.aiAgentId ?? s.ai_agent_id ?? 0,
    title: s.title,
    status: s.status ?? "ACTIVE",
    lastMessagePreview: s.lastMessagePreview ?? s.last_message_preview ?? "",
    messageCount: s.messageCount ?? s.message_count ?? 0,
    createdAt: s.createdAt ?? s.created_at ?? new Date().toISOString(),
    updatedAt: s.updatedAt ?? s.updated_at ?? new Date().toISOString(),
  };
}

function chatMessageFromBackend(m: BackendChatMessage): AIChatMessage {
  const sender = m.senderType ?? m.sender_type ?? m.role;
  return {
    id: m.id,
    sessionId: m.sessionId ?? m.session_id ?? 0,
    role: sender === "USER" ? "user" : sender === "AI" ? "assistant" : (m.role as AIChatMessage["role"]),
    content: m.content,
    status: m.status ?? "DONE",
    sequenceNo: m.sequenceNo ?? m.sequence_no ?? m.id,
    requestId: m.requestId ?? m.request_id ?? null,
    errorMessage: m.errorMessage ?? m.error_message ?? null,
    createdAt: m.createdAt ?? m.created_at ?? new Date().toISOString(),
    updatedAt: m.updatedAt ?? m.updated_at ?? m.createdAt ?? m.created_at ?? new Date().toISOString(),
  };
}

async function chatFromBackend(agentId: number, chat: BackendChat): Promise<AIChat> {
  const session = chatSessionFromBackend(chat.session);
  const agent = chat.agent ? agentFromBackend(chat.agent) : await realApi.agents.get(agentId);
  return {
    session,
    agent,
    messages: chat.messages.map(chatMessageFromBackend),
  };
}

function chatSessionSummaryFromBackend(row: BackendChatSessionSummary): AIChatSessionSummary {
  return {
    session: chatSessionFromBackend(row.session),
    agent: agentFromBackend(row.agent),
    lastMessage: row.lastMessage ?? row.last_message ?? "",
    messageCount: row.messageCount ?? row.message_count ?? 0,
  };
}

function chatSessionPageFromBackend(page: BackendChatSessionPage): AIChatSessionPage {
  return {
    items: page.items.map(chatSessionSummaryFromBackend),
    page: page.page,
    pageSize: page.pageSize ?? page.page_size ?? 20,
    total: page.total,
  };
}

function streamEventFromBackend(event: string, data: unknown): AIChatStreamEvent {
  const payload = data as any;
  if (event === "conversation_created") {
    return { event, data: { ...payload, session: chatSessionFromBackend(payload.session) } };
  }
  if (event === "user_message_saved" || event === "ai_message_created") {
    return { event, data: { ...payload, message: chatMessageFromBackend(payload.message) } } as AIChatStreamEvent;
  }
  if (event === "done") {
    return { event, data: { ...payload, session: chatSessionFromBackend(payload.session), message: chatMessageFromBackend(payload.message) } };
  }
  if (event === "error") {
    return { event, data: { ...payload, aiMessage: payload.aiMessage ? chatMessageFromBackend(payload.aiMessage) : undefined } };
  }
  return { event: "token", data: payload };
}

async function postChatStream(
  path: string,
  body: unknown,
  onEvent?: AIChatStreamHandler,
): Promise<{ session?: AIChatSession; userMessage?: AIChatMessage; assistantMessage?: AIChatMessage }> {
  const headers = new Headers({ "Content-Type": "application/json" });
  const token = getAuthToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const response = await fetch(apiURL(path), { method: "POST", headers, body: JSON.stringify(body) });
  if (!response.ok) throw new HttpError(response.status, await response.text());
  if (!response.body) throw new HttpError(500, "stream unavailable");

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  const state: { session?: AIChatSession; userMessage?: AIChatMessage; assistantMessage?: AIChatMessage } = {};

  const handleBlock = (block: string) => {
    let event = "message";
    const dataLines: string[] = [];
    for (const line of block.split("\n")) {
      if (line.startsWith("event:")) event = line.slice(6).trim();
      if (line.startsWith("data:")) dataLines.push(line.slice(5).trimStart());
    }
    if (!dataLines.length) return;
    const parsed = streamEventFromBackend(event, JSON.parse(dataLines.join("\n")));
    if (parsed.event === "conversation_created") state.session = parsed.data.session;
    if (parsed.event === "user_message_saved") state.userMessage = parsed.data.message;
    if (parsed.event === "ai_message_created") state.assistantMessage = parsed.data.message;
    if (parsed.event === "token" && state.assistantMessage) {
      state.assistantMessage = { ...state.assistantMessage, content: `${state.assistantMessage.content}${parsed.data.content}` };
    }
    if (parsed.event === "done") {
      state.session = parsed.data.session;
      state.assistantMessage = parsed.data.message;
    }
    if (parsed.event === "error" && parsed.data.aiMessage) state.assistantMessage = parsed.data.aiMessage;
    onEvent?.(parsed);
  };

  for (;;) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const blocks = buffer.split("\n\n");
    buffer = blocks.pop() ?? "";
    blocks.forEach(handleBlock);
  }
  if (buffer.trim()) handleBlock(buffer);
  return state;
}

function taskFromBackend(t: BackendTask): AIReplyTask {
  return {
    id: t.id,
    postId: t.postId ?? t.post_id ?? 0,
    parentCommentId: t.parentCommentId ?? t.parent_comment_id ?? null,
    targetCommentId: t.targetCommentId ?? t.commentId ?? t.comment_id ?? null,
    aiAgentId: t.aiAgentId ?? t.ai_agent_id ?? 0,
    triggerType: t.triggerType ?? t.trigger_type ?? "POST_AUTO",
    status: t.status,
    prompt: t.prompt ?? "",
    result: t.result ?? "",
    errorMessage: t.errorMessage ?? t.error_message ?? "",
    retryCount: t.retryCount ?? t.retry_count ?? t.attempt_count ?? 0,
    createdAt: t.createdAt ?? t.created_at ?? new Date().toISOString(),
    startedAt: t.startedAt ?? t.started_at ?? null,
    finishedAt: t.finishedAt ?? t.finished_at ?? null,
  };
}

function decisionLogFromBackend(l: BackendDecisionLog): AIDecisionLog {
  const score = l.willingnessScore ?? l.willingness_score ?? 0;
  const threshold = l.thresholdValue ?? l.threshold_value ?? 0;
  return {
    id: l.id,
    postId: l.postId ?? l.post_id ?? 0,
    commentId: l.commentId ?? l.comment_id ?? null,
    aiAgentId: l.aiAgentId ?? l.ai_agent_id ?? 0,
    aiAgentName: l.aiAgentName ?? l.ai_agent_name ?? "AI",
    triggerType: l.triggerType ?? l.trigger_type ?? "POST_AUTO",
    willingnessScore: score > 1 ? score : score * 100,
    thresholdValue: threshold > 1 ? threshold : threshold * 100,
    decision: l.decision === "FALLBACK" ? "REPLY" : l.decision,
    reason: l.reason ?? "",
    hitTags: l.hitTags ?? l.hit_tags ?? [],
    createdAt: l.createdAt ?? l.created_at ?? new Date().toISOString(),
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
      const normalizedTag = tag?.trim().toLowerCase();
      const posts = query.trim()
        ? (await http<BackendPost[]>(`/api/search/posts?q=${encodeURIComponent(query.trim())}`)).map(postFromBackend)
        : await listPosts();
      const sorted = [...posts].sort((a, b) => {
        if (tab === "hottest") return b.viewCount - a.viewCount;
        if (tab === "ai_most") return b.aiResponsesCount - a.aiResponsesCount;
        return +new Date(b.createdAt) - +new Date(a.createdAt);
      });
      return sorted.filter((post) => {
        if (tab === "unanswered" && post.aiResponsesCount > 0) return false;
        const matchesTag =
          !normalizedTag || post.tags.some((postTag) => postTag.toLowerCase() === normalizedTag);
        return matchesTag;
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
    create: async (comment) => {
      const created = await http<BackendComment>(`/api/posts/${comment.postId}/comments`, {
        method: "POST",
        body: JSON.stringify({
          content: comment.content,
          parent_comment_id: comment.parentId,
        }),
      });
      return commentFromBackend(created, comment.postId, comment.author);
    },
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
    list: async (): Promise<AIAgent[]> =>
      (await http<BackendAgent[]>("/api/agents")).map(agentFromBackend),
    get: async (id: number): Promise<AIAgent> => {
      const agent = (await realApi.agents.list()).find((a) => a.id === id);
      if (!agent) throw new Error(`Agent ${id} not found`);
      return agent;
    },
    update: async (): Promise<AIAgent> => {
      throw new Error("Agent update unavailable in web real mode");
    },
  },

  chat: {
    list: async (input = {}): Promise<AIChatSessionPage> => {
      const params = new URLSearchParams();
      if (input.page) params.set("page", String(input.page));
      if (input.pageSize) params.set("pageSize", String(input.pageSize));
      if (input.agentId) params.set("agentId", String(input.agentId));
      const qs = params.toString();
      return chatSessionPageFromBackend(
        await http<BackendChatSessionPage>(`/api/ai-chat/conversations${qs ? `?${qs}` : ""}`, {
          skipAuthRedirect: true,
        }),
      );
    },
    get: async (conversationId: number): Promise<AIChat> => {
      const chat = await http<BackendChat>(`/api/ai-chat/conversations/${conversationId}/messages`, {
        skipAuthRedirect: true,
      });
      return chatFromBackend(chat.session.aiAgentId ?? chat.session.ai_agent_id ?? 0, chat);
    },
    sendMessage: async (agentId, content, conversationId, requestId, onEvent): Promise<AIChatSendResult> => {
      const result = await postChatStream(
        "/api/ai-chat/messages/stream",
        { conversationId, agentId, content, requestId },
        onEvent,
      );
      if (!result.session || !result.userMessage || !result.assistantMessage) {
        throw new HttpError(502, "incomplete chat stream");
      }
      return {
        session: result.session,
        userMessage: result.userMessage,
        assistantMessage: result.assistantMessage,
      };
    },
    retryMessage: async (messageId, requestId, onEvent): Promise<AIChatMessage> => {
      const result = await postChatStream(`/api/ai-chat/messages/${messageId}/retry`, { requestId }, onEvent);
      if (!result.assistantMessage) throw new HttpError(502, "incomplete retry stream");
      return result.assistantMessage;
    },
    deleteConversation: (conversationId: number) =>
      http<{ success: boolean }>(`/api/ai-chat/conversations/${conversationId}`, {
        method: "DELETE",
        skipAuthRedirect: true,
      }),
  },

  tasks: { list: async (): Promise<AIReplyTask[]> => (await http<BackendTask[]>("/api/ai-tasks")).map(taskFromBackend) },
  decisionLogs: {
    list: async (): Promise<AIDecisionLog[]> =>
      (await http<BackendDecisionLog[]>("/api/decision-logs")).map(decisionLogFromBackend),
    listForPost: async (postId: number): Promise<AIDecisionLog[]> =>
      (await http<BackendDecisionLog[]>(`/api/posts/${postId}/decision-logs`)).map(decisionLogFromBackend),
  },
  activities: { list: () => http<AIActivity[]>("/api/ai-activity") },

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
    getStats: async (): Promise<UserStats> => http<UserStats>("/api/me/stats"),
    updateProfile: async (updates: Partial<UserProfile>) =>
      userFromBackend(
        await http<BackendUser>("/api/me", {
          method: "PATCH",
          body: JSON.stringify({ nickname: updates.nickname }),
        }),
      ),
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
    retry: (postId: number) => http(`/api/posts/${postId}/ai-retry`, { method: "POST" }),
  },
};

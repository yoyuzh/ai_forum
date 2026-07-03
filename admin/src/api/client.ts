import {
  AGENTS,
  TASKS,
  DECISION_LOGS,
  POSTS,
  DASHBOARD_STATS,
  WEEKLY_POST_TREND,
  TASK_STATUS_BREAKDOWN,
  SERVICES,
  RECENT_POSTS,
  RECENT_TASKS,
  DECISION_TIMELINE,
  DECISION_POST_CONTEXT,
  TASK_SUMMARY,
} from "./mockData";
import type {
  AdminAIAgent,
  AdminAITask,
  AdminDecisionLog,
  AdminPost,
  AdminPreference,
  AdminSession,
  AdminTag,
  AdminUser,
} from "./types";

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? "http://127.0.0.1:19091").replace(/\/$/, "");
const API_MODE = import.meta.env.VITE_API_MODE ?? "real";
const TOKEN_KEY = "ai_forum_admin_token";
const SESSION_KEY = "ai_forum_admin_session";

export class AdminHttpError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
    this.name = "AdminHttpError";
  }
}

const delay = <T>(value: T): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(value), 120));

export function getAdminToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setAdminToken(token: string | null): void {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

export function getStoredSession(): AdminSession | null {
  const raw = localStorage.getItem(SESSION_KEY);
  return raw ? (JSON.parse(raw) as AdminSession) : null;
}

function setStoredSession(session: AdminSession | null): void {
  if (session) localStorage.setItem(SESSION_KEY, JSON.stringify(session));
  else localStorage.removeItem(SESSION_KEY);
}

async function http<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const token = getAdminToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers });
  if (response.status === 401) {
    setAdminToken(null);
    setStoredSession(null);
    if (location.pathname !== "/login") location.assign("/login");
    throw new AdminHttpError(401, "请先登录");
  }
  if (!response.ok) {
    throw new AdminHttpError(response.status, (await response.text()) || `HTTP ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

const mockApi = {
  dashboard: {
    stats: async () => delay(DASHBOARD_STATS),
    weeklyTrend: async () => delay(WEEKLY_POST_TREND),
    taskStatusBreakdown: async () => delay(TASK_STATUS_BREAKDOWN),
    services: async () => delay(SERVICES),
    recentPosts: async () => delay(RECENT_POSTS),
    recentTasks: async () => delay(RECENT_TASKS),
    decisionTimeline: async () => delay(DECISION_TIMELINE),
  },
  decisionContext: async () => delay(DECISION_POST_CONTEXT),
  taskSummary: async () => delay(TASK_SUMMARY),
  auth: {
    login: async (username: string): Promise<AdminSession> => {
      const session = { username, role: "ADMIN", permissions: ["post:delete-any", "ai_task:retry", "ai_agent:update", "decision_log:read"] };
      setAdminToken("mock-admin-token");
      setStoredSession(session);
      return delay(session);
    },
    me: async () => delay(getStoredSession() ?? { username: "admin", role: "ADMIN", permissions: [] }),
    logout: async () => {
      setAdminToken(null);
      setStoredSession(null);
    },
  },
  users: { list: async (): Promise<AdminUser[]> => delay([]) },
  agents: {
    list: async (): Promise<AdminAIAgent[]> => delay(AGENTS),
    get: async (id: string | number): Promise<AdminAIAgent> => {
      const a = AGENTS.find((x) => String(x.id) === String(id));
      if (!a) throw new Error(`Agent ${id} not found`);
      return delay(a);
    },
    update: async (id: string | number, updates: Partial<AdminAIAgent>): Promise<AdminAIAgent> => {
      const index = AGENTS.findIndex((x) => String(x.id) === String(id));
      if (index === -1) throw new Error(`Agent ${id} not found`);
      AGENTS[index] = { ...AGENTS[index], ...updates };
      return delay(AGENTS[index]);
    },
  },
  tasks: {
    list: async (): Promise<AdminAITask[]> => delay(TASKS),
    get: async (id: string | number): Promise<AdminAITask> => {
      const t = TASKS.find((x) => String(x.id) === String(id));
      if (!t) throw new Error(`Task ${id} not found`);
      return delay(t);
    },
    retry: async (id: string | number): Promise<AdminAITask> => {
      const t = TASKS.find((x) => String(x.id) === String(id));
      if (!t) throw new Error(`Task ${id} not found`);
      Object.assign(t, { status: "PENDING", statusLabel: "Waiting", retryCount: 0, errorMessage: null });
      return delay(t);
    },
    terminate: async (id: string | number): Promise<AdminAITask> => mockApi.tasks.retry(id),
    markProcessed: async (id: string | number): Promise<AdminAITask> => mockApi.tasks.retry(id),
  },
  decisionLogs: { list: async (): Promise<AdminDecisionLog[]> => delay(DECISION_LOGS) },
  posts: { list: async (): Promise<AdminPost[]> => delay(POSTS) },
  comments: { list: async () => delay([]) },
  tags: { list: async (): Promise<AdminTag[]> => delay([]) },
  preferences: { list: async (): Promise<AdminPreference[]> => delay([]) },
};

function agentFromBackend(row: AdminAIAgent): AdminAIAgent {
  return {
    icon: row.fallback ? "support_agent" : "smart_toy",
    displayName: row.name,
    description: row.fallback ? "Fallback decision agent" : "AI reply decision agent",
    traits: row.fallback ? ["fallback"] : ["decision"],
    specialties: [],
    temperature: 0.6,
    systemPrompt: "",
    ...row,
    id: row.id,
  };
}

const realApi = {
  dashboard: mockApi.dashboard,
  decisionContext: mockApi.decisionContext,
  taskSummary: async () => {
    const rows = await http<AdminAITask[]>("/api/admin/ai-tasks");
    return {
      pending: rows.filter((t) => t.status === "PENDING").length,
      running: rows.filter((t) => t.status === "PROCESSING").length,
      success: rows.filter((t) => t.status === "COMPLETED").length,
      failed: rows.filter((t) => t.status === "FAILED").length,
      skipped: 0,
    };
  },
  auth: {
    login: async (username: string, password: string): Promise<AdminSession> => {
      const { token } = await http<{ token: string }>("/api/login", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });
      setAdminToken(token);
      const session = await realApi.auth.me();
      setStoredSession(session);
      return session;
    },
    me: async (): Promise<AdminSession> => {
      const profile = await http<{ id?: number; username: string; role: string; permissions?: string[] }>("/api/me");
      let permissions: string[];
      try {
        permissions = (await http<{ permissions: string[] }>("/api/admin/permissions")).permissions;
      } catch {
        permissions = profile.permissions ?? [];
      }
      const session = { id: profile.id, username: profile.username, role: profile.role, permissions };
      setStoredSession(session);
      return session;
    },
    logout: async () => {
      setAdminToken(null);
      setStoredSession(null);
    },
  },
  users: { list: () => http<AdminUser[]>("/api/admin/users") },
  posts: { list: () => http<AdminPost[]>("/api/admin/posts") },
  comments: { list: () => http<unknown[]>("/api/admin/comments") },
  agents: {
    list: async () => (await http<AdminAIAgent[]>("/api/admin/ai-agents")).map(agentFromBackend),
    get: async (id: string | number) => (await realApi.agents.list()).find((a) => String(a.id) === String(id)) ?? Promise.reject(new Error(`Agent ${id} not found`)),
    update: (id: string | number, updates: Partial<AdminAIAgent>) =>
      http<AdminAIAgent>(`/api/admin/ai-agents/${id}`, { method: "PATCH", body: JSON.stringify(updates) }),
  },
  tasks: {
    list: () => http<AdminAITask[]>("/api/admin/ai-tasks"),
    get: async (id: string | number) => (await realApi.tasks.list()).find((t) => String(t.id) === String(id)) ?? Promise.reject(new Error(`Task ${id} not found`)),
    retry: (id: string | number) => http<AdminAITask>(`/api/admin/ai-tasks/${id}/retry`, { method: "POST" }),
    terminate: (id: string | number) => http<AdminAITask>(`/api/admin/ai-tasks/${id}/terminate`, { method: "POST" }),
    markProcessed: (id: string | number) => http<AdminAITask>(`/api/admin/ai-tasks/${id}/mark-processed`, { method: "POST" }),
  },
  decisionLogs: { list: () => http<AdminDecisionLog[]>("/api/admin/decision-logs") },
  tags: { list: () => http<AdminTag[]>("/api/admin/tags") },
  preferences: { list: () => http<AdminPreference[]>("/api/admin/preferences") },
};

export const adminApi = API_MODE === "mock" ? mockApi : realApi;

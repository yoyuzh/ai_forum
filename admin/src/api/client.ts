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
} from "./types";

const delay = <T>(value: T): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(value), 200));

/**
 * Thin mock API client. The dashboard reads directly from here; the agent/task
 * tables go through the Refine dataProvider which wraps these calls.
 *
 * When the real backend lands, replace each function body with a fetch call —
 * signatures stay the same so call sites don't change.
 */
export const adminApi = {
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

  agents: {
    list: async (): Promise<AdminAIAgent[]> => delay(AGENTS),
    get: async (id: string): Promise<AdminAIAgent> => {
      const a = AGENTS.find((x) => x.id === id);
      if (!a) throw new Error(`Agent ${id} not found`);
      return delay(a);
    },
    update: async (id: string, updates: Partial<AdminAIAgent>): Promise<AdminAIAgent> => {
      const index = AGENTS.findIndex((x) => x.id === id);
      if (index === -1) throw new Error(`Agent ${id} not found`);
      AGENTS[index] = { ...AGENTS[index], ...updates };
      return delay(AGENTS[index]);
    },
  },

  tasks: {
    list: async (): Promise<AdminAITask[]> => delay(TASKS),
    get: async (id: string): Promise<AdminAITask> => {
      const t = TASKS.find((x) => x.id === id);
      if (!t) throw new Error(`Task ${id} not found`);
      return delay(t);
    },
    retry: async (id: string): Promise<AdminAITask> => {
      const t = TASKS.find((x) => x.id === id);
      if (!t) throw new Error(`Task ${id} not found`);
      // Reset the task into the pending queue — backend RBAC would gate this.
      Object.assign(t, { status: "PENDING", statusLabel: "Waiting", retryCount: 0, errorMessage: null });
      return delay(t);
    },
  },

  decisionLogs: {
    list: async (): Promise<AdminDecisionLog[]> => delay(DECISION_LOGS),
  },

  posts: {
    list: async (): Promise<AdminPost[]> => delay(POSTS),
  },
};

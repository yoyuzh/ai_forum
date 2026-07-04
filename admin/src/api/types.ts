// Admin domain shapes. Several reuse the web app's contract so the two apps
// stay in sync; admin adds dashboard/operational types.

export type AIStatus = "PENDING" | "PROCESSING" | "COMPLETED";

export interface AdminUser {
  id: number;
  username: string;
  avatar?: string;
  role: string;
  status: "active" | "banned" | "ACTIVE" | "BANNED";
  postCount: number;
  createdAt: string;
}

export interface AdminPost {
  id: number;
  title: string;
  author: string;
  category?: string;
  status: "published" | "review" | "draft" | "NORMAL" | "HIDDEN" | "DELETED";
  viewCount: number;
  commentCount: number;
  aiResponsesCount: number;
  createdAt: string;
}

export interface AdminAIAgent {
  id: string | number; // e.g. "A001" or backend id 1001
  name: string;
  displayName?: string;
  avatar?: string;
  icon?: string;
  description?: string;
  traits?: string[];
  specialties?: string[];
  replyThreshold: number; // 0–1
  activityLevel: number; // 0–1
  temperature?: number;
  systemPrompt?: string;
  allowAutoReply: boolean;
  allowMentionReply: boolean;
  allowFollowupReply: boolean;
  active: boolean;
  fallback?: boolean;
  replyCount: number;
  createdAt?: string;
}

export type TaskStatus = "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED";

export interface AdminAITask {
  id: string | number; // e.g. "tsk_928a4b" or backend id
  agentId?: string | number;
  aiAgentId?: number;
  agentName?: string;
  aiAgentName?: string;
  agentInitials?: string;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP" | "SCHEDULED";
  triggerLabel: string;
  targetPostId?: string;
  postId?: number;
  status: TaskStatus;
  statusLabel: string;
  durationMs?: number | null;
  tokens?: number | null;
  retryCount: number;
  maxRetries?: number;
  errorMessage: string | null;
  prompt?: string;
  result?: string;
  createdAt: string;
  updatedAt?: string;
  timeline?: { time: string; label: string; detail: string; state: "ok" | "active" | "error" }[];
}

export type DecisionResult = "REPLY" | "IGNORE" | "FAILED" | "FALLBACK";

export interface AdminDecisionLog {
  id: number;
  postId: string | number;
  commentId?: number | null;
  aiAgentId: string | number;
  aiAgentName: string;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  willingnessScore: number; // 0–100
  thresholdValue: number; // 0–100
  decision: DecisionResult;
  decisionLabel?: string;
  reason: string;
  fallback?: boolean;
  traits?: string[];
  hitTags: string[];
  taskId?: number | null;
  commentLink?: number | null;
  createdAt: string;
}

export interface AdminTag {
  id: number;
  postId: number;
  type: string;
  name: string;
  createdAt: string;
}

export interface AdminPreference {
  id: number;
  agentId: number;
  tagType: string;
  tagName: string;
  weight: number;
  createdAt: string;
}

export interface AdminSession {
  id?: number;
  username: string;
  role: string;
  permissions: string[];
}

export interface DashboardStats {
  totalUsers: number;
  totalPosts: number;
  aiReplies: number;
  todayAiTasks: number;
  failedTasks: number;
}

export interface TrendPoint {
  label: string;
  value: number;
}

export interface TaskStatusBreakdown {
  success: number;
  running: number;
  failed: number;
}

export interface ServiceStatus {
  name: string;
  metric: string;
  healthy: boolean;
}

export interface RecentPostRow {
  id: number;
  title: string;
  author: string;
  relativeTime: string;
  status: "published" | "review";
}

export interface DecisionTimelineEntry {
  time: string;
  message: string;
}

export interface RecentTaskRow {
  id: string | number;
  label: string;
  icon: string;
  status: TaskStatus;
}

export interface DecisionPostContext {
  postId: string | number;
  title: string;
  body: string;
  tags: string[];
  timestamp: string;
}

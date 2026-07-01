// Admin domain shapes. Several reuse the web app's contract so the two apps
// stay in sync; admin adds dashboard/operational types.

export type AIStatus = "PENDING" | "PROCESSING" | "COMPLETED";

export interface AdminUser {
  id: number;
  username: string;
  avatar: string;
  role: string;
  status: "active" | "banned";
  postCount: number;
  createdAt: string;
}

export interface AdminPost {
  id: number;
  title: string;
  author: string;
  category: string;
  status: "published" | "review" | "draft";
  viewCount: number;
  commentCount: number;
  aiResponsesCount: number;
  createdAt: string;
}

export interface AdminAIAgent {
  id: string; // e.g. "A001"
  name: string;
  displayName: string;
  avatar: string;
  icon: string;
  description: string;
  traits: string[];
  specialties: string[];
  replyThreshold: number; // 0–1
  activityLevel: number; // 0–1
  temperature: number;
  systemPrompt: string;
  allowAutoReply: boolean;
  allowMentionReply: boolean;
  allowFollowupReply: boolean;
  active: boolean;
  replyCount: number;
}

export type TaskStatus = "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED";

export interface AdminAITask {
  id: string; // e.g. "tsk_928a4b"
  agentId: string;
  agentName: string;
  agentInitials: string;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP" | "SCHEDULED";
  triggerLabel: string;
  targetPostId: string;
  status: TaskStatus;
  statusLabel: string;
  durationMs: number | null;
  tokens: number | null;
  retryCount: number;
  maxRetries: number;
  errorMessage: string | null;
  prompt: string;
  result: string;
  createdAt: string;
  timeline: { time: string; label: string; detail: string; state: "ok" | "active" | "error" }[];
}

export type DecisionResult = "REPLY" | "IGNORE" | "FAILED";

export interface AdminDecisionLog {
  id: number;
  postId: string;
  aiAgentId: string;
  aiAgentName: string;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  willingnessScore: number; // 0–100
  thresholdValue: number; // 0–100
  decision: DecisionResult;
  decisionLabel: string;
  reason: string;
  traits: string[];
  hitTags: string[];
  createdAt: string;
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

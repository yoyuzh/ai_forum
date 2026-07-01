// Shared domain shapes for the user-facing web app.
//
// These mirror the backend contract documented in ai_forum_architecture_v1.md
// and the Stitch prototypes. While the backend is incomplete, the mock layer in
// src/api/db.ts returns these exact shapes so swapping to a real API is a
// drop-in change inside src/api/client.ts.

export type AIStatus = "PENDING" | "PROCESSING" | "COMPLETED";

export type FeedTab = "latest" | "hottest" | "unanswered" | "ai_most";

export interface Author {
  username: string;
  avatar: string;
  role?: string;
}

export interface Post {
  id: number;
  title: string;
  content: string;
  category: string;
  tags: string[];
  author: Author;
  aiStatus: AIStatus;
  aiResponsesCount: number;
  aiAvatars: string[];
  viewCount: number;
  commentCount: number;
  likeCount: number;
  createdAt: string;
}

export interface CommentAuthor {
  username: string;
  avatar: string;
  isAi: boolean;
  aiAgentId?: number;
  role?: string;
}

export interface Comment {
  id: number;
  postId: number;
  parentId: number | null;
  content: string;
  author: CommentAuthor;
  likeCount: number;
  /** For AI comments: 0-100 willingness score that produced this reply. */
  willingnessScore?: number;
  /** For AI comments: how the reply was triggered. */
  triggerType?: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  createdAt: string;
}

export interface AIAgent {
  id: number;
  name: string;
  displayName: string;
  avatar: string;
  icon: string; // Material Symbols name
  description: string;
  ageViewpoint: string;
  personality: string;
  valueOrientation: string;
  speakingStyle: string;
  systemPrompt: string;
  stylePrompt: string;
  traits: string[];
  specialties: string[];
  replyThreshold: number; // e.g. 0.60
  activityLevel: number; // e.g. 0.50
  temperature: number;
  allowAutoReply: boolean;
  allowMentionReply: boolean;
  allowFollowupReply: boolean;
  maxAutoRepliesPerPost: number;
  maxFollowupRepliesPerPost: number;
  isFallback: boolean;
  active: boolean;
}

export type TaskStatus = "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED";

export interface AIReplyTask {
  id: number;
  postId: number;
  parentCommentId: number | null;
  targetCommentId: number | null;
  aiAgentId: number;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  status: TaskStatus;
  prompt: string;
  result: string;
  errorMessage: string;
  retryCount: number;
  createdAt: string;
  startedAt: string | null;
  finishedAt: string | null;
}

export type DecisionResult = "REPLY" | "IGNORE" | "FAILED";

export interface AIDecisionLog {
  id: number;
  postId: number;
  commentId: number | null;
  aiAgentId: number;
  aiAgentName: string;
  triggerType: "POST_AUTO" | "MENTION" | "FOLLOWUP";
  willingnessScore: number;
  thresholdValue: number;
  decision: DecisionResult;
  reason: string;
  /** Tags on the post that matched the agent's specialties (optional in
   *  simulated decisions — the live simulator doesn't compute them). */
  hitTags?: string[];
  createdAt: string;
}

/** Sidebar activity item — recent AI actions across the forum. */
export interface AIActivity {
  id: number;
  agentName: string;
  agentAvatar: string;
  action: string;
  target: string;
  targetId: number;
  relativeTime: string;
}

export interface RelatedDiscussion {
  id: number;
  title: string;
}

/** AI processing pipeline steps shown in the post-detail sidebar. */
export interface ProcessingStep {
  key: string;
  label: string;
  detail: string;
  icon: string;
  state: "done" | "active" | "pending";
}

export interface DatabaseState {
  posts: Post[];
  comments: Comment[];
  agents: AIAgent[];
  tasks: AIReplyTask[];
  decisionLogs: AIDecisionLog[];
  activities: AIActivity[];
  users: UserProfile[];
  /** Username of the locally "logged-in" user. null after explicit logout. */
  currentUserId: string | null;
}

// --- Auth & user profile (mock layer; no real backend session) ---

export interface UserPreferences {
  /** Receive notifications when an AI replies to the user's posts/comments. */
  aiReplyNotifications: boolean;
  /** Show real-time AI activity feed / live SSE status. */
  liveActivity: boolean;
  /** Theme preference — display-only placeholder, no live theme switching yet. */
  themePreference: "system" | "light" | "dark";
}

export interface UserProfile {
  /** Stable login handle, also used by Header/PostCard consumers. */
  username: string;
  /** Display name shown across the forum. */
  nickname: string;
  email: string;
  /** Avatar URL (dicebear/preset or custom). */
  avatar: string;
  bio: string;
  /** Frontend display-only role label; NOT a security capability. */
  role: string;
  /** Frontend display-only user id; NOT authoritative. */
  uid: string;
  joinedAt: string;
  /** Display-only badge; real email verification lives on the backend. */
  emailVerified: boolean;
  preferences: UserPreferences;
}

export interface AuthResult {
  user: UserProfile;
}

/** Aggregated profile statistics computed from posts/comments, never stored. */
export interface UserStats {
  postCount: number;
  commentCount: number;
  likeCount: number;
  aiReplyCount: number;
}

import { useMemo, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { usePostDetail } from "../hooks/usePosts";
import { useComments } from "../hooks/useComments";
import { useDecisionLogsForPost } from "../hooks/useDecisionLogs";
import { usePostActions } from "../hooks/usePostActions";
import { useAIStatus } from "../hooks/useAIStatus";
import { useRelatedDiscussions } from "../hooks/useRelatedDiscussions";
import { useUserStore } from "../stores/useUserStore";
import { Comment, ProcessingStep } from "../api/types";
import { formatRelativeTime, formatCount } from "../utils/format";
import MaterialIcon from "../components/ui/MaterialIcon";
import SafeMarkdown from "../components/ui/SafeMarkdown";
import CommentEditor from "../components/comments/CommentEditor";
import HumanComment from "../components/comments/HumanComment";
import AIComment from "../components/comments/AIComment";
import AlertBar from "../components/ui/AlertBar";
import ParticipatingAI from "../components/sidebar/ParticipatingAI";
import PostTags from "../components/sidebar/PostTags";
import AIProcessingStatus from "../components/sidebar/AIProcessingStatus";
import RelatedDiscussions from "../components/sidebar/RelatedDiscussions";

/** Derive the 4-step processing pipeline from the post's current AI status. */
function buildSteps(status: "IDLE" | "RUNNING" | "COMPLETED" | "FAILED", hasTags: boolean): ProcessingStep[] {
  const base: ProcessingStep[] = [
    { key: "tags", label: "分析帖子标签", detail: "已提取核心概念", icon: "label", state: "done" },
    { key: "score", label: "计算 AI 回答意愿分", detail: "评估各代理阈值", icon: "analytics", state: "done" },
    { key: "generate", label: "生成 AI 回复", detail: "推理引擎运行中…", icon: "hourglass_empty", state: "active" },
    { key: "write", label: "写入评论区", detail: "等待挂起", icon: "edit_note", state: "pending" },
  ];

  if (status === "IDLE") {
    return [
      { ...base[0], state: hasTags ? "done" : "active", detail: hasTags ? "已提取核心概念" : "等待标签分析" },
      { ...base[1], state: "pending", detail: "等待挂起" },
      { ...base[2], state: "pending", detail: "等待挂起" },
      { ...base[3], state: "pending", detail: "等待挂起" },
    ];
  }
  if (status === "COMPLETED") {
    return [
      { ...base[0], detail: "已提取核心概念" },
      { ...base[1], detail: "已计算意愿分" },
      { ...base[2], state: "done", detail: "回复已生成", icon: "check" },
      { ...base[3], state: "done", detail: "已写入评论区", icon: "check" },
    ];
  }
  if (status === "FAILED") {
    return [
      { ...base[0], state: hasTags ? "done" : "active", detail: hasTags ? "已提取核心概念" : "等待标签分析" },
      { ...base[1], detail: "已计算意愿分" },
      { ...base[2], state: "done", detail: "生成失败", icon: "error" },
      { ...base[3], state: "pending", detail: "未写入评论区" },
    ];
  }
  return base;
}

function statusFromPost(status: "PENDING" | "PROCESSING" | "COMPLETED"): "IDLE" | "RUNNING" | "COMPLETED" {
  if (status === "COMPLETED") return "COMPLETED";
  if (status === "PROCESSING") return "RUNNING";
  return "IDLE";
}

type CommentNode = Comment & { children: CommentNode[] };
type ReplyExpansion = "preview" | "all";
type FlatReply = { comment: CommentNode; replyTo: string };

const PREVIEW_REPLY_COUNT = 5;

function authorLabel(comment: Comment): string {
  return comment.author.username;
}

function buildCommentTree(comments: Comment[]): CommentNode[] {
  const byID = new Map<number, CommentNode>();
  const roots: CommentNode[] = [];

  for (const comment of comments) {
    byID.set(comment.id, { ...comment, children: [] });
  }

  for (const comment of comments) {
    const node = byID.get(comment.id)!;
    const parent = comment.parentId == null ? null : byID.get(comment.parentId);
    if (parent) {
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  }

  return roots;
}

function flattenReplies(comment: CommentNode): FlatReply[] {
  const replies: FlatReply[] = [];
  for (const child of comment.children) {
    replies.push({ comment: child, replyTo: authorLabel(comment) });
    replies.push(...flattenReplies(child));
  }
  return replies;
}

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const postId = Number(id);
  const { currentUser } = useUserStore();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [replyingTo, setReplyingTo] = useState<Comment | null>(null);
  const [expandedReplies, setExpandedReplies] = useState<Record<number, ReplyExpansion>>({});

  const { data: post, isLoading } = usePostDetail(postId);
  const { comments, isLoading: commentsLoading, createComment, isSubmitting } = useComments(postId);
  const { likePost, favoritePost, isLiking, isFavoriting } = usePostActions(postId);
  const { data: decisionLogs = [] } = useDecisionLogsForPost(postId);
  const { data: aiStatusSnapshot } = useAIStatus(postId);
  const { data: relatedDiscussions = [] } = useRelatedDiscussions(post);
  const retryAIReplies = useMutation({
    mutationFn: () => api.aiStatus.retry(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["aiStatus", postId] });
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
    },
    onError: (err) => setError(err instanceof Error ? err.message : "AI 回复重试失败"),
  });

  const overallStatus = aiStatusSnapshot?.overallStatus ?? (post ? statusFromPost(post.aiStatus) : "IDLE");
  const steps = useMemo(() => (post ? buildSteps(overallStatus, post.tags.length > 0) : []), [overallStatus, post]);
  const progress = useMemo(() => {
    if (!post) return 0;
    if (overallStatus === "COMPLETED") return 1;
    if (overallStatus === "RUNNING") return 0.75;
    if (overallStatus === "FAILED") return 0.5;
    return post.tags.length > 0 ? 0.25 : 0;
  }, [overallStatus, post]);
  const commentTree = useMemo(() => buildCommentTree(comments), [comments]);
  const aiReplyScoreByAgent = useMemo(() => {
    const scores = new Map<number, number>();
    for (const log of decisionLogs) {
      if (log.decision === "REPLY") scores.set(log.aiAgentId, log.willingnessScore);
    }
    return scores;
  }, [decisionLogs]);

  if (isLoading) {
    return (
      <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-xl md:px-margin-desktop">
        <p className="font-body-main text-cohere-muted">加载帖子中…</p>
      </main>
    );
  }

  if (!post) {
    return (
      <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-xl md:px-margin-desktop">
        <div className="card-base flex flex-col items-center gap-md p-xl text-center">
          <MaterialIcon name="search_off" size={48} className="text-cohere-muted" />
          <h2 className="font-feature-title text-cohere-primary">帖子不存在</h2>
          <Link to="/posts" className="btn-primary">
            返回帖子广场
          </Link>
        </div>
      </main>
    );
  }

  const handleSubmit = async (content: string, parentId: number | null = null) => {
    try {
      setError(null);
      await createComment({
        postId,
        parentId,
        content,
        author: { username: currentUser!.username, avatar: currentUser!.avatar, isAi: false },
      });
      setReplyingTo(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "评论发布失败");
    }
  };

  const withFallbackScore = (comment: CommentNode): CommentNode => {
    const fallbackScore = comment.author.aiAgentId ? aiReplyScoreByAgent.get(comment.author.aiAgentId) : undefined;
    return (
      comment.author.isAi && comment.willingnessScore === undefined && fallbackScore !== undefined
        ? { ...comment, willingnessScore: fallbackScore }
        : comment
    );
  };

  const renderCommentCard = (comment: CommentNode, replyTo?: string) => {
    const displayComment = withFallbackScore(comment);
    return comment.author.isAi ? (
      <AIComment comment={displayComment} replyTo={replyTo} onFollowup={setReplyingTo} />
    ) : (
      <HumanComment comment={displayComment} replyTo={replyTo} onReply={setReplyingTo} />
    );
  };

  const renderReplyEditor = (comment: CommentNode) =>
    replyingTo?.id === comment.id ? (
      <div className="ml-12">
        <CommentEditor
          onSubmit={(content) => handleSubmit(content, comment.id)}
          isSubmitting={isSubmitting}
          placeholder={`回复 ${comment.author.role ?? comment.author.username}…`}
        />
      </div>
    ) : null;

  const renderComment = (comment: CommentNode) => {
    const expansion = expandedReplies[comment.id];
    const replies = flattenReplies(comment);
    const visibleChildren =
      expansion === "all"
        ? replies
        : expansion === "preview"
          ? replies.slice(0, PREVIEW_REPLY_COUNT)
          : [];

    return (
      <div className="space-y-md">
        {renderCommentCard(comment)}
        {renderReplyEditor(comment)}

        {replies.length > 0 && (
          <>
            {!expansion && (
              <button
                type="button"
                onClick={() => setExpandedReplies((current) => ({ ...current, [comment.id]: "preview" }))}
                className="ml-12 inline-flex items-center gap-1 font-label-mono text-micro text-cohere-muted transition-colors hover:text-cohere-ink focus:outline-none focus-visible:underline"
              >
                <MaterialIcon name="expand_more" size={16} />
                展开 {replies.length} 条回复
              </button>
            )}
            {visibleChildren.length > 0 && (
              <div className="ml-6 space-y-md border-l border-cohere-hairline pl-md md:ml-12">
                {visibleChildren.map(({ comment: reply, replyTo }) => (
                  <div key={reply.id} className="space-y-md">
                    {renderCommentCard(reply, replyTo)}
                    {renderReplyEditor(reply)}
                  </div>
                ))}
                <div className="flex flex-wrap gap-md">
                  {expansion === "preview" && replies.length > PREVIEW_REPLY_COUNT && (
                    <button
                      type="button"
                      onClick={() => setExpandedReplies((current) => ({ ...current, [comment.id]: "all" }))}
                      className="inline-flex items-center gap-1 font-label-mono text-micro text-cohere-muted transition-colors hover:text-cohere-ink focus:outline-none focus-visible:underline"
                    >
                      <MaterialIcon name="unfold_more" size={16} />
                      继续展开全部
                    </button>
                  )}
                  <button
                    type="button"
                    onClick={() =>
                      setExpandedReplies((current) => {
                        const next = { ...current };
                        delete next[comment.id];
                        return next;
                      })
                    }
                    className="inline-flex items-center gap-1 font-label-mono text-micro text-cohere-muted transition-colors hover:text-cohere-ink focus:outline-none focus-visible:underline"
                  >
                    <MaterialIcon name="expand_less" size={16} />
                    收起回复
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    );
  };

  return (
    <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-lg md:px-margin-desktop animate-reveal-up">
      <div className="grid grid-cols-1 gap-gutter md:grid-cols-12">
        {/* Left column */}
        <div className="flex flex-col gap-section md:col-span-8">
          <article className="card-base p-lg md:p-xl">
            <div className="mb-lg flex items-center gap-sm">
              <span className="rounded border border-cohere-hairline bg-cohere-surface-container px-1 py-0.5 font-label-mono text-micro text-cohere-on-surface">
                {post.category}
              </span>
              <span className="font-label-mono text-micro text-cohere-on-surface-variant">
                {formatRelativeTime(post.createdAt)}发布
              </span>
            </div>

            <h1 className="mb-md font-headline-xl text-cohere-ink">{post.title}</h1>

            <div className="mb-lg flex items-center justify-between border-b border-cohere-hairline pb-md">
              <div className="flex items-center gap-sm">
                <img
                  src={post.author.avatar}
                  alt={post.author.username}
                  width={32}
                  height={32}
                  className="h-8 w-8 rounded-full bg-cohere-surface-container"
                />
                <div>
                  <div className="font-label-mono-bold text-cohere-on-surface">
                    {post.author.username}
                  </div>
                  <div className="font-micro text-cohere-on-surface-variant">{post.author.role ?? "研究员"}</div>
                </div>
              </div>
              <div className="flex gap-md font-label-mono text-micro text-cohere-on-surface-variant">
                <span className="flex items-center gap-1">
                  <MaterialIcon name="visibility" size={16} /> {formatCount(post.viewCount)}
                </span>
                <span className="flex items-center gap-1">
                  <MaterialIcon name="forum" size={16} /> {post.commentCount}
                </span>
                <button
                  type="button"
                  className="flex items-center gap-1 transition-colors hover:text-cohere-ink disabled:opacity-60"
                  disabled={isLiking}
                  onClick={() => likePost().catch((err) => setError(err instanceof Error ? err.message : "点赞失败"))}
                >
                  <MaterialIcon name="thumb_up" size={16} /> 点赞
                </button>
                <button
                  type="button"
                  className="flex items-center gap-1 transition-colors hover:text-cohere-ink disabled:opacity-60"
                  disabled={isFavoriting}
                  onClick={() =>
                    favoritePost().catch((err) => setError(err instanceof Error ? err.message : "收藏失败"))
                  }
                >
                  <MaterialIcon name="bookmark" size={16} /> 收藏
                </button>
              </div>
            </div>

            <SafeMarkdown content={post.content} className="font-body-large text-cohere-ink space-y-md" />
          </article>

          {/* Comments */}
          <section className="mt-md">
            <h3 className="mb-lg font-headline-lg text-cohere-ink">讨论区</h3>

            {error && (
              <div className="mb-md">
                <AlertBar tone="error" message={error} onClose={() => setError(null)} />
              </div>
            )}

            <CommentEditor
              onSubmit={(content) => handleSubmit(content)}
              isSubmitting={isSubmitting}
            />

            {commentsLoading ? (
              <p className="font-body-main text-cohere-on-surface-variant">加载评论中…</p>
            ) : commentTree.length === 0 ? (
              <p className="font-body-main text-cohere-on-surface-variant">
                还没有评论。发布第一条评论，或 @某个 AI 角色让它参与。
              </p>
            ) : (
              <div className="space-y-lg">
                {commentTree.length > 20 ? (
                  <Virtuoso
                    useWindowScroll
                    data={commentTree}
                    itemContent={(_i, comment) => renderComment(comment)}
                    style={{ minHeight: 300 }}
                  />
                ) : (
                  commentTree.map((comment) => <div key={comment.id}>{renderComment(comment)}</div>)
                )}
              </div>
            )}
          </section>
        </div>

        {/* Right column */}
        <div className="flex flex-col gap-lg md:col-span-4">
          <Link
            to={`/posts/${postId}#decision-logs`}
            className="btn-primary flex items-center justify-center gap-sm"
          >
            <MaterialIcon name="terminal" /> 查看 AI 决策日志
          </Link>

          <div id="decision-logs" className="scroll-mt-24">
            <ParticipatingAI logs={decisionLogs} comments={comments} />
          </div>

          <PostTags tags={post.tags} />

          <AIProcessingStatus
            steps={steps}
            progress={progress}
            status={overallStatus}
            canRetry={(aiStatusSnapshot?.retryableCount ?? 0) > 0}
            retrying={retryAIReplies.isPending}
            onRetry={() => retryAIReplies.mutate()}
            summary={{
              done: aiStatusSnapshot?.completedCount ?? decisionLogs.filter((l) => l.decision === "REPLY").length,
              running: aiStatusSnapshot?.runningCount ?? (overallStatus === "RUNNING" ? 1 : 0),
              failed: aiStatusSnapshot?.failedCount ?? decisionLogs.filter((l) => l.decision === "FAILED").length,
            }}
          />

          <RelatedDiscussions discussions={relatedDiscussions} />
        </div>
      </div>
    </main>
  );
}

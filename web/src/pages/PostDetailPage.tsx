import { useMemo, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { usePostDetail } from "../hooks/usePosts";
import { useComments } from "../hooks/useComments";
import { useDecisionLogsForPost } from "../hooks/useDecisionLogs";
import { usePostActions } from "../hooks/usePostActions";
import { useUserStore } from "../stores/useUserStore";
import { ProcessingStep } from "../api/types";
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

const RELATED = [
  { id: 901, title: "RAG 系统中检索器与生成器的上下文匹配问题" },
  { id: 902, title: "Mamba 模型在长文本摘要任务上的初步评测" },
  { id: 903, title: "如何优雅地在生产环境部署 1M Context 模型" },
];

/** Derive the 4-step processing pipeline from the post's current AI status. */
function buildSteps(status: "PENDING" | "PROCESSING" | "COMPLETED"): ProcessingStep[] {
  const base: ProcessingStep[] = [
    { key: "tags", label: "分析帖子标签", detail: "已提取核心概念", icon: "label", state: "done" },
    { key: "score", label: "计算 AI 回答意愿分", detail: "评估各代理阈值", icon: "analytics", state: "done" },
    { key: "generate", label: "生成 AI 回复", detail: "推理引擎运行中…", icon: "hourglass_empty", state: "active" },
    { key: "write", label: "写入评论区", detail: "等待挂起", icon: "edit_note", state: "pending" },
  ];

  if (status === "PENDING") {
    return [
      { ...base[0], state: "active", detail: "正在提取核心概念…" },
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
  // PROCESSING — the default above already shows steps 1-2 done, 3 active, 4 pending.
  return base;
}

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const postId = Number(id);
  const { currentUser } = useUserStore();
  const [error, setError] = useState<string | null>(null);

  const { data: post, isLoading } = usePostDetail(postId);
  const { comments, isLoading: commentsLoading, createComment, isSubmitting } = useComments(postId);
  const { likePost, favoritePost, isLiking, isFavoriting } = usePostActions(postId);
  const { data: decisionLogs = [] } = useDecisionLogsForPost(postId);

  const steps = useMemo(() => (post ? buildSteps(post.aiStatus) : []), [post]);
  const progress = useMemo(() => {
    if (!post) return 0;
    if (post.aiStatus === "COMPLETED") return 1;
    if (post.aiStatus === "PENDING") return 0.25;
    return 0.75;
  }, [post]);

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

  const handleSubmit = async (content: string) => {
    try {
      setError(null);
      await createComment({
        postId,
        parentId: null,
        content,
        author: { username: currentUser!.username, avatar: currentUser!.avatar, isAi: false },
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "评论发布失败");
    }
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

            <CommentEditor onSubmit={handleSubmit} isSubmitting={isSubmitting} />

            {commentsLoading ? (
              <p className="font-body-main text-cohere-on-surface-variant">加载评论中…</p>
            ) : comments.length === 0 ? (
              <p className="font-body-main text-cohere-on-surface-variant">
                还没有评论。发布第一条评论，或 @某个 AI 角色让它参与。
              </p>
            ) : (
              <div className="space-y-lg">
                {comments.length > 20 ? (
                  <Virtuoso
                    useWindowScroll
                    data={comments}
                    itemContent={(_i, comment) =>
                      comment.author.isAi ? (
                        <AIComment key={comment.id} comment={comment} />
                      ) : (
                        <HumanComment key={comment.id} comment={comment} />
                      )
                    }
                    style={{ minHeight: 300 }}
                  />
                ) : (
                  comments.map((comment) =>
                    comment.author.isAi ? (
                      <AIComment key={comment.id} comment={comment} />
                    ) : (
                      <HumanComment key={comment.id} comment={comment} />
                    ),
                  )
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
            <ParticipatingAI logs={decisionLogs} />
          </div>

          <PostTags tags={post.tags} />

          <AIProcessingStatus
            steps={steps}
            progress={progress}
            active={post.aiStatus === "PROCESSING"}
            summary={{
              done: decisionLogs.filter((l) => l.decision === "REPLY").length,
              running: post.aiStatus === "PROCESSING" ? 1 : 0,
              failed: decisionLogs.filter((l) => l.decision === "FAILED").length,
            }}
          />

          <RelatedDiscussions discussions={RELATED} />
        </div>
      </div>
    </main>
  );
}

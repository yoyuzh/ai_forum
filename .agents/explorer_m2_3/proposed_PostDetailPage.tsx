import React, { useState, useRef } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import ReactMarkdown from "react-markdown";
import DOMPurify from "dompurify";
import { usePostDetail } from "../hooks/usePosts";
import { useComments } from "../hooks/useComments";
import { useUserStore } from "../stores/useUserStore";
import { useConnectionStore } from "../stores/useConnectionStore";
import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

export function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const postId = Number(id);

  const { data: post, isLoading: isPostLoading } = usePostDetail(postId);
  const { comments, isLoading: isCommentsLoading, createComment, isSubmitting } = useComments(postId);
  const { currentUser } = useUserStore();
  const { sseStatus } = useConnectionStore();

  const [commentText, setCommentText] = useState("");
  const [parentId, setParentId] = useState<number | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Fetch decision logs and tasks for real-time AI status logs
  const { data: decisionLogs = [], refetch: refetchLogs } = useQuery({
    queryKey: ["decisionLogs", postId],
    queryFn: api.decisionLogs.list,
  });

  const { data: allTasks = [], refetch: refetchTasks } = useQuery({
    queryKey: ["tasks", postId],
    queryFn: api.tasks.list,
  });

  // Filter logs and tasks for this specific post
  const postLogs = decisionLogs.filter((log) => log.postId === postId);
  const postTasks = allTasks.filter((task) => task.postId === postId);

  if (isPostLoading) {
    return <div className="py-xl text-center text-muted font-body-main">正在加载帖子...</div>;
  }

  if (!post) {
    return (
      <div className="py-xl text-center text-muted font-body-main">
        未找到该帖子
        <button onClick={() => navigate("/")} className="block mx-auto mt-md btn-pill-outline">
          返回首页
        </button>
      </div>
    );
  }

  const handleTextareaInsert = (before: string, after: string = "") => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const text = textarea.value;
    const selected = text.substring(start, end);
    const replacement = before + selected + after;

    setCommentText(text.substring(0, start) + replacement + text.substring(end));
    
    // Restore focus
    setTimeout(() => {
      textarea.focus();
      textarea.setSelectionRange(start + before.length, start + before.length + selected.length);
    }, 0);
  };

  const handleMentionAI = (agentName: string) => {
    handleTextareaInsert(`@${agentName} `);
  };

  const handleAskFollowup = (agentName: string, commentId: number) => {
    setParentId(commentId);
    handleTextareaInsert(`@${agentName} `);
  };

  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;

    try {
      await createComment({
        postId,
        parentId,
        content: commentText,
        author: {
          username: currentUser.username,
          avatar: currentUser.avatar,
          isAi: false,
        },
      });
      setCommentText("");
      setParentId(null);
      
      // Proactively refresh logs/tasks
      setTimeout(() => {
        refetchLogs();
        refetchTasks();
      }, 500);
    } catch (err) {
      console.error(err);
    }
  };

  // Safe markdown render helper
  const renderMarkdown = (markdown: string) => {
    const cleanMarkdown = DOMPurify.sanitize(markdown);
    return (
      <div className="markdown-body">
        <ReactMarkdown>{cleanMarkdown}</ReactMarkdown>
      </div>
    );
  };

  // Determine current active simulation step
  const activeTaskCount = postTasks.filter((t) => t.status === "PENDING" || t.status === "PROCESSING").length;
  const completedTaskCount = postTasks.filter((t) => t.status === "COMPLETED").length;
  const failedTaskCount = postTasks.filter((t) => t.status === "FAILED").length;

  return (
    <main className="flex-grow max-w-7xl mx-auto w-full px-margin-mobile md:px-margin-desktop py-lg grid grid-cols-1 lg:grid-cols-12 gap-gutter">
      {/* Left Column (Content, Comments) */}
      <div className="lg:col-span-8 flex flex-col gap-section">
        {/* Post Content */}
        <article className="bg-surface-container-lowest border border-hairline rounded-[16px] p-lg md:p-xl">
          <div className="flex items-center gap-sm mb-lg">
            <span className="bg-surface-container text-on-surface font-label-mono text-micro px-sm py-[2px] rounded border border-hairline">
              {post.category}
            </span>
            <span className="text-muted font-label-mono text-micro">
              {new Date(post.createdAt).toLocaleString()}
            </span>
          </div>

          <h1 className="font-headline-xl text-headline-xl text-ink mb-md">{post.title}</h1>

          <div className="flex items-center justify-between border-b border-hairline pb-md mb-lg">
            <div className="flex items-center gap-sm">
              <img
                className="w-8 h-8 rounded-full bg-surface-container object-cover border"
                src={post.author.avatar}
                alt={post.author.username}
              />
              <div>
                <div className="font-label-mono-bold text-label-mono-bold text-on-surface">
                  {post.author.username}
                </div>
                <div className="font-micro text-muted">发帖者</div>
              </div>
            </div>
            <div className="flex gap-md font-label-mono text-micro text-muted">
              <span className="flex items-center gap-xs">
                <span className="material-symbols-outlined text-[16px]">forum</span>{" "}
                {comments.length}
              </span>
            </div>
          </div>

          {/* Body Render with markdown support */}
          <div className="font-body-large text-body-large text-ink space-y-md">
            {renderMarkdown(post.content)}
          </div>
        </article>

        {/* Discussion Area */}
        <section className="mt-md">
          <h3 className="font-headline-lg text-headline-lg text-ink mb-lg">讨论区</h3>

          {/* Comment Input */}
          <form
            onSubmit={handleSubmitComment}
            className="bg-surface-container-lowest border border-hairline rounded-[16px] p-md mb-xl flex flex-col focus-within:border-primary focus-within:shadow-[0_0_0_1px_rgba(0,0,0,1)] transition-all"
          >
            {parentId && (
              <div className="flex items-center justify-between bg-surface-container-low px-sm py-xs rounded mb-sm text-micro text-muted">
                <span>正在回复特定评论 (子评论模式)</span>
                <button
                  type="button"
                  onClick={() => setParentId(null)}
                  className="text-error hover:underline"
                >
                  取消
                </button>
              </div>
            )}
            <textarea
              ref={textareaRef}
              value={commentText}
              onChange={(e) => setCommentText(e.target.value)}
              className="w-full bg-transparent border-none outline-none resize-none font-body-main text-body-main text-ink placeholder:text-muted min-h-[80px] p-0 focus:ring-0"
              placeholder="输入你的评论，输入 @ 唤醒特定 AI Agent..."
            />
            <div className="flex justify-between items-center mt-sm pt-sm border-t border-surface-variant">
              <div className="flex gap-sm">
                <button
                  type="button"
                  onClick={() => handleTextareaInsert("**", "**")}
                  className="text-muted hover:text-ink transition-colors p-xs"
                  title="加粗"
                >
                  <span className="material-symbols-outlined">format_bold</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleTextareaInsert("`", "`")}
                  className="text-muted hover:text-ink transition-colors p-xs"
                  title="代码"
                >
                  <span className="material-symbols-outlined">code</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleTextareaInsert("@")}
                  className="text-on-tertiary-fixed-variant hover:text-tertiary transition-colors p-xs font-label-mono-bold text-xs"
                  title="艾特"
                >
                  @AI
                </button>
              </div>
              <button
                type="submit"
                disabled={isSubmitting || !commentText.trim()}
                className="bg-primary text-on-primary rounded-full px-lg py-sm font-label-mono-bold hover:bg-ink transition-colors disabled:opacity-50"
              >
                {isSubmitting ? "发布中..." : "发布评论"}
              </button>
            </div>
          </form>

          {/* Comment List using Virtuoso */}
          {isCommentsLoading ? (
            <div className="py-md text-center text-muted font-body-main">加载评论中...</div>
          ) : comments.length === 0 ? (
            <div className="py-md text-center text-muted font-body-main">暂无讨论，快来发表第一条评论吧！</div>
          ) : (
            <div className="space-y-lg">
              <Virtuoso
                useWindowScroll
                data={comments}
                itemContent={(index, comment) => {
                  const isAi = comment.author.isAi;
                  
                  // Render thread connection line if it's a child comment
                  return (
                    <div key={comment.id} className="flex gap-md relative mb-lg">
                      {comment.parentId && (
                        <div className="absolute -left-[17px] -top-12 bottom-0 w-[1px] border-l border-dotted border-hairline hidden md:block"></div>
                      )}
                      
                      <div className="flex-shrink-0 z-10">
                        {isAi ? (
                          <div className="w-10 h-10 rounded-[22px] bg-secondary text-on-secondary flex items-center justify-center border-2 border-surface-container-lowest shadow-sm">
                            <span className="material-symbols-outlined">psychology</span>
                          </div>
                        ) : (
                          <img
                            className="w-10 h-10 rounded-full object-cover border border-hairline"
                            src={comment.author.avatar}
                            alt={comment.author.username}
                          />
                        )}
                      </div>

                      <div
                        className={`flex-1 border rounded-tr-[16px] rounded-br-[16px] rounded-bl-[16px] p-md transition-colors ${
                          isAi
                            ? "bg-[#f5f7f6] border-[#e0e5e3]"
                            : "bg-surface-container-lowest border-hairline"
                        }`}
                      >
                        <div className="flex items-center flex-wrap gap-sm mb-md">
                          <div
                            className={`font-label-mono-bold text-label-mono-bold ${
                              isAi ? "text-secondary" : "text-ink"
                            }`}
                          >
                            {comment.author.username} {isAi && "· AI"}
                          </div>
                          {isAi && (
                            <span className="bg-secondary-container text-on-secondary-container font-label-mono text-[10px] px-sm py-[2px] rounded">
                              系统自动回复
                            </span>
                          )}
                          <div className="flex-1 text-right font-micro text-muted">
                            {new Date(comment.createdAt).toLocaleTimeString([], {
                              hour: "2-digit",
                              minute: "2-digit",
                            })}
                          </div>
                        </div>

                        <div className="font-body-main text-body-main text-ink space-y-sm">
                          {renderMarkdown(comment.content)}
                        </div>

                        <div className="mt-md pt-sm border-t border-hairline flex justify-between items-center">
                          {isAi && (
                            <div className="font-label-mono text-micro text-muted flex items-center gap-xs">
                              <span className="material-symbols-outlined text-[14px]">analytics</span>
                              意愿分: {Math.round((postLogs.find(l => l.aiAgentId === comment.author.aiAgentId)?.willingnessScore || 0.85) * 100)}/100
                            </div>
                          )}
                          <div className="flex gap-md ml-auto">
                            <button
                              type="button"
                              onClick={() => handleAskFollowup(comment.author.username, comment.id)}
                              className="bg-surface-container-lowest border border-hairline text-ink rounded-full px-md py-sm font-label-mono-bold text-micro hover:border-secondary hover:text-secondary transition-colors flex items-center gap-xs"
                            >
                              {isAi ? "继续追问" : "回复"}
                              <span className="material-symbols-outlined text-[14px]">arrow_forward</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                }}
              />
            </div>
          )}
        </section>
      </div>

      {/* Right Column (Sidebar) */}
      <div className="lg:col-span-4 flex flex-col gap-lg">
        {/* Primary Action */}
        <button
          onClick={() => navigate("/ai-agents")}
          className="w-full bg-primary text-on-primary rounded-full py-md font-label-mono-bold flex items-center justify-center gap-sm hover:bg-ink transition-colors shadow-none border-none cursor-pointer"
        >
          <span className="material-symbols-outlined">terminal</span>
          管理 AI 决策规则
        </button>

        {/* Participating AI Agents status roster */}
        <div className="bg-surface-container-lowest border border-hairline rounded-[16px] p-lg">
          <h3 className="font-feature-title text-[18px] text-ink mb-md">参与本帖的 AI</h3>
          <div className="space-y-md">
            {postLogs.map((log) => {
              const task = postTasks.find((t) => t.aiAgentId === log.aiAgentId);
              const isReplied = log.decision === "REPLY";
              const isIgnored = log.decision === "IGNORE";

              return (
                <div key={log.id} className="flex items-center justify-between group">
                  <div className="flex items-center gap-sm">
                    <div className="w-8 h-8 rounded-[16px] bg-secondary-container text-on-secondary-container flex items-center justify-center">
                      <span className="material-symbols-outlined text-[16px]">psychology</span>
                    </div>
                    <div>
                      <div className="font-label-mono-bold text-label-mono-bold text-ink">{log.aiAgentName}</div>
                      <div className="font-micro text-muted flex items-center gap-xs">
                        {isReplied && (
                          <>
                            <span className="w-1.5 h-1.5 rounded-full bg-secondary"></span>
                            {task?.status === "PROCESSING" ? "正在响应" : "已回复"}
                          </>
                        )}
                        {isIgnored && (
                          <>
                            <span className="w-1.5 h-1.5 rounded-full bg-outline"></span>
                            忽略 (意愿低)
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-label-mono text-[14px] text-ink">{Math.round(log.willingnessScore * 100)}</div>
                    <div className="font-micro text-muted">意愿分</div>
                  </div>
                </div>
              );
            })}
            {postLogs.length === 0 && (
              <div className="text-center font-micro text-muted py-xs">暂无 AI 匹配该帖子</div>
            )}
          </div>
        </div>

        {/* AI Status / Processing Timeline Panel */}
        <section className="bg-surface border border-hairline rounded-[16px] p-md relative overflow-hidden mb-lg">
          <div className="absolute top-0 left-0 w-full h-[4px] bg-surface-variant">
            {post.aiStatus === "PROCESSING" && (
              <div className="h-full bg-on-tertiary-fixed-variant w-3/4 animate-pulse"></div>
            )}
            {post.aiStatus === "COMPLETED" && <div className="h-full bg-secondary w-full"></div>}
          </div>
          <div className="flex justify-between items-center mb-md mt-sm">
            <h2 className="font-feature-title text-[18px] text-ink">AI 处理状态</h2>
            {post.aiStatus === "PROCESSING" && (
              <div className="bg-tertiary-fixed text-on-tertiary-fixed font-label-mono-bold text-[10px] px-sm py-xs rounded-full flex items-center gap-xs">
                <span className="w-1.5 h-1.5 rounded-full bg-on-tertiary-fixed-variant animate-ping"></span>
                分析中
              </div>
            )}
            {post.aiStatus === "COMPLETED" && (
              <div className="bg-success-green text-secondary font-label-mono-bold text-[10px] px-sm py-xs rounded-full flex items-center gap-xs">
                已完成
              </div>
            )}
            {post.aiStatus === "PENDING" && (
              <div className="bg-surface-container text-muted font-label-mono-bold text-[10px] px-sm py-xs rounded-full flex items-center gap-xs">
                队列中
              </div>
            )}
          </div>

          <div className="flex flex-col gap-md relative z-10">
            {/* Step 1: Matching / Scanning */}
            <div className="flex items-center gap-md">
              <div className="w-8 h-8 rounded-full bg-surface-container-highest text-muted border border-hairline flex-shrink-0 flex items-center justify-center">
                <span className="material-symbols-outlined text-[16px] text-secondary">check</span>
              </div>
              <div>
                <div className="font-label-mono-bold text-micro text-on-surface">分析与标签提取</div>
                <div className="font-caption text-[12px] text-muted">语义解析已完成</div>
              </div>
            </div>

            {/* Step 2: Willingness evaluation */}
            <div className="flex items-center gap-md">
              <div className={`w-8 h-8 rounded-full border flex-shrink-0 flex items-center justify-center ${
                postLogs.length > 0 ? "bg-surface-container-highest text-secondary border-hairline" : "bg-surface border-hairline text-muted"
              }`}>
                {postLogs.length > 0 ? (
                  <span className="material-symbols-outlined text-[16px] text-secondary">check</span>
                ) : (
                  <span className="w-1.5 h-1.5 rounded-full bg-current"></span>
                )}
              </div>
              <div>
                <div className="font-label-mono-bold text-micro text-on-surface">评估回答意愿</div>
                <div className="font-caption text-[12px] text-muted">
                  {postLogs.length > 0 ? `匹配到 ${postLogs.length} 个角色判定` : "评估中"}
                </div>
              </div>
            </div>

            {/* Step 3: Response generation */}
            <div className="flex items-center gap-md">
              <div className={`w-8 h-8 rounded-full border flex-shrink-0 flex items-center justify-center ${
                post.aiStatus === "COMPLETED"
                  ? "bg-surface-container-highest text-secondary border-hairline"
                  : post.aiStatus === "PROCESSING"
                  ? "bg-on-tertiary-fixed-variant text-on-tertiary border-on-tertiary-fixed-variant"
                  : "bg-surface border-hairline text-muted"
              }`}>
                {post.aiStatus === "COMPLETED" ? (
                  <span className="material-symbols-outlined text-[16px] text-secondary">check</span>
                ) : post.aiStatus === "PROCESSING" ? (
                  <span className="material-symbols-outlined text-[16px] animate-spin">hourglass_empty</span>
                ) : (
                  <span className="w-1.5 h-1.5 rounded-full bg-current"></span>
                )}
              </div>
              <div>
                <div className="font-label-mono-bold text-micro text-ink">生成 AI 回复</div>
                <div className="font-caption text-[12px] text-muted">
                  {post.aiStatus === "COMPLETED"
                    ? "回复生成完毕"
                    : post.aiStatus === "PROCESSING"
                    ? "推理引擎运行中"
                    : "等待触发"}
                </div>
              </div>
            </div>

            {/* Step 4: Write-back Comments */}
            <div className={`flex items-center gap-md ${post.aiStatus !== "COMPLETED" ? "opacity-50" : ""}`}>
              <div className={`w-8 h-8 rounded-full border flex-shrink-0 flex items-center justify-center ${
                post.aiStatus === "COMPLETED" ? "bg-surface-container-highest text-secondary border-hairline" : "bg-surface border-hairline text-muted"
              }`}>
                {post.aiStatus === "COMPLETED" ? (
                  <span className="material-symbols-outlined text-[16px] text-secondary">check</span>
                ) : (
                  <span className="material-symbols-outlined text-[16px]">edit_note</span>
                )}
              </div>
              <div>
                <div className="font-label-mono-bold text-micro text-on-surface">写入评论区</div>
                <div className="font-caption text-[12px] text-muted">
                  {post.aiStatus === "COMPLETED" ? "评论已写入" : "等待生成中"}
                </div>
              </div>
            </div>
          </div>

          {/* Stats Footer */}
          <div className="mt-md pt-md border-t border-hairline flex justify-between font-label-mono text-[10px]">
            <div className="flex flex-col">
              <span className="text-muted">已完成任务</span>
              <span className="text-ink font-label-mono-bold">{completedTaskCount}</span>
            </div>
            <div className="flex flex-col">
              <span className="text-muted">运行中</span>
              <span className="text-on-tertiary-fixed-variant font-label-mono-bold">{activeTaskCount}</span>
            </div>
            <div className="flex flex-col">
              <span className="text-muted">失败</span>
              <span className="text-error font-label-mono-bold">{failedTaskCount}</span>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}

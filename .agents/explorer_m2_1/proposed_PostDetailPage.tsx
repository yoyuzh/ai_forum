import React, { useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Virtuoso } from "react-virtuoso";
import ReactMarkdown from "react-markdown";
import DOMPurify from "dompurify";
import { usePostDetail } from "../hooks/usePosts";
import { useComments } from "../hooks/useComments";
import { useAgents } from "../hooks/useAgents";
import { api } from "../api/client";
import { useUserStore } from "../stores/useUserStore";
import { ArrowLeft, MessageSquare, Send, CheckCircle2, AlertCircle, HelpCircle, Loader2 } from "lucide-react";

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const postId = Number(id);
  const navigate = useNavigate();
  const { currentUser } = useUserStore();

  const { data: post, isLoading: postLoading, error: postError } = usePostDetail(postId);
  const { comments, isLoading: commentsLoading, createComment, isSubmitting } = useComments(postId);
  const { agents } = useAgents();

  const [commentText, setCommentText] = useState("");

  // Fetch tasks and decision logs to show real-time processing details
  const { data: allTasks = [], refetch: refetchTasks } = useQuery({
    queryKey: ["tasks"],
    queryFn: api.tasks.list,
    refetchInterval: 1000 // Poll tasks during simulation
  });

  const { data: allLogs = [] } = useQuery({
    queryKey: ["decisionLogs"],
    queryFn: api.decisionLogs.list,
    refetchInterval: 1000 // Poll decision logs during simulation
  });

  const postTasks = allTasks.filter(t => t.postId === postId);
  const postLogs = allLogs.filter(t => t.postId === postId);

  if (postLoading) {
    return (
      <div className="min-h-screen bg-cohere-canvas flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-cohere-action-blue" />
        <span className="ml-2 font-sans text-sm text-cohere-slate">Loading post details...</span>
      </div>
    );
  }

  if (postError || !post) {
    return (
      <div className="min-h-screen bg-cohere-canvas flex flex-col items-center justify-center p-6 text-center">
        <AlertCircle className="w-12 h-12 text-cohere-error mb-4" />
        <h2 className="font-card-heading text-cohere-primary font-semibold mb-2">Post Not Found</h2>
        <p className="font-body text-cohere-slate mb-6">The discussion topic you are looking for does not exist or has been removed.</p>
        <button onClick={() => navigate("/")} className="btn-primary">
          Back to Feed
        </button>
      </div>
    );
  }

  // Pre-flatten comments for virtualized rendering of hierarchical comments
  const buildFlatComments = () => {
    const flat: Array<{ comment: any; depth: number }> = [];
    const root = comments.filter(c => c.parentId === null);
    
    const recurse = (parentComment: any, depth: number) => {
      flat.push({ comment: parentComment, depth });
      const children = comments.filter(c => c.parentId === parentComment.id);
      children.forEach(child => recurse(child, depth + 1));
    };

    root.forEach(c => recurse(c, 0));
    return flat;
  };

  const flatComments = buildFlatComments();

  // Handle Comment Submission
  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;

    try {
      await createComment({
        postId,
        parentId: null, // Simple root comments from user
        content: commentText,
        author: {
          username: currentUser.username,
          avatar: currentUser.avatar,
          isAi: false
        }
      });
      setCommentText("");
      refetchTasks();
    } catch (err) {
      console.error("Failed to add comment:", err);
    }
  };

  // Helper to trigger followup question simulation
  const handleFollowup = async (parentCommentId: number, agentId: number, agentName: string, agentAvatar: string) => {
    try {
      await createComment({
        postId,
        parentId: parentCommentId,
        content: `I'd like to ask @${agentName} to elaborate on this further. Can you dive deeper into this point?`,
        author: {
          username: currentUser.username,
          avatar: currentUser.avatar,
          isAi: false
        }
      });
      refetchTasks();
    } catch (err) {
      console.error("Failed to trigger followup:", err);
    }
  };

  // Sanitize Markdown
  const renderMarkdown = (text: string) => {
    const cleanHTML = DOMPurify.sanitize(text);
    return <ReactMarkdown className="prose max-w-none font-body text-cohere-ink text-sm leading-relaxed space-y-3">{cleanHTML}</ReactMarkdown>;
  };

  return (
    <div className="min-h-screen bg-cohere-canvas flex flex-col text-cohere-ink font-sans">
      
      {/* Header breadcrumb bar */}
      <nav className="border-b border-cohere-hairline py-4 px-6 md:px-12 max-w-7xl mx-auto w-full flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 text-xs font-mono-label text-cohere-slate hover:text-cohere-primary transition-colors">
          <ArrowLeft className="w-4 h-4" />
          Back to feed
        </Link>
        <span className="font-mono-label text-xs text-cohere-muted">
          Post #{post.id}
        </span>
      </nav>

      {/* Main Grid */}
      <main className="max-w-7xl mx-auto w-full px-6 md:px-12 py-8 grid grid-cols-1 lg:grid-cols-12 gap-8">
        
        {/* Left Column: Post Content & Comments (8 cols) */}
        <div className="lg:col-span-8 flex flex-col gap-8">
          
          {/* Post Content Area */}
          <article className="bg-white border border-cohere-hairline rounded-md p-6 md:p-8">
            <div className="flex items-center gap-2 mb-4">
              <span className="font-mono-label text-xs bg-cohere-soft-stone px-2 py-0.5 rounded-sm text-cohere-slate">
                {post.category}
              </span>
              <span className="text-cohere-hairline font-thin">•</span>
              <span className="font-micro text-xs text-cohere-muted">
                Published {new Date(post.createdAt).toLocaleString()}
              </span>
            </div>

            <h1 className="font-sans text-3xl md:text-4xl font-normal leading-[1.2] -tracking-[0.5px] text-cohere-black mb-6">
              {post.title}
            </h1>

            {/* Author */}
            <div className="flex items-center gap-3 border-b border-cohere-hairline pb-6 mb-6">
              <img
                src={post.author.avatar}
                alt={post.author.username}
                className="w-10 h-10 rounded-full object-cover border border-cohere-hairline"
              />
              <div>
                <div className="font-mono-label text-xs font-semibold text-cohere-primary">
                  {post.author.username}
                </div>
                <div className="font-micro text-cohere-muted">Human Contributor</div>
              </div>
            </div>

            {/* Markdown Body */}
            <div className="min-h-[150px] leading-relaxed">
              {renderMarkdown(post.content)}
            </div>
          </article>

          {/* Discussion comments section */}
          <section className="flex flex-col gap-6">
            <h3 className="font-sans text-xl font-semibold text-cohere-black">
              Comments ({comments.length})
            </h3>

            {/* New Comment Input */}
            <form onSubmit={handleSubmitComment} className="bg-white border border-cohere-hairline rounded-md p-4 flex flex-col gap-3 focus-within:border-cohere-primary transition-colors">
              <textarea
                value={commentText}
                onChange={(e) => setCommentText(e.target.value)}
                placeholder="Write a comment... Add '@ArchTechLead' or other agent name to trigger manual mention response."
                rows={3}
                className="w-full bg-transparent resize-none outline-none border-none font-sans text-sm focus:ring-0 p-0 text-cohere-ink"
              />
              <div className="flex justify-between items-center border-t border-cohere-hairline pt-3">
                <div className="flex gap-2">
                  {agents.filter(a => a.active).map(a => (
                    <button
                      key={a.id}
                      type="button"
                      onClick={() => setCommentText(prev => prev + ` @${a.name} `)}
                      className="font-mono-label text-[11px] bg-cohere-soft-stone text-cohere-slate hover:bg-cohere-hairline px-2 py-0.5 rounded-sm"
                    >
                      @{a.name}
                    </button>
                  ))}
                </div>
                <button
                  type="submit"
                  disabled={isSubmitting || !commentText.trim()}
                  className="btn-primary font-button text-xs py-2 px-4 flex items-center gap-1.5 disabled:opacity-40"
                >
                  {isSubmitting ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Send className="w-3.5 h-3.5" />
                  )}
                  Post Comment
                </button>
              </div>
            </form>

            {/* Virtualized Comment List */}
            {commentsLoading ? (
              <div className="text-center text-cohere-muted font-sans py-8">
                Loading comments...
              </div>
            ) : flatComments.length === 0 ? (
              <div className="text-center text-cohere-muted font-sans py-8 border border-dashed border-cohere-hairline rounded-md">
                No replies yet. Be the first to start the discussion!
              </div>
            ) : (
              <div className="h-[600px] w-full">
                <Virtuoso
                  useWindowScroll
                  data={flatComments}
                  itemContent={(index, item) => {
                    const c = item.comment;
                    const isAi = c.author.isAi;
                    const log = postLogs.find(l => l.commentId === c.id);
                    
                    return (
                      <div className="pb-4" style={{ paddingLeft: `${item.depth * 32}px` }}>
                        <div className={`relative flex gap-4 p-4 rounded-md border ${
                          isAi 
                            ? "bg-cohere-pale-green/20 border-cohere-pale-green text-cohere-ink" 
                            : "bg-white border-cohere-hairline"
                        }`}>
                          {/* Thread guide line if nested */}
                          {item.depth > 0 && (
                            <div className="absolute -left-4 top-0 bottom-0 w-[1px] border-l border-dotted border-cohere-hairline" />
                          )}

                          <img
                            src={c.author.avatar}
                            alt={c.author.username}
                            className={`w-9 h-9 rounded-full object-cover border border-cohere-hairline ${
                              isAi ? "bg-cohere-secondary-container" : "bg-cohere-soft-stone"
                            }`}
                          />
                          
                          <div className="flex-grow min-w-0">
                            <div className="flex flex-wrap items-center justify-between gap-2 mb-2">
                              <div className="flex items-center gap-2">
                                <span className={`font-mono-label text-xs font-semibold ${
                                  isAi ? "text-cohere-deep-green" : "text-cohere-primary"
                                }`}>
                                  {c.author.username} {isAi && "• AI"}
                                </span>
                                {isAi && log && (
                                  <span className="font-micro text-[10px] bg-cohere-secondary-container text-cohere-deep-green px-1.5 py-0.5 rounded-sm">
                                    Trigger: {log.triggerType}
                                  </span>
                                )}
                              </div>
                              <span className="font-micro text-xs text-cohere-muted">
                                {new Date(c.createdAt).toLocaleTimeString()}
                              </span>
                            </div>

                            {/* Markdown Render */}
                            <div className="mb-3">
                              {renderMarkdown(c.content)}
                            </div>

                            {/* Actions / AI Stats Panel */}
                            <div className="flex flex-wrap items-center justify-between pt-2 border-t border-dashed border-cohere-hairline/60">
                              {isAi && log ? (
                                <div className="font-mono-label text-[10px] text-cohere-slate flex items-center gap-1">
                                  Willingness Score: <strong>{(log.willingnessScore * 100).toFixed(0)}/100</strong>
                                </div>
                              ) : (
                                <div />
                              )}

                              {isAi && (
                                <button
                                  onClick={() => handleFollowup(c.id, c.author.aiAgentId!, c.author.username, c.author.avatar)}
                                  className="font-mono-label text-[11px] text-cohere-action-blue border border-cohere-action-blue/20 bg-cohere-pale-blue px-2.5 py-1 rounded-sm hover:bg-cohere-action-blue hover:text-white transition-colors flex items-center gap-1"
                                >
                                  Ask Followup
                                </button>
                              )}
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

        {/* Right Column: AI Activity Tracker & Stats (4 cols) */}
        <aside className="lg:col-span-4 flex flex-col gap-6">
          
          {/* AI Processing Timeline Tracker */}
          <div className="bg-cohere-soft-stone border border-cohere-card-border rounded-md p-6 relative overflow-hidden">
            
            {/* Pulsing state bar at the top */}
            <div className="absolute top-0 left-0 w-full h-[4px] bg-cohere-hairline">
              <div 
                className={`h-full bg-cohere-action-blue transition-all duration-500 ${
                  post.aiStatus === "PROCESSING" ? "w-2/3 animate-pulse" : post.aiStatus === "COMPLETED" ? "w-full" : "w-1/6"
                }`}
              />
            </div>

            <div className="flex justify-between items-center mb-4 mt-2">
              <h3 className="font-mono-label text-sm text-cohere-black font-semibold">
                AI Processing State
              </h3>
              <span className={`font-mono-label text-[11px] px-2 py-0.5 rounded-sm ${
                post.aiStatus === "PROCESSING" 
                  ? "bg-cohere-pale-blue text-cohere-action-blue animate-pulse" 
                  : post.aiStatus === "COMPLETED" 
                  ? "bg-cohere-pale-green text-cohere-deep-green" 
                  : "bg-cohere-hairline text-cohere-slate"
              }`}>
                {post.aiStatus}
              </span>
            </div>

            {/* Checklist Flow */}
            <div className="flex flex-col gap-4 relative pl-5 border-l border-dotted border-cohere-hairline mt-6">
              
              {/* Step 1 */}
              <div className="relative">
                <span className="absolute -left-[29px] bg-white text-cohere-deep-green w-5 h-5 rounded-full flex items-center justify-center border border-cohere-hairline text-xs font-bold">
                  ✓
                </span>
                <h4 className="font-mono-label text-[11px] font-semibold text-cohere-black">
                  1. Tag & Context Parsed
                </h4>
                <p className="font-sans text-xs text-cohere-slate">
                  Extracted topics: {post.tags.join(", ") || "None"}
                </p>
              </div>

              {/* Step 2 */}
              <div className="relative">
                <span className={`absolute -left-[29px] w-5 h-5 rounded-full flex items-center justify-center border text-xs font-bold ${
                  postLogs.length > 0 ? "bg-white text-cohere-deep-green border-cohere-hairline" : "bg-cohere-soft-stone text-cohere-slate border-cohere-hairline"
                }`}>
                  {postLogs.length > 0 ? "✓" : "2"}
                </span>
                <h4 className="font-mono-label text-[11px] font-semibold text-cohere-black">
                  2. Compute AI Willingness
                </h4>
                <p className="font-sans text-xs text-cohere-slate">
                  {postLogs.length > 0 
                    ? `${postLogs.filter(l => l.decision === "REPLY").length} agent(s) accepted response trigger.`
                    : "Calculating agent willingness scores..."
                  }
                </p>
              </div>

              {/* Step 3 */}
              <div className="relative">
                <span className={`absolute -left-[29px] w-5 h-5 rounded-full flex items-center justify-center border text-xs font-bold ${
                  postTasks.length > 0 && postTasks.every(t => t.status === "COMPLETED") 
                    ? "bg-white text-cohere-deep-green border-cohere-hairline" 
                    : post.aiStatus === "PROCESSING"
                    ? "bg-cohere-action-blue text-white border-cohere-action-blue animate-pulse"
                    : "bg-cohere-soft-stone text-cohere-slate border-cohere-hairline"
                }`}>
                  {postTasks.length > 0 && postTasks.every(t => t.status === "COMPLETED") ? "✓" : "3"}
                </span>
                <h4 className="font-mono-label text-[11px] font-semibold text-cohere-black">
                  3. Generate Agent Replies
                </h4>
                <p className="font-sans text-xs text-cohere-slate">
                  {postTasks.length > 0
                    ? `Tasks: ${postTasks.filter(t => t.status === "COMPLETED").length}/${postTasks.length} replies completed.`
                    : "Pending response generation triggers..."
                  }
                </p>
              </div>
            </div>

            {/* Micro Counter logs */}
            <div className="mt-6 pt-4 border-t border-cohere-hairline grid grid-cols-3 text-center text-xs font-mono-label">
              <div>
                <span className="block text-[10px] text-cohere-slate">Pending</span>
                <strong className="text-cohere-black">{postTasks.filter(t => t.status === "PENDING").length}</strong>
              </div>
              <div>
                <span className="block text-[10px] text-cohere-slate">Processing</span>
                <strong className="text-cohere-action-blue animate-pulse">{postTasks.filter(t => t.status === "PROCESSING").length}</strong>
              </div>
              <div>
                <span className="block text-[10px] text-cohere-slate">Completed</span>
                <strong className="text-cohere-deep-green">{postTasks.filter(t => t.status === "COMPLETED").length}</strong>
              </div>
            </div>
          </div>

          {/* AI Decision Log History Viewer */}
          <div className="bg-white border border-cohere-hairline rounded-md p-6">
            <h3 className="font-mono-label text-sm text-cohere-black font-semibold mb-4 border-b border-cohere-hairline pb-2">
              AI Decision Logs
            </h3>
            {postLogs.length === 0 ? (
              <p className="font-sans text-xs text-cohere-slate">No decision logs generated for this post.</p>
            ) : (
              <div className="flex flex-col gap-4">
                {postLogs.map((log) => (
                  <div key={log.id} className="border-b border-dashed border-cohere-hairline pb-3 last:border-0 last:pb-0">
                    <div className="flex justify-between items-center gap-2 mb-1">
                      <span className="font-mono-label text-[11px] font-semibold text-cohere-black">
                        {log.aiAgentName}
                      </span>
                      <span className={`font-mono-label text-[10px] px-1.5 py-0.2 rounded-sm ${
                        log.decision === "REPLY" 
                          ? "bg-cohere-pale-green text-cohere-deep-green" 
                          : "bg-cohere-soft-stone text-cohere-slate"
                      }`}>
                        {log.decision}
                      </span>
                    </div>
                    <p className="font-sans text-xs text-cohere-body-muted mb-1">
                      {log.reason}
                    </p>
                    <div className="font-mono-label text-[10px] text-cohere-muted">
                      Willingness: {(log.willingnessScore * 100).toFixed(0)}% (Thresh: {(log.thresholdValue * 100).toFixed(0)}%)
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

        </aside>
      </main>
    </div>
  );
}

import React from "react";
import { useNavigate } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { usePosts } from "../hooks/usePosts";
import { useAgents } from "../hooks/useAgents";
import { useFilterStore } from "../stores/useFilterStore";
import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";
import { Post } from "../api/types";

export function FeedPage() {
  const navigate = useNavigate();
  const { posts, isLoading } = usePosts();
  const { agents } = useAgents();
  const {
    selectedCategory,
    searchQuery,
    selectedTags,
    setCategory,
    toggleTag,
  } = useFilterStore();

  // Load decision logs for the "recent AI activities" sidebar
  const { data: decisionLogs = [] } = useQuery({
    queryKey: ["decisionLogs"],
    queryFn: api.decisionLogs.list,
  });

  const categories = ["所有领域", "后端开发", "前端开发", "人工智能", "日常交流", "求助问答"];

  // Filter posts based on store states
  const filteredPosts = posts.filter((post) => {
    if (selectedCategory && selectedCategory !== "所有领域" && post.category !== selectedCategory) {
      return false;
    }
    if (selectedTags.length > 0 && !selectedTags.every((t) => post.tags.includes(t))) {
      return false;
    }
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      const titleMatch = post.title.toLowerCase().includes(query);
      const contentMatch = post.content.toLowerCase().includes(query);
      const categoryMatch = post.category.toLowerCase().includes(query);
      return titleMatch || contentMatch || categoryMatch;
    }
    return true;
  });

  // Extract all unique tags for the popular tags sidebar
  const allTags = Array.from(new Set(posts.flatMap((p) => p.tags || []))).slice(0, 10);

  // Active agents list
  const activeAgents = agents.filter((a) => a.active);

  // Take the 5 most recent AI decision logs that resulted in a reply
  const recentLogs = decisionLogs
    .filter((log) => log.decision === "REPLY")
    .slice(0, 4);

  return (
    <main className="flex-grow max-w-7xl mx-auto w-full px-margin-mobile md:px-margin-desktop py-xl grid grid-cols-1 lg:grid-cols-12 gap-lg">
      {/* Left Content Area (8 cols) */}
      <div className="lg:col-span-8 flex flex-col gap-section">
        {/* Hero Section */}
        <section className="flex flex-col gap-md items-start pt-lg">
          <h1 className="font-headline-xl text-headline-xl text-primary tracking-tight">AI Forum</h1>
          <p className="font-body-large text-body-large text-surface-tint max-w-2xl">
            一个让多个 AI 角色参与讨论的智能论坛
          </p>
          <div className="flex gap-md mt-sm">
            <button
              onClick={() => navigate("/create-post")}
              className="bg-primary text-on-primary rounded-full px-lg py-sm font-label-mono-bold text-label-mono-bold hover:bg-ink transition-colors shadow-none"
            >
              发布帖子
            </button>
            <button
              onClick={() => navigate("/ai-agents")}
              className="bg-surface-container-lowest border border-hairline text-primary rounded-full px-lg py-sm font-label-mono-bold text-label-mono-bold hover:border-secondary hover:text-secondary transition-colors"
            >
              查看 AI 角色
            </button>
          </div>
        </section>

        {/* Main Feed */}
        <section className="flex flex-col gap-md">
          {/* Feed Tabs / Categories */}
          <div className="flex gap-md border-b border-hairline pb-xs overflow-x-auto">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setCategory(cat === "所有领域" ? null : cat)}
                className={`pb-sm px-xs whitespace-nowrap transition-colors ${
                  (selectedCategory === cat || (cat === "所有领域" && !selectedCategory))
                    ? "font-label-mono-bold text-label-mono-bold text-primary border-b-2 border-primary"
                    : "font-label-mono text-label-mono text-muted hover:text-primary"
                }`}
              >
                {cat}
              </button>
            ))}
          </div>

          {/* Post Lists using Virtuoso */}
          {isLoading ? (
            <div className="py-xl text-center text-muted font-body-main">加载中...</div>
          ) : filteredPosts.length === 0 ? (
            <div className="py-xl text-center text-muted font-body-main">暂无帖子</div>
          ) : (
            <div className="flex flex-col gap-md">
              <Virtuoso
                useWindowScroll
                data={filteredPosts}
                itemContent={(index, post) => (
                  <article
                    key={post.id}
                    onClick={() => navigate(`/posts/${post.id}`)}
                    className="bg-surface-container-lowest border border-hairline rounded-[16px] p-lg flex flex-col gap-sm hover:border-secondary transition-colors cursor-pointer mb-md"
                  >
                    <div className="flex justify-between items-start gap-md">
                      <h2 className="font-feature-title text-feature-title text-primary hover:text-secondary transition-colors">
                        {post.title}
                      </h2>
                      <span className="border border-coral text-coral rounded px-xs py-[2px] font-micro text-micro whitespace-nowrap">
                        {post.category}
                      </span>
                    </div>
                    <p className="font-body-main text-body-main text-on-surface-variant line-clamp-2">
                      {post.content}
                    </p>
                    <div className="flex flex-wrap gap-sm items-center mt-sm">
                      <div className="flex items-center gap-xs mr-md">
                        <img
                          src={post.author.avatar}
                          alt={post.author.username}
                          className="w-6 h-6 rounded-full border"
                        />
                        <span className="font-caption text-caption text-muted">{post.author.username}</span>
                      </div>
                      {post.tags.map((tag) => (
                        <span
                          key={tag}
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleTag(tag);
                          }}
                          className={`px-sm py-[2px] rounded font-micro text-micro transition-colors ${
                            selectedTags.includes(tag)
                              ? "bg-secondary text-white"
                              : "bg-surface-container text-on-surface-variant hover:bg-surface-variant"
                          }`}
                        >
                          #{tag}
                        </span>
                      ))}
                      <div className="flex-grow"></div>
                      
                      {/* AI Response Indicators */}
                      <div className="flex items-center gap-xs">
                        {post.aiAvatars && post.aiAvatars.length > 0 && (
                          <div className="flex -space-x-2 mr-xs">
                            {post.aiAvatars.map((avatar, idx) => (
                              <img
                                key={idx}
                                alt="AI Avatar"
                                className="w-6 h-6 rounded-full border border-surface-container-lowest object-cover"
                                src={avatar}
                              />
                            ))}
                          </div>
                        )}
                        {post.aiStatus === "PROCESSING" && (
                          <span className="bg-secondary-container text-on-secondary-container px-xs py-[2px] rounded font-micro text-micro flex items-center gap-xs">
                            <span className="material-symbols-outlined text-[14px]">psychology</span>
                            AI 分析中
                          </span>
                        )}
                        {post.aiStatus === "COMPLETED" && (
                          <span className="bg-success-green text-secondary px-xs py-[2px] rounded font-micro text-micro flex items-center gap-xs border border-secondary/10">
                            <span className="material-symbols-outlined text-[14px]">check_circle</span>
                            AI 已回复 ({post.aiResponsesCount || 0})
                          </span>
                        )}
                        {post.aiStatus === "PENDING" && (
                          <span className="bg-surface-container text-muted px-xs py-[2px] rounded font-micro text-micro flex items-center gap-xs">
                            <span className="material-symbols-outlined text-[14px]">schedule</span>
                            待分析
                          </span>
                        )}
                      </div>
                    </div>
                  </article>
                )}
              />
            </div>
          )}
        </section>
      </div>

      {/* Right Sidebar (4 cols) */}
      <aside className="lg:col-span-4 flex flex-col gap-xl pt-lg">
        {/* Popular Tags */}
        {allTags.length > 0 && (
          <div className="bg-surface-container-low rounded-[16px] p-md border border-hairline">
            <h3 className="font-label-mono-bold text-label-mono-bold text-primary mb-md border-b border-hairline pb-xs">
              热门标签
            </h3>
            <div className="flex flex-wrap gap-sm">
              {allTags.map((tag) => {
                const isActive = selectedTags.includes(tag);
                return (
                  <span
                    key={tag}
                    onClick={() => toggleTag(tag)}
                    className={`px-sm py-xs rounded font-caption text-caption cursor-pointer transition-colors ${
                      isActive
                        ? "bg-secondary text-white"
                        : "bg-surface-container text-on-surface-variant hover:bg-surface-variant"
                    }`}
                  >
                    {tag}
                  </span>
                );
              })}
            </div>
          </div>
        )}

        {/* Active AI Agents */}
        <div className="bg-surface-container-low rounded-[16px] p-md border border-hairline flex flex-col gap-sm">
          <h3 className="font-label-mono-bold text-label-mono-bold text-primary mb-xs border-b border-hairline pb-xs">
            活跃 AI 角色
          </h3>
          {activeAgents.length === 0 ? (
            <div className="py-sm text-center text-muted font-micro text-micro">暂无活跃角色</div>
          ) : (
            activeAgents.map((agent) => (
              <div
                key={agent.id}
                onClick={() => navigate("/ai-agents")}
                className="flex items-center gap-sm p-sm rounded-lg hover:bg-surface-container cursor-pointer transition-colors"
              >
                <img
                  alt={agent.name}
                  className="w-10 h-10 rounded-[22px] object-cover border"
                  src={agent.avatar}
                />
                <div className="flex flex-col">
                  <span className="font-label-mono-bold text-label-mono-bold text-primary">{agent.name}</span>
                  <span className="font-micro text-micro text-muted line-clamp-1">{agent.description}</span>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Recent AI Activities */}
        <div className="bg-surface-container-low rounded-[16px] p-md border border-hairline flex flex-col gap-sm">
          <h3 className="font-label-mono-bold text-label-mono-bold text-primary mb-xs border-b border-hairline pb-xs">
            最近 AI 动态
          </h3>
          {recentLogs.length === 0 ? (
            <div className="py-sm text-center text-muted font-micro text-micro">暂无 AI 动态</div>
          ) : (
            <div className="flex flex-col mt-sm">
              {recentLogs.map((log, index) => {
                const isLast = index === recentLogs.length - 1;
                const formattedTime = new Date(log.createdAt).toLocaleTimeString([], {
                  hour: "2-digit",
                  minute: "2-digit",
                });
                return (
                  <div
                    key={log.id}
                    onClick={() => navigate(`/posts/${log.postId}`)}
                    className="flex flex-col gap-xs relative pl-md border-l border-dotted border-hairline pb-md last:pb-0 cursor-pointer"
                  >
                    <span className="absolute w-2 h-2 rounded-full bg-secondary -left-[5px] top-[4px]"></span>
                    <span className="font-label-mono text-[10px] text-muted">{formattedTime}</span>
                    <p className="font-caption text-caption text-primary">
                      <span className="font-semibold">{log.aiAgentName}</span>{" "}
                      {log.triggerType === "POST_AUTO" ? "自动回复了帖子" : "跟进回复了帖子"}
                    </p>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </aside>
    </main>
  );
}

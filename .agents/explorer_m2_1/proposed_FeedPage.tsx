import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { usePosts } from "../hooks/usePosts";
import { useAgents } from "../hooks/useAgents";
import { useFilterStore } from "../stores/useFilterStore";
import { useUserStore } from "../stores/useUserStore";
import { MessageSquare, Eye, Search, PlusCircle, User, Bot, Sparkles, Activity } from "lucide-react";

export default function FeedPage() {
  const { posts, isLoading } = usePosts();
  const { agents } = useAgents();
  const { currentUser } = useUserStore();
  const navigate = useNavigate();

  const {
    selectedCategory,
    searchQuery,
    selectedTags,
    setCategory,
    setSearchQuery,
    toggleTag,
    resetFilters
  } = useFilterStore();

  // Filter logic
  const filteredPosts = posts.filter((post) => {
    if (selectedCategory && post.category !== selectedCategory) {
      return false;
    }
    if (selectedTags.length > 0 && !selectedTags.every(tag => post.tags.includes(tag))) {
      return false;
    }
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchTitle = post.title.toLowerCase().includes(q);
      const matchContent = post.content.toLowerCase().includes(q);
      const matchAuthor = post.author.username.toLowerCase().includes(q);
      return matchTitle || matchContent || matchAuthor;
    }
    return true;
  });

  // Extract all unique tags for the sidebar tag cloud
  const allTags = Array.from(
    new Set(posts.flatMap((post) => post.tags || []))
  ).slice(0, 10);

  // Active agents list
  const activeAgents = agents.filter((agent) => agent.active);

  // Categories list based on prototype and db options
  const categories = ["后端开发", "前端开发", "人工智能", "技术分享", "技术探讨", "日常交流", "求助问答"];

  return (
    <div className="min-h-screen bg-cohere-canvas flex flex-col text-cohere-ink font-sans">
      {/* Hero Header */}
      <section className="bg-cohere-canvas border-b border-cohere-hairline py-12 px-6 md:px-12 max-w-7xl mx-auto w-full">
        <div className="max-w-3xl">
          <h1 className="font-display text-5xl md:text-7xl font-normal leading-[1.1] -tracking-[1.5px] text-cohere-black mb-4">
            AI Forum
          </h1>
          <p className="font-sans text-lg md:text-xl text-cohere-body-muted leading-relaxed max-w-2xl mb-8">
            A collaborative multi-agent discussion platform where human perspectives align with specialized AI models.
          </p>
          <div className="flex flex-wrap gap-4">
            <button
              onClick={() => navigate("/create-post")}
              className="btn-primary flex items-center gap-2"
            >
              <PlusCircle className="w-4 h-4" />
              Publish Post
            </button>
            <button
              onClick={() => navigate("/agents")}
              className="btn-pill-outline bg-white hover:bg-cohere-soft-stone"
            >
              View AI Agents
            </button>
          </div>
        </div>
      </section>

      {/* Main Container */}
      <main className="max-w-7xl mx-auto w-full px-6 md:px-12 py-10 grid grid-cols-1 lg:grid-cols-12 gap-8">
        
        {/* Left Feed Section (8 cols) */}
        <div className="lg:col-span-8 flex flex-col gap-6">
          
          {/* Filters Bar */}
          <div className="flex flex-col gap-4 border-b border-cohere-hairline pb-4">
            
            {/* Category Tabs */}
            <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-none">
              <button
                onClick={() => setCategory(null)}
                className={`font-mono-label text-xs px-3 py-1.5 rounded-sm border transition-colors whitespace-nowrap ${
                  selectedCategory === null
                    ? "bg-cohere-primary text-white border-cohere-primary"
                    : "bg-transparent text-cohere-slate border-cohere-hairline hover:bg-cohere-soft-stone"
                }`}
              >
                All Posts
              </button>
              {categories.map((cat) => (
                <button
                  key={cat}
                  onClick={() => setCategory(cat)}
                  className={`font-mono-label text-xs px-3 py-1.5 rounded-sm border transition-colors whitespace-nowrap ${
                    selectedCategory === cat
                      ? "bg-cohere-primary text-white border-cohere-primary"
                      : "bg-transparent text-cohere-slate border-cohere-hairline hover:bg-cohere-soft-stone"
                  }`}
                >
                  {cat}
                </button>
              ))}
            </div>

            {/* Search and Quick Filters */}
            <div className="flex flex-col sm:flex-row gap-3 items-center justify-between">
              <div className="relative w-full sm:w-80">
                <Search className="absolute left-3 top-2.5 w-4 h-4 text-cohere-muted" />
                <input
                  type="text"
                  placeholder="Search discussion..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-9 pr-4 py-2 bg-cohere-canvas border border-cohere-hairline rounded-sm font-sans text-sm focus:outline-none focus:border-cohere-form-focus transition-colors"
                />
              </div>
              
              {selectedTags.length > 0 && (
                <button
                  onClick={resetFilters}
                  className="text-xs text-cohere-action-blue hover:underline whitespace-nowrap"
                >
                  Clear all filters
                </button>
              )}
            </div>
          </div>

          {/* Virtualized Post Feed */}
          {isLoading ? (
            <div className="py-12 text-center text-cohere-muted font-sans">
              Loading conversations...
            </div>
          ) : filteredPosts.length === 0 ? (
            <div className="py-12 text-center text-cohere-muted font-sans border border-dashed border-cohere-hairline rounded-md">
              No matching topics found. Try refining your filters.
            </div>
          ) : (
            <div className="h-[750px] w-full">
              <Virtuoso
                useWindowScroll
                data={filteredPosts}
                itemContent={(index, post) => (
                  <div className="pb-6">
                    <article className="bg-white border border-cohere-hairline rounded-md p-6 hover:border-cohere-slate transition-all duration-300">
                      <div className="flex justify-between items-start gap-4 mb-3">
                        <Link to={`/post/${post.id}`}>
                          <h2 className="font-feature-heading text-cohere-black font-semibold hover:text-cohere-action-blue transition-colors">
                            {post.title}
                          </h2>
                        </Link>
                        <span className="font-mono-label text-xs border border-cohere-coral text-cohere-coral px-2.5 py-0.5 rounded-sm whitespace-nowrap">
                          {post.category}
                        </span>
                      </div>
                      
                      <p className="font-body text-cohere-body-muted line-clamp-2 mb-4">
                        {post.content}
                      </p>

                      <div className="flex flex-wrap gap-2 mb-4">
                        {post.tags.map((tag) => (
                          <span
                            key={tag}
                            onClick={() => toggleTag(tag)}
                            className={`font-mono-label text-[11px] px-2 py-0.5 rounded-sm cursor-pointer transition-colors ${
                              selectedTags.includes(tag)
                                ? "bg-cohere-coral text-white"
                                : "bg-cohere-soft-stone text-cohere-slate hover:bg-cohere-hairline"
                            }`}
                          >
                            #{tag}
                          </span>
                        ))}
                      </div>

                      <div className="flex flex-wrap items-center justify-between gap-4 pt-4 border-t border-cohere-hairline">
                        <div className="flex items-center gap-2">
                          <img
                            src={post.author.avatar}
                            alt={post.author.username}
                            className="w-6 h-6 rounded-full object-cover border border-cohere-hairline"
                          />
                          <span className="font-caption text-xs text-cohere-slate">
                            {post.author.username}
                          </span>
                          <span className="text-cohere-hairline font-thin">•</span>
                          <span className="font-micro text-xs text-cohere-muted">
                            {new Date(post.createdAt).toLocaleDateString()}
                          </span>
                        </div>

                        <div className="flex items-center gap-4">
                          {/* AI Status Badge */}
                          {post.aiStatus === "PROCESSING" ? (
                            <span className="bg-cohere-pale-blue text-cohere-action-blue border border-cohere-action-blue/20 px-2 py-0.5 rounded-sm font-mono-label text-[11px] flex items-center gap-1">
                              <Sparkles className="w-3 h-3 animate-pulse" />
                              Analyzing
                            </span>
                          ) : post.aiStatus === "COMPLETED" ? (
                            <span className="bg-cohere-pale-green text-cohere-deep-green border border-cohere-pale-green px-2 py-0.5 rounded-sm font-mono-label text-[11px] flex items-center gap-1">
                              <Bot className="w-3 h-3" />
                              AI Replied ({post.aiResponsesCount})
                            </span>
                          ) : (
                            <span className="bg-cohere-soft-stone text-cohere-slate border border-cohere-hairline px-2 py-0.5 rounded-sm font-mono-label text-[11px] flex items-center gap-1">
                              Pending
                            </span>
                          )}

                          {/* AI Avatars engaged */}
                          {post.aiAvatars.length > 0 && (
                            <div className="flex -space-x-1.5">
                              {post.aiAvatars.map((av, i) => (
                                <img
                                  key={i}
                                  src={av}
                                  alt="AI agent engaged"
                                  className="w-5 h-5 rounded-full border border-white object-cover"
                                />
                              ))}
                            </div>
                          )}
                        </div>
                      </div>
                    </article>
                  </div>
                )}
              />
            </div>
          )}
        </div>

        {/* Right Sidebar (4 cols) */}
        <aside className="lg:col-span-4 flex flex-col gap-8">
          
          {/* Active AI Agents Plaza Teaser */}
          <div className="bg-cohere-soft-stone border border-cohere-card-border rounded-md p-6">
            <h3 className="font-mono-label text-sm text-cohere-black font-semibold mb-4 border-b border-cohere-hairline pb-2 flex items-center gap-2">
              <Bot className="w-4 h-4 text-cohere-deep-green" />
              Active AI Agents
            </h3>
            <div className="flex flex-col gap-4">
              {activeAgents.map((agent) => (
                <div key={agent.id} className="flex items-center gap-3">
                  <img
                    src={agent.avatar}
                    alt={agent.name}
                    className="w-10 h-10 rounded-lg object-cover border border-cohere-hairline"
                  />
                  <div className="flex-1 min-w-0">
                    <h4 className="font-sans text-sm font-semibold text-cohere-primary truncate">
                      {agent.name}
                    </h4>
                    <p className="font-sans text-xs text-cohere-slate truncate">
                      {agent.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
            <button
              onClick={() => navigate("/agents")}
              className="mt-6 w-full text-center font-button text-xs text-cohere-action-blue hover:underline"
            >
              Configure Agents Plaza →
            </button>
          </div>

          {/* Popular Tags Cloud */}
          <div className="bg-white border border-cohere-hairline rounded-md p-6">
            <h3 className="font-mono-label text-sm text-cohere-black font-semibold mb-4 border-b border-cohere-hairline pb-2">
              Popular Tags
            </h3>
            <div className="flex flex-wrap gap-2">
              {allTags.map((tag) => {
                const isSelected = selectedTags.includes(tag);
                return (
                  <span
                    key={tag}
                    onClick={() => toggleTag(tag)}
                    className={`font-mono-label text-xs px-2.5 py-1 rounded-sm cursor-pointer transition-colors ${
                      isSelected
                        ? "bg-cohere-coral text-white"
                        : "bg-cohere-soft-stone text-cohere-slate hover:bg-cohere-hairline"
                    }`}
                  >
                    #{tag}
                  </span>
                );
              })}
            </div>
          </div>
          
        </aside>
      </main>
    </div>
  );
}

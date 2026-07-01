import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePosts } from "../hooks/usePosts";
import { useUserStore } from "../stores/useUserStore";
import { ArrowLeft, Send, Sparkles, AlertCircle, HelpCircle, Loader2 } from "lucide-react";

export default function CreatePostPage() {
  const { createPost, isCreating } = usePosts();
  const { currentUser } = useUserStore();
  const navigate = useNavigate();

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [category, setCategory] = useState("后端开发");
  const [tagsInput, setTagsInput] = useState("");
  
  // Mock field for AI Response modes matching prototype (low-ai, humans-only, standard-ai, busy-ai)
  const [aiMode, setAiMode] = useState("low-ai");

  const categories = ["后端开发", "前端开发", "人工智能", "技术分享", "技术探讨", "日常交流", "求助问答"];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !content.trim()) return;

    // Split tag input by comma or whitespace
    const tags = tagsInput
      .split(/[,，\s]+/)
      .map(tag => tag.trim())
      .filter(tag => tag.length > 0);

    try {
      const newPost = await createPost({
        title,
        content,
        category,
        tags: tags.length > 0 ? tags : ["General"],
        author: {
          username: currentUser.username,
          avatar: currentUser.avatar
        }
      });
      // Redirect to the newly created post's detail view
      navigate(`/post/${newPost.id}`);
    } catch (err) {
      console.error("Failed to publish post:", err);
    }
  };

  return (
    <div className="min-h-screen bg-cohere-canvas flex flex-col text-cohere-ink font-sans">
      
      {/* Top Breadcrumb Nav */}
      <nav className="border-b border-cohere-hairline py-4 px-6 md:px-12 max-w-7xl mx-auto w-full flex items-center justify-between">
        <button onClick={() => navigate("/")} className="flex items-center gap-2 text-xs font-mono-label text-cohere-slate hover:text-cohere-primary transition-colors">
          <ArrowLeft className="w-4 h-4" />
          Cancel and return
        </button>
        <span className="font-mono-label text-xs text-cohere-muted">
          New Topic
        </span>
      </nav>

      {/* Main Grid */}
      <main className="max-w-7xl mx-auto w-full px-6 md:px-12 py-10 grid grid-cols-1 lg:grid-cols-12 gap-8">
        
        {/* Left Form Area (8 cols) */}
        <section className="lg:col-span-8 flex flex-col gap-6">
          
          <div className="mb-2">
            <h1 className="font-sans text-3xl font-normal leading-[1.2] -tracking-[0.5px] text-cohere-black mb-2">
              Write New Topic
            </h1>
            <p className="font-sans text-sm text-cohere-slate">
              Share your engineering critique, architectural ideas, or code questions with the community and active AI agents.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            
            {/* Title */}
            <div className="flex flex-col gap-2">
              <label className="font-mono-label text-xs font-semibold text-cohere-slate uppercase tracking-wider">
                Title
              </label>
              <input
                type="text"
                placeholder="Enter a descriptive topic title..."
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
                className="w-full bg-white border border-cohere-hairline rounded-sm px-4 py-2.5 font-sans text-sm focus:outline-none focus:border-cohere-form-focus transition-colors"
              />
            </div>

            {/* Category & Tags Row */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              
              {/* Category */}
              <div className="flex flex-col gap-2">
                <label className="font-mono-label text-xs font-semibold text-cohere-slate uppercase tracking-wider">
                  Category
                </label>
                <select
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className="w-full bg-white border border-cohere-hairline rounded-sm px-4 py-2.5 font-sans text-sm focus:outline-none focus:border-cohere-form-focus transition-colors appearance-none"
                >
                  {categories.map(cat => (
                    <option key={cat} value={cat}>{cat}</option>
                  ))}
                </select>
              </div>

              {/* Tags */}
              <div className="flex flex-col gap-2">
                <label className="font-mono-label text-xs font-semibold text-cohere-slate uppercase tracking-wider">
                  Tags (comma separated)
                </label>
                <input
                  type="text"
                  placeholder="e.g. Go, Rust, SystemDesign"
                  value={tagsInput}
                  onChange={(e) => setTagsInput(e.target.value)}
                  className="w-full bg-white border border-cohere-hairline rounded-sm px-4 py-2.5 font-sans text-sm focus:outline-none focus:border-cohere-form-focus transition-colors"
                />
              </div>

            </div>

            {/* Editor Box */}
            <div className="flex flex-col gap-2">
              <label className="font-mono-label text-xs font-semibold text-cohere-slate uppercase tracking-wider">
                Body Content (Supports Markdown)
              </label>
              <textarea
                placeholder="Write your topic description here. Markdown format is supported..."
                value={content}
                onChange={(e) => setContent(e.target.value)}
                required
                rows={10}
                className="w-full bg-white border border-cohere-hairline rounded-sm p-4 font-sans text-sm focus:outline-none focus:border-cohere-form-focus transition-colors resize-none"
              />
            </div>

            {/* AI Participation Mode Selector */}
            <div className="flex flex-col gap-3 mt-2">
              <div className="flex items-center gap-1.5 text-cohere-coral">
                <Sparkles className="w-4.5 h-4.5" />
                <label className="font-mono-label text-xs font-semibold uppercase tracking-wider">
                  AI Participation Mode
                </label>
              </div>
              
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                
                {/* Humans Only */}
                <label className="relative cursor-pointer">
                  <input
                    type="radio"
                    name="ai-mode"
                    value="humans-only"
                    checked={aiMode === "humans-only"}
                    onChange={() => setAiMode("humans-only")}
                    className="sr-only peer"
                  />
                  <div className="p-4 rounded-md border border-cohere-hairline bg-white peer-checked:border-cohere-primary peer-checked:bg-cohere-soft-stone/20 hover:bg-cohere-soft-stone/10 transition-all flex flex-col gap-1.5 h-full">
                    <span className="font-sans text-sm font-semibold text-cohere-primary">Humans Only</span>
                    <p className="font-sans text-xs text-cohere-slate">
                      Disable all auto-replies from AI agents. A quiet zone reserved for human discussion.
                    </p>
                  </div>
                </label>

                {/* Low AI */}
                <label className="relative cursor-pointer">
                  <input
                    type="radio"
                    name="ai-mode"
                    value="low-ai"
                    checked={aiMode === "low-ai"}
                    onChange={() => setAiMode("low-ai")}
                    className="sr-only peer"
                  />
                  <div className="p-4 rounded-md border border-cohere-hairline bg-white peer-checked:border-cohere-primary peer-checked:bg-cohere-soft-stone/20 hover:bg-cohere-soft-stone/10 transition-all flex flex-col gap-1.5 h-full">
                    <span className="font-sans text-sm font-semibold text-cohere-primary">Low AI Intervention</span>
                    <p className="font-sans text-xs text-cohere-slate">
                      Only highly relevant active agents or direct @mentions will trigger an AI response.
                    </p>
                  </div>
                </label>

                {/* Standard AI */}
                <label className="relative cursor-pointer">
                  <input
                    type="radio"
                    name="ai-mode"
                    value="standard"
                    checked={aiMode === "standard"}
                    onChange={() => setAiMode("standard")}
                    className="sr-only peer"
                  />
                  <div className="p-4 rounded-md border border-cohere-hairline bg-white peer-checked:border-cohere-primary peer-checked:bg-cohere-soft-stone/20 hover:bg-cohere-soft-stone/10 transition-all flex flex-col gap-1.5 h-full">
                    <span className="font-sans text-sm font-semibold text-cohere-primary">Standard AI</span>
                    <p className="font-sans text-xs text-cohere-slate">
                      Agents will treat the topic like a standard forum post and respond based on normal thresholds.
                    </p>
                  </div>
                </label>

                {/* Busy AI */}
                <label className="relative cursor-pointer">
                  <input
                    type="radio"
                    name="ai-mode"
                    value="busy-ai"
                    checked={aiMode === "busy-ai"}
                    onChange={() => setAiMode("busy-ai")}
                    className="sr-only peer"
                  />
                  <div className="p-4 rounded-md border border-cohere-hairline bg-white peer-checked:border-cohere-coral peer-checked:bg-cohere-coral/5 hover:bg-cohere-soft-stone/10 transition-all flex flex-col gap-1.5 h-full">
                    <span className="font-sans text-sm font-semibold text-cohere-coral">Spirited Mode</span>
                    <p className="font-sans text-xs text-cohere-slate">
                      Maximize agent willingness scores. Highly active discussion and multiple angles of critique.
                    </p>
                  </div>
                </label>

              </div>
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-3 mt-4 pt-4 border-t border-cohere-hairline">
              <button
                type="button"
                onClick={() => navigate("/")}
                className="px-6 py-2 border border-cohere-hairline rounded-sm font-button text-xs hover:bg-cohere-soft-stone"
              >
                Save Draft
              </button>
              <button
                type="submit"
                disabled={isCreating || !title.trim() || !content.trim()}
                className="btn-primary font-button text-xs py-2 px-6 flex items-center gap-1.5 disabled:opacity-40"
              >
                {isCreating ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <Send className="w-3.5 h-3.5" />
                )}
                Publish Topic
              </button>
            </div>

          </form>
        </section>

        {/* Right Preview Sidebar (4 cols) */}
        <aside className="lg:col-span-4 flex flex-col gap-6">
          <div className="bg-white border border-cohere-hairline rounded-md p-6 flex flex-col gap-4">
            <h3 className="font-mono-label text-sm text-cohere-black font-semibold mb-2 border-b border-cohere-hairline pb-2">
              Card Preview
            </h3>
            
            <div className="bg-white border border-cohere-hairline rounded-md p-6">
              <div className="flex justify-between items-start gap-4 mb-3">
                <h4 className="font-feature-heading text-cohere-primary truncate max-w-[180px]">
                  {title || "Enter title..."}
                </h4>
                <span className="font-mono-label text-[10px] border border-cohere-coral text-cohere-coral px-2 py-0.5 rounded-sm">
                  {category}
                </span>
              </div>
              <p className="font-body text-cohere-body-muted line-clamp-3 text-xs mb-4">
                {content || "Your post content preview will show here..."}
              </p>
              <div className="flex items-center justify-between pt-3 border-t border-cohere-hairline">
                <span className="font-mono-label text-[10px] text-cohere-muted">
                  Just now • {currentUser.username}
                </span>
                <span className="font-mono-label text-[10px] bg-cohere-soft-stone text-cohere-slate px-2 py-0.5 rounded-sm">
                  Pending AI
                </span>
              </div>
            </div>
          </div>
        </aside>

      </main>
    </div>
  );
}

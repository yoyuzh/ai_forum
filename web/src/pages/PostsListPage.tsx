import { useState } from "react";
import { Link, createSearchParams, useNavigate, useSearchParams } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { usePosts } from "../hooks/usePosts";
import { useUserStore } from "../stores/useUserStore";
import { useFilterStore } from "../stores/useFilterStore";
import PostCard from "../components/cards/PostCard";
import CategoryBadge from "../components/ui/CategoryBadge";
import { FeedTab } from "../api/types";

const TABS: { key: FeedTab; label: string }[] = [
  { key: "latest", label: "最新" },
  { key: "hottest", label: "最热" },
  { key: "unanswered", label: "待回复" },
  { key: "ai_most", label: "AI 参与最多" },
];

const CATEGORIES = ["后端开发", "前端开发", "人工智能", "架构设计", "DevOps", "技术探讨"];

export default function PostsListPage() {
  const { feedTab, setFeedTab } = useFilterStore();
  const [searchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const tag = searchParams.get("tag") ?? undefined;
  const requestedTab = searchParams.get("tab") as FeedTab | null;
  const activeTab = requestedTab && TABS.some((tab) => tab.key === requestedTab) ? requestedTab : feedTab;
  const { posts, isLoading, createPost, isCreating } = usePosts(activeTab, query, tag);
  const { currentUser } = useUserStore();
  const navigate = useNavigate();

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [category, setCategory] = useState(CATEGORIES[0]);
  const [tags, setTags] = useState("");

  const selectTab = (tab: FeedTab) => {
    setFeedTab(tab);
    const next = createSearchParams();
    if (tab !== "latest") next.set("tab", tab);
    if (query) next.set("q", query);
    if (tag) next.set("tag", tag);
    navigate({ search: next.toString() ? `?${next}` : "" }, { replace: true });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !content.trim()) return;
    const post = await createPost({
      title: title.trim(),
      content: content.trim(),
      category,
      tags: tags
        .split(/[,，\s]+/)
        .map((t) => t.trim())
        .filter(Boolean),
      author: { username: currentUser!.username, avatar: currentUser!.avatar, role: "研究员" },
    });
    setTitle("");
    setContent("");
    setTags("");
    navigate(`/posts/${post.id}`);
  };

  return (
    <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-xl md:px-margin-desktop animate-reveal-up">
      <div className="grid grid-cols-1 gap-lg lg:grid-cols-12">
        <div className="flex flex-col gap-section lg:col-span-8">
          <section className="flex flex-col items-start gap-md pt-lg">
            <h1 className="font-headline-xl text-cohere-primary">帖子广场</h1>
            <p className="max-w-2xl font-body-large text-cohere-on-surface-variant">
              发布你的技术问题或研究观察，AI 角色会自主评估是否参与讨论。
            </p>
          </section>

          <section className="flex flex-col gap-md">
            <div className="flex gap-md overflow-x-auto border-b border-cohere-hairline pb-xs">
              {TABS.map((tab) => (
                <button
                  key={tab.key}
                  type="button"
                  onClick={() => selectTab(tab.key)}
                  aria-pressed={activeTab === tab.key}
                  className={`feed-tab ${activeTab === tab.key ? "active" : ""}`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            {isLoading ? (
              <p className="font-body-main text-cohere-on-surface-variant">加载帖子中…</p>
            ) : posts.length === 0 ? (
              <div className="card-base p-lg">
                <p className="font-body-main text-cohere-muted">
                  {query || tag ? "没有找到匹配的帖子。" : "该筛选下暂无帖子。"}
                </p>
                {(query || tag) && (
                  <Link to="/posts" className="btn-link mt-sm inline-flex">
                    清除筛选
                  </Link>
                )}
              </div>
            ) : (
              <Virtuoso
                useWindowScroll
                data={posts}
                itemContent={(_index, post) => (
                  <div className="pb-md">
                    <PostCard post={post} />
                  </div>
                )}
                style={{ minHeight: 400 }}
              />
            )}
          </section>
        </div>

        {/* Composer */}
        <aside className="flex flex-col gap-xl pt-lg lg:col-span-4">
          <div className="card-surface-low p-xl">
            <div className="mb-md flex items-center gap-sm">
              <CategoryBadge category="发帖" />
              <h2 className="font-feature-title text-cohere-primary">创建新话题</h2>
            </div>
            <form onSubmit={handleSubmit} className="flex flex-col gap-md">
              <div>
                <label htmlFor="post-title" className="mb-1 block font-caption text-cohere-on-surface-variant">
                  标题
                </label>
                <input
                  id="post-title"
                  name="title"
                  type="text"
                  autoComplete="off"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="例如：分布式事务选型的思考…"
                  required
                  className="form-control"
                />
              </div>

              <div>
                <label htmlFor="post-category" className="mb-1 block font-caption text-cohere-on-surface-variant">
                  分类
                </label>
                <select
                  id="post-category"
                  name="category"
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className="form-control"
                >
                  {CATEGORIES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label htmlFor="post-tags" className="mb-1 block font-caption text-cohere-on-surface-variant">
                  标签（逗号分隔）
                </label>
                <input
                  id="post-tags"
                  name="tags"
                  type="text"
                  autoComplete="off"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                  placeholder="Go, 微服务, 性能…"
                  className="form-control"
                />
              </div>

              <div>
                <label htmlFor="post-content" className="mb-1 block font-caption text-cohere-on-surface-variant">
                  正文
                </label>
                <textarea
                  id="post-content"
                  name="content"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="描述你的设计问题或研究观察…（支持 Markdown）"
                  rows={5}
                  required
                  className="form-control resize-y"
                />
              </div>

              <button type="submit" disabled={isCreating} className="btn-primary">
                {isCreating ? "发布中…" : "发布帖子"}
              </button>
            </form>
          </div>

          <div className="card-base p-lg">
            <Link to="/" className="btn-link flex items-center gap-1">
              <span className="material-symbols-outlined text-[16px]">arrow_back</span>
              返回首页
            </Link>
          </div>
        </aside>
      </div>
    </main>
  );
}

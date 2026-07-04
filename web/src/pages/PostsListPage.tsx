import { Link, createSearchParams, useNavigate, useSearchParams } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { usePosts } from "../hooks/usePosts";
import { useFilterStore } from "../stores/useFilterStore";
import PostCard from "../components/cards/PostCard";
import { FeedTab } from "../api/types";

const TABS: { key: FeedTab; label: string }[] = [
  { key: "latest", label: "最新" },
  { key: "hottest", label: "最热" },
  { key: "unanswered", label: "待回复" },
  { key: "ai_most", label: "AI 参与最多" },
];

export default function PostsListPage() {
  const { feedTab, setFeedTab } = useFilterStore();
  const [searchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const tag = searchParams.get("tag") ?? undefined;
  const requestedTab = searchParams.get("tab") as FeedTab | null;
  const activeTab = requestedTab && TABS.some((tab) => tab.key === requestedTab) ? requestedTab : feedTab;
  const { posts, isLoading } = usePosts(activeTab, query, tag);
  const navigate = useNavigate();

  const selectTab = (tab: FeedTab) => {
    setFeedTab(tab);
    const next = createSearchParams();
    if (tab !== "latest") next.set("tab", tab);
    if (query) next.set("q", query);
    if (tag) next.set("tag", tag);
    navigate({ search: next.toString() ? `?${next}` : "" }, { replace: true });
  };

  return (
    <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-xl md:px-margin-desktop animate-reveal-up">
      <div className="flex flex-col gap-section">
        <section className="flex flex-col items-start gap-md pt-lg md:flex-row md:items-end md:justify-between">
          <div>
            <h1 className="font-headline-xl text-cohere-primary">帖子广场</h1>
            <p className="max-w-2xl font-body-large text-cohere-on-surface-variant">
              发布你的技术问题或研究观察，AI 角色会自主评估是否参与讨论。
            </p>
          </div>
          <Link to="/posts/new" className="btn-primary">
            发布帖子
          </Link>
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
    </main>
  );
}

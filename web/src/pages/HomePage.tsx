import { Link, createSearchParams, useNavigate, useSearchParams } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { usePosts } from "../hooks/usePosts";
import { useHotTags } from "../hooks/useHotTags";
import { useAgents } from "../hooks/useAgents";
import { useActivities } from "../hooks/useActivities";
import { useFilterStore } from "../stores/useFilterStore";
import { FeedTab } from "../api/types";
import PostCard from "../components/cards/PostCard";
import HotTags from "../components/sidebar/HotTags";
import ActiveAIRoles from "../components/sidebar/ActiveAIRoles";
import RecentAIActivity from "../components/sidebar/RecentAIActivity";
import BrandIcon from "../components/ui/BrandIcon";

const TABS: { key: FeedTab; label: string }[] = [
  { key: "latest", label: "最新" },
  { key: "hottest", label: "最热" },
  { key: "unanswered", label: "待回复" },
  { key: "ai_most", label: "AI 参与最多" },
];

export default function HomePage() {
  const { feedTab, setFeedTab } = useFilterStore();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const query = searchParams.get("q") ?? "";
  const tag = searchParams.get("tag") ?? undefined;
  const { posts, isLoading } = usePosts(feedTab, query, tag);
  const { tags: hotTags } = useHotTags();
  const { agents } = useAgents();
  const { activities } = useActivities();
  const activeAgents = agents.filter((a) => a.active).slice(0, 3);

  const selectTab = (tab: FeedTab) => {
    setFeedTab(tab);
    const next = createSearchParams();
    if (tab !== "latest") next.set("tab", tab);
    if (query) next.set("q", query);
    if (tag) next.set("tag", tag);
    navigate({ search: next.toString() ? `?${next}` : "" }, { replace: true });
  };

  const selectTag = (nextTag: string) => {
    const next = createSearchParams();
    if (feedTab !== "latest") next.set("tab", feedTab);
    if (query) next.set("q", query);
    next.set("tag", nextTag);
    navigate({ pathname: "/posts", search: `?${next}` });
  };

  return (
    <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-xl md:px-margin-desktop animate-reveal-up">
      <div className="grid grid-cols-1 gap-lg lg:grid-cols-12">
        {/* Left content area (8 cols) */}
        <div className="flex flex-col gap-section lg:col-span-8">
          {/* Hero */}
          <section className="flex flex-col items-start gap-md pt-lg">
            <span className="flex items-center gap-1 font-label-mono text-cohere-secondary">
              <span className="status-dot bg-cohere-secondary animate-pulse-soft" />
              AI Forum Research Lab
            </span>
            <h1 className="font-display-lg text-display-lg text-cohere-primary">AI Forum</h1>
            <p className="max-w-2xl font-body-large text-cohere-on-surface-variant">
              一个让多个 AI 角色参与讨论的智能研究论坛。发帖后，AI 代理会根据意愿分与阈值自主决定是否回复，
              每一次决策都可追溯、可解释。
            </p>
            <div className="mt-sm flex flex-wrap gap-md">
              <Link to="/posts/new" className="btn-primary inline-flex items-center gap-xs">
                <BrandIcon name="compose" size={18} className="brightness-0 invert" />
                发布帖子
              </Link>
              <Link to="/agents" className="btn-pill-outline inline-flex items-center gap-xs">
                <BrandIcon name="ai" size={18} />
                查看 AI 角色
              </Link>
            </div>
          </section>

          {/* Feed */}
          <section className="flex flex-col gap-md">
            <div className="flex gap-md overflow-x-auto border-b border-cohere-hairline pb-xs">
              {TABS.map((tab) => (
                <button
                  key={tab.key}
                  type="button"
                  onClick={() => selectTab(tab.key)}
                  aria-pressed={feedTab === tab.key}
                  className={`feed-tab ${feedTab === tab.key ? "active" : ""}`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            {isLoading ? (
              <div className="flex flex-col gap-md">
                {[0, 1, 2].map((i) => (
                  <div
                    key={i}
                    className="h-36 animate-pulse-soft rounded-lg border border-cohere-hairline bg-cohere-surface-low"
                  />
                ))}
              </div>
            ) : posts.length === 0 ? (
              <div className="card-base flex flex-col items-center gap-md p-xl text-center">
                <BrandIcon name="compose" size={56} />
                <h3 className="font-feature-title text-cohere-primary">还没有帖子</h3>
                <p className="font-body-main text-cohere-on-surface-variant">
                  发布第一篇帖子，看看 AI 角色会如何回应。
                </p>
                <Link to="/posts/new" className="btn-primary">
                  发布帖子
                </Link>
              </div>
            ) : (
              // Long feeds use Virtuoso so the DOM stays light as the list grows.
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

        {/* Right sidebar (4 cols) */}
        <aside className="flex flex-col gap-xl pt-lg lg:col-span-4">
          <HotTags tags={hotTags} onSelect={selectTag} />
          <ActiveAIRoles agents={activeAgents} />
          <RecentAIActivity activities={activities.slice(0, 10)} />
        </aside>
      </div>
    </main>
  );
}

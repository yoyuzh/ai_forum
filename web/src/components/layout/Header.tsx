import { NavLink, useNavigate } from "react-router-dom";
import { useState } from "react";
import { useUserStore } from "../../stores/useUserStore";
import { useConnectionStore } from "../../stores/useConnectionStore";
import { useNotifications } from "../../hooks/useNotifications";
import { forumLogo } from "../../assets/brand";
import BrandIcon from "../ui/BrandIcon";

const NAV_LINKS = [
  { to: "/", label: "首页", end: true },
  { to: "/posts", label: "帖子" },
  { to: "/agents", label: "AI 角色" },
];

export default function Header() {
  const currentUser = useUserStore((s) => s.currentUser);
  const { sseStatus } = useConnectionStore();
  const { notifications, unreadCount, markRead, markAllRead } = useNotifications();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);

  const onSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;
    navigate(`/posts?q=${encodeURIComponent(query.trim())}`);
  };

  return (
    <header className="bg-cohere-surface border-b border-cohere-hairline sticky top-0 z-50 w-full">
      <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-margin-mobile md:px-margin-desktop">
        <div className="flex items-center gap-lg">
          <NavLink to="/" className="flex items-center" aria-label="AI Forum 首页">
            <img src={forumLogo} alt="AI Forum Research Lab" className="h-9 w-auto" />
          </NavLink>

          <form onSubmit={onSearch} className="relative hidden items-center md:flex">
            <BrandIcon name="search" size={18} className="absolute left-2 opacity-70" />
            <input
              id="site-search"
              name="q"
              type="text"
              autoComplete="off"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="搜索帖子、标签、AI 角色…"
              aria-label="搜索"
              className="w-56 focus:w-64 rounded-sm border border-cohere-hairline bg-cohere-surface-low py-1 pl-8 pr-3 font-label-mono text-label-mono text-cohere-on-surface placeholder:text-cohere-muted transition-all duration-300 ease-cohere focus:border-cohere-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
            />
          </form>
        </div>

        <nav className="hidden items-center gap-lg md:flex" aria-label="主导航">
          {NAV_LINKS.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.end}
              className={({ isActive }) =>
                `border-b-2 pb-1 transition-all duration-300 ease-cohere hover:-translate-y-[1px] active:translate-y-0 ${
                  isActive
                    ? "border-cohere-primary text-cohere-primary font-bold"
                    : "border-transparent text-cohere-on-surface-variant hover:text-cohere-primary"
                }`
              }
            >
              {link.label}
            </NavLink>
          ))}
        </nav>

        <div className="flex items-center gap-md">
          {/* SSE live status — discreet dot, matches the "AI 分析中" language. */}
          <span
            className="hidden items-center gap-1 font-micro text-cohere-on-surface-variant sm:flex"
            title={`SSE 连接状态：${sseStatus}`}
          >
            <span
              className={`status-dot ${
                sseStatus === "connected"
                  ? "bg-cohere-secondary"
                  : sseStatus === "connecting"
                    ? "bg-cohere-coral animate-pulse-soft"
                    : "bg-cohere-muted"
              }`}
            />
            {sseStatus === "connected" ? "实时在线" : sseStatus}
          </span>

          <div className="relative">
            <button
              type="button"
              onClick={() => setOpen((value) => !value)}
              className="relative flex items-center justify-center rounded-full p-sm text-cohere-on-surface-variant hover:bg-cohere-surface-variant hover:text-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
              aria-label={`通知 ${unreadCount}`}
            >
              <BrandIcon name="notification" size={22} />
              {unreadCount > 0 && (
                <span className="absolute -right-1 -top-1 min-w-5 rounded-full bg-cohere-error px-1 text-center font-micro text-[11px] text-white">
                  {unreadCount}
                </span>
              )}
            </button>
            {open && (
              <div className="absolute right-0 top-12 z-50 w-72 rounded-sm border border-cohere-hairline bg-cohere-surface-lowest p-sm shadow-sm">
                <div className="mb-sm flex items-center justify-between">
                  <span className="font-label-mono-bold text-cohere-primary">通知</span>
                  <button
                    type="button"
                    onClick={() => markAllRead()}
                    className="font-micro text-cohere-on-surface-variant hover:text-cohere-primary"
                  >
                    全部已读
                  </button>
                </div>
                <div className="flex max-h-80 flex-col overflow-y-auto">
                  {notifications.length === 0 ? (
                    <p className="p-sm font-caption text-cohere-on-surface-variant">暂无通知</p>
                  ) : (
                    notifications.map((item) => (
                      <button
                        key={item.id}
                        type="button"
                        onClick={() => markRead(item.id)}
                        className="rounded-sm p-sm text-left transition-colors hover:bg-cohere-surface-low"
                      >
                        <div className="font-caption text-cohere-primary">{item.title}</div>
                        {item.body && (
                          <div className="mt-xxs line-clamp-2 font-micro text-cohere-on-surface-variant">
                            {item.body}
                          </div>
                        )}
                      </button>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>

          <NavLink
            to="/profile"
            className={({ isActive }) =>
              `flex items-center gap-xs rounded-pill px-md py-xs font-label-mono-bold transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue focus-visible:ring-offset-2 focus-visible:ring-offset-cohere-background ${
                isActive
                  ? "bg-cohere-secondary text-cohere-on-secondary"
                  : "bg-cohere-primary text-cohere-on-primary hover:bg-cohere-ink"
              }`
            }
            aria-label="个人中心"
          >
            {currentUser && (
              <img
                src={currentUser.avatar}
                alt={currentUser.nickname || currentUser.username}
                width={20}
                height={20}
                className="h-5 w-5 rounded-full border border-cohere-on-primary/30"
              />
            )}
            {!currentUser && <BrandIcon name="profile" size={18} />}
            个人中心
          </NavLink>
        </div>
      </div>
    </header>
  );
}

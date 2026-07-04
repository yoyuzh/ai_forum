import { NavLink } from "react-router-dom";
import MaterialIcon from "../MaterialIcon";
import { adminApi } from "../../api/client";

interface NavItem {
  to: string;
  label: string;
  icon: string;
}

const NAV: NavItem[] = [
  { to: "/", label: "仪表盘", icon: "dashboard" },
  { to: "/users", label: "用户管理", icon: "group" },
  { to: "/posts", label: "帖子管理", icon: "article" },
  { to: "/comments", label: "评论管理", icon: "comment" },
  { to: "/agents", label: "AI 角色管理", icon: "smart_toy" },
  { to: "/tasks", label: "AI 任务", icon: "assignment" },
  { to: "/decisions", label: "AI 决策", icon: "psychology" },
  { to: "/tags", label: "标签", icon: "sell" },
  { to: "/preferences", label: "偏好", icon: "tune" },
];

/** Fixed left sidebar — operational nav, dense and console-styled. */
export default function SideNav() {
  return (
    <aside className="fixed left-0 top-16 z-40 hidden h-[calc(100vh-64px)] w-64 flex-col gap-md overflow-y-auto bg-cohere-surface-low p-md md:flex">
      <div className="mb-lg px-xs">
        <div className="font-headline-lg font-bold text-cohere-primary">后台管理</div>
        <div className="mt-xs font-caption text-cohere-muted">系统管理</div>
      </div>

      <nav className="flex flex-col gap-xs">
        {NAV.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === "/"}
            className={({ isActive }) =>
              `flex items-center gap-sm rounded-lg px-sm py-xs font-label-mono transition-transform transition-colors active:scale-95 focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue ${
                isActive
                  ? "bg-cohere-secondary-container text-cohere-on-secondary-container font-bold"
                  : "text-cohere-on-surface-variant hover:bg-cohere-surface-variant hover:text-cohere-primary"
              }`
            }
          >
            {({ isActive }) => (
              <>
                <MaterialIcon name={item.icon} fill={isActive} size={20} />
                {item.label}
              </>
            )}
          </NavLink>
        ))}
      </nav>

      <div className="mt-auto border-t border-cohere-hairline pt-lg">
        <button
          type="button"
          onClick={() => {
            void adminApi.auth.logout().then(() => location.assign("/login"));
          }}
          className="flex items-center gap-sm rounded-lg px-sm py-xs font-label-mono text-cohere-on-surface-variant transition-colors hover:bg-cohere-surface-variant hover:text-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
        >
          <MaterialIcon name="logout" size={20} />
          退出登录
        </button>
      </div>
    </aside>
  );
}

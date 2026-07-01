import { NavLink } from "react-router-dom";
import MaterialIcon from "../MaterialIcon";

interface NavItem {
  to: string;
  label: string;
  icon: string;
}

const NAV: NavItem[] = [
  { to: "/", label: "仪表盘", icon: "dashboard" },
  { to: "/users", label: "用户管理", icon: "group" },
  { to: "/posts", label: "帖子管理", icon: "article" },
  { to: "/agents", label: "AI 角色管理", icon: "smart_toy" },
  { to: "/tasks", label: "AI 任务", icon: "assignment" },
  { to: "/decisions", label: "AI 决策", icon: "psychology" },
];

/** Fixed left sidebar — operational nav, dense and console-styled. */
export default function SideNav() {
  return (
    <aside className="fixed left-0 top-16 z-40 hidden h-[calc(100vh-64px)] w-64 flex-col gap-md overflow-y-auto border-r border-cohere-hairline bg-cohere-surface-low p-md md:flex">
      <div className="mb-lg px-sm">
        <div className="font-label-mono-bold text-cohere-on-surface">后台管理</div>
        <div className="font-micro text-cohere-muted">系统管理</div>
      </div>

      {NAV.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.to === "/"}
          className={({ isActive }) =>
            `flex items-center gap-md rounded-lg px-md py-sm font-label-mono transition-transform transition-colors active:scale-95 focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue ${
              isActive
                ? "bg-cohere-secondary-container text-cohere-on-secondary-container font-bold"
                : "text-cohere-on-surface-variant hover:bg-cohere-surface-variant"
            }`
          }
        >
          {({ isActive }) => (
            <>
              <MaterialIcon name={item.icon} fill={isActive} />
              {item.label}
            </>
          )}
        </NavLink>
      ))}

      <div className="mt-auto border-t border-cohere-hairline pt-lg">
        <button
          type="button"
          className="flex items-center gap-md rounded-lg px-md py-sm font-label-mono text-cohere-on-surface-variant transition-colors hover:bg-cohere-surface-variant focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
        >
          <MaterialIcon name="logout" />
          退出登录
        </button>
      </div>
    </aside>
  );
}

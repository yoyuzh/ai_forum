import MaterialIcon from "../MaterialIcon";

const publicAppUrl = (import.meta.env.VITE_PUBLIC_APP_URL ?? "http://localhost:5173").replace(/\/$/, "");
const publicHref = (path: string) => `${publicAppUrl}${path}`;

/** Sticky top nav — compact product rail above the admin console. */
export default function TopNav() {
  return (
    <header className="sticky top-0 z-50 flex h-16 w-full items-center justify-between border-b border-cohere-hairline bg-cohere-surface px-margin-mobile md:grid md:grid-cols-[1fr_auto_1fr] md:px-margin-desktop">
      <div className="flex items-center">
        <a
          href={publicHref("/")}
          className="whitespace-nowrap text-[28px] font-black leading-none text-cohere-primary focus:outline-none focus-visible:underline md:font-headline-lg"
        >
          AI Forum
        </a>
      </div>

      <nav className="hidden items-center gap-lg font-body-main text-cohere-on-surface-variant md:flex" aria-label="主站导航">
        <a className="transition-colors hover:text-cohere-primary focus:outline-none focus-visible:underline" href={publicHref("/")}>
          首页
        </a>
        <a className="transition-colors hover:text-cohere-primary focus:outline-none focus-visible:underline" href={publicHref("/posts")}>
          帖子
        </a>
        <a className="transition-colors hover:text-cohere-primary focus:outline-none focus-visible:underline" href={publicHref("/agents")}>
          AI 角色
        </a>
      </nav>

      <div className="flex items-center justify-end gap-sm">
        <span className="p-xs text-cohere-on-surface-variant" aria-label="通知" role="img">
          <MaterialIcon name="notifications" size={20} />
        </span>
        <a
          href={publicHref("/profile")}
          className="rounded-pill border border-cohere-hairline bg-cohere-surface px-md py-xs font-label-mono-bold text-cohere-primary transition-colors hover:border-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
        >
          个人中心
        </a>
      </div>
    </header>
  );
}

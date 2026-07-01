import MaterialIcon from "../MaterialIcon";

/** Sticky top nav — logo, public-site link, notifications, profile. */
export default function TopNav() {
  return (
    <header className="sticky top-0 z-50 flex h-16 w-full items-center justify-between border-b border-cohere-hairline bg-cohere-surface px-margin-mobile md:px-margin-desktop">
      <div className="flex items-center gap-lg">
        <span className="font-headline-lg font-black text-cohere-primary">AI Forum</span>
        <span className="hidden font-micro text-cohere-muted md:inline">Admin Console</span>
      </div>

      <div className="flex items-center gap-md">
        <button
          type="button"
          className="rounded-full p-xs text-cohere-on-surface-variant transition-colors hover:bg-cohere-surface-variant hover:text-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
          aria-label="通知"
        >
          <MaterialIcon name="notifications" />
        </button>
        <button
          type="button"
          className="rounded-pill border border-cohere-hairline px-md py-xs font-label-mono-bold text-cohere-primary transition-colors hover:border-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
        >
          个人中心
        </button>
      </div>
    </header>
  );
}

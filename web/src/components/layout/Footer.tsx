const FOOTER_LINKS = [
  { label: "服务条款", href: "#" },
  { label: "隐私政策", href: "#" },
  { label: "API 文档", href: "#" },
  { label: "联系我们", href: "#" },
];

export default function Footer() {
  return (
    <footer className="mt-auto w-full border-t border-cohere-hairline bg-cohere-surface py-section">
      <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-lg px-margin-mobile px-margin-desktop md:flex-row md:px-margin-desktop">
        <div className="flex items-baseline gap-md">
          <span className="font-headline-lg text-cohere-primary">AI Forum</span>
          <span className="font-micro text-cohere-muted">© 2024 AI Forum Research Lab</span>
        </div>
        <div className="flex flex-wrap justify-center gap-md">
          {FOOTER_LINKS.map((link) => (
            <a
              key={link.label}
              href={link.href}
              className="font-micro text-cohere-muted transition-colors hover:text-cohere-primary"
            >
              {link.label}
            </a>
          ))}
        </div>
      </div>
    </footer>
  );
}

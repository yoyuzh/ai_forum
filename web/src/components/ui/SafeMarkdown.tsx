import DOMPurify from "dompurify";
import ReactMarkdown from "react-markdown";

interface SafeMarkdownProps {
  content: string;
  className?: string;
}

/**
 * Renders Markdown with all HTML sanitized through DOMPurify first.
 *
 * Per web/AGENTS.md: never render user-supplied rich text without sanitization.
 * We sanitize the raw source, then let react-markdown parse the cleaned string
 * into safe React elements (no dangerouslySetInnerHTML).
 */
export default function SafeMarkdown({ content, className }: SafeMarkdownProps) {
  const cleaned = DOMPurify.sanitize(content, {
    USE_PROFILES: { html: false, mathMl: false, svg: false },
    ALLOWED_ATTR: [],
  });

  return (
    <div className={`${className ?? "prose-cohere"} [&_p]:text-cohere-ink`}>
      <ReactMarkdown>{cleaned}</ReactMarkdown>
    </div>
  );
}

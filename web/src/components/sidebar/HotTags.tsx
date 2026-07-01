interface HotTagsProps {
  tags: string[];
  onSelect?: (tag: string) => void;
}

const TAG_WEIGHTS: Record<string, "lg" | "base" | "sm"> = {
  Rust: "lg",
  架构设计: "lg",
  "大模型微调": "base",
  k8s: "base",
  Vite: "sm",
};

/** Tag cloud sidebar card. */
export default function HotTags({ tags, onSelect }: HotTagsProps) {
  return (
    <div className="card-surface-low p-md">
      <h3 className="mb-md border-b border-cohere-hairline pb-xs font-label-mono-bold text-cohere-primary">
        热门标签
      </h3>
      <div className="flex flex-wrap gap-sm">
        {tags.map((tag) => {
          const weight = TAG_WEIGHTS[tag] ?? "base";
          const size =
            weight === "lg" ? "text-feature-title" : weight === "base" ? "text-body-main" : "text-caption";
          return (
            <button
              key={tag}
              type="button"
              onClick={() => onSelect?.(tag)}
              className={`rounded font-caption ${size} text-cohere-on-surface-variant transition-colors hover:bg-cohere-surface-variant hover:text-cohere-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue`}
            >
              {tag}
            </button>
          );
        })}
      </div>
    </div>
  );
}

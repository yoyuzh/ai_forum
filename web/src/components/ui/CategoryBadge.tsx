interface CategoryBadgeProps {
  category: string;
}

/** Coral-outlined taxonomy label — the editorial category marker from Cohere. */
export default function CategoryBadge({ category }: CategoryBadgeProps) {
  return (
    <span className="whitespace-nowrap rounded border border-cohere-on-surface-variant px-1 py-0.5 font-micro text-micro text-cohere-on-surface-variant">
      {category}
    </span>
  );
}

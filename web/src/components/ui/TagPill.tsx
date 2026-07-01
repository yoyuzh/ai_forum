interface TagPillProps {
  tag: string;
  variant?: "default" | "coral";
  onClick?: () => void;
}

/** Content tag pill. Default variant is the neutral surface chip; coral is the
 *  post-detail sidebar style with a coral outline. */
export default function TagPill({ tag, variant = "default", onClick }: TagPillProps) {
  const base =
    "rounded px-sm py-0.5 font-caption text-caption transition-colors cursor-pointer";
  const styles =
    variant === "coral"
      ? "border border-cohere-coral text-cohere-coral bg-cohere-surface-lowest hover:bg-cohere-coral hover:text-white"
      : "bg-cohere-surface-container text-cohere-on-surface-variant hover:bg-cohere-surface-variant";

  const Comp = onClick ? "button" : "span";
  return (
    <Comp
      type={onClick ? "button" : undefined}
      onClick={onClick}
      className={`${base} ${styles}`}
    >
      {variant === "coral" ? tag : `#${tag}`}
    </Comp>
  );
}

interface MaterialIconProps {
  name: string;
  className?: string;
  /** Render the filled variant of the symbol. */
  fill?: boolean;
  /** Pixel size — defaults to inherit from surrounding font. */
  size?: number;
  "aria-hidden"?: boolean;
}

/**
 * Wrapper around Material Symbols Outlined. The font is loaded in index.html;
 * base styling lives in src/styles/index.css under `.material-symbols-outlined`.
 */
export default function MaterialIcon({
  name,
  className = "",
  fill = false,
  size,
  ...rest
}: MaterialIconProps) {
  const style = size ? { fontSize: `${size}px` } : undefined;
  return (
    <span
      aria-hidden={rest["aria-hidden"] ?? true}
      className={`material-symbols-outlined ${fill ? "fill" : ""} ${className}`.trim()}
      style={style}
    >
      {name}
    </span>
  );
}

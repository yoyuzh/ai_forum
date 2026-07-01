interface MaterialIconProps {
  name: string;
  className?: string;
  fill?: boolean;
  size?: number;
}

export default function MaterialIcon({
  name,
  className = "",
  fill = false,
  size,
}: MaterialIconProps) {
  const style = size ? { fontSize: `${size}px` } : undefined;
  return (
    <span
      aria-hidden
      className={`material-symbols-outlined ${fill ? "fill" : ""} ${className}`.trim()}
      style={style}
    >
      {name}
    </span>
  );
}

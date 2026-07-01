import MaterialIcon from "./MaterialIcon";

interface StatCardProps {
  label: string;
  value: string;
  icon: string;
  variant?: "default" | "secondary" | "error";
}

const VARIANT: Record<NonNullable<StatCardProps["variant"]>, string> = {
  default: "bg-cohere-surface-container text-cohere-primary",
  secondary: "bg-cohere-secondary text-cohere-on-secondary",
  error: "bg-cohere-error-container text-cohere-error",
};

const ICON_COLOR: Record<NonNullable<StatCardProps["variant"]>, string> = {
  default: "text-cohere-muted",
  secondary: "text-cohere-on-secondary opacity-80",
  error: "text-cohere-error",
};

/** Bento stat card — label, oversized value, icon. Matches the dashboard prototype. */
export default function StatCard({ label, value, icon, variant = "default" }: StatCardProps) {
  return (
    <div
      className={`flex h-32 flex-col justify-between rounded-xl border border-cohere-hairline p-md ${VARIANT[variant]}`}
    >
      <div className="flex items-center justify-between">
        <span className="font-label-mono uppercase tracking-wider opacity-80">{label}</span>
        <MaterialIcon name={icon} className={ICON_COLOR[variant]} size={20} />
      </div>
      <div className="font-display-lg text-display-lg leading-none">{value}</div>
    </div>
  );
}

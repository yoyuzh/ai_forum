import { TaskStatus } from "../api/types";

interface StatusBadgeProps {
  status: TaskStatus;
}

const CONFIG: Record<
  TaskStatus,
  { label: string; className: string; icon: string; dotClass: string }
> = {
  PENDING: {
    label: "Waiting",
    className:
      "bg-cohere-surface-highest text-cohere-on-surface border border-cohere-hairline",
    icon: "hourglass_empty",
    dotClass: "bg-cohere-muted",
  },
  PROCESSING: {
    label: "Generating",
    className: "bg-cohere-action-blue text-white",
    icon: "autorenew",
    dotClass: "bg-white animate-pulse-soft",
  },
  COMPLETED: {
    label: "Success",
    className:
      "bg-cohere-success text-cohere-secondary border border-cohere-secondary-container",
    icon: "check",
    dotClass: "bg-cohere-secondary",
  },
  FAILED: {
    label: "Failed",
    className: "bg-cohere-error-container text-cohere-error border border-cohere-error",
    icon: "error",
    dotClass: "bg-cohere-error",
  },
};

export default function StatusBadge({ status }: StatusBadgeProps) {
  const config = CONFIG[status];
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-1 py-0.5 font-label-mono text-[11px] ${config.className}`}
    >
      <span className={`status-dot ${config.dotClass}`} />
      {config.label}
    </span>
  );
}

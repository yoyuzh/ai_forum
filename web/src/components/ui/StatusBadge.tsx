import { AIStatus } from "../../api/types";

interface StatusBadgeProps {
  status: AIStatus;
  responsesCount?: number;
}

const CONFIG: Record<
  AIStatus,
  { label: string; className: string; icon: string; dotClass: string }
> = {
  PENDING: {
    label: "AI 待处理",
    className: "bg-cohere-surface-container text-cohere-on-surface-variant border border-cohere-hairline",
    icon: "schedule",
    dotClass: "bg-cohere-on-surface-variant",
  },
  PROCESSING: {
    label: "AI 分析中",
    className: "bg-cohere-secondary-container text-cohere-on-secondary-container",
    icon: "psychology",
    dotClass: "bg-cohere-secondary animate-pulse-soft",
  },
  COMPLETED: {
    label: "AI 已回复",
    className: "bg-cohere-success text-cohere-deep-green",
    icon: "check_circle",
    dotClass: "bg-cohere-secondary",
  },
};

/** Pill showing where a post sits in the AI reply pipeline. */
export default function StatusBadge({ status, responsesCount }: StatusBadgeProps) {
  const config = CONFIG[status];
  const label =
    status === "COMPLETED" && responsesCount !== undefined
      ? `AI 已回复 (${responsesCount})`
      : config.label;

  return (
    <span
      className={`inline-flex items-center gap-1 rounded px-1 py-0.5 font-micro text-micro ${config.className}`}
    >
      <span className={`status-dot ${config.dotClass}`} />
      {label}
    </span>
  );
}

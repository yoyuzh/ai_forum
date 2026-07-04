import { AIResponder, AIStatus } from "../../api/types";

interface StatusBadgeProps {
  status: AIStatus;
  responsesCount?: number;
  responders?: AIResponder[];
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
    className: "bg-cohere-success text-cohere-deep-green border border-cohere-deep-green/10",
    icon: "check_circle",
    dotClass: "bg-cohere-secondary",
  },
};

/** Pill showing where a post sits in the AI reply pipeline. */
export default function StatusBadge({ status, responsesCount, responders = [] }: StatusBadgeProps) {
  const config = CONFIG[status];
  const label =
    status === "COMPLETED" && responsesCount !== undefined
      ? `AI 已回复 (${responsesCount})`
      : config.label;

  return (
    <span
      className={`group/status relative inline-flex items-center gap-1 rounded px-1 py-0.5 font-micro text-micro ${config.className}`}
    >
      {status === "COMPLETED" && responders.length > 0 ? (
        <span className="flex -space-x-1">
          {responders.slice(0, 3).map((responder, idx) => (
            <img
              key={`${responder.name}-${idx}`}
              src={responder.avatar}
              alt={responder.name}
              width={18}
              height={18}
              title={responder.name}
              className="h-[18px] w-[18px] rounded-full border-2 border-cohere-surface-lowest object-cover"
              style={{ borderColor: responder.accentColor ?? undefined }}
            />
          ))}
        </span>
      ) : (
        <span className={`status-dot ${config.dotClass}`} />
      )}
      {label}
      {status === "COMPLETED" && responders.length > 0 && (
        <span className="pointer-events-none absolute right-0 top-full z-20 mt-xs hidden min-w-max rounded-sm border border-cohere-hairline bg-cohere-surface-lowest px-sm py-xs text-left font-micro text-cohere-on-surface-variant shadow-sm group-hover/status:block">
          {responders.slice(0, 3).map((responder) => responder.name).join("、")}
        </span>
      )}
    </span>
  );
}

import { ProcessingStep } from "../../api/types";
import MaterialIcon from "../ui/MaterialIcon";

interface AIProcessingStatusProps {
  steps: ProcessingStep[];
  summary: { done: number; running: number; failed: number };
  /** Top-bar progress fraction (0–1). */
  progress: number;
  active: boolean;
}

const STEP_STYLE: Record<ProcessingStep["state"], string> = {
  done:
    "border-cohere-hairline bg-cohere-surface-highest text-cohere-on-surface-variant",
  active:
    "border-cohere-action-blue bg-cohere-action-blue text-white",
  pending:
    "border-cohere-hairline bg-cohere-surface text-cohere-on-surface-variant",
};

/** 4-step AI pipeline stepper — analyzes tags → scores willingness → generates
 *  reply → writes to comments. Matches the _2 prototype "AI 处理状态" panel. */
export default function AIProcessingStatus({
  steps,
  summary,
  progress,
  active,
}: AIProcessingStatusProps) {
  return (
    <section className="card-base relative mb-lg overflow-hidden p-md">
      <div className="absolute left-0 top-0 h-1 w-full bg-cohere-surface-variant">
        <div
          className={`h-full bg-cohere-action-blue transition-all duration-500 ease-cohere ${active ? "animate-pulse-soft" : ""}`}
          style={{ width: `${Math.round(progress * 100)}%` }}
        />
      </div>

      <div className="mb-md mt-sm flex items-center justify-between">
        <h2 className="font-feature-title text-[18px] text-cohere-ink">AI 处理状态</h2>
        {active ? (
          <div className="flex items-center gap-1 rounded-full bg-cohere-pale-blue px-1 py-0.5 font-label-mono-bold text-[10px] text-cohere-action-blue">
            <span className="status-dot bg-cohere-action-blue animate-pulse-soft" />
            生成中
          </div>
        ) : (
          <div className="flex items-center gap-1 rounded-full bg-cohere-success px-1 py-0.5 font-label-mono-bold text-[10px] text-cohere-secondary">
            <MaterialIcon name="check" size={12} />
            已完成
          </div>
        )}
      </div>

      <div className="relative z-10 flex flex-col gap-md">
        {steps.map((step) => (
          <div
            key={step.key}
            className="flex items-center gap-md transition-all duration-300 ease-cohere"
          >
            <div
              className={`flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full border transition-all duration-300 ease-cohere ${STEP_STYLE[step.state]}`}
            >
              <MaterialIcon name={step.icon} size={16} />
            </div>
            <div>
              <div className="font-label-mono-bold text-micro text-cohere-on-surface">
                {step.label}
              </div>
              <div
                className={`font-caption text-[12px] ${
                  step.state === "active" ? "text-cohere-action-blue" : "text-cohere-on-surface-variant"
                }`}
              >
                {step.detail}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-md flex justify-between border-t border-cohere-hairline pt-md font-label-mono text-[10px]">
        <div className="flex flex-col">
          <span className="text-cohere-on-surface-variant">已完成</span>
          <span className="font-label-mono-bold text-cohere-ink">{summary.done}</span>
        </div>
        <div className="flex flex-col">
          <span className="text-cohere-on-surface-variant">运行中</span>
          <span className="font-label-mono-bold text-cohere-action-blue">{summary.running}</span>
        </div>
        <div className="flex flex-col">
          <span className="text-cohere-on-surface-variant">失败</span>
          <span className="font-label-mono-bold text-cohere-error">{summary.failed}</span>
        </div>
      </div>
    </section>
  );
}

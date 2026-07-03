import { Drawer } from "antd";
import type { AdminDecisionLog } from "../../api/types";

interface Props {
  log: AdminDecisionLog | null;
  onClose: () => void;
}

export default function DecisionDetailDrawer({ log, onClose }: Props) {
  if (!log) return null;
  const below = log.willingnessScore < log.thresholdValue;

  return (
    <Drawer title="Decision detail" open={!!log} onClose={onClose} width={560}>
      <div className="space-y-lg">
        <section>
          <div className="font-label-mono text-cohere-muted">Agent</div>
          <div className="font-feature-title text-cohere-primary">{log.aiAgentName}</div>
          <div className="font-caption text-cohere-muted">Post #{log.postId} · {log.triggerType}</div>
        </section>
        <WillingnessGauge score={log.willingnessScore} threshold={log.thresholdValue} />
        <section>
          <div className="font-label-mono text-cohere-muted">Decision</div>
          <div className="font-body-main text-cohere-on-surface">
            {log.decision} · {log.fallback ? "fallback true" : "fallback false"} · {below ? "below threshold" : "above threshold"}
          </div>
        </section>
        <HitTagsViewer tags={log.hitTags} />
        <SkipReasonBlock reason={log.reason} />
        <section>
          <div className="font-label-mono text-cohere-muted">Result link</div>
          <div className="font-body-main text-cohere-on-surface">
            {log.taskId ? `Task #${log.taskId}` : log.commentLink ? `Comment #${log.commentLink}` : "no reply produced"}
          </div>
        </section>
      </div>
    </Drawer>
  );
}

export function WillingnessGauge({ score, threshold }: { score: number; threshold: number }) {
  const scorePct = score <= 1 ? score * 100 : score;
  const thresholdPct = threshold <= 1 ? threshold * 100 : threshold;
  const below = scorePct < thresholdPct;
  return (
    <section aria-label={`willingness ${scorePct.toFixed(0)} threshold ${thresholdPct.toFixed(0)} ${below ? "below threshold" : "above threshold"}`}>
      <div className="mb-1 flex justify-between font-label-mono">
        <span>{(score <= 1 ? score : score / 100).toFixed(2)} / {(threshold <= 1 ? threshold : threshold / 100).toFixed(2)}</span>
        <span className={below ? "text-cohere-coral" : "text-cohere-action-blue"}>
          {below ? "below threshold" : "above threshold"}
        </span>
      </div>
      <div className="relative h-3 rounded-full bg-cohere-surface-variant">
        <div className={`h-3 rounded-full ${below ? "bg-cohere-coral" : "bg-cohere-action-blue"}`} style={{ width: `${Math.min(scorePct, 100)}%` }} />
        <div className="absolute top-[-4px] h-5 w-0.5 bg-cohere-primary" style={{ left: `${Math.min(thresholdPct, 100)}%` }} />
      </div>
    </section>
  );
}

export function HitTagsViewer({ tags }: { tags: string[] }) {
  return (
    <section>
      <div className="font-label-mono text-cohere-muted">Hit tags</div>
      <div className="mt-1 flex flex-wrap gap-1">
        {tags.length ? tags.map((tag) => <span key={tag} className="rounded bg-cohere-surface-container px-1 py-0.5 font-label-mono text-cohere-on-surface">{tag}</span>) : <span className="font-caption text-cohere-muted">None</span>}
      </div>
    </section>
  );
}

export function SkipReasonBlock({ reason }: { reason: string }) {
  return (
    <section className="rounded-lg border border-cohere-hairline bg-cohere-surface-low p-md">
      <div className="font-label-mono text-cohere-muted">Skip reason</div>
      <p className="mt-1 font-body-main text-cohere-on-surface">{reason || "no skip reason"}</p>
    </section>
  );
}

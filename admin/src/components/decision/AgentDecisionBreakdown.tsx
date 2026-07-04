import type { AdminDecisionLog } from "../../api/types";

export default function AgentDecisionBreakdown({ logs }: { logs: AdminDecisionLog[] }) {
  const rows = Object.values(
    logs.reduce<Record<string, { name: string; total: number; replies: number; fallback: number; score: number }>>((acc, log) => {
      const key = String(log.aiAgentId);
      acc[key] ??= { name: log.aiAgentName, total: 0, replies: 0, fallback: 0, score: 0 };
      acc[key].total += 1;
      acc[key].replies += log.decision === "REPLY" ? 1 : 0;
      acc[key].fallback += log.decision === "FALLBACK" || log.fallback ? 1 : 0;
      acc[key].score += log.willingnessScore <= 1 ? log.willingnessScore : log.willingnessScore / 100;
      return acc;
    }, {}),
  );

  return (
    <div className="mb-lg grid grid-cols-1 gap-md md:grid-cols-3">
      {rows.map((row) => (
        <div key={row.name} className="rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md">
          <div className="font-label-mono-bold text-cohere-primary">{row.name}</div>
          <div className="mt-sm grid grid-cols-3 gap-sm font-caption text-cohere-muted">
            <span>reply-rate {(row.replies / row.total).toFixed(2)}</span>
            <span>avg willingness {(row.score / row.total).toFixed(2)}</span>
            <span>fallback-rate {(row.fallback / row.total).toFixed(2)}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

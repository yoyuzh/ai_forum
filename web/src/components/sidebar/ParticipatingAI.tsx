import { useMemo } from "react";
import { AIDecisionLog } from "../../api/types";
import { useAgents } from "../../hooks/useAgents";
import MaterialIcon from "../ui/MaterialIcon";

interface ParticipatingAIProps {
  logs: AIDecisionLog[];
}

const STATE_STYLE: Record<string, { dot: string; label: string; opacity: string }> = {
  REPLY: { dot: "bg-cohere-action-blue", label: "已回复", opacity: "opacity-100" },
  IGNORE: { dot: "bg-cohere-muted", label: "忽略", opacity: "opacity-40" },
  FAILED: { dot: "bg-cohere-error", label: "失败", opacity: "opacity-70" },
};

/** Participating AI panel — per-agent willingness score and decision state. */
export default function ParticipatingAI({ logs }: ParticipatingAIProps) {
  const { agents } = useAgents();

  const agentMap = useMemo(() => {
    return new Map(agents.map((a) => [a.id, { avatar: a.avatar, icon: a.icon }]));
  }, [agents]);

  return (
    <div className="card-base p-lg">
      <h3 className="mb-md font-feature-title text-[18px] text-cohere-ink">参与本帖的 AI</h3>
      <div className="space-y-md">
        {logs.length === 0 && (
          <p className="font-micro text-cohere-muted">暂无 AI 决策记录。</p>
        )}
        {logs.map((log) => {
          const style = STATE_STYLE[log.decision] ?? STATE_STYLE.IGNORE;
          const score = Math.round(log.willingnessScore * 100);
          const agentInfo = agentMap.get(log.aiAgentId);

          return (
            <div key={log.id} className={`group flex items-center justify-between transition-all duration-300 ease-cohere hover:bg-cohere-surface-container/20 p-xs -mx-xs rounded-sm ${style.opacity}`}>
              <div className="flex items-center gap-sm">
                <div className="relative h-8 w-8 flex-shrink-0 overflow-hidden rounded-md border border-cohere-hairline shadow-sm">
                  {agentInfo?.avatar ? (
                    <img
                      src={agentInfo.avatar}
                      alt={log.aiAgentName}
                      width={32}
                      height={32}
                      className="h-full w-full object-cover transition-transform duration-500 ease-spring group-hover:scale-110"
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center bg-cohere-secondary-container text-cohere-on-secondary-container">
                      <MaterialIcon name={agentInfo?.icon || "psychology"} size={16} />
                    </div>
                  )}
                </div>
                <div>
                  <div className="font-label-mono-bold text-cohere-ink">{log.aiAgentName}</div>
                  <div className="flex items-center gap-1 font-micro text-cohere-muted">
                    <span className={`status-dot ${style.dot}`} />
                    {style.label}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="font-label-mono text-[14px] text-cohere-ink">{score}</div>
                <div className="font-micro text-cohere-muted">意愿分</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

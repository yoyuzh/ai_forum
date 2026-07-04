import { useMemo } from "react";
import { AIDecisionLog, Comment } from "../../api/types";
import { useAgents } from "../../hooks/useAgents";
import MaterialIcon from "../ui/MaterialIcon";

interface ParticipatingAIProps {
  logs: AIDecisionLog[];
  comments: Comment[];
}

interface Participant {
  key: string;
  aiAgentId?: number;
  name: string;
  avatar?: string;
  score?: number;
  hasAnswered: boolean;
}

function scoreLabel(score?: number): string {
  if (score === undefined) return "—";
  return String(Math.round(score > 1 ? score : score * 100));
}

/** Participating AI panel — only selected or already-answering agents. */
export default function ParticipatingAI({ logs, comments }: ParticipatingAIProps) {
  const { agents } = useAgents();

  const agentMap = useMemo(() => {
    return new Map(agents.map((a) => [a.id, { avatar: a.avatar, icon: a.icon }]));
  }, [agents]);

  const participants = useMemo(() => {
    const byKey = new Map<string, Participant>();

    for (const log of logs
      .filter((item) => item.decision === "REPLY")
      .sort((a, b) => b.willingnessScore - a.willingnessScore)) {
      const key = `agent:${log.aiAgentId}`;
      if (byKey.has(key)) continue;
      byKey.set(key, {
        key,
        aiAgentId: log.aiAgentId,
        name: log.aiAgentName,
        score: log.willingnessScore,
        hasAnswered: false,
      });
    }

    for (const comment of comments) {
      if (!comment.author.isAi) continue;
      const key = comment.author.aiAgentId ? `agent:${comment.author.aiAgentId}` : `comment:${comment.id}`;
      const current = byKey.get(key);
      byKey.set(key, {
        key,
        aiAgentId: comment.author.aiAgentId,
        name: current?.name ?? comment.author.role ?? comment.author.username,
        avatar: comment.author.avatar,
        score: comment.willingnessScore ?? current?.score,
        hasAnswered: true,
      });
    }

    return [...byKey.values()];
  }, [comments, logs]);

  if (participants.length === 0) {
    return null;
  }

  return (
    <div className="card-base p-lg">
      <h3 className="mb-md font-feature-title text-[18px] text-cohere-ink">参与本帖的 AI</h3>
      <div className="space-y-md">
        {participants.map((participant) => {
          const agentInfo = participant.aiAgentId ? agentMap.get(participant.aiAgentId) : undefined;
          const avatar = participant.avatar ?? agentInfo?.avatar;

          return (
            <div
              key={participant.key}
              className="group flex items-center justify-between rounded-sm p-xs -mx-xs transition-all duration-300 ease-cohere hover:bg-cohere-surface-container/20"
            >
              <div className="flex items-center gap-sm">
                <div className="relative h-8 w-8 flex-shrink-0 overflow-hidden rounded-md border border-cohere-hairline shadow-sm">
                  {avatar ? (
                    <img
                      src={avatar}
                      alt={participant.name}
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
                  <div className="font-label-mono-bold text-cohere-ink">{participant.name}</div>
                  <div className="flex items-center gap-1 font-micro text-cohere-on-surface-variant">
                    <span className="status-dot bg-cohere-action-blue" />
                    {participant.hasAnswered ? "已回答" : "已入选"}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="font-label-mono text-[14px] text-cohere-ink">{scoreLabel(participant.score)}</div>
                <div className="font-micro text-cohere-on-surface-variant">意愿分</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

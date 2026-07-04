import { Link } from "react-router-dom";
import { AIActivity } from "../../api/types";

interface RecentAIActivityProps {
  activities: AIActivity[];
}

/** Dotted-border timeline of recent AI actions across the forum. */
export default function RecentAIActivity({ activities }: RecentAIActivityProps) {
  return (
    <div className="card-surface-low flex flex-col gap-sm p-md">
      <h3 className="mb-xs border-b border-cohere-hairline pb-xs font-label-mono-bold text-cohere-primary">
        最近 AI 动态
      </h3>
      {activities.map((activity) => {
        const action = normalizeAction(activity.action);
        const title = normalizeTarget(activity.target);
        return (
          <div
            key={activity.id}
            className="group relative flex flex-col gap-xs border-l border-dotted border-cohere-hairline pl-md pb-md last:border-l-transparent last:pb-0 transition-all duration-300 ease-cohere hover:bg-cohere-surface-container/30 rounded-r-sm"
          >
            <span className="absolute -left-[5px] top-[4px] h-2 w-2 rounded-full bg-cohere-secondary transition-all duration-300 ease-spring group-hover:scale-125" />
            <span className="font-label-mono text-[10px] text-cohere-muted">
              {activity.relativeTime}
            </span>
            <p className="font-caption text-cohere-primary">
              <span className="font-semibold">{activity.agentName}</span> {action}了帖子
              <Link
                to={`/posts/${activity.targetId}`}
                className="text-cohere-action-blue hover:underline"
              >
                《{title}》
              </Link>
            </p>
          </div>
        );
      })}
    </div>
  );
}

function normalizeAction(action: string): "评论" | "点赞" | "参与讨论" {
  if (/like|点赞/i.test(action)) return "点赞";
  if (/follow|参与|讨论/i.test(action)) return "参与讨论";
  if (/reply|comment|评论|回复/i.test(action)) return "评论";
  return "参与讨论";
}

function normalizeTarget(target: string): string {
  const cleaned = target.replace(/\bIGNORE\b/gi, "").replace(/\bpost\b/gi, "").trim();
  return cleaned || "这篇讨论";
}

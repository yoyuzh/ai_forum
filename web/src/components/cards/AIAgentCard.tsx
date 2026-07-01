import { AIAgent } from "../../api/types";
import MaterialIcon from "../ui/MaterialIcon";

interface AIAgentCardProps {
  agent: AIAgent;
}

/** Agent showcase card — avatar, name, description, traits, activity status. */
export default function AIAgentCard({ agent }: AIAgentCardProps) {
  return (
    <article className="card-base group flex flex-col p-lg hover:border-cohere-secondary hover:-translate-y-[2px] hover:shadow-sm transition-all duration-300 ease-cohere">
      <div className="flex items-start gap-md">
        <div className="relative h-12 w-12 flex-shrink-0 overflow-hidden rounded-ai">
          <img
            src={agent.avatar}
            alt={agent.displayName}
            width={48}
            height={48}
            onError={(e) => {
              e.currentTarget.style.display = "none";
              const fb = e.currentTarget.parentElement?.querySelector(".avatar-fallback");
              if (fb) fb.classList.remove("hidden");
            }}
            className="h-12 w-12 rounded-ai object-cover border border-cohere-hairline transition-all duration-500 ease-spring group-hover:scale-105"
          />
          <div className="avatar-fallback hidden absolute inset-0 flex items-center justify-center rounded-ai bg-cohere-secondary-container text-cohere-on-secondary-container">
            <MaterialIcon name={agent.icon} size={24} />
          </div>
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-sm">
            <h3 className="font-label-mono-bold text-cohere-primary">{agent.displayName}</h3>
            <span
              className="inline-flex items-center gap-1 rounded border border-cohere-hairline px-1 py-0.5 font-micro text-micro text-cohere-muted"
              title={agent.active ? "运行中" : "已停用"}
            >
              <span
                className={`status-dot ${agent.active ? "bg-cohere-secondary" : "bg-cohere-muted"}`}
              />
              {agent.active ? "运行中" : "已停用"}
            </span>
          </div>
          <p className="mt-1 font-caption text-cohere-muted">{agent.description}</p>
        </div>
      </div>

      <div className="mt-md flex flex-wrap gap-1">
        {agent.traits.map((trait) => (
          <span
            key={trait}
            className="rounded border border-cohere-hairline bg-cohere-surface-container px-1 py-0.5 font-micro text-micro text-cohere-on-surface-variant"
          >
            {trait}
          </span>
        ))}
      </div>

      <div className="mt-md grid grid-cols-2 gap-md border-t border-cohere-hairline pt-md">
        <div>
          <div className="font-micro text-cohere-muted">专业领域</div>
          <div className="mt-1 flex flex-wrap gap-1">
            {agent.specialties.map((s) => (
              <span key={s} className="font-caption text-cohere-on-surface-variant">
                {s}
              </span>
            ))}
          </div>
        </div>
        <div>
          <div className="font-micro text-cohere-muted">回答意愿阈值</div>
          <div className="mt-1 flex items-center gap-1">
            <div className="h-1 w-full overflow-hidden rounded-full bg-cohere-surface-variant">
              <div
                className="h-full rounded-full bg-cohere-primary transition-all duration-700 ease-spring"
                style={{ width: `${agent.replyThreshold * 100}%` }}
              />
            </div>
            <span className="font-label-mono text-micro text-cohere-muted">
              {agent.replyThreshold.toFixed(2)}
            </span>
          </div>
        </div>
      </div>
    </article>
  );
}

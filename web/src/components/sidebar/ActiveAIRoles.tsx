import { Link } from "react-router-dom";
import { AIAgent } from "../../api/types";
import MaterialIcon from "../ui/MaterialIcon";

interface ActiveAIRolesProps {
  agents: AIAgent[];
}

/** Compact list of active AI agents shown in the home sidebar. */
export default function ActiveAIRoles({ agents }: ActiveAIRolesProps) {
  const primaryAccent = agents[0]?.accentColor ?? "#7FBFA0";

  return (
    <div
      className="flex flex-col gap-sm rounded-lg border border-cohere-hairline p-md"
      style={{
        background: `linear-gradient(135deg, ${primaryAccent}08 0%, rgba(255,255,255,0.96) 48%, ${primaryAccent}05 100%)`,
      }}
    >
      <h3 className="mb-xs border-b border-cohere-hairline pb-xs font-label-mono-bold text-cohere-primary">
        活跃 AI 角色
      </h3>
      {agents.map((agent) => (
        <Link
          key={agent.id}
          to="/agents"
          className="group flex items-center gap-sm rounded-sm p-sm transition-all duration-300 ease-cohere hover:bg-cohere-surface-container hover:-translate-y-[1px] active:translate-y-0"
        >
          <div className="relative h-10 w-10 flex-shrink-0 overflow-hidden rounded-ai">
            <img
              src={agent.avatar}
              alt={agent.displayName}
              width={40}
              height={40}
              onError={(e) => {
                e.currentTarget.style.display = "none";
                const fb = e.currentTarget.parentElement?.querySelector(".avatar-fallback");
                if (fb) fb.classList.remove("hidden");
              }}
              className="h-10 w-10 rounded-ai object-cover border-2 transition-all duration-500 ease-spring group-hover:scale-105"
              style={{ borderColor: agent.accentColor }}
            />
            <div className="avatar-fallback hidden absolute inset-0 flex items-center justify-center rounded-ai bg-cohere-secondary-container text-cohere-on-secondary-container">
              <MaterialIcon name={agent.icon} size={20} />
            </div>
          </div>
          <div className="flex flex-col">
            <span className="font-label-mono-bold" style={{ color: agent.accentColor }}>
              {agent.displayName}
            </span>
            <span className="font-micro text-cohere-muted">{agent.description}</span>
          </div>
        </Link>
      ))}
    </div>
  );
}

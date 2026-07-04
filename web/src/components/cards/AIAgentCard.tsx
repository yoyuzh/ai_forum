import { Link } from "react-router-dom";
import { AIAgent } from "../../api/types";
import MaterialIcon from "../ui/MaterialIcon";

interface AIAgentCardProps {
  agent: AIAgent;
}

export default function AIAgentCard({ agent }: AIAgentCardProps) {
  const age = agent.ageViewpoint.match(/\d+岁/)?.[0] ?? `${Math.max(22, Math.min(45, Math.round(22 + agent.activityLevel * 18)))}岁`;
  const traits = [agent.traits[0], agent.traits[1]].filter(Boolean);

  return (
    <Link
      to={`/agents/${agent.id}/chat`}
      aria-label={`和 ${agent.displayName} 对话`}
      className="group grid min-h-[285px] grid-cols-[minmax(0,1fr)_112px] overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-low transition-colors hover:border-cohere-slate focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue sm:grid-cols-[minmax(0,1fr)_168px]"
    >
      <div className="flex min-w-0 flex-col gap-md p-lg pr-md">
        <div className="flex items-start justify-between gap-md">
          <div className="min-w-0">
            <h2 className="truncate font-headline-lg-bold" style={{ color: agent.accentColor }}>
              {agent.displayName}
            </h2>
            <p className="mt-xxs font-body-main text-cohere-muted">{viewpoint(agent)}</p>
          </div>

        </div>

        <p className="line-clamp-3 font-body-main text-cohere-on-surface-variant">{agent.description}</p>

        <div className="flex flex-wrap gap-xs">
          <Tag accentColor={agent.accentColor}>{age}设定</Tag>
          {traits.map((trait) => (
            <Tag key={trait} accentColor={agent.accentColor}>{trait}</Tag>
          ))}
        </div>

        <div className="mt-auto flex flex-wrap items-center gap-sm border-t border-cohere-hairline pt-md">
          {agent.active ? (
            agent.allowAutoReply ? (
              <StatusPill tone="green" iconDot>
                自动回复
              </StatusPill>
            ) : (
              <StatusPill tone="blue" iconDot>
                运行中
              </StatusPill>
            )
          ) : (
            <StatusPill>已停用</StatusPill>
          )}

          {agent.allowMentionReply && <StatusPill icon="alternate_email">支持</StatusPill>}
          {agent.allowFollowupReply && <StatusPill icon="psychology">跟进</StatusPill>}
        </div>
      </div>

      <div className="relative min-h-full overflow-hidden bg-cohere-surface-low">
        <img
          src={agent.avatar}
          alt={agent.displayName}
          width={180}
          height={260}
          onError={(e) => {
            e.currentTarget.style.display = "none";
            const fallback = e.currentTarget.parentElement?.querySelector(".avatar-fallback");
            if (fallback) fallback.classList.remove("hidden");
          }}
          className="relative z-10 h-full min-h-[285px] w-full border-l-2 object-cover object-center transition-transform duration-500 ease-cohere [mask-image:linear-gradient(to_right,transparent_0%,black_32%,black_100%)] [-webkit-mask-image:linear-gradient(to_right,transparent_0%,black_32%,black_100%)] group-hover:scale-[1.03]"
          style={{ borderColor: agent.accentColor }}
        />
        <div className="pointer-events-none absolute inset-y-0 left-0 z-20 w-1/2 bg-[linear-gradient(to_right,var(--c-surface-low)_0%,rgba(245,243,238,0.86)_28%,rgba(245,243,238,0)_100%)]" />
        <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 h-20 bg-[linear-gradient(to_top,var(--c-surface-low)_0%,rgba(245,243,238,0)_100%)]" />
        <div className="avatar-fallback absolute inset-0 z-10 hidden items-center justify-center text-cohere-deep-green">
          <MaterialIcon name={agent.icon || "smart_toy"} size={56} />
        </div>
      </div>
    </Link>
  );
}

function viewpoint(agent: AIAgent): string {
  const role = agent.ageViewpoint.split("·")[1]?.trim();
  if (role) return role;
  const primary = agent.specialties[0] ?? agent.traits[0] ?? "AI";
  if (/产品|用户|增长|体验/.test(primary)) return "互联网观察家";
  if (/心理|情感|共情|沟通/.test(primary)) return "心理咨询师视角";
  if (/代码|安全|性能|架构|系统/.test(primary)) return "系统架构师视角";
  return `${primary}视角`;
}

function Tag({ children, accentColor }: { children: React.ReactNode; accentColor?: string }) {
  return (
    <span
      className="rounded-sm px-xs py-xxs font-label-mono text-cohere-on-surface"
      style={accentColor ? { backgroundColor: `${accentColor}1A` } : undefined}
    >
      {children}
    </span>
  );
}

function StatusPill({
  children,
  icon,
  iconDot = false,
  tone = "stone",
}: {
  children: React.ReactNode;
  icon?: string;
  iconDot?: boolean;
  tone?: "green" | "blue" | "stone";
}) {
  const styles = {
    green: "border-cohere-deep-green/20 bg-cohere-success text-cohere-deep-green",
    blue: "border-cohere-action-blue bg-cohere-action-blue text-white",
    stone: "border-cohere-hairline bg-cohere-surface-lowest text-cohere-muted",
  }[tone];

  return (
    <span className={`inline-flex items-center gap-xs rounded-pill border px-sm py-xxs font-label-mono ${styles}`}>
      {iconDot && <span className={`h-2 w-2 rounded-full ${tone === "blue" ? "bg-white" : "bg-current"}`} />}
      {icon && <MaterialIcon name={icon} size={14} />}
      {children}
    </span>
  );
}

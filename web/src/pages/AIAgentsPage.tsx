import { useMemo, useState } from "react";
import { useAgents } from "../hooks/useAgents";
import type { AIAgent } from "../api/types";
import AIAgentCard from "../components/cards/AIAgentCard";
import MaterialIcon from "../components/ui/MaterialIcon";

type DomainFilter = "all" | "tech" | "life" | "emotion";
type PersonalityFilter = "all" | "calm" | "direct" | "warm";

const DOMAINS: Array<{ key: DomainFilter; label: string }> = [
  { key: "all", label: "所有领域" },
  { key: "tech", label: "技术" },
  { key: "life", label: "生活" },
  { key: "emotion", label: "情感" },
];

const PERSONALITIES: Array<{ key: PersonalityFilter; label: string }> = [
  { key: "all", label: "性格筛选" },
  { key: "calm", label: "冷静理性" },
  { key: "direct", label: "直接犀利" },
  { key: "warm", label: "温柔支持" },
];

export default function AIAgentsPage() {
  const { agents, isLoading } = useAgents();
  const [domain, setDomain] = useState<DomainFilter>("all");
  const [personality, setPersonality] = useState<PersonalityFilter>("all");
  const [mentionOnly, setMentionOnly] = useState(false);

  const filtered = useMemo(
    () =>
      agents.filter((agent) => {
        if (mentionOnly && !agent.allowMentionReply) return false;
        if (domain !== "all" && classifyDomain(agent) !== domain) return false;
        if (personality !== "all" && classifyPersonality(agent) !== personality) return false;
        return true;
      }),
    [agents, domain, mentionOnly, personality],
  );

  return (
    <main className="mx-auto flex w-full max-w-7xl flex-grow flex-col gap-section px-margin-mobile py-section md:px-margin-desktop">
      <section className="mx-auto flex max-w-3xl flex-col gap-md text-center">
        <h1 className="font-headline-xl text-cohere-primary">AI 角色广场</h1>
        <p className="font-body-large text-cohere-muted">
          不同年龄视角、性格和价值倾向的 AI 会参与不同类型的讨论
        </p>
      </section>

      <section className="flex flex-col items-center justify-between gap-md rounded-ai border border-cohere-hairline bg-cohere-surface-lowest p-md md:flex-row">
        <div className="flex w-full flex-wrap gap-sm md:w-auto">
          {DOMAINS.map((item) => (
            <button
              key={item.key}
              type="button"
              onClick={() => setDomain(item.key)}
              aria-pressed={domain === item.key}
              className={`rounded-pill border px-md py-sm font-label-mono transition-colors ${
                domain === item.key
                  ? "border-cohere-hairline bg-cohere-surface text-cohere-primary"
                  : "border-cohere-hairline text-cohere-muted hover:bg-cohere-surface-variant"
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>

        <div className="flex w-full flex-wrap gap-sm md:w-auto md:justify-end">
          <label className="relative">
            <select
              value={personality}
              onChange={(e) => setPersonality(e.target.value as PersonalityFilter)}
              className="appearance-none rounded-pill border border-cohere-hairline bg-cohere-surface-lowest py-sm pl-md pr-xl font-label-mono text-cohere-muted transition-colors hover:bg-cohere-surface-variant focus:border-cohere-secondary focus:outline-none"
              aria-label="性格筛选"
            >
              {PERSONALITIES.map((item) => (
                <option key={item.key} value={item.key}>
                  {item.label}
                </option>
              ))}
            </select>
            <MaterialIcon
              name="expand_more"
              size={18}
              className="pointer-events-none absolute right-md top-1/2 -translate-y-1/2 text-cohere-muted"
            />
          </label>

          <button
            type="button"
            onClick={() => setMentionOnly((value) => !value)}
            aria-pressed={mentionOnly}
            className={`flex items-center gap-xs rounded-pill border px-md py-sm font-label-mono transition-colors ${
              mentionOnly
                ? "border-cohere-secondary bg-cohere-secondary-container text-cohere-on-secondary-container"
                : "border-cohere-hairline text-cohere-muted hover:bg-cohere-surface-variant"
            }`}
          >
            支持 @AI
            <MaterialIcon name="filter_list" size={18} />
          </button>
        </div>
      </section>

      <section>
        {isLoading ? (
          <div className="grid grid-cols-1 gap-lg lg:grid-cols-2">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className="h-72 animate-pulse-soft rounded-lg border border-cohere-hairline bg-cohere-surface-low"
              />
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <div className="card-base flex flex-col items-center gap-md p-xl text-center">
            <MaterialIcon name="smart_toy" size={48} className="text-cohere-muted" />
            <h2 className="font-feature-title text-cohere-primary">没有匹配的 AI 角色</h2>
            <p className="font-body-main text-cohere-on-surface-variant">调整筛选条件后再试。</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-lg lg:grid-cols-2">
            {filtered.map((agent) => (
              <AIAgentCard key={agent.id} agent={agent} />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

function classifyDomain(agent: AIAgent): DomainFilter {
  const text = [agent.description, ...agent.traits, ...agent.specialties].join(" ");
  if (/心理|情感|共情|温柔|陪伴|人文|倾听/.test(text)) return "emotion";
  if (/生活|用户|产品|体验|增长|沟通/.test(text)) return "life";
  return "tech";
}

function classifyPersonality(agent: AIAgent): PersonalityFilter {
  const text = [agent.displayName, agent.description, ...agent.traits].join(" ");
  if (/批判|辛辣|犀利|直接|怀疑|毒舌/.test(text)) return "direct";
  if (/温柔|鼓励|共情|支持|倾听|包容/.test(text)) return "warm";
  return "calm";
}

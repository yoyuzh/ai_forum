import { useAgents } from "../hooks/useAgents";
import AIAgentCard from "../components/cards/AIAgentCard";
import MaterialIcon from "../components/ui/MaterialIcon";

export default function AIAgentsPage() {
  const { agents, isLoading } = useAgents();

  return (
    <main className="mx-auto w-full max-w-7xl flex-grow px-margin-mobile py-xl md:px-margin-desktop animate-reveal-up">
      <section className="flex flex-col items-start gap-md pt-lg">
        <span className="flex items-center gap-1 font-label-mono text-cohere-secondary">
          <span className="status-dot bg-cohere-secondary" />
          AI Agent Roster
        </span>
        <h1 className="font-headline-xl text-cohere-primary">AI 角色</h1>
        <p className="max-w-2xl font-body-large text-cohere-on-surface-variant">
          论坛中的每一个 AI 代理都拥有独立的人格、专业领域与回答意愿阈值。
          它们会根据帖子内容自主评估是否参与讨论，每一次决策都记录在可解释的决策日志中。
        </p>
      </section>

      <section className="mt-xl">
        {isLoading ? (
          <div className="grid grid-cols-1 gap-lg md:grid-cols-2 xl:grid-cols-3">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className="h-48 animate-pulse-soft rounded-lg border border-cohere-hairline bg-cohere-surface-low"
              />
            ))}
          </div>
        ) : agents.length === 0 ? (
          <div className="card-base flex flex-col items-center gap-md p-xl text-center">
            <MaterialIcon name="smart_toy" size={48} className="text-cohere-muted" />
            <h3 className="font-feature-title text-cohere-primary">暂无可用 AI 角色</h3>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-lg md:grid-cols-2 xl:grid-cols-3">
            {agents.map((agent) => (
              <AIAgentCard key={agent.id} agent={agent} />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

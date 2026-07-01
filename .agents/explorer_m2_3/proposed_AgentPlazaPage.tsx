import React from "react";
import { useAgents } from "../hooks/useAgents";
import { AIAgent } from "../api/types";

export function AgentPlazaPage() {
  const { agents, isLoading, updateAgent } = useAgents();

  const handleToggleActive = async (id: number, currentActive: boolean) => {
    try {
      await updateAgent({ id, updates: { active: !currentActive } });
    } catch (err) {
      console.error("Failed to update active state:", err);
    }
  };

  const handleUpdateThreshold = async (id: number, field: keyof AIAgent, val: number) => {
    try {
      await updateAgent({ id, updates: { [field]: val } });
    } catch (err) {
      console.error(`Failed to update ${field}:`, err);
    }
  };

  const handleToggleRule = async (id: number, field: keyof AIAgent, currentVal: boolean) => {
    try {
      await updateAgent({ id, updates: { [field]: !currentVal } });
    } catch (err) {
      console.error(`Failed to update ${field}:`, err);
    }
  };

  const handleUpdateRepliesLimit = async (id: number, field: keyof AIAgent, val: number) => {
    try {
      await updateAgent({ id, updates: { [field]: val } });
    } catch (err) {
      console.error(`Failed to update ${field}:`, err);
    }
  };

  return (
    <main className="flex-grow max-w-7xl mx-auto w-full px-margin-mobile md:px-margin-desktop py-section flex flex-col gap-section">
      {/* Header */}
      <header className="flex flex-col gap-md text-center max-w-3xl mx-auto">
        <h1 className="font-headline-xl text-headline-xl text-primary">AI 角色广场</h1>
        <p className="font-body-large text-body-large text-surface-tint">
          不同领域视角、性格和价值倾向的 AI 会参与不同类型的讨论。您可以自定义他们的行为规则和阈值。
        </p>
      </header>

      {/* Grid of Agent Console Cards */}
      {isLoading ? (
        <div className="py-xl text-center text-muted font-body-main">加载角色配置中...</div>
      ) : (
        <section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-lg">
          {agents.map((agent) => (
            <article
              key={agent.id}
              className={`bg-surface-container-low rounded-[16px] p-lg border transition-all flex flex-col gap-md hover:border-outline-variant ${
                agent.active ? "border-hairline" : "border-hairline opacity-60"
              }`}
            >
              {/* Profile Row */}
              <div className="flex items-start justify-between">
                <div className="flex gap-md items-center">
                  <img
                    alt={agent.name}
                    className="w-[44px] h-[44px] rounded-full object-cover border border-hairline"
                    src={agent.avatar}
                  />
                  <div>
                    <h3 className="font-feature-title text-feature-title text-primary flex items-center gap-xs">
                      {agent.name}
                      {agent.active && <span className="material-symbols-outlined text-coral text-sm">verified</span>}
                    </h3>
                    <p className="font-caption text-caption text-surface-tint">{agent.description}</p>
                  </div>
                </div>
                
                {/* Active Toggle Switch */}
                <button
                  onClick={() => handleToggleActive(agent.id, agent.active)}
                  className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                    agent.active ? "bg-secondary" : "bg-hairline"
                  }`}
                >
                  <span
                    className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                      agent.active ? "translate-x-5" : "translate-x-0"
                    }`}
                  />
                </button>
              </div>

              {/* Personality Tags */}
              <div className="flex flex-wrap gap-xs">
                <span className="px-xs py-[2px] bg-secondary-container text-on-secondary-container rounded-sm font-label-mono text-label-mono">
                  {agent.ageViewpoint}
                </span>
                <span className="px-xs py-[2px] bg-surface-container-highest text-on-surface rounded-sm font-label-mono text-label-mono">
                  {agent.personality}
                </span>
                <span className="px-xs py-[2px] bg-surface-container-highest text-on-surface rounded-sm font-label-mono text-label-mono">
                  {agent.valueOrientation}
                </span>
              </div>

              {/* Collapsible/Detail System configuration */}
              <div className="space-y-sm text-xs border-t border-hairline pt-md">
                <h4 className="font-label-mono-bold text-primary">AI 决策配置</h4>
                
                {/* Threshold slider */}
                <div className="space-y-xs">
                  <div className="flex justify-between font-label-mono text-muted">
                    <span>回答触发阈值 (Willingness Threshold)</span>
                    <span className="text-primary font-semibold">{agent.replyThreshold.toFixed(2)}</span>
                  </div>
                  <input
                    type="range"
                    min="0.0"
                    max="1.0"
                    step="0.05"
                    value={agent.replyThreshold}
                    onChange={(e) => handleUpdateThreshold(agent.id, "replyThreshold", Number(e.target.value))}
                    disabled={!agent.active}
                    className="w-full h-1 bg-surface-variant rounded-lg appearance-none cursor-pointer disabled:opacity-50 accent-secondary"
                  />
                </div>

                {/* Activity level slider */}
                <div className="space-y-xs">
                  <div className="flex justify-between font-label-mono text-muted">
                    <span>活跃度 (Activity Level)</span>
                    <span className="text-primary font-semibold">{agent.activityLevel.toFixed(2)}</span>
                  </div>
                  <input
                    type="range"
                    min="0.0"
                    max="1.0"
                    step="0.05"
                    value={agent.activityLevel}
                    onChange={(e) => handleUpdateThreshold(agent.id, "activityLevel", Number(e.target.value))}
                    disabled={!agent.active}
                    className="w-full h-1 bg-surface-variant rounded-lg appearance-none cursor-pointer disabled:opacity-50 accent-secondary"
                  />
                </div>

                {/* Rule Toggles */}
                <div className="grid grid-cols-2 gap-sm pt-xs text-[11px] font-label-mono">
                  <label className="flex items-center gap-xs cursor-pointer">
                    <input
                      type="checkbox"
                      checked={agent.allowAutoReply}
                      onChange={() => handleToggleRule(agent.id, "allowAutoReply", agent.allowAutoReply)}
                      disabled={!agent.active}
                      className="rounded border-hairline text-secondary focus:ring-secondary w-3.5 h-3.5"
                    />
                    <span>允许发帖自动回复</span>
                  </label>
                  <label className="flex items-center gap-xs cursor-pointer">
                    <input
                      type="checkbox"
                      checked={agent.allowMentionReply}
                      onChange={() => handleToggleRule(agent.id, "allowMentionReply", agent.allowMentionReply)}
                      disabled={!agent.active}
                      className="rounded border-hairline text-secondary focus:ring-secondary w-3.5 h-3.5"
                    />
                    <span>支持 @ 唤醒回复</span>
                  </label>
                  <label className="flex items-center gap-xs cursor-pointer">
                    <input
                      type="checkbox"
                      checked={agent.allowFollowupReply}
                      onChange={() => handleToggleRule(agent.id, "allowFollowupReply", agent.allowFollowupReply)}
                      disabled={!agent.active}
                      className="rounded border-hairline text-secondary focus:ring-secondary w-3.5 h-3.5"
                    />
                    <span>允许跟进追问回复</span>
                  </label>
                </div>

                {/* Limits Config */}
                <div className="flex gap-md pt-xs text-[11px] font-label-mono">
                  <div className="flex items-center gap-xs">
                    <span>自动回复上限:</span>
                    <input
                      type="number"
                      min="0"
                      max="5"
                      value={agent.maxAutoRepliesPerPost}
                      onChange={(e) => handleUpdateRepliesLimit(agent.id, "maxAutoRepliesPerPost", Number(e.target.value))}
                      disabled={!agent.active}
                      className="w-10 border border-hairline rounded px-xs py-[2px] text-center"
                    />
                  </div>
                  <div className="flex items-center gap-xs">
                    <span>追问上限:</span>
                    <input
                      type="number"
                      min="0"
                      max="5"
                      value={agent.maxFollowupRepliesPerPost}
                      onChange={(e) => handleUpdateRepliesLimit(agent.id, "maxFollowupRepliesPerPost", Number(e.target.value))}
                      disabled={!agent.active}
                      className="w-10 border border-hairline rounded px-xs py-[2px] text-center"
                    />
                  </div>
                </div>
              </div>

              {/* Prompt System details */}
              <div className="flex flex-col gap-xs text-[11px] font-label-mono pt-xs border-t border-dotted border-hairline mt-auto">
                <span className="text-muted">核心提示词 (System Prompt):</span>
                <p className="bg-surface-container-lowest p-xs rounded border border-hairline text-ink italic max-h-[60px] overflow-y-auto">
                  {agent.systemPrompt}
                </p>
                <span className="text-muted mt-xs">风格提示词 (Speaking Style):</span>
                <p className="bg-surface-container-lowest p-xs rounded border border-hairline text-ink italic max-h-[60px] overflow-y-auto">
                  {agent.speakingStyle}
                </p>
              </div>

              {/* Status footer inside card */}
              <div className="pt-sm border-t border-hairline flex gap-sm items-center mt-sm text-micro text-muted">
                {agent.active ? (
                  <span className="bg-[#edfce9] text-[#003c33] px-sm py-[2px] rounded-full border border-[#003c33]/20 flex items-center gap-xs">
                    <span className="w-1.5 h-1.5 rounded-full bg-[#003c33]"></span> 自动运行中
                  </span>
                ) : (
                  <span className="bg-surface-container text-muted px-sm py-[2px] rounded-full border flex items-center gap-xs">
                    <span className="w-1.5 h-1.5 rounded-full bg-outline"></span> 已禁用
                  </span>
                )}
              </div>
            </article>
          ))}
        </section>
      )}
    </main>
  );
}

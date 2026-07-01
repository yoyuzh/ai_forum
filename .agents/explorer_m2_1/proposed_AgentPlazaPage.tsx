import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAgents } from "../hooks/useAgents";
import { AIAgent } from "../api/types";
import { ArrowLeft, Sliders, Settings2, ShieldCheck, Check, Save, Loader2, Bot } from "lucide-react";

export default function AgentPlazaPage() {
  const { agents, isLoading, updateAgent } = useAgents();
  const navigate = useNavigate();
  
  // Track which agent card is currently expanded for configuration edits
  const [editingAgentId, setEditingAgentId] = useState<number | null>(null);
  
  // Keep local form state for the agent being edited
  const [formState, setFormState] = useState<Partial<AIAgent> | null>(null);
  const [savingId, setSavingId] = useState<number | null>(null);

  const startEditing = (agent: AIAgent) => {
    setEditingAgentId(agent.id);
    setFormState({
      systemPrompt: agent.systemPrompt,
      stylePrompt: agent.stylePrompt,
      replyThreshold: agent.replyThreshold,
      maxAutoRepliesPerPost: agent.maxAutoRepliesPerPost,
      allowAutoReply: agent.allowAutoReply,
      allowMentionReply: agent.allowMentionReply,
      allowFollowupReply: agent.allowFollowupReply
    });
  };

  const handleToggleActive = async (agent: AIAgent) => {
    setSavingId(agent.id);
    try {
      await updateAgent({
        id: agent.id,
        updates: { active: !agent.active }
      });
    } catch (err) {
      console.error(err);
    } finally {
      setSavingId(null);
    }
  };

  const handleFieldChange = (key: keyof AIAgent, value: any) => {
    if (formState) {
      setFormState({ ...formState, [key]: value });
    }
  };

  const handleSaveSettings = async (id: number) => {
    if (!formState) return;
    setSavingId(id);
    try {
      await updateAgent({
        id,
        updates: formState
      });
      setEditingAgentId(null);
      setFormState(null);
    } catch (err) {
      console.error(err);
    } finally {
      setSavingId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-cohere-canvas flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-cohere-action-blue" />
        <span className="ml-2 font-sans text-sm text-cohere-slate">Loading agents roster...</span>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cohere-canvas flex flex-col text-cohere-ink font-sans">
      
      {/* Top Breadcrumb Bar */}
      <nav className="border-b border-cohere-hairline py-4 px-6 md:px-12 max-w-7xl mx-auto w-full flex items-center justify-between">
        <button onClick={() => navigate("/")} className="flex items-center gap-2 text-xs font-mono-label text-cohere-slate hover:text-cohere-primary transition-colors">
          <ArrowLeft className="w-4 h-4" />
          Back to Feed
        </button>
        <span className="font-mono-label text-xs text-cohere-muted flex items-center gap-1.5">
          <Bot className="w-4.5 h-4.5 text-cohere-deep-green" />
          AI Config Plaza
        </span>
      </nav>

      {/* Hero Header */}
      <header className="py-10 px-6 md:px-12 max-w-7xl mx-auto w-full border-b border-cohere-hairline">
        <h1 className="font-display text-4xl md:text-5xl font-normal leading-[1.1] -tracking-[1px] text-cohere-black mb-3">
          AI Agents Plaza
        </h1>
        <p className="font-sans text-sm md:text-base text-cohere-body-muted max-w-2xl">
          Customize agent personas, toggle response modes, and tweak willingness thresholds. These changes directly govern their real-time SSE replies.
        </p>
      </header>

      {/* Roster Grid */}
      <main className="max-w-7xl mx-auto w-full px-6 md:px-12 py-10">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {agents.map((agent) => {
            const isEditing = editingAgentId === agent.id;
            const isSaving = savingId === agent.id;

            return (
              <article 
                key={agent.id} 
                className={`bg-white border rounded-md p-6 flex flex-col justify-between transition-all duration-300 ${
                  agent.active 
                    ? "border-cohere-hairline hover:border-cohere-slate" 
                    : "border-cohere-hairline opacity-60 hover:opacity-80"
                }`}
              >
                {/* Agent Profile Header */}
                <div>
                  <div className="flex justify-between items-start mb-4">
                    <div className="flex items-center gap-3">
                      <img 
                        src={agent.avatar} 
                        alt={agent.name} 
                        className="w-12 h-12 rounded-lg border border-cohere-hairline object-cover"
                      />
                      <div>
                        <h3 className="font-sans text-base font-semibold text-cohere-primary">
                          {agent.name}
                        </h3>
                        <span className="font-mono-label text-[10px] text-cohere-muted uppercase">
                          ID: {agent.id}
                        </span>
                      </div>
                    </div>

                    {/* Active Toggle Switch */}
                    <button
                      onClick={() => handleToggleActive(agent)}
                      disabled={isSaving}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ${
                        agent.active ? "bg-cohere-deep-green" : "bg-cohere-hairline"
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          agent.active ? "translate-x-6" : "translate-x-1"
                        }`}
                      />
                    </button>
                  </div>

                  {/* Descriptions */}
                  <p className="font-sans text-xs text-cohere-body-muted mb-4 leading-relaxed">
                    {agent.description}
                  </p>

                  <div className="border-t border-cohere-hairline pt-3 mt-3 flex flex-col gap-2">
                    <div className="flex justify-between text-xs">
                      <span className="text-cohere-slate">Speaking Tone:</span>
                      <strong className="text-cohere-primary">{agent.personality}</strong>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-cohere-slate">Trigger Threshold:</span>
                      <strong className="text-cohere-primary">{(agent.replyThreshold * 100).toFixed(0)}% Willingness</strong>
                    </div>
                  </div>
                </div>

                {/* Inline Editing Drawer / Settings Panel */}
                {isEditing && formState ? (
                  <div className="mt-6 border-t border-cohere-hairline pt-4 flex flex-col gap-4">
                    <h4 className="font-mono-label text-xs font-semibold text-cohere-black flex items-center gap-1">
                      <Settings2 className="w-3.5 h-3.5" />
                      Agent Settings
                    </h4>
                    
                    {/* System Prompt */}
                    <div className="flex flex-col gap-1">
                      <label className="font-mono-label text-[10px] text-cohere-slate">System Instructions</label>
                      <textarea
                        value={formState.systemPrompt}
                        onChange={(e) => handleFieldChange("systemPrompt", e.target.value)}
                        rows={2}
                        className="border border-cohere-hairline rounded-sm p-2 text-xs font-sans focus:outline-none focus:border-cohere-form-focus bg-cohere-soft-stone/30"
                      />
                    </div>

                    {/* Style Prompt */}
                    <div className="flex flex-col gap-1">
                      <label className="font-mono-label text-[10px] text-cohere-slate">Speaking Style / Layout</label>
                      <textarea
                        value={formState.stylePrompt}
                        onChange={(e) => handleFieldChange("stylePrompt", e.target.value)}
                        rows={2}
                        className="border border-cohere-hairline rounded-sm p-2 text-xs font-sans focus:outline-none focus:border-cohere-form-focus bg-cohere-soft-stone/30"
                      />
                    </div>

                    {/* Reply Threshold */}
                    <div className="flex flex-col gap-1">
                      <label className="font-mono-label text-[10px] text-cohere-slate flex justify-between">
                        <span>Min Interest Threshold</span>
                        <span>{Math.round((formState.replyThreshold || 0) * 100)}%</span>
                      </label>
                      <input
                        type="range"
                        min="0"
                        max="100"
                        value={Math.round((formState.replyThreshold || 0) * 100)}
                        onChange={(e) => handleFieldChange("replyThreshold", Number(e.target.value) / 100)}
                        className="w-full accent-cohere-deep-green"
                      />
                    </div>

                    {/* Max Replies */}
                    <div className="flex justify-between items-center gap-2">
                      <label className="font-mono-label text-[10px] text-cohere-slate">Max replies/post</label>
                      <input
                        type="number"
                        min="1"
                        max="5"
                        value={formState.maxAutoRepliesPerPost || 1}
                        onChange={(e) => handleFieldChange("maxAutoRepliesPerPost", Number(e.target.value))}
                        className="w-16 border border-cohere-hairline rounded-sm px-2 py-1 text-xs focus:outline-none focus:border-cohere-form-focus"
                      />
                    </div>

                    {/* Flags */}
                    <div className="flex flex-col gap-2 pt-2">
                      <label className="flex items-center gap-2 text-xs text-cohere-primary cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formState.allowAutoReply}
                          onChange={(e) => handleFieldChange("allowAutoReply", e.target.checked)}
                          className="rounded text-cohere-deep-green focus:ring-cohere-deep-green border-cohere-hairline"
                        />
                        <span>Auto-reply on new posts</span>
                      </label>

                      <label className="flex items-center gap-2 text-xs text-cohere-primary cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formState.allowMentionReply}
                          onChange={(e) => handleFieldChange("allowMentionReply", e.target.checked)}
                          className="rounded text-cohere-deep-green focus:ring-cohere-deep-green border-cohere-hairline"
                        />
                        <span>Respond to @Mentions</span>
                      </label>
                    </div>

                    <div className="flex gap-2 justify-end mt-2">
                      <button
                        onClick={() => { setEditingAgentId(null); setFormState(null); }}
                        className="px-3 py-1.5 border border-cohere-hairline rounded-sm text-xs font-button text-cohere-slate hover:bg-cohere-soft-stone"
                      >
                        Cancel
                      </button>
                      <button
                        onClick={() => handleSaveSettings(agent.id)}
                        disabled={isSaving}
                        className="px-3 py-1.5 bg-cohere-primary text-white rounded-sm text-xs font-button hover:bg-cohere-black flex items-center gap-1 disabled:opacity-40"
                      >
                        {isSaving ? <Loader2 className="w-3 animate-spin" /> : <Save className="w-3.5 h-3.5" />}
                        Save
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="mt-6 flex justify-end">
                    <button
                      onClick={() => startEditing(agent)}
                      className="font-button text-xs text-cohere-action-blue hover:underline flex items-center gap-1"
                    >
                      <Sliders className="w-3.5 h-3.5" />
                      Configure Persona
                    </button>
                  </div>
                )}
              </article>
            );
          })}
        </div>
      </main>
    </div>
  );
}

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Input, Select, Button, Table } from "antd";
import type { ColumnsType } from "antd/es/table";
import { adminApi } from "../api/client";
import { AdminDecisionLog } from "../api/types";
import MaterialIcon from "../components/MaterialIcon";
import DecisionDetailDrawer from "../components/decision/DecisionDetailDrawer";
import AgentDecisionBreakdown from "../components/decision/AgentDecisionBreakdown";

export default function AIDecisionLogsPage() {
  const { data: logs = [] } = useQuery({
    queryKey: ["decisionLogs"],
    queryFn: adminApi.decisionLogs.list,
  });
  const { data: context } = useQuery({
    queryKey: ["decisionContext"],
    queryFn: adminApi.decisionContext,
  });

  const [postId, setPostId] = useState("");
  const [agent, setAgent] = useState("ALL");
  const [decision, setDecision] = useState("ALL");
  const [detail, setDetail] = useState<AdminDecisionLog | null>(null);
  const filteredLogs = logs.filter((log) => {
    const normalizedPostId = postId.trim().toLowerCase();
    if (normalizedPostId && !String(log.postId).toLowerCase().includes(normalizedPostId)) return false;
    if (agent !== "ALL" && log.aiAgentName !== agent) return false;
    if (decision !== "ALL" && log.decision !== decision) return false;
    return true;
  });

  const columns: ColumnsType<AdminDecisionLog> = [
    {
      title: "AI Name",
      dataIndex: "aiAgentName",
      width: 180,
      render: (v: string) => <span className="pl-md font-label-mono font-bold text-cohere-primary">{v}</span>,
    },
    {
      title: "Traits",
      dataIndex: "traits",
      width: 160,
      render: (traits: string[]) => (
        <span className="font-caption text-cohere-on-surface-variant">{(traits ?? []).join(", ")}</span>
      ),
    },
    {
      title: "Hit Tags",
      dataIndex: "hitTags",
      width: 200,
      render: (tags: string[]) =>
        (tags ?? []).length > 0 ? (
          <div className="flex flex-wrap gap-1">
            {tags.map((t) => (
              <span key={t} className="mb-1 rounded bg-cohere-surface-variant px-1 py-0.5 font-label-mono text-[11px]">
                {t}
              </span>
            ))}
          </div>
        ) : (
          <span className="font-label-mono text-cohere-muted">None Strong</span>
        ),
    },
    {
      title: "Score",
      dataIndex: "willingnessScore",
      width: 80,
      align: "right",
      render: (v: number, record: AdminDecisionLog) => (
        <span className={`font-label-mono font-bold ${v >= record.thresholdValue ? "text-cohere-action-blue" : "text-cohere-coral"}`}>
          {(v <= 1 ? v : v / 100).toFixed(2)}
        </span>
      ),
    },
    {
      title: "Thresh.",
      dataIndex: "thresholdValue",
      width: 80,
      align: "right",
      render: (v: number) => <span className="font-label-mono text-cohere-muted">{(v <= 1 ? v : v / 100).toFixed(2)}</span>,
    },
    {
      title: "Decision",
      dataIndex: "decision",
      width: 120,
      render: (_: string, record: AdminDecisionLog) =>
        record.decision === "REPLY" ? (
          <span className="inline-flex items-center gap-1 rounded-full border border-cohere-secondary/20 bg-cohere-success px-1 py-0.5 font-label-mono-bold text-[11px] text-cohere-secondary">
            <span className="status-dot bg-cohere-secondary" /> 参与回复
          </span>
        ) : record.decision === "FALLBACK" ? (
          <span className="inline-flex items-center gap-1 rounded-full border border-cohere-coral bg-cohere-coral-soft px-1 py-0.5 font-label-mono-bold text-[11px] text-cohere-primary">
            <span className="status-dot bg-cohere-coral" /> fallback
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 rounded-full border border-cohere-hairline bg-cohere-surface-variant px-1 py-0.5 font-label-mono-bold text-[11px] text-cohere-muted">
            <span className="status-dot bg-cohere-muted" /> 跳过
          </span>
        ),
    },
    {
      title: "Detail",
      key: "detail",
      width: 96,
      render: (_: unknown, record: AdminDecisionLog) => (
        <Button size="small" aria-label={`decision detail ${record.id}`} onClick={() => setDetail(record)}>详情</Button>
      ),
    },
    {
      title: "Reasoning Log",
      dataIndex: "reason",
      render: (reason: string) => (
        <span className="border-l border-dotted border-cohere-hairline pl-md font-caption text-cohere-on-surface-variant">
          {reason}
        </span>
      ),
    },
  ];

  return (
    <div className="mx-auto max-w-[1440px] overflow-x-hidden px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-lg">
        <h1 className="font-headline-xl font-bold text-cohere-primary">AI 决策日志</h1>
        <p className="mt-1 font-body-main text-cohere-muted">
          监控和分析 AI 代理在论坛帖子中的响应决策意图与阈值表现。
        </p>
      </div>

      <AgentDecisionBreakdown logs={filteredLogs} />

      {/* Filter bar */}
      <div className="mb-lg flex flex-col gap-sm rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-sm md:flex-row md:items-end">
        <div className="flex w-full flex-1 flex-col gap-1">
          <label htmlFor="decision-post-id" className="font-label-mono uppercase tracking-wider text-cohere-muted">Post ID</label>
          <Input
            id="decision-post-id"
            name="postId"
            value={postId}
            onChange={(e) => setPostId(e.target.value)}
            className="font-label-mono"
            allowClear
          />
        </div>
        <div className="flex w-full flex-1 flex-col gap-1">
          <label htmlFor="decision-agent" className="font-label-mono uppercase tracking-wider text-cohere-muted">AI Agent</label>
          <Select
            id="decision-agent"
            value={agent}
            onChange={setAgent}
            className="w-full"
            options={[
              { value: "ALL", label: "All Agents" },
              { value: "Tech_Guru_v4", label: "Tech_Guru_v4" },
              { value: "DebateBot_Core", label: "DebateBot_Core" },
            ]}
          />
        </div>
        <div className="flex w-full flex-1 flex-col gap-1">
          <label htmlFor="decision-outcome" className="font-label-mono uppercase tracking-wider text-cohere-muted">Decision</label>
          <Select
            id="decision-outcome"
            value={decision}
            onChange={setDecision}
            className="w-full"
            options={[
              { value: "ALL", label: "All Outcomes" },
              { value: "REPLY", label: "参与回复" },
              { value: "IGNORE", label: "跳过" },
            ]}
          />
        </div>
        <Button type="primary" className="h-9 whitespace-nowrap">Apply Filters</Button>
      </div>

      {/* Bento: post context + score visualization */}
      <div className="mb-section grid grid-cols-1 gap-gutter xl:grid-cols-12">
        {/* Post context card */}
        <div className="relative flex h-full flex-col overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md xl:col-span-4">
          <div className="absolute left-0 top-0 h-full w-1 bg-cohere-coral" />
          <div className="mb-sm flex items-start justify-between pl-1">
            <span className="rounded border border-cohere-coral px-1 py-0.5 font-label-mono-bold uppercase text-[10px] text-cohere-coral">
              Context Trigger
            </span>
            {context && (
              <span className="font-label-mono text-cohere-muted">{context.timestamp}</span>
            )}
          </div>
          {context && (
            <>
              <h2 className="mb-1 pl-1 font-feature-title font-bold leading-tight text-cohere-primary">
                {context.title}
              </h2>
              <p className="mb-md line-clamp-4 flex-grow pl-1 font-body-main text-cohere-on-surface-variant">
                {context.body}
              </p>
              <div className="mt-auto flex flex-wrap gap-2 pl-1">
                {context.tags.map((t) => (
                  <span key={t} className="rounded bg-cohere-surface-container px-1 py-0.5 font-label-mono text-cohere-muted">
                    {t}
                  </span>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Score vs threshold visualization */}
        <div className="relative flex h-full flex-col rounded-lg border border-[#1b1b20] bg-[#1b1b20] p-md text-white xl:col-span-8">
          <div className="mb-md flex items-center justify-between border-b border-white/10 pb-sm">
            <div className="flex items-center gap-1">
              <h2 className="font-feature-title font-bold text-white">AI 回答意愿分</h2>
              <span className="ml-1 font-caption font-normal text-white/50">Score vs Threshold</span>
            </div>
            <MaterialIcon name="bar_chart" className="text-white/50" />
          </div>

          <div className="flex flex-col gap-md">
            {filteredLogs.map((log) => {
              const passed = log.willingnessScore >= log.thresholdValue;
              return (
                <div key={log.id}>
                  <div className="mb-1 flex justify-between font-label-mono">
                    <span className="text-white">{log.aiAgentName}</span>
                    <span className={passed ? "text-[#9dd1c4]" : "text-white/70"}>
                      {log.willingnessScore} / {log.thresholdValue}
                    </span>
                  </div>
                  <div className="relative h-2 w-full overflow-hidden rounded-full bg-white/10">
                    <div
                      className={`h-full rounded-full ${passed ? "bg-[#9dd1c4]" : "bg-white/30"}`}
                      style={{ width: `${(log.willingnessScore <= 1 ? log.willingnessScore * 100 : log.willingnessScore)}%` }}
                    />
                    {/* Coral threshold marker */}
                    <div
                      className="absolute top-0 bottom-0 z-10 w-0.5 bg-cohere-coral"
                      style={{ left: `${(log.thresholdValue <= 1 ? log.thresholdValue * 100 : log.thresholdValue)}%` }}
                    />
                  </div>
                </div>
              );
            })}
            {filteredLogs.length === 0 && (
              <div className="rounded border border-white/10 p-md font-label-mono text-white/60">
                No matching decision logs.
              </div>
            )}
          </div>

          <div className="mt-md flex gap-md border-t border-white/10 pt-sm font-label-mono text-[11px] uppercase tracking-wide">
            <span className="flex items-center gap-1">
              <span className="h-3 w-3 rounded-sm bg-[#9dd1c4]" /> 参与回复 (Score &gt; Threshold)
            </span>
            <span className="flex items-center gap-1">
              <span className="h-3 w-3 rounded-sm bg-white/30" /> 跳过 (Score &lt; Threshold)
            </span>
            <span className="flex items-center gap-1">
              <span className="h-3 w-0.5 bg-cohere-coral" /> 决策阈值
            </span>
          </div>
        </div>
      </div>

      {/* Decision matrix table */}
      <div className="mb-section overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest">
        <div className="flex items-center justify-between border-b border-cohere-hairline bg-cohere-surface-low p-md">
          <h2 className="text-[20px] font-bold text-cohere-primary">详细决策矩阵</h2>
          <button type="button" className="flex items-center gap-1 font-label-mono text-cohere-primary transition-colors hover:text-cohere-coral focus:outline-none focus-visible:underline">
            <MaterialIcon name="download" size={16} /> Export CSV
          </button>
        </div>
        <Table<AdminDecisionLog>
          columns={columns}
          dataSource={filteredLogs}
          rowKey="id"
          pagination={false}
          scroll={{ x: 900 }}
          size="middle"
        />
      </div>
      <DecisionDetailDrawer log={detail} onClose={() => setDetail(null)} />
    </div>
  );
}

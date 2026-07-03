import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Table, Select, Input, Button } from "antd";
import type { ColumnsType } from "antd/es/table";
import { adminApi } from "../api/client";
import { AdminAITask, TaskStatus } from "../api/types";
import MaterialIcon from "../components/MaterialIcon";
import StatusBadge from "../components/StatusBadge";
import TaskDetailDrawer from "../components/TaskDetailDrawer";
import { usePermission } from "../hooks/usePermission";

export default function AITasksPage() {
  const { data: tasks = [], isLoading } = useQuery({
    queryKey: ["tasks"],
    queryFn: adminApi.tasks.list,
  });
  const { data: summary } = useQuery({
    queryKey: ["taskSummary"],
    queryFn: adminApi.taskSummary,
  });

  const [statusFilter, setStatusFilter] = useState<TaskStatus | "ALL">("ALL");
  const [agentFilter, setAgentFilter] = useState<string>("ALL");
  const [triggerFilter, setTriggerFilter] = useState<string>("ALL");
  const [search, setSearch] = useState("");
  const [detailId, setDetailId] = useState<string | null>(null);
  const canRetry = usePermission("ai_task:retry");

  const filtered = tasks.filter((t) => {
    const agentName = t.agentName ?? t.aiAgentName ?? "";
    const targetPostId = t.targetPostId ?? String(t.postId ?? "");
    if (statusFilter !== "ALL" && t.status !== statusFilter) return false;
    if (agentFilter !== "ALL" && agentName !== agentFilter) return false;
    if (triggerFilter !== "ALL" && t.triggerType !== triggerFilter) return false;
    if (search && !String(t.id).toLowerCase().includes(search.toLowerCase()) && !targetPostId.toLowerCase().includes(search.toLowerCase()))
      return false;
    return true;
  });

  const summaryCards = summary
    ? ([
        { label: "Pending", value: summary.pending, icon: "hourglass_empty", variant: "default" as const, status: "PENDING" as const },
        { label: "Running", value: summary.running, icon: "autorenew", variant: "secondary" as const, status: "PROCESSING" as const },
        { label: "Success", value: summary.success, icon: "check_circle", variant: "success" as const, status: "COMPLETED" as const },
        { label: "Failed", value: summary.failed, icon: "error", variant: "error" as const, status: "FAILED" as const },
        { label: "Skipped", value: summary.skipped, icon: "skip_next", variant: "default" as const, status: null },
      ])
    : [];

  const columns: ColumnsType<AdminAITask> = [
    {
      title: "Task ID",
      dataIndex: "id",
      width: 140,
      render: (id: string, record: AdminAITask) => (
        <div>
          <span className="rounded border border-cohere-hairline bg-cohere-surface-container px-1 py-0.5 font-label-mono text-cohere-ink">
            {id}
          </span>
          <div className="font-micro text-cohere-muted">{record.createdAt}</div>
        </div>
      ),
    },
    {
      title: "Agent / Trigger",
      dataIndex: "agentName",
      width: 220,
      render: (_: string, record: AdminAITask) => (
        <div className="flex items-center gap-1">
          <div className="flex h-6 w-6 items-center justify-center rounded bg-cohere-secondary-container font-label-mono-bold text-[10px] text-cohere-secondary">
            {record.agentInitials ?? (record.agentName ?? record.aiAgentName ?? "AI").slice(0, 2)}
          </div>
          <div>
            <div className="font-label-mono-bold text-cohere-on-surface">{record.agentName ?? record.aiAgentName}</div>
            <div className="font-micro text-cohere-muted">{record.triggerLabel}</div>
          </div>
        </div>
      ),
    },
    {
      title: "Target",
      dataIndex: "targetPostId",
      width: 120,
      render: (v: string) =>
        v === "System" ? (
          <span className="font-label-mono text-cohere-muted">System</span>
        ) : (
          <span className="font-label-mono text-cohere-action-blue hover:underline">{v}</span>
        ),
    },
    {
      title: "Status",
      dataIndex: "status",
      width: 130,
      render: (status: TaskStatus, record: AdminAITask) => (
        <div>
          <StatusBadge status={status} />
          {record.status === "FAILED" && record.errorMessage && (
            <div className="mt-1 font-micro text-cohere-error">{record.errorMessage}</div>
          )}
        </div>
      ),
    },
    {
      title: "Metrics",
      key: "metrics",
      width: 140,
      align: "right",
      render: (_: unknown, record: AdminAITask) => (
        <div className="font-label-mono text-micro text-cohere-muted">
          <div>{record.durationMs !== null && record.durationMs !== undefined ? `${(record.durationMs / 1000).toFixed(1)}s` : "--"}</div>
          <div>{record.tokens !== null ? `Tokens: ${record.tokens}` : `Retry: ${record.retryCount} (Max)`}</div>
        </div>
      ),
    },
    {
      title: "",
      key: "chevron",
      width: 56,
      render: (_: unknown, record: AdminAITask) => (
        <Button type="text" onClick={() => setDetailId(String(record.id))} icon={<MaterialIcon name="chevron_right" size={20} />} />
      ),
    },
  ];

  const summaryVariantClass = (variant: string) => {
    switch (variant) {
      case "secondary":
        return "bg-cohere-action-blue text-white";
      case "success":
        return "bg-cohere-success text-cohere-secondary border-cohere-secondary-container";
      case "error":
        return "bg-cohere-error-container text-cohere-error border-cohere-error";
      default:
        return "bg-cohere-surface-lowest text-cohere-ink";
    }
  };

  return (
    <div className="mx-auto max-w-[1440px] px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-xl flex flex-col items-start justify-between gap-md md:flex-row md:items-center">
        <div>
          <h1 className="font-headline-xl font-bold tracking-tight text-cohere-ink">AI Reply Tasks</h1>
          <p className="mt-1 font-body-large text-cohere-muted">
            监控并管理自动化 AI 回复队列。
          </p>
        </div>
        <div className="flex gap-sm">
          <Button icon={<MaterialIcon name="refresh" size={18} />}>Refresh</Button>
          {canRetry && (
            <Button type="primary" icon={<MaterialIcon name="play_arrow" size={18} />}>
              Resume All Failed
            </Button>
          )}
        </div>
      </div>

      {/* Status summary bento */}
      <div className="mb-xl grid grid-cols-2 gap-md md:grid-cols-5">
        {summaryCards.map((c) => (
          <div
            key={c.label}
            className={`flex h-32 flex-col justify-between rounded-lg border border-cohere-hairline p-md ${summaryVariantClass(c.variant)}`}
          >
            <div className="flex items-center justify-between">
              <span className="font-label-mono opacity-80">{c.label}</span>
              <MaterialIcon name={c.icon} className={c.variant === "secondary" ? "text-white" : "text-cohere-muted"} size={20} />
            </div>
            <div className="font-headline-lg font-bold">{c.value.toLocaleString()}</div>
          </div>
        ))}
      </div>

      <div className="overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest">
        <div className="flex flex-col items-center justify-between gap-md border-b border-cohere-hairline bg-cohere-surface p-md md:flex-row">
          <div className="flex flex-wrap gap-md">
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              className="w-full md:w-40"
              options={[
                { value: "ALL", label: "All Statuses" },
                { value: "PENDING", label: "Pending" },
                { value: "PROCESSING", label: "Running" },
                { value: "COMPLETED", label: "Success" },
                { value: "FAILED", label: "Failed" },
              ]}
            />
            <Select
              value={agentFilter}
              onChange={setAgentFilter}
              className="w-full md:w-40"
              options={[
                { value: "ALL", label: "All Agents" },
                { value: "Alpha-7", label: "Alpha-7" },
                { value: "Beta-Tutor", label: "Beta-Tutor" },
                { value: "Gamma-Critic", label: "Gamma-Critic" },
              ]}
            />
            <Select
              value={triggerFilter}
              onChange={setTriggerFilter}
              className="w-full md:w-48"
              options={[
                { value: "ALL", label: "All Triggers" },
                { value: "POST_AUTO", label: "Auto-Reply" },
                { value: "MENTION", label: "User Mention" },
                { value: "SCHEDULED", label: "Scheduled Task" },
              ]}
            />
          </div>
          <Input
            placeholder="Search Task ID or Post ID…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            prefix={<MaterialIcon name="search" size={18} className="text-cohere-muted" />}
            className="w-full md:w-64"
            allowClear
          />
        </div>

        <Table<AdminAITask>
          columns={columns}
          dataSource={filtered}
          rowKey="id"
          loading={isLoading}
          pagination={{
            pageSize: 10,
            showTotal: (total, range) => `显示 ${range[0]}-${range[1]} 共 ${total} 个任务`,
          }}
          scroll={{ x: 900 }}
          size="middle"
        />
      </div>

      <TaskDetailDrawer taskId={detailId} onClose={() => setDetailId(null)} canRetry={canRetry} />
    </div>
  );
}

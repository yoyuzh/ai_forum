import { Drawer, Button, App as AntdApp } from "antd";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { adminApi } from "../api/client";
import { AdminAITask } from "../api/types";
import MaterialIcon from "./MaterialIcon";
import StatusBadge from "./StatusBadge";

interface TaskDetailDrawerProps {
  taskId: string | null;
  onClose: () => void;
  canRetry?: boolean;
}

/** Slide-in drawer showing a task's metadata, error log, I/O payload, and
 *  execution timeline. Includes a retry affordance gated on backend RBAC. */
export default function TaskDetailDrawer({ taskId, onClose, canRetry = true }: TaskDetailDrawerProps) {
  const queryClient = useQueryClient();
  const { message } = AntdApp.useApp();

  const { data: task } = useQuery<AdminAITask>({
    queryKey: ["task", taskId],
    queryFn: () => adminApi.tasks.get(taskId!),
    enabled: !!taskId,
  });

  const retryMutation = useMutation({
    mutationFn: () => adminApi.tasks.retry(taskId!),
    onSuccess: (updated) => {
      queryClient.setQueryData(["task", taskId], updated);
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      message.success(`任务 ${updated.id} 已重新入队`);
    },
    onError: () => message.error("重试失败，请稍后再试"),
  });
  const terminateMutation = useMutation({
    mutationFn: () => adminApi.tasks.terminate(taskId!),
    onSuccess: (updated) => {
      queryClient.setQueryData(["task", taskId], updated);
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      message.success(`任务 ${updated.id} 已终止`);
    },
    onError: () => message.error("终止失败，请稍后再试"),
  });
  const markProcessedMutation = useMutation({
    mutationFn: () => adminApi.tasks.markProcessed(taskId!),
    onSuccess: (updated) => {
      queryClient.setQueryData(["task", taskId], updated);
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      message.success(`任务 ${updated.id} 已标记完成`);
    },
    onError: () => message.error("标记失败，请稍后再试"),
  });

  if (!task) return null;

  const metadata = [
    { label: "AI Agent", value: `${task.agentInitials} · ${task.agentName}` },
    { label: "触发类型", value: task.triggerLabel },
    { label: "目标帖子", value: task.targetPostId, link: true },
    {
      label: "执行时长",
      value:
        task.durationMs !== null && task.durationMs !== undefined
          ? `${(task.durationMs / 1000).toFixed(1)}s${task.status === "FAILED" ? " (Timeout)" : ""}`
          : "—",
      sub: `重试次数: ${task.retryCount}/${task.maxRetries}`,
    },
  ];

  return (
    <Drawer
      title={
        <div className="mb-1 flex items-center gap-1">
          <span className="font-feature-title text-cohere-ink">Task Details</span>
          <StatusBadge status={task.status} />
        </div>
      }
      extra={
        <div className="flex items-center gap-1">
          {canRetry && task.status === "FAILED" && (
            <Button onClick={() => retryMutation.mutate()} loading={retryMutation.isPending}>
              Retry Task
            </Button>
          )}
          {canRetry && (task.status === "PENDING" || task.status === "PROCESSING") && (
            <Button danger onClick={() => terminateMutation.mutate()} loading={terminateMutation.isPending}>
              Terminate
            </Button>
          )}
          {canRetry && task.status !== "COMPLETED" && (
            <Button onClick={() => markProcessedMutation.mutate()} loading={markProcessedMutation.isPending}>
              Mark Processed
            </Button>
          )}
        </div>
      }
      placement="right"
      width={672}
      open={!!taskId}
      onClose={onClose}
      styles={{ body: { padding: 24 } }}
    >
      <div className="font-label-mono text-cohere-muted">
        ID: {task.id} · 创建于 {task.createdAt}
      </div>

      {/* Metadata bento */}
      <div className="mt-lg grid grid-cols-2 gap-md">
        {metadata.map((m) => (
          <div key={m.label} className="rounded-lg border border-cohere-hairline bg-cohere-surface-container p-md">
            <div className="font-label-mono text-cohere-muted">{m.label}</div>
            <div className="mt-1 font-body-main text-cohere-on-surface">
              {m.link ? (
                <span className="inline-flex items-center gap-1 text-cohere-action-blue">
                  {m.value} <MaterialIcon name="open_in_new" size={14} />
                </span>
              ) : (
                m.value
              )}
            </div>
            {m.sub && <div className="font-micro text-cohere-muted">{m.sub}</div>}
          </div>
        ))}
      </div>

      {/* Error log */}
      {task.status === "FAILED" && task.errorMessage && (
        <div className="mt-lg overflow-hidden rounded-lg border border-cohere-error bg-[#fff5f5]">
          <div className="flex items-center gap-1 border-b border-cohere-error bg-cohere-error-container px-md py-sm font-label-mono-bold text-cohere-error">
            <MaterialIcon name="warning" size={18} /> Error Log
          </div>
          <pre className="overflow-x-auto bg-[#1e1e1e] p-md font-mono text-[13px] leading-relaxed text-[#d4d4d4]">
            {`[2024-05-20 14:32:31] ERROR: LLMProviderTimeout\n  File "provider/llm.go", line 142\n    TimeoutError("LLM API non-responsive")\n  ${task.errorMessage}`}
          </pre>
        </div>
      )}

      {/* I/O payload */}
      <h3 className="mt-lg border-b border-cohere-hairline pb-xs font-feature-title text-[18px] text-cohere-ink">
        I/O Payload
      </h3>
      <div className="mt-md overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest">
        <div className="flex items-center justify-between border-b border-cohere-hairline bg-cohere-surface-container px-md py-sm">
          <span className="font-label-mono-bold text-cohere-on-surface">System Prompt & Context</span>
          <MaterialIcon name="content_copy" size={16} className="text-cohere-muted" />
        </div>
        <div className="whitespace-pre-wrap break-words bg-cohere-surface p-md font-body-main text-cohere-ink">
          {task.prompt}
        </div>
      </div>

      <div className="mt-md overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest">
        <div className="border-b border-cohere-hairline bg-cohere-surface-container px-md py-sm font-label-mono-bold text-cohere-on-surface">
          Generation Result {task.status === "FAILED" ? "(Partial/Failed)" : ""}
        </div>
        <div
          className={`flex items-center justify-center py-xl ${
            task.result ? "bg-cohere-surface p-md font-body-main text-cohere-ink" : "bg-cohere-surface-variant p-md font-label-mono italic text-cohere-muted"
          }`}
        >
          {task.result || "No content generated due to timeout."}
        </div>
      </div>

      {/* Execution timeline */}
      <h3 className="mt-lg border-b border-cohere-hairline pb-xs font-feature-title text-[18px] text-cohere-ink">
        Execution Timeline
      </h3>
      <div className="relative ml-sm mt-sm space-y-lg border-l border-dotted border-cohere-hairline pl-md">
        {(task.timeline ?? []).map((step, idx) => (
          <div key={idx} className="relative">
            <span
              className={`absolute -left-[25px] top-1 h-4 w-4 rounded-full border-2 ${
                step.state === "error"
                  ? "border-cohere-error bg-cohere-error shadow-[0_0_8px_rgba(186,26,26,0.5)]"
                  : "border-cohere-primary bg-cohere-surface"
              }`}
            />
            <div className={`font-label-mono-bold text-micro ${step.state === "error" ? "text-cohere-error" : "text-cohere-on-surface"}`}>
              {step.time} - {step.label}
            </div>
            <div className="font-caption text-cohere-muted">{step.detail}</div>
          </div>
        ))}
      </div>
    </Drawer>
  );
}

import { useEffect, useState } from "react";
import { Drawer, Switch, Slider, Input, Button, App as AntdApp } from "antd";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { adminApi } from "../api/client";
import { AdminAIAgent } from "../api/types";
import MaterialIcon from "./MaterialIcon";

interface AgentEditDrawerProps {
  agentId: string | null;
  onClose: () => void;
}

type Draft = Pick<
  AdminAIAgent,
  | "systemPrompt"
  | "activityLevel"
  | "temperature"
  | "allowAutoReply"
  | "allowMentionReply"
  | "allowFollowupReply"
>;

const INITIAL_DRAFT: Draft = {
  systemPrompt: "",
  activityLevel: 0.5,
  temperature: 0.6,
  allowAutoReply: false,
  allowMentionReply: false,
  allowFollowupReply: false,
};

/** Slide-in drawer for editing an agent's system prompt, thresholds, and permissions. */
export default function AgentEditDrawer({ agentId, onClose }: AgentEditDrawerProps) {
  const queryClient = useQueryClient();
  const { message } = AntdApp.useApp();
  const [draft, setDraft] = useState<Draft>(INITIAL_DRAFT);

  const { data: agent } = useQuery({
    queryKey: ["agent", agentId],
    queryFn: () => adminApi.agents.get(agentId!),
    enabled: !!agentId,
  });

  useEffect(() => {
    if (agent) {
      setDraft({
        systemPrompt: agent.systemPrompt,
        activityLevel: agent.activityLevel,
        temperature: agent.temperature,
        allowAutoReply: agent.allowAutoReply,
        allowMentionReply: agent.allowMentionReply,
        allowFollowupReply: agent.allowFollowupReply,
      });
    }
  }, [agent]);

  const updateMutation = useMutation({
    mutationFn: (updates: Partial<AdminAIAgent>) => adminApi.agents.update(agentId!, updates),
    onSuccess: () => {
      // Permissions are display-only here; backend RBAC is authoritative.
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      queryClient.invalidateQueries({ queryKey: ["agent", agentId] });
      message.success("代理配置已保存");
      onClose();
    },
    onError: () => message.error("保存失败，请重试"),
  });

  const handleSave = () => updateMutation.mutate(draft);

  return (
    <Drawer
      title={
        agent ? (
          <div>
            <div className="font-feature-title text-cohere-primary">编辑代理配置</div>
            <div className="mt-1 font-label-mono text-cohere-muted">
              #{agent.id} · {agent.name}
            </div>
          </div>
        ) : null
      }
      placement="right"
      width={480}
      open={!!agentId}
      onClose={onClose}
      extra={
        <Button type="primary" onClick={handleSave} loading={updateMutation.isPending}>
          保存更改
        </Button>
      }
      styles={{ body: { padding: 24 } }}
    >
      <div className="flex flex-col gap-xl">
        <section>
          <div className="mb-1 flex items-center gap-1 font-label-mono-bold text-cohere-on-surface">
            <MaterialIcon name="terminal" size={16} /> 系统提示词 (System Prompt)
          </div>
          <Input.TextArea
            value={draft.systemPrompt}
            onChange={(e) => setDraft((d) => ({ ...d, systemPrompt: e.target.value }))}
            rows={8}
            spellCheck={false}
            className="font-mono"
          />
        </section>

        <section className="flex flex-col gap-lg rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md">
          <div className="flex items-center gap-1 border-b border-cohere-hairline pb-sm font-label-mono-bold text-cohere-on-surface">
            <MaterialIcon name="tune" size={16} /> 行为阈值与活跃度
          </div>

          <div>
            <div className="mb-1 flex items-center justify-between font-caption">
              <span>活跃度阈值 (Activity Threshold)</span>
              <span className="rounded bg-cohere-surface-variant px-1 font-label-mono text-cohere-primary">
                {draft.activityLevel.toFixed(2)}
              </span>
            </div>
            <Slider
              min={0}
              max={1}
              step={0.01}
              value={draft.activityLevel}
              onChange={(v) => setDraft((d) => ({ ...d, activityLevel: v }))}
            />
            <p className="font-micro text-cohere-muted">
              决定代理主动插入对话的频率，值越高越不活跃。
            </p>
          </div>

          <div>
            <div className="mb-1 flex items-center justify-between font-caption">
              <span>温度值 (Temperature)</span>
              <span className="rounded bg-cohere-surface-variant px-1 font-label-mono text-cohere-primary">
                {draft.temperature.toFixed(2)}
              </span>
            </div>
            <Slider
              min={0}
              max={2}
              step={0.1}
              value={draft.temperature}
              onChange={(v) => setDraft((d) => ({ ...d, temperature: v }))}
            />
            <p className="font-micro text-cohere-muted">
              控制生成内容的随机性，较低值更确定，较高值更发散。
            </p>
          </div>
        </section>

        <section className="overflow-hidden rounded-lg border border-cohere-hairline">
          <div className="flex items-center gap-1 border-b border-cohere-hairline bg-cohere-surface-low p-md font-label-mono-bold text-cohere-on-surface">
            <MaterialIcon name="admin_panel_settings" size={16} /> 权限配置
          </div>
          <div className="flex flex-col divide-y divide-cohere-hairline">
            <PermissionRow
              label="允许自动回复"
              desc="代理可以无需提及即回应上下文"
              checked={draft.allowAutoReply}
              onChange={(v) => setDraft((d) => ({ ...d, allowAutoReply: v }))}
            />
            <PermissionRow
              label="启用 @提及 响应"
              desc="用户可以通过 @名称 强制唤醒代理"
              checked={draft.allowMentionReply}
              onChange={(v) => setDraft((d) => ({ ...d, allowMentionReply: v }))}
            />
            <PermissionRow
              label="读取历史上下文"
              desc="允许代理访问线程内的过往消息"
              checked={draft.allowFollowupReply}
              onChange={(v) => setDraft((d) => ({ ...d, allowFollowupReply: v }))}
            />
          </div>
        </section>
      </div>
    </Drawer>
  );
}

interface PermissionRowProps {
  label: string;
  desc: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}

function PermissionRow({ label, desc, checked, onChange }: PermissionRowProps) {
  return (
    <div className="flex items-center justify-between bg-cohere-surface-lowest p-md">
      <div>
        <div className="font-body-main text-cohere-on-surface">{label}</div>
        <div className="font-micro text-cohere-muted">{desc}</div>
      </div>
      <Switch checked={checked} onChange={onChange} />
    </div>
  );
}

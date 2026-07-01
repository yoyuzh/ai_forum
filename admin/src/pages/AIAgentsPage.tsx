import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Table, Input, Button, Switch, App as AntdApp } from "antd";
import type { ColumnsType } from "antd/es/table";
import { adminApi } from "../api/client";
import { AdminAIAgent } from "../api/types";
import MaterialIcon from "../components/MaterialIcon";
import AgentEditDrawer from "../components/AgentEditDrawer";

export default function AIAgentsPage() {
  const { data: agents = [], isLoading } = useQuery({
    queryKey: ["agents"],
    queryFn: adminApi.agents.list,
  });
  const { message } = AntdApp.useApp();
  const [search, setSearch] = useState("");
  const [editId, setEditId] = useState<string | null>(null);

  const filtered = agents.filter(
    (a) =>
      a.name.toLowerCase().includes(search.toLowerCase()) ||
      a.id.toLowerCase().includes(search.toLowerCase()) ||
      a.displayName.includes(search),
  );
  const runningCount = filtered.filter((agent) => agent.active).length;

  const columns: ColumnsType<AdminAIAgent> = [
    {
      title: "ID",
      dataIndex: "id",
      width: 80,
      render: (id: string) => <span className="font-label-mono text-cohere-muted">#{id}</span>,
    },
    {
      title: "头像",
      dataIndex: "icon",
      width: 64,
      render: (_: string, record: AdminAIAgent) => (
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-cohere-secondary-container text-cohere-on-secondary-container">
          <MaterialIcon name={record.icon} size={20} />
        </div>
      ),
    },
    {
      title: "名称 / 简介",
      dataIndex: "displayName",
      width: 220,
      render: (_: string, record: AdminAIAgent) => (
        <div className="min-w-0">
          <div className="font-label-mono-bold text-cohere-on-surface">{record.name}</div>
          <div className="truncate text-cohere-muted">{record.description}</div>
        </div>
      ),
    },
    {
      title: "特质",
      dataIndex: "traits",
      width: 160,
      render: (traits: string[]) => (
        <div className="flex flex-wrap gap-1">
          {traits.map((t) => (
            <span
              key={t}
              className="rounded border border-cohere-hairline bg-cohere-surface-variant px-1 py-0.5 font-label-mono text-[10px] text-cohere-on-surface-variant"
            >
              {t}
            </span>
          ))}
        </div>
      ),
    },
    {
      title: "活跃度阈值",
      dataIndex: "activityLevel",
      width: 140,
      render: (v: number) => (
        <div className="flex items-center gap-1">
          <div className="h-1 w-full overflow-hidden rounded-full bg-cohere-surface-variant">
            <div className="h-full rounded-full bg-cohere-primary" style={{ width: `${v * 100}%` }} />
          </div>
          <span className="font-label-mono text-[10px] text-cohere-muted">{v.toFixed(2)}</span>
        </div>
      ),
    },
    {
      title: "自动回复",
      dataIndex: "allowAutoReply",
      width: 90,
      render: (v: boolean) => <Switch size="small" defaultChecked={v} disabled />,
    },
    {
      title: "@支持",
      dataIndex: "allowMentionReply",
      width: 90,
      render: (v: boolean) => <Switch size="small" defaultChecked={v} disabled />,
    },
    {
      title: "状态",
      dataIndex: "active",
      width: 110,
      render: (active: boolean) =>
        active ? (
          <span className="inline-flex items-center gap-1 rounded-sm border border-cohere-secondary-container bg-cohere-success px-1 py-0.5 font-label-mono text-[11px] text-cohere-secondary">
            <span className="status-dot bg-cohere-secondary" /> 运行中
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 rounded-sm border border-cohere-hairline px-1 py-0.5 font-label-mono text-[11px] text-cohere-muted">
            <span className="status-dot bg-cohere-slate" /> 空闲
          </span>
        ),
    },
    {
      title: "操作",
      key: "actions",
      width: 180,
      render: (_: unknown, record: AdminAIAgent) => (
        <div className="flex gap-1">
          <Button size="small" onClick={() => setEditId(record.id)} icon={<MaterialIcon name="edit" size={16} />}>
            编辑
          </Button>
          <Button
            size="small"
            type="text"
            icon={<MaterialIcon name={record.active ? "block" : "play_arrow"} size={16} />}
            onClick={() => message.info(`${record.active ? "停用" : "启用"} ${record.name}（需后端 RBAC 校验）`)}
          />
        </div>
      ),
    },
  ];

  return (
    <div className="mx-auto max-w-[1440px] px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-xl flex flex-col items-start justify-between gap-md md:flex-row md:items-center">
        <div>
          <h1 className="font-headline-xl text-cohere-primary">AI 代理管理</h1>
          <p className="mt-1 font-body-large text-cohere-muted">
            监控、配置和控制系统范围内的所有人工智能代理。
          </p>
        </div>
        <Button type="primary" icon={<MaterialIcon name="add" size={18} />} onClick={() => message.info("新建代理需后端 RBAC 校验")}>
          新建代理
        </Button>
      </div>

      <div className="overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest">
        <div className="flex flex-col items-center justify-between gap-sm border-b border-cohere-hairline bg-cohere-surface-low p-md md:flex-row">
          <Input
            placeholder="搜索 ID, 名称…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            prefix={<MaterialIcon name="search" size={18} className="text-cohere-muted" />}
            className="w-full md:w-64"
            allowClear
          />
          <span className="font-caption text-cohere-muted">
            显示 {filtered.length} 个代理，{runningCount} 个运行中
          </span>
        </div>

        <Table<AdminAIAgent>
          columns={columns}
          dataSource={filtered}
          rowKey="id"
          loading={isLoading}
          pagination={{ pageSize: 10, showSizeChanger: false }}
          scroll={{ x: 1100 }}
          size="middle"
        />
      </div>

      <AgentEditDrawer agentId={editId} onClose={() => setEditId(null)} />
    </div>
  );
}

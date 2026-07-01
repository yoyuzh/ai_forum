import { Table, Tag, Avatar } from "antd";
import type { ColumnsType } from "antd/es/table";
import { AdminUser } from "../api/types";

// Inline mock — users are a thin resource not yet wired to the full mock client.
const USERS: AdminUser[] = [
  { id: 1, username: "NeoResearcher", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=Neo", role: "资深研究员", status: "active", postCount: 42, createdAt: "2024-01-12" },
  { id: 2, username: "DevMaster", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=DevMaster", role: "前端架构师", status: "active", postCount: 87, createdAt: "2024-01-20" },
  { id: 3, username: "spammer", avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=spam", role: "—", status: "banned", postCount: 3, createdAt: "2024-05-01" },
];

export default function UsersManagePage() {
  const columns: ColumnsType<AdminUser> = [
    {
      title: "用户",
      dataIndex: "username",
      render: (_: string, record: AdminUser) => (
        <div className="flex items-center gap-2">
          <Avatar src={record.avatar} size="small" />
          <span className="font-label-mono-bold text-cohere-on-surface">{record.username}</span>
        </div>
      ),
    },
    { title: "角色", dataIndex: "role", width: 140 },
    { title: "发帖数", dataIndex: "postCount", width: 100, align: "right" },
    { title: "注册时间", dataIndex: "createdAt", width: 140 },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (status: AdminUser["status"]) => (
        <Tag color={status === "active" ? "green" : "red"}>
          {status === "active" ? "正常" : "已封禁"}
        </Tag>
      ),
    },
  ];

  return (
    <div className="mx-auto max-w-[1440px] px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-xl">
        <h1 className="font-headline-xl text-cohere-primary">用户管理</h1>
        <p className="mt-1 font-body-large text-cohere-muted">查看与管理论坛注册用户。</p>
      </div>
      <div className="overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md">
        <Table<AdminUser>
          columns={columns}
          dataSource={USERS}
          rowKey="id"
          pagination={{ pageSize: 10 }}
          size="middle"
        />
      </div>
    </div>
  );
}

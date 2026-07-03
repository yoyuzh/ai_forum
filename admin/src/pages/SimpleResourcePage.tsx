import { useQuery } from "@tanstack/react-query";
import { Table } from "antd";
import { adminApi } from "../api/client";

export function CommentsPage() {
  const { data = [] } = useQuery({ queryKey: ["comments"], queryFn: adminApi.comments.list });
  return <SimpleResource title="评论管理" rows={data as Record<string, unknown>[]} />;
}

export function TagsPage() {
  const { data = [] } = useQuery({ queryKey: ["tags"], queryFn: adminApi.tags.list });
  return <SimpleResource title="标签管理" rows={data as unknown as Record<string, unknown>[]} />;
}

export function PreferencesPage() {
  const { data = [] } = useQuery({ queryKey: ["preferences"], queryFn: adminApi.preferences.list });
  return <SimpleResource title="偏好管理" rows={data as unknown as Record<string, unknown>[]} />;
}

function SimpleResource({ title, rows }: { title: string; rows: Record<string, unknown>[] }) {
  const keys = Array.from(new Set(rows.flatMap((row) => Object.keys(row)))).slice(0, 8);
  return (
    <div className="mx-auto max-w-[1440px] px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-xl">
        <h1 className="font-headline-xl text-cohere-primary">{title}</h1>
      </div>
      <div className="overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md">
        <Table
          dataSource={rows}
          rowKey={(row) => String(row.id)}
          columns={keys.map((key) => ({ title: key, dataIndex: key, render: (v: unknown) => String(v ?? "") }))}
          size="middle"
          scroll={{ x: 900 }}
        />
      </div>
    </div>
  );
}

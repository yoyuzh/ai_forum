import { useQuery } from "@tanstack/react-query";
import { Table, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import { adminApi } from "../api/client";
import { AdminPost } from "../api/types";

export default function PostsManagePage() {
  const { data: posts = [], isLoading } = useQuery({
    queryKey: ["adminPosts"],
    queryFn: adminApi.posts.list,
  });

  const columns: ColumnsType<AdminPost> = [
    { title: "ID", dataIndex: "id", width: 70 },
    { title: "标题", dataIndex: "title" },
    { title: "作者", dataIndex: "author", width: 140 },
    { title: "分类", dataIndex: "category", width: 120 },
    {
      title: "浏览",
      dataIndex: "viewCount",
      width: 90,
      align: "right",
    },
    {
      title: "评论",
      dataIndex: "commentCount",
      width: 90,
      align: "right",
    },
    {
      title: "AI 回复",
      dataIndex: "aiResponsesCount",
      width: 100,
      align: "right",
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      render: (status: AdminPost["status"]) => (
        <Tag color={status === "published" ? "green" : "default"}>
          {status === "published" ? "已发布" : status === "review" ? "审核中" : "草稿"}
        </Tag>
      ),
    },
    { title: "发布时间", dataIndex: "createdAt", width: 120 },
  ];

  return (
    <div className="mx-auto max-w-[1440px] px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-xl">
        <h1 className="font-headline-xl text-cohere-primary">帖子管理</h1>
        <p className="mt-1 font-body-large text-cohere-muted">审核、检索并管理论坛中的所有帖子。</p>
      </div>
      <div className="overflow-hidden rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-md">
        <Table<AdminPost>
          columns={columns}
          dataSource={posts}
          rowKey="id"
          loading={isLoading}
          pagination={{ pageSize: 10 }}
          scroll={{ x: 900 }}
          size="middle"
        />
      </div>
    </div>
  );
}

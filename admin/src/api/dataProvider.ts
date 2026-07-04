import type { DataProvider } from "@refinedev/core";
import { adminApi } from "./client";

class ReadOnlyResourceError extends Error {
  constructor(resource: string) {
    super(`${resource} is read-only in real mode`);
    this.name = "ReadOnlyResourceError";
  }
}

const listByResource: Record<string, () => Promise<unknown[]>> = {
  users: adminApi.users.list,
  posts: adminApi.posts.list,
  comments: adminApi.comments.list,
  agents: adminApi.agents.list,
  tasks: adminApi.tasks.list,
  decisionLogs: adminApi.decisionLogs.list,
  tags: adminApi.tags.list,
  preferences: adminApi.preferences.list,
};

const provider: Record<string, any> = {
  getList: async ({ resource, pagination, filters, sorters }: any) => {
    const current = pagination?.current ?? 1;
    const pageSize = pagination?.pageSize ?? 10;
    let rows = ((await listByResource[resource]?.()) ?? []) as Record<string, unknown>[];

    for (const f of filters ?? []) {
      if ("field" in f && f.operator === "eq" && f.value !== undefined) {
        rows = rows.filter((row) => String(row[f.field as string]) === String(f.value));
      }
    }
    const sorter = sorters?.[0];
    if (sorter) {
      rows = [...rows].sort((a, b) => {
        const av = String(a[sorter.field]);
        const bv = String(b[sorter.field]);
        return sorter.order === "desc" ? bv.localeCompare(av) : av.localeCompare(bv);
      });
    }
    const total = rows.length;
    const start = (current - 1) * pageSize;
    return { data: rows.slice(start, start + pageSize), total };
  },
  getOne: async ({ resource, id }: any) => {
    const rows = ((await listByResource[resource]?.()) ?? []) as Record<string, unknown>[];
    return { data: rows.find((row) => String(row.id) === String(id)) ?? null };
  },
  update: async ({ resource, id, variables }: any) => {
    if (resource === "agents") return { data: await adminApi.agents.update(id, variables as Record<string, unknown>) };
    throw new ReadOnlyResourceError(resource);
  },
  create: async ({ resource }: any) => {
    throw new ReadOnlyResourceError(resource);
  },
  deleteOne: async ({ resource }: any) => {
    throw new ReadOnlyResourceError(resource);
  },
  getMany: async () => ({ data: [] }),
  updateMany: async () => ({ data: [] }),
  deleteMany: async () => ({ data: [] }),
  getApiUrl: () => "",
  custom: async () => ({ data: {} }),
};

export const dataProvider = provider as unknown as DataProvider;

import type { DataProvider } from "@refinedev/core";
import { adminApi } from "../api/client";

/**
 * Refine dataProvider that wraps the mock API client.
 *
 * Only `getList` / `getOne` / `update` are wired — the admin pages that mutate
 * (agent config, task retry) go through `adminApi` directly with TanStack Query
 * mutations, since those flows need bespoke payloads (sliders, retry semantics)
 * that a generic CRUD provider would only awkwardly express. When the real
 * backend lands, swap this for `@refinedev/simple-rest` against the api-server.
 *
 * The object is built untyped then cast because Refine's `DataProvider` is
 * invariant on its per-call `TData` generic; the mock layer intentionally
 * returns heterogeneous record shapes per resource.
 */
const provider = {
  getList: async ({ resource, pagination, filters, sorters }: { resource: string; pagination?: { current?: number; pageSize?: number }; filters?: Array<{ field?: string; operator?: string; value?: unknown }>; sorters?: Array<{ field: string; order: "asc" | "desc" }> }) => {
    const current = pagination?.current ?? 1;
    const pageSize = pagination?.pageSize ?? 10;

    let rows: Record<string, unknown>[] = [];
    if (resource === "agents") rows = (await adminApi.agents.list()) as unknown as Record<string, unknown>[];
    else if (resource === "tasks") rows = (await adminApi.tasks.list()) as unknown as Record<string, unknown>[];
    else if (resource === "decisionLogs") rows = (await adminApi.decisionLogs.list()) as unknown as Record<string, unknown>[];
    else if (resource === "posts") rows = (await adminApi.posts.list()) as unknown as Record<string, unknown>[];

    for (const f of filters ?? []) {
      if (f.field && f.operator === "eq" && f.value !== undefined) {
        rows = rows.filter((row) => String(row[f.field as string]) === String(f.value));
      }
    }

    if (sorters && sorters.length > 0) {
      const { field, order } = sorters[0];
      rows = [...rows].sort((a, b) => {
        const av = String(a[field]);
        const bv = String(b[field]);
        if (av === bv) return 0;
        const cmp = av > bv ? 1 : -1;
        return order === "desc" ? -cmp : cmp;
      });
    }

    const total = rows.length;
    const start = (current - 1) * pageSize;
    return { data: rows.slice(start, start + pageSize), total };
  },

  getOne: async ({ resource, id }: { resource: string; id: string | number }) => {
    let data: unknown = null;
    if (resource === "agents") data = await adminApi.agents.get(String(id));
    else if (resource === "tasks") data = await adminApi.tasks.get(String(id));
    return { data };
  },

  update: async ({ resource, id, variables }: { resource: string; id: string | number; variables: unknown }) => {
    if (resource === "agents") {
      const data = await adminApi.agents.update(String(id), variables as Record<string, unknown>);
      return { data };
    }
    throw new Error(`update not implemented for resource ${resource}`);
  },

  create: async () => {
    throw new Error("create not implemented in mock data provider");
  },
  deleteOne: async () => {
    throw new Error("deleteOne not implemented in mock data provider");
  },
  getMany: async () => ({ data: [] }),
  updateMany: async () => ({ data: [] }),
  deleteMany: async () => ({ data: [] }),
  getApiUrl: () => "",
  custom: async () => ({ data: null }),
};

export const mockDataProvider = provider as unknown as DataProvider;

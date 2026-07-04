import type { AccessControlProvider } from "@refinedev/core";
import { getStoredSession } from "../api/client";

const resourcePermissions: Record<string, string> = {
  agents: "ai_agent:update",
  tasks: "ai_task:retry",
  decisionLogs: "decision_log:read",
  posts: "post:delete-any",
  users: "user:ban",
};

export const accessControlProvider: AccessControlProvider = {
  can: async ({ resource, action }) => {
    if (action === "list" || action === "show") return { can: true };
    const permission = resource ? resourcePermissions[resource] : undefined;
    return { can: !permission || (getStoredSession()?.permissions ?? []).includes(permission) };
  },
};

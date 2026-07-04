import type { AuthProvider } from "@refinedev/core";
import { adminApi, getAdminToken, getStoredSession } from "./client";

export const authProvider: AuthProvider = {
  login: async ({ username, password }) => {
    await adminApi.auth.login(String(username), String(password));
    return { success: true, redirectTo: "/" };
  },
  logout: async () => {
    await adminApi.auth.logout();
    return { success: true, redirectTo: "/login" };
  },
  check: async () => {
    if (!getAdminToken()) return { authenticated: false, redirectTo: "/login" };
    try {
      await adminApi.auth.me();
      return { authenticated: true };
    } catch {
      return { authenticated: false, redirectTo: "/login" };
    }
  },
  getIdentity: async () => getStoredSession(),
  getPermissions: async () => getStoredSession()?.permissions ?? [],
  onError: async (error) => {
    if (error?.status === 401) return { logout: true, redirectTo: "/login" };
    return {};
  },
};

import { getStoredSession } from "../api/client";

export function usePermission(permission: string): boolean {
  return (getStoredSession()?.permissions ?? []).includes(permission);
}

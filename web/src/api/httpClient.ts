import { clearAuthToken, getAuthToken } from "./auth";
import { useUserStore } from "../stores/useUserStore";

export class HttpError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "HttpError";
  }
}

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");

type HttpInit = RequestInit & { skipAuthRedirect?: boolean };

function resolveURL(path: string): string {
  return `${API_BASE_URL}${path.startsWith("/") ? path : `/${path}`}`;
}

async function errorMessage(response: Response): Promise<string> {
  const text = await response.text();
  if (response.status === 403) return text || "没有权限执行该操作";
  if (response.status === 429) return text || "请求过快，请稍后再试";
  return text || `HTTP ${response.status}`;
}

export async function http<T>(path: string, init: HttpInit = {}): Promise<T> {
  const { skipAuthRedirect, ...requestInit } = init;
  const headers = new Headers(init.headers);
  if (!headers.has("Content-Type") && init.body) headers.set("Content-Type", "application/json");
  const token = getAuthToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const response = await fetch(resolveURL(path), { ...requestInit, headers });
  if (response.status === 401) {
    if (skipAuthRedirect) {
      throw new HttpError(401, "请先登录");
    }
    clearAuthToken();
    useUserStore.getState().clearAuthed();
    if (window.location.pathname !== "/login") window.location.assign("/login");
    throw new HttpError(401, "请先登录");
  }
  if (!response.ok) {
    throw new HttpError(response.status, await errorMessage(response));
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

export function apiURL(path: string): string {
  return resolveURL(path);
}

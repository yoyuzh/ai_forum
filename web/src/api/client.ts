import type { ApiClient } from "./types";
import { mockApi } from "./mockClient";
import { realApi } from "./realClient";

export const api: ApiClient = import.meta.env.VITE_API_MODE === "real" ? realApi : mockApi;
export const apiMode = import.meta.env.VITE_API_MODE === "real" ? "real" : "mock";

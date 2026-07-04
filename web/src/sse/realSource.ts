import { api } from "../api/client";
import { apiURL } from "../api/httpClient";
import type { AIStatusSnapshot } from "../api/types";
import { useConnectionStore } from "../stores/useConnectionStore";
import { sseEmitter } from "./emitter";

const sources = new Map<number, EventSource>();
const lastEventIds = new Map<number, string>();

function reconcile(postId: number) {
  api.aiStatus
    .get(postId)
    .then((status: AIStatusSnapshot) => {
      sseEmitter.emit("ai-status.updated", { postId, ...status });
    })
    .catch(() => undefined);
}

function emitServerEvent(postId: number, event: MessageEvent) {
  if (event.lastEventId) lastEventIds.set(postId, event.lastEventId);
  let data: any;
  try {
    data = event.data ? JSON.parse(event.data) : {};
  } catch {
    data = {};
  }
  const type = event.type === "message" ? data.type : event.type;
  if (!type) return;
  sseEmitter.emit(type, { postId, ...data });
}

export function subscribePostEvents(postId: number): () => void {
  if (!Number.isFinite(postId) || sources.has(postId)) return () => undefined;
  useConnectionStore.getState().setSSEStatus("connecting");
  const lastId = lastEventIds.get(postId);
  const url = apiURL(
    `/api/posts/${postId}/events${lastId ? `?lastEventId=${encodeURIComponent(lastId)}` : ""}`,
  );
  const source = new EventSource(url);
  sources.set(postId, source);

  source.onopen = () => {
    useConnectionStore.getState().setSSEStatus("connected");
    reconcile(postId);
  };
  source.onerror = () => {
    useConnectionStore.getState().setSSEStatus("disconnected");
    reconcile(postId);
  };
  for (const type of ["ai_reply_completed", "comment.created", "post.updated", "task.updated"]) {
    source.addEventListener(type, (event) => emitServerEvent(postId, event as MessageEvent));
  }
  source.onmessage = (event) => emitServerEvent(postId, event);

  return () => {
    source.close();
    sources.delete(postId);
  };
}

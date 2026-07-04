import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { useSSE } from "../sse/useSSE";

export function useAIStatus(postId: number) {
  const queryClient = useQueryClient();

  useSSE("ai-status.updated", (status: { postId?: number }) => {
    if (status.postId === postId) {
      queryClient.invalidateQueries({ queryKey: ["aiStatus", postId] });
    }
  });

  useSSE("task.updated", (event: { postId?: number }) => {
    if (event.postId === postId) {
      queryClient.invalidateQueries({ queryKey: ["aiStatus", postId] });
    }
  });

  useSSE("ai_reply_completed", (event: { postId?: number }) => {
    if (event.postId === postId) {
      queryClient.invalidateQueries({ queryKey: ["aiStatus", postId] });
    }
  });

  return useQuery({
    queryKey: ["aiStatus", postId],
    queryFn: () => api.aiStatus.get(postId),
    enabled: !Number.isNaN(postId),
  });
}

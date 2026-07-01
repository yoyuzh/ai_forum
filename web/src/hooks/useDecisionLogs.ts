import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";
import { useSSE } from "../sse/useSSE";
import { useQueryClient } from "@tanstack/react-query";

/** Decision logs for a single post — powers the post-detail sidebar. */
export function useDecisionLogsForPost(postId: number) {
  const queryClient = useQueryClient();

  useSSE("decision_log.created", (log: { postId?: number }) => {
    if (log.postId === postId) {
      queryClient.invalidateQueries({ queryKey: ["decisionLogs", postId] });
    }
  });

  return useQuery({
    queryKey: ["decisionLogs", postId],
    queryFn: () => api.decisionLogs.listForPost(postId),
    enabled: !Number.isNaN(postId),
  });
}

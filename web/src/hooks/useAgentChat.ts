import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";

export function useAgentChat(agentId: number) {
  const queryClient = useQueryClient();
  const queryKey = ["agent-chat", agentId] as const;
  const historyQueryKey = ["agent-chat-history"] as const;

  const historyQuery = useQuery({
    queryKey: historyQueryKey,
    queryFn: api.chat.list,
    enabled: !Number.isNaN(agentId),
  });

  const chatQuery = useQuery({
    queryKey,
    queryFn: () => api.chat.get(agentId),
    enabled: !Number.isNaN(agentId),
  });

  const sendMutation = useMutation({
    mutationFn: (content: string) => api.chat.sendMessage(agentId, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
      queryClient.invalidateQueries({ queryKey: historyQueryKey });
    },
  });

  return {
    history: historyQuery.data ?? [],
    isHistoryLoading: historyQuery.isLoading,
    chat: chatQuery.data,
    isLoading: chatQuery.isLoading,
    error: chatQuery.error,
    sendMessage: sendMutation.mutateAsync,
    isSending: sendMutation.isPending,
    sendError: sendMutation.error,
  };
}

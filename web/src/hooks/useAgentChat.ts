import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";

export function useAgentChat(agentId: number, sessionId?: number) {
  const queryClient = useQueryClient();
  const queryKey = ["agent-chat", agentId, sessionId ?? null] as const;
  const historyQueryKey = ["agent-chat-history"] as const;

  const historyQuery = useQuery({
    queryKey: historyQueryKey,
    queryFn: api.chat.list,
    enabled: !Number.isNaN(agentId),
  });

  const chatQuery = useQuery({
    queryKey,
    queryFn: () => api.chat.get(agentId, sessionId),
    enabled: !Number.isNaN(agentId),
  });

  const createMutation = useMutation({
    mutationFn: () => api.chat.create(agentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: historyQueryKey });
    },
  });

  const sendMutation = useMutation({
    mutationFn: (content: string) => api.chat.sendMessage(agentId, content, sessionId),
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
    createChat: createMutation.mutateAsync,
    isCreatingChat: createMutation.isPending,
    sendMessage: sendMutation.mutateAsync,
    isSending: sendMutation.isPending,
    sendError: sendMutation.error,
  };
}

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import type { AIChat, AIChatSessionSummary } from "../api/types";

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
    onSuccess: (next) => {
      queryClient.setQueryData(["agent-chat", agentId, next.session.id] as const, next);
      queryClient.setQueryData<AIChatSessionSummary[]>(historyQueryKey, (old = []) => [
        { session: next.session, agent: next.agent, lastMessage: "", messageCount: 0 },
        ...old.filter((item) => item.session.id !== next.session.id),
      ]);
      queryClient.invalidateQueries({ queryKey: historyQueryKey });
    },
  });

  const sendMutation = useMutation({
    mutationFn: (content: string) => api.chat.sendMessage(agentId, content, sessionId),
    onSuccess: (result) => {
      queryClient.setQueryData<AIChat>(queryKey, (old) =>
        old
          ? {
              ...old,
              session: result.session,
              messages: [
                ...old.messages,
                result.userMessage,
                ...(result.assistantMessage ? [result.assistantMessage] : []),
              ],
            }
          : old,
      );
      queryClient.setQueryData<AIChatSessionSummary[]>(historyQueryKey, (old = []) =>
        old.map((item) =>
          item.session.id === result.session.id
            ? {
                ...item,
                session: result.session,
                lastMessage: result.assistantMessage?.content ?? result.userMessage.content,
                messageCount: item.messageCount + (result.assistantMessage ? 2 : 1),
              }
            : item,
        ),
      );
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

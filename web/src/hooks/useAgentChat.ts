import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import type { AIAgent, AIChat, AIChatMessage, AIChatSessionSummary, AIChatStreamEvent } from "../api/types";

export function useAgentChat(agentId: number, sessionId?: number) {
  const queryClient = useQueryClient();
  const queryKey = ["agent-chat", sessionId ?? "draft", agentId] as const;
  const historyQueryKey = ["agent-chat-history"] as const;
  const [streamingChat, setStreamingChat] = useState<AIChat | null>(null);

  useEffect(() => setStreamingChat(null), [agentId, sessionId]);

  const historyQuery = useQuery({
    queryKey: historyQueryKey,
    queryFn: () => api.chat.list({ page: 1, pageSize: 20 }),
    enabled: !Number.isNaN(agentId),
  });

  const agentQuery = useQuery({
    queryKey: ["agent", agentId],
    queryFn: () => api.agents.get(agentId),
    enabled: !Number.isNaN(agentId) && !sessionId,
  });

  const chatQuery = useQuery({
    queryKey,
    queryFn: () => api.chat.get(sessionId!),
    enabled: !Number.isNaN(agentId) && Boolean(sessionId),
  });

  const draftChat = useMemo(() => (agentQuery.data ? makeDraftChat(agentQuery.data) : undefined), [agentQuery.data]);
  const chat = streamingChat ?? chatQuery.data ?? draftChat;

  const applyStreamEvent = (event: AIChatStreamEvent) => {
    setStreamingChat((current) => applyChatEvent(current ?? chat ?? draftChat, event));
  };

  const sendMutation = useMutation({
    mutationFn: (content: string) =>
      api.chat.sendMessage(agentId, content, sessionId ?? null, crypto.randomUUID(), applyStreamEvent),
    onSuccess: (result) => {
      const nextChat = streamingChat ?? {
        session: result.session,
        agent: chat!.agent,
        messages: [result.userMessage, result.assistantMessage].filter((message): message is AIChatMessage => Boolean(message)),
      };
      queryClient.setQueryData(["agent-chat", result.session.id, result.session.aiAgentId] as const, nextChat);
      queryClient.setQueryData<Awaited<ReturnType<typeof api.chat.list>>>(historyQueryKey, (old) => upsertHistory(old, nextChat));
      queryClient.invalidateQueries({ queryKey: historyQueryKey });
      queryClient.invalidateQueries({ queryKey: ["agent-chat", result.session.id, result.session.aiAgentId] });
    },
  });

  const retryMutation = useMutation({
    mutationFn: (messageId: number) => api.chat.retryMessage(messageId, crypto.randomUUID(), applyStreamEvent),
    onSuccess: () => {
      if (chat?.session.id) queryClient.invalidateQueries({ queryKey: ["agent-chat", chat.session.id, chat.session.aiAgentId] });
      queryClient.invalidateQueries({ queryKey: historyQueryKey });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (conversationId: number) => api.chat.deleteConversation(conversationId),
    onSuccess: (_result, conversationId) => {
      queryClient.setQueryData<Awaited<ReturnType<typeof api.chat.list>>>(historyQueryKey, (old) =>
        old ? { ...old, items: old.items.filter((item) => item.session.id !== conversationId), total: Math.max(0, old.total - 1) } : old,
      );
      queryClient.removeQueries({ queryKey: ["agent-chat", conversationId] });
      queryClient.invalidateQueries({ queryKey: historyQueryKey });
    },
  });

  return {
    history: historyQuery.data?.items ?? [],
    isHistoryLoading: historyQuery.isLoading,
    chat,
    isLoading: Boolean(sessionId ? chatQuery.isLoading : agentQuery.isLoading),
    error: chatQuery.error ?? agentQuery.error,
    sendMessage: sendMutation.mutateAsync,
    isSending: sendMutation.isPending,
    sendError: sendMutation.error,
    retryMessage: retryMutation.mutateAsync,
    isRetrying: retryMutation.isPending,
    deleteConversation: deleteMutation.mutateAsync,
    isDeleting: deleteMutation.isPending,
  };
}

function makeDraftChat(agent: AIAgent): AIChat {
  const now = new Date().toISOString();
  return {
    agent,
    session: {
      id: 0,
      userId: 0,
      aiAgentId: agent.id,
      title: agent.displayName,
      status: "ACTIVE",
      lastMessagePreview: "",
      messageCount: 0,
      createdAt: now,
      updatedAt: now,
    },
    messages: [],
  };
}

function applyChatEvent(chat: AIChat | undefined, event: AIChatStreamEvent): AIChat | null {
  if (!chat) return null;
  if (event.event === "conversation_created") {
    return { ...chat, session: event.data.session };
  }
  if (event.event === "user_message_saved") {
    return appendOrReplace(chat, event.data.message);
  }
  if (event.event === "ai_message_created") {
    return appendOrReplace(chat, event.data.message);
  }
  if (event.event === "token") {
    const messages = [...chat.messages];
    const index = findLastAssistant(messages);
    if (index >= 0) {
      messages[index] = { ...messages[index], content: `${messages[index].content}${event.data.content}`, status: "STREAMING" };
      return { ...chat, messages };
    }
  }
  if (event.event === "done") {
    return { ...appendOrReplace(chat, event.data.message), session: event.data.session };
  }
  if (event.event === "error" && event.data.aiMessage) {
    return appendOrReplace(chat, event.data.aiMessage);
  }
  return chat;
}

function appendOrReplace(chat: AIChat, message: AIChatMessage): AIChat {
  const index = chat.messages.findIndex((item) => item.id === message.id);
  const messages = index >= 0 ? [...chat.messages] : [...chat.messages, message];
  if (index >= 0) messages[index] = message;
  messages.sort((a, b) => a.sequenceNo - b.sequenceNo);
  return { ...chat, messages };
}

function findLastAssistant(messages: AIChatMessage[]): number {
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i].role === "assistant") return i;
  }
  return -1;
}

function upsertHistory(
  old: Awaited<ReturnType<typeof api.chat.list>> | undefined,
  chat: AIChat,
): Awaited<ReturnType<typeof api.chat.list>> | undefined {
  if (!old) return old;
  const summary: AIChatSessionSummary = {
    session: chat.session,
    agent: chat.agent,
    lastMessage: chat.session.lastMessagePreview || chat.messages.at(-1)?.content || "",
    messageCount: chat.session.messageCount || chat.messages.length,
  };
  return { ...old, items: [summary, ...old.items.filter((item) => item.session.id !== chat.session.id)] };
}

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { HttpError } from "../api/httpClient";
import { api } from "../api/client";
import { useUserStore } from "../stores/useUserStore";

export function useNotifications() {
  const queryClient = useQueryClient();
  const currentUser = useUserStore((s) => s.currentUser);
  const list = useQuery({
    queryKey: ["notifications"],
    queryFn: async () => {
      try {
        return await api.notifications.list();
      } catch (error) {
        if (error instanceof HttpError && error.status === 401) return [];
        throw error;
      }
    },
    enabled: Boolean(currentUser),
  });
  const unread = useQuery({
    queryKey: ["notifications", "unread"],
    queryFn: async () => {
      try {
        return await api.notifications.unreadCount();
      } catch (error) {
        if (error instanceof HttpError && error.status === 401) return 0;
        throw error;
      }
    },
    enabled: Boolean(currentUser),
    refetchInterval: 30_000,
  });
  const markRead = useMutation({
    mutationFn: api.notifications.markRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });
  const markAllRead = useMutation({
    mutationFn: api.notifications.markAllRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });
  return {
    notifications: list.data ?? [],
    unreadCount: unread.data ?? 0,
    markRead: markRead.mutateAsync,
    markAllRead: markAllRead.mutateAsync,
  };
}

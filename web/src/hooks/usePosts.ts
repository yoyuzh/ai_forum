import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { api, apiMode } from "../api/client";
import { FeedTab, Post } from "../api/types";
import { subscribePostEvents } from "../sse/realSource";
import { useSSE } from "../sse/useSSE";

export function usePosts(tab: FeedTab = "latest", query = "", tag?: string) {
  const queryClient = useQueryClient();

  // Real-time cache refresh when the simulated SSE pushes post updates.
  useSSE("post.updated", () => {
    queryClient.invalidateQueries({ queryKey: ["posts"] });
  });

  const postsQuery = useQuery({
    queryKey: ["posts", tab, query, tag],
    queryFn: () => api.posts.listByFilter(tab, query, tag),
  });

  const createPostMutation = useMutation({
    mutationFn: api.posts.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
    },
  });

  return {
    posts: postsQuery.data ?? [],
    isLoading: postsQuery.isLoading,
    createPost: createPostMutation.mutateAsync,
    isCreating: createPostMutation.isPending,
  };
}

export function usePostDetail(id: number) {
  const queryClient = useQueryClient();

  useSSE("post.updated", (updatedPost: { id?: number }) => {
    if (updatedPost.id === id) {
      queryClient.setQueryData(["post", id], updatedPost);
    }
  });

  useSSE("ai-status.updated", (status: { postId?: number; overallStatus?: string }) => {
    if (status.postId === id) {
      queryClient.invalidateQueries({ queryKey: ["comments", id] });
      queryClient.setQueryData(["post", id], (current: Post | undefined) =>
        current
          ? {
              ...current,
              aiStatus: status.overallStatus === "COMPLETED" ? "COMPLETED" : current.aiStatus,
            }
          : current,
      );
    }
  });

  useSSE("ai_reply_completed", (event: { postId?: number }) => {
    if (event.postId === id) {
      queryClient.invalidateQueries({ queryKey: ["comments", id] });
      queryClient.invalidateQueries({ queryKey: ["post", id] });
    }
  });

  useEffect(() => {
    if (apiMode !== "real" || Number.isNaN(id)) return undefined;
    return subscribePostEvents(id);
  }, [id]);

  return useQuery({
    queryKey: ["post", id],
    queryFn: () => api.posts.get(id),
    enabled: !Number.isNaN(id),
  });
}

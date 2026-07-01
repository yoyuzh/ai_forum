import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { FeedTab } from "../api/types";
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

  return useQuery({
    queryKey: ["post", id],
    queryFn: () => api.posts.get(id),
    enabled: !Number.isNaN(id),
  });
}

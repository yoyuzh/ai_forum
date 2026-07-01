import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { useSSE } from "../sse/useSSE";

export function useComments(postId: number) {
  const queryClient = useQueryClient();

  useSSE("comment.created", (newComment: { postId?: number }) => {
    if (newComment.postId === postId) {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
    }
  });

  const commentsQuery = useQuery({
    queryKey: ["comments", postId],
    queryFn: () => api.comments.list(postId),
    enabled: !Number.isNaN(postId),
  });

  const createCommentMutation = useMutation({
    mutationFn: api.comments.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["comments", postId] });
    },
  });

  return {
    comments: commentsQuery.data ?? [],
    isLoading: commentsQuery.isLoading,
    createComment: createCommentMutation.mutateAsync,
    isSubmitting: createCommentMutation.isPending,
  };
}

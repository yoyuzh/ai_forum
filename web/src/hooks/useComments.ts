import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import type { Comment } from "../api/types";
import { useSSE } from "../sse/useSSE";

function uniqueById(comments: Comment[]): Comment[] {
  const seen = new Set<number>();
  return comments.filter((comment) => {
    if (seen.has(comment.id)) return false;
    seen.add(comment.id);
    return true;
  });
}

export function useComments(postId: number) {
  const queryClient = useQueryClient();

  useSSE("comment.created", (newComment: Comment & { postId?: number }) => {
    if (newComment.postId === postId) {
      queryClient.setQueryData(["comments", postId], (current: Comment[] = []) =>
        uniqueById([...current, newComment]),
      );
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

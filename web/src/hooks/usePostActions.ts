import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";

export function usePostActions(postId: number) {
  const queryClient = useQueryClient();
  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["post", postId] });
    queryClient.invalidateQueries({ queryKey: ["posts"] });
  };

  const like = useMutation({
    mutationFn: () => api.likes.likePost(postId),
    onSuccess: invalidate,
  });

  const favorite = useMutation({
    mutationFn: () => api.favorites.favoritePost(postId),
    onSuccess: invalidate,
  });

  return {
    likePost: like.mutateAsync,
    favoritePost: favorite.mutateAsync,
    isLiking: like.isPending,
    isFavoriting: favorite.isPending,
  };
}

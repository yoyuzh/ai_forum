import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";
import type { Post, RelatedDiscussion } from "../api/types";

export function useRelatedDiscussions(post: Post | undefined) {
  return useQuery({
    queryKey: ["relatedDiscussions", post?.id, post?.tags],
    enabled: Boolean(post),
    queryFn: async (): Promise<RelatedDiscussion[]> => {
      if (!post) return [];
      const posts = await api.posts.list();
      const tagSet = new Set(post.tags.map((tag) => tag.toLowerCase()));
      return posts
        .filter((candidate) => candidate.id !== post.id)
        .map((candidate) => ({
          post: candidate,
          overlap: candidate.tags.filter((tag) => tagSet.has(tag.toLowerCase())).length,
        }))
        .filter((candidate) => tagSet.size === 0 || candidate.overlap > 0)
        .sort((a, b) => b.overlap - a.overlap || +new Date(b.post.createdAt) - +new Date(a.post.createdAt))
        .slice(0, 3)
        .map(({ post }) => ({ id: post.id, title: post.title }));
    },
  });
}

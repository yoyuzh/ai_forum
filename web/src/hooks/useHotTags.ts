import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

export function useHotTags() {
  const hotTagsQuery = useQuery({
    queryKey: ["hotTags"],
    queryFn: api.tags.hot,
  });

  return {
    tags: hotTagsQuery.data?.map((tag) => tag.name) ?? [],
    isLoading: hotTagsQuery.isLoading,
  };
}

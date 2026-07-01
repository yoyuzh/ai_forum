import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

/** Recent AI actions across the forum — powers the home sidebar timeline. */
export function useActivities() {
  const query = useQuery({
    queryKey: ["activities"],
    queryFn: api.activities.list,
  });
  return {
    activities: query.data ?? [],
    isLoading: query.isLoading,
  };
}

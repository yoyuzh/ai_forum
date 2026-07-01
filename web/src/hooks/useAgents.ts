import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api/client";
import { AIAgent } from "../api/types";

export function useAgents() {
  const queryClient = useQueryClient();

  const agentsQuery = useQuery({
    queryKey: ["agents"],
    queryFn: api.agents.list,
  });

  const updateAgentMutation = useMutation({
    mutationFn: ({ id, updates }: { id: number; updates: Partial<AIAgent> }) =>
      api.agents.update(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["agents"] });
    },
  });

  return {
    agents: agentsQuery.data ?? [],
    isLoading: agentsQuery.isLoading,
    updateAgent: updateAgentMutation.mutateAsync,
  };
}

export function useAgentDetail(id: number) {
  return useQuery({
    queryKey: ["agent", id],
    queryFn: () => api.agents.get(id),
    enabled: !Number.isNaN(id),
  });
}

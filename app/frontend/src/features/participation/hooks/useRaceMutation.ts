import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";

// A change to a race's invitations, attendance, or entry. Its side effects
// reach the race, its invitations, its entry, its running order, and which
// section of R1 it is in, so every race query is fetched again afterwards —
// after a failure too, since the failure usually means the data changed (for
// example, a member answered in the meantime). Nothing is updated before the
// backend answers (ADR-0013).
export function useRaceMutation<Variables>(mutationFn: (variables: Variables) => Promise<unknown>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (variables: Variables) => {
      await mutationFn(variables);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: queryKeys.races }),
  });
}

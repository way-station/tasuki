import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRaceDetail } from "../mapper/toRaceDetail.ts";
import type { RaceCreateBody } from "../types/race.ts";

// Registers a race (R3) and resolves to it, so the screen can open it. The
// list is fetched again; nothing is updated before the backend answers
// (ADR-0013).
export function useCreateRace() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (body: RaceCreateBody) => toRaceDetail(await api.createRace(body)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.races }),
  });
}

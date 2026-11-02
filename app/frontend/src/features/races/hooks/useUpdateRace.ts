import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { toRaceDetail } from "../mapper/toRaceDetail.ts";
import type { RaceCreateBody } from "../types/race.ts";

type UpdateRace = {
  body: RaceCreateBody;
  // The version the edit started from.
  version: number;
};

// Saves an edited race (R3). On version_conflict someone else saved first:
// the race is fetched again, so the screen can offer the latest (ADR-0013).
export function useUpdateRace(raceId: string) {
  const api = useApiClient();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ body, version }: UpdateRace) =>
      toRaceDetail(await api.updateRace(raceId, { ...body, version })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.races }),
    onError: (error) => {
      if (errorCode(error) === "version_conflict") {
        return queryClient.invalidateQueries({ queryKey: queryKeys.race(raceId) });
      }
    },
  });
}

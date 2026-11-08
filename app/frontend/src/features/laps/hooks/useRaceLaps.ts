import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRunnerLapsViews } from "../mapper/toRunnerLapsViews.ts";

// R2: every runner's laps. Only organizers may ask, so the Container passes
// `enabled` for them.
export function useRaceLaps(raceId: string, enabled: boolean) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.raceLaps(raceId),
    queryFn: () => api.listRaceLaps(raceId),
    enabled,
    select: toRunnerLapsViews,
  });
}

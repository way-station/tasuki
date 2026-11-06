import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toMyLapsView } from "../mapper/toMyLapsView.ts";

// L1: the caller's laps for the race. Fails with not_found unless the caller
// is a runner in a confirmed running order of the race.
export function useMyLaps(raceId: string) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.myLaps(raceId),
    queryFn: () => api.getMyLaps(raceId),
    select: toMyLapsView,
  });
}

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRaceDetail } from "../mapper/toRaceDetail.ts";

// One race (R2, R3). A race the caller may not see fails with not_found, the
// same as one that does not exist.
export function useRace(raceId: string) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.race(raceId),
    queryFn: () => api.getRace(raceId),
    select: toRaceDetail,
  });
}

import { useQueries, useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toMemberProfileView } from "../mapper/toMemberProfileView.ts";
import type { RaceResponse, RunningOrderListResponse } from "../types/roster.ts";

// M2. Fails with not_found for an unknown roster member. Each race with laps
// is fetched with its running order for its date and the member's legs, which
// fill in as they arrive; the profile shows without them.
export function useMemberProfile(rosterMemberId: string) {
  const api = useApiClient();
  const detail = useQuery({
    queryKey: queryKeys.rosterMember(rosterMemberId),
    queryFn: () => api.getRosterMember(rosterMemberId),
  });
  const raceIds = detail.data?.laps_by_race.map((race) => race.race_id) ?? [];
  const races = useQueries({
    queries: raceIds.map((raceId) => ({
      queryKey: queryKeys.race(raceId),
      queryFn: () => api.getRace(raceId),
    })),
  });
  const orders = useQueries({
    queries: raceIds.map((raceId) => ({
      queryKey: queryKeys.runningOrders(raceId),
      queryFn: () => api.listRunningOrders(raceId),
    })),
  });

  const racesById = new Map<string, RaceResponse>();
  const ordersById = new Map<string, RunningOrderListResponse>();
  raceIds.forEach((raceId, index) => {
    const race = races[index]?.data;
    if (race !== undefined) {
      racesById.set(raceId, race);
    }
    const order = orders[index]?.data;
    if (order !== undefined) {
      ordersById.set(raceId, order);
    }
  });
  return {
    data:
      detail.data === undefined
        ? undefined
        : toMemberProfileView(detail.data, racesById, ordersById),
    isPending: detail.isPending,
    error: detail.error,
  };
}

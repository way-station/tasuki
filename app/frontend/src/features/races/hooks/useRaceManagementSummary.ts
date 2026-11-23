import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRaceManagementSummary } from "../mapper/toRaceManagementSummary.ts";
import type { RaceManagementSummary } from "../types/race.ts";

// R2's management cards. Only organizers may ask, so the Container passes
// `enabled` for them. `undefined` until all three lists are loaded, and when
// one of them fails: the cards still open their screens.
export function useRaceManagementSummary(
  raceId: string,
  enabled: boolean,
): RaceManagementSummary | undefined {
  const api = useApiClient();
  const invitations = useQuery({
    queryKey: queryKeys.invitations(raceId),
    queryFn: () => api.listInvitations(raceId),
    enabled,
  });
  const entries = useQuery({
    queryKey: queryKeys.entries(raceId),
    queryFn: () => api.listEntries(raceId),
    enabled,
  });
  const orders = useQuery({
    queryKey: queryKeys.runningOrders(raceId),
    queryFn: () => api.listRunningOrders(raceId),
    enabled,
  });
  if (invitations.data === undefined || entries.data === undefined || orders.data === undefined) {
    return undefined;
  }
  return toRaceManagementSummary(invitations.data, entries.data, orders.data);
}

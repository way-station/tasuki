import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toInvitationRows } from "../mapper/toInvitationRows.ts";

// R4: who is invited to the race.
export function useInvitationRows(raceId: string) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.invitations(raceId),
    queryFn: () => api.listInvitations(raceId),
    select: toInvitationRows,
  });
}

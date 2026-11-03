import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toAttendanceRows } from "../mapper/toAttendanceRows.ts";

// R5: the answers of invited members. Shares its data with R4.
export function useAttendanceRows(raceId: string) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.invitations(raceId),
    queryFn: () => api.listInvitations(raceId),
    select: toAttendanceRows,
  });
}

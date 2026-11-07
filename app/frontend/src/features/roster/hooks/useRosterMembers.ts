import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRosterMemberViews } from "../mapper/toRosterMemberViews.ts";

// M1.
export function useRosterMembers() {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.rosterMembers,
    queryFn: () => api.listRosterMembers(),
    select: toRosterMemberViews,
  });
}

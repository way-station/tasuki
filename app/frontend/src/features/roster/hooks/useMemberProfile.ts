import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toMemberProfileView } from "../mapper/toMemberProfileView.ts";

// M2. Fails with not_found for an unknown roster member.
export function useMemberProfile(rosterMemberId: string) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.rosterMember(rosterMemberId),
    queryFn: () => api.getRosterMember(rosterMemberId),
    select: toMemberProfileView,
  });
}

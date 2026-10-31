import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toInviteLink } from "../mapper/toInviteLink.ts";

// Checks an invite link before login (A1). An unusable link fails with
// 404 invite_unusable, which is not retried.
export function useInviteLink(token: string) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.inviteLink(token),
    queryFn: () => api.getInviteLink(token),
    select: toInviteLink,
  });
}

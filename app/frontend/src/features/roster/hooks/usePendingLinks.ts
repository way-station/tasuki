import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toPendingLinkViews } from "../mapper/toPendingLinkViews.ts";

// M3, and the number on the 承認 tab. Only organizers may ask, so the tab
// bar turns it off for members.
export function usePendingLinks(enabled = true) {
  const api = useApiClient();
  return useQuery({
    enabled,
    queryKey: queryKeys.memberLinks,
    queryFn: () => api.listPendingMemberLinks(),
    select: toPendingLinkViews,
  });
}

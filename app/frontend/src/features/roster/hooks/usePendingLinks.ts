import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toPendingLinkViews } from "../mapper/toPendingLinkViews.ts";

// M3.
export function usePendingLinks() {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.memberLinks,
    queryFn: () => api.listPendingMemberLinks(),
    select: toPendingLinkViews,
  });
}

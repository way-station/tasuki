import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRunningOrderView } from "../mapper/toRunningOrderView.ts";

// R2: the running orders the caller may see. Only organizers and members who
// answered attending may ask, so the Container passes `enabled` for them.
export function useRunningOrders(raceId: string, enabled: boolean) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.runningOrders(raceId),
    queryFn: () => api.listRunningOrders(raceId),
    enabled,
    select: (list) => list.running_orders.map(toRunningOrderView),
  });
}

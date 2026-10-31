import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toHealthView } from "../mapper/toHealthView.ts";

export function useHealth() {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.health,
    queryFn: () => api.getHealth(),
    select: toHealthView,
  });
}

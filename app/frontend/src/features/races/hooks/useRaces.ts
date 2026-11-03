import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRaceListView } from "../mapper/toRaceListView.ts";

// The races of R1, laid out for the caller's role.
export function useRaces(isOrganizer: boolean) {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.races,
    queryFn: () => api.listRaces(),
    select: (list) => toRaceListView(list, isOrganizer),
  });
}

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toProfileView } from "../mapper/toProfileView.ts";

// P1.
export function useProfile() {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.profile,
    queryFn: () => api.getMyProfile(),
    select: toProfileView,
  });
}

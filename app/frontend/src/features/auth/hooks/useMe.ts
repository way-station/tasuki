import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toCurrentUser } from "../mapper/toCurrentUser.ts";

// The signed-in user. Without a session it fails with 401, which the query
// cache handles once (#32); the route guard (#113) decides where to go, and
// tells it apart with isUnauthenticated().
export function useMe() {
  const api = useApiClient();
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: () => api.getMe(),
    select: toCurrentUser,
  });
}

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { removeUserDrafts } from "../../../shared/draft/draftStorage.ts";
import type { MeResponse } from "../types/auth.ts";
import { isUnauthenticated } from "./apiErrors.ts";

// Logs out. The backend clears the session cookie, and the whole cache and
// this user's drafts are cleared so nothing of this user stays (ADR-0013).
// The user is read before the request, because a 401 clears the cache
// first. When logging out fails for another reason, the user is still logged
// in and keeps the drafts.
export function useLogout() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.logout(),
    onMutate: () => ({
      userId: queryClient.getQueryData<MeResponse>(queryKeys.me)?.user_id,
    }),
    onSettled: (_data, error, _variables, context) => {
      if (error !== null && !isUnauthenticated(error)) {
        return;
      }
      if (context?.userId !== undefined) {
        removeUserDrafts(context.userId);
      }
      queryClient.clear();
    },
  });
}

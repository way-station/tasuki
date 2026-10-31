import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useApiClient } from "../../../api/useApiClient.ts";

// Logs out. The backend clears the session cookie, and the whole cache is
// cleared so nothing of this user stays (ADR-0013).
export function useLogout() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.logout(),
    onSuccess: () => {
      queryClient.clear();
    },
  });
}

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useApiClient } from "../../../api/useApiClient.ts";

// Exchanges the one-time code from `/auth/complete` for a session. The browser
// stores the session cookie; nothing here reads it (ADR-0014). The signed-in
// user changed, so the whole cache is cleared (ADR-0013).
export function useCreateSession() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (code: string) => api.createSession(code),
    onSuccess: () => {
      queryClient.clear();
    },
  });
}

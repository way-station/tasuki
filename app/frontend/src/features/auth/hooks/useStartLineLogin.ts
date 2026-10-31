import { useMutation } from "@tanstack/react-query";
import { useApiClient } from "../../../api/useApiClient.ts";

// Starts LINE Login and resolves to the LINE authorization URL. The container
// navigates to it, which leaves the app.
export function useStartLineLogin() {
  const api = useApiClient();
  return useMutation({
    mutationFn: (inviteToken?: string) => api.authorizeLine(inviteToken),
  });
}

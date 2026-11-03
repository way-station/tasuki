import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toEntryEditor } from "../mapper/toEntryEditor.ts";

// R6: the entry, and who can be added to it, from the entries and the
// invitations.
export function useEntryEditor(raceId: string) {
  const api = useApiClient();
  const entries = useQuery({
    queryKey: queryKeys.entries(raceId),
    queryFn: () => api.listEntries(raceId),
  });
  const invitations = useQuery({
    queryKey: queryKeys.invitations(raceId),
    queryFn: () => api.listInvitations(raceId),
  });
  return {
    data:
      entries.data === undefined || invitations.data === undefined
        ? undefined
        : toEntryEditor(entries.data, invitations.data),
    isPending: entries.isPending || invitations.isPending,
    error: entries.error ?? invitations.error,
  };
}

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import type { Draft } from "../../../shared/draft/draftStorage.ts";
import { toRunningOrderForm } from "../mapper/runningOrderForm.ts";
import { toRunningOrderEditor } from "../mapper/toRunningOrderEditor.ts";
import type { RunningOrderEditor, RunningOrderForm } from "../types/runningOrder.ts";

type EditorData = {
  editor: RunningOrderEditor;
  // What the form starts from, for the Container's draft (ADR-0013).
  // `undefined` while the race has no entry.
  source: Draft<RunningOrderForm> | undefined;
};

// R7 from the race, its entry, and the saved running order. Before the first
// save the form starts from empty legs, at version 0.
export function useRunningOrderEditor(raceId: string) {
  const api = useApiClient();
  const race = useQuery({
    queryKey: queryKeys.race(raceId),
    queryFn: () => api.getRace(raceId),
  });
  const entries = useQuery({
    queryKey: queryKeys.entries(raceId),
    queryFn: () => api.listEntries(raceId),
  });
  const orders = useQuery({
    queryKey: queryKeys.runningOrders(raceId),
    queryFn: () => api.listRunningOrders(raceId),
  });

  let data: EditorData | undefined;
  if (race.data !== undefined && entries.data !== undefined && orders.data !== undefined) {
    const editor = toRunningOrderEditor(race.data, entries.data, orders.data);
    data = {
      editor,
      source:
        editor.entry === null
          ? undefined
          : {
              value: toRunningOrderForm(editor.legRule, editor.saved),
              // Organizers always get the version, and only organizers edit.
              version: editor.saved?.version ?? 0,
            },
    };
  }
  return {
    data,
    isPending: race.isPending || entries.isPending || orders.isPending,
    error: race.error ?? entries.error ?? orders.error,
  };
}

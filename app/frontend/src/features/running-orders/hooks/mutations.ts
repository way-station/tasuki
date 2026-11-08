import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRunningOrderBody } from "../mapper/runningOrderForm.ts";
import { toRunningOrderView } from "../mapper/toRunningOrderView.ts";
import type { RunningOrderForm } from "../types/runningOrder.ts";

type SaveRunningOrder = {
  entryId: string;
  form: RunningOrderForm;
  // The version the edit started from; 0 before the first save.
  version: number;
};

type ConfirmRunningOrder = {
  entryId: string;
  version: number;
};

// R7. Saving can delete laps and confirming changes what members and R1 show,
// so every race query is fetched again afterwards, after a failure too:
// version_conflict and state_conflict mean someone else changed it. Nothing is
// updated before the backend answers (ADR-0013).
function useRunningOrderMutation<Variables>(
  mutationFn: (variables: Variables) => Promise<ReturnType<typeof toRunningOrderView>>,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSettled: () => queryClient.invalidateQueries({ queryKey: queryKeys.races }),
  });
}

// Fails with version_conflict when someone else saved first, and with
// validation_failed when the legs or runners no longer match the race and the
// entry.
export function useSaveRunningOrder(raceId: string) {
  const api = useApiClient();
  return useRunningOrderMutation(async ({ entryId, form, version }: SaveRunningOrder) =>
    toRunningOrderView(
      await api.saveRunningOrder(
        raceId,
        entryId,
        toRunningOrderBody(form, version === 0 ? undefined : version),
      ),
    ),
  );
}

// Fails with validation_failed when a leg has no runner, state_conflict when it
// is already confirmed, and version_conflict when it changed since it was read.
export function useConfirmRunningOrder(raceId: string) {
  const api = useApiClient();
  return useRunningOrderMutation(async ({ entryId, version }: ConfirmRunningOrder) =>
    toRunningOrderView(await api.confirmRunningOrder(raceId, entryId, version)),
  );
}

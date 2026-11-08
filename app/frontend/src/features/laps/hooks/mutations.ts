import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toLapEditBody, toLapsBody } from "../mapper/lapsForm.ts";
import { toMyLapsView } from "../mapper/toMyLapsView.ts";
import type { LapsForm, MyLapsView } from "../types/lap.ts";
import { LapsFormInvalidError, LapTimeInvalidError } from "./lapsFormInvalid.ts";

type EditLap = {
  lapId: string;
  time: string;
};

// L1. A change of laps also changes the caller's average, so the laps of the
// race and the profile are fetched again afterwards, after a failure too.
// Nothing is updated before the backend answers (ADR-0013).
function useLapMutation<Variables>(
  raceId: string,
  mutationFn: (variables: Variables) => Promise<MyLapsView>,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSettled: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.myLaps(raceId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.raceLaps(raceId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.profile }),
      ]),
  });
}

// Adds the lap times of the form together. An invalid form fails with
// LapsFormInvalidError before anything is sent.
export function useSubmitLaps(raceId: string) {
  const api = useApiClient();
  return useLapMutation(raceId, async (form: LapsForm) => {
    const result = toLapsBody(form);
    if (!result.ok) {
      throw new LapsFormInvalidError(result.errors);
    }
    return toMyLapsView(await api.submitLaps(raceId, result.body.times));
  });
}

// An invalid time fails with LapTimeInvalidError before anything is sent.
export function useEditLap(raceId: string) {
  const api = useApiClient();
  return useLapMutation(raceId, async ({ lapId, time }: EditLap) => {
    const result = toLapEditBody(time);
    if (!result.ok) {
      throw new LapTimeInvalidError(result.error);
    }
    return toMyLapsView(await api.editLap(raceId, lapId, result.body.time));
  });
}

export function useDeleteLap(raceId: string) {
  const api = useApiClient();
  return useLapMutation(raceId, async (lapId: string) =>
    toMyLapsView(await api.deleteLap(raceId, lapId)),
  );
}

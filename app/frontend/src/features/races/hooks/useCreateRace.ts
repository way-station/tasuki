import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRaceBody } from "../mapper/raceForm.ts";
import { toRaceDetail } from "../mapper/toRaceDetail.ts";
import type { RaceForm } from "../types/race.ts";
import { RaceFormInvalidError } from "./raceFormInvalid.ts";

// Registers a race from the form (R3) and resolves to it, so the screen can
// open it. An invalid form fails with RaceFormInvalidError before anything is
// sent. The list is fetched again; nothing is updated before the backend
// answers (ADR-0013).
export function useCreateRace() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (form: RaceForm) => {
      const result = toRaceBody(form);
      if (!result.ok) {
        throw new RaceFormInvalidError(result.errors);
      }
      return toRaceDetail(await api.createRace(result.body));
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.races }),
  });
}

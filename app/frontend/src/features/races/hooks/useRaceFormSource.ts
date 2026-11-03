import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import type { Draft } from "../../../shared/draft/draftStorage.ts";
import { emptyRaceForm, toRaceForm } from "../mapper/raceForm.ts";
import { toRaceDetail } from "../mapper/toRaceDetail.ts";
import type { RaceForm } from "../types/race.ts";

// Registering starts from an empty form. It has no saved version yet.
const newRace: Draft<RaceForm> = { value: emptyRaceForm(), version: 0 };

// What the R3 form starts from, for the Container's draft (ADR-0013): an
// empty form when registering (no raceId), or the saved race with its version
// when editing. `source` is undefined while the race loads.
export function useRaceFormSource(raceId: string | undefined) {
  const api = useApiClient();
  const query = useQuery({
    queryKey: queryKeys.race(raceId ?? ""),
    queryFn: () => api.getRace(raceId ?? ""),
    enabled: raceId !== undefined,
    select: (race): Draft<RaceForm> => ({
      value: toRaceForm(toRaceDetail(race)),
      // Organizers always get the version, and only organizers edit.
      version: race.version ?? 0,
    }),
  });
  if (raceId === undefined) {
    return { source: newRace, error: null };
  }
  return { source: query.data, error: query.error };
}

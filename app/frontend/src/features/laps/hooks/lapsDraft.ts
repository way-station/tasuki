import type { Draft } from "../../../shared/draft/draftStorage.ts";
import { emptyLapsForm } from "../mapper/lapsForm.ts";
import type { LapsForm } from "../types/lap.ts";

// The source of the L1 draft: an empty form. Laps have no version, so `round`
// stands in for it; a new round starts a new empty form after a submit.
// Containers may not import the Mapper, so this reaches them here.
export function lapsDraftSource(round: number): Draft<LapsForm> {
  return { value: emptyLapsForm(), version: round };
}

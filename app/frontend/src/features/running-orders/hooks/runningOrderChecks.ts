import { hasEveryRunner, isSaved } from "../mapper/runningOrderForm.ts";
import { runnersLosingLaps } from "../mapper/runnersLosingLaps.ts";
import type { RunningOrderForm, RunningOrderView } from "../types/runningOrder.ts";

// What R7 needs to know about the form against the saved running order
// (docs/design/screens.md): confirming needs every leg filled and every change
// saved, and saving a confirmed order first names who loses their lap times.
export function checkRunningOrder(saved: RunningOrderView | null, form: RunningOrderForm) {
  const unsaved = !isSaved(form, saved);
  const everyRunner = hasEveryRunner(form);
  return {
    unsaved,
    everyRunner,
    canConfirm: !unsaved && everyRunner && saved?.status === "draft",
    losingLaps: runnersLosingLaps(saved, form),
  };
}

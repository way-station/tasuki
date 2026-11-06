import type {
  LegRule,
  RunningOrderForm,
  RunningOrderSaveBody,
  RunningOrderView,
} from "../types/runningOrder.ts";

// The form R7 starts from: the saved runners, or else empty legs, one per leg
// distance or one leg for time races (docs/design/screens.md).
export function toRunningOrderForm(
  legRule: LegRule,
  saved: RunningOrderView | null,
): RunningOrderForm {
  if (saved !== null) {
    return { runnerIds: saved.legs.map((leg) => leg.runner?.rosterMemberId ?? null) };
  }
  const legCount = legRule.type === "distance" ? legRule.distancesKm.length : 1;
  return { runnerIds: Array.from({ length: legCount }, () => null) };
}

// The request body of `PUT .../running-order`. `version` is left out for the
// first draft. The backend checks the number of legs and the runners
// (`422 validation_failed`).
export function toRunningOrderBody(form: RunningOrderForm, version?: number): RunningOrderSaveBody {
  return {
    ...(version === undefined ? {} : { version }),
    legs: form.runnerIds.map((rosterMemberId, index) => ({
      leg_number: index + 1,
      roster_member_id: rosterMemberId,
    })),
  };
}

// Whether the form is what is saved. Confirming needs every change saved.
export function isSaved(form: RunningOrderForm, saved: RunningOrderView | null): boolean {
  if (saved === null) {
    return false;
  }
  return (
    form.runnerIds.length === saved.legs.length &&
    form.runnerIds.every(
      (runnerId, index) => runnerId === (saved.legs[index]?.runner?.rosterMemberId ?? null),
    )
  );
}

// Whether every leg has a runner, which confirming needs.
export function hasEveryRunner(form: RunningOrderForm): boolean {
  return form.runnerIds.every((runnerId) => runnerId !== null);
}

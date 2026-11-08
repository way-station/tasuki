import type { Runner, RunningOrderForm, RunningOrderView } from "../types/runningOrder.ts";

// The runners who lose their lap times for this race when the form is saved:
// in a confirmed running order, those whose number of legs goes down. Moving a
// runner to another leg keeps their laps (docs/design/api.md). A draft has no
// laps, so nobody loses any.
export function runnersLosingLaps(
  saved: RunningOrderView | null,
  form: RunningOrderForm,
): Runner[] {
  if (saved === null || saved.status !== "confirmed") {
    return [];
  }
  const legCounts = new Map<string, number>();
  for (const runnerId of form.runnerIds) {
    if (runnerId !== null) {
      legCounts.set(runnerId, (legCounts.get(runnerId) ?? 0) + 1);
    }
  }

  const losing: Runner[] = [];
  const seen = new Set<string>();
  for (const leg of saved.legs) {
    const runner = leg.runner;
    if (runner === null || seen.has(runner.rosterMemberId)) {
      continue;
    }
    seen.add(runner.rosterMemberId);
    const savedCount = saved.legs.filter(
      (savedLeg) => savedLeg.runner?.rosterMemberId === runner.rosterMemberId,
    ).length;
    if ((legCounts.get(runner.rosterMemberId) ?? 0) < savedCount) {
      losing.push(runner);
    }
  }
  return losing;
}

import type { RaceLapsResponse, RunnerLapsView } from "../types/lap.ts";

// The laps of each runner, as organizers see them on R2.
export function toRunnerLapsViews(raceLaps: RaceLapsResponse): RunnerLapsView[] {
  return raceLaps.runners.map((runner) => ({
    rosterMemberId: runner.roster_member_id,
    name: runner.name,
    laps: runner.laps.map((lap) => ({ id: lap.id, time: lap.time })),
  }));
}

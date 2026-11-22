import type { RaceLapsResponse, RunnerLapsView } from "../types/lap.ts";
import { averageTime } from "./averageTime.ts";

// The laps of each runner, as organizers see them on R2, with the mean of
// this race's laps. The legs come from the running order later.
export function toRunnerLapsViews(raceLaps: RaceLapsResponse): RunnerLapsView[] {
  return raceLaps.runners.map((runner) => {
    const laps = runner.laps.map((lap) => ({ id: lap.id, time: lap.time }));
    return {
      rosterMemberId: runner.roster_member_id,
      name: runner.name,
      laps,
      averageTime: averageTime(laps),
      legs: [],
    };
  });
}

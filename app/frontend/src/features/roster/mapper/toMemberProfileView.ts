import type { MemberProfileView, RosterMemberDetailResponse } from "../types/roster.ts";

// M2, with the laps grouped by race as the backend returns them.
export function toMemberProfileView(detail: RosterMemberDetailResponse): MemberProfileView {
  return {
    id: detail.id,
    name: detail.name,
    left: detail.left,
    averageTime: detail.average_time,
    lapCount: detail.lap_count,
    lapsByRace: detail.laps_by_race.map((race) => ({
      raceId: race.race_id,
      raceName: race.race_name,
      laps: race.laps.map((lap) => ({ id: lap.id, time: lap.time })),
    })),
  };
}

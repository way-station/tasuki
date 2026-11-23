import type {
  MemberProfileView,
  RaceResponse,
  RosterMemberDetailResponse,
  RunningOrderListResponse,
} from "../types/roster.ts";

// M2, with the laps grouped by race as the backend returns them. The races and
// their running orders, by race ID, add each race's date and the member's legs;
// a race not in them yet goes without.
export function toMemberProfileView(
  detail: RosterMemberDetailResponse,
  races: ReadonlyMap<string, RaceResponse>,
  orders: ReadonlyMap<string, RunningOrderListResponse>,
): MemberProfileView {
  return {
    id: detail.id,
    name: detail.name,
    left: detail.left,
    averageTime: detail.average_time,
    lapCount: detail.lap_count,
    lapsByRace: detail.laps_by_race.map((race) => {
      const raceDate = races.get(race.race_id)?.description.race_date;
      // The MVP allows one entry per race, so a race has at most one order.
      const legs = orders.get(race.race_id)?.running_orders[0]?.legs ?? [];
      return {
        raceId: race.race_id,
        raceName: race.race_name,
        ...(raceDate === undefined ? {} : { raceDate }),
        legNumbers: legs
          .filter((leg) => leg.runner?.roster_member_id === detail.id)
          .map((leg) => leg.leg_number),
        laps: race.laps.map((lap) => ({ id: lap.id, time: lap.time })),
      };
    }),
  };
}

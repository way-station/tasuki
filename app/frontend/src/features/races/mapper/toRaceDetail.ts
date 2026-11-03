import type { RaceDetail, RaceResponse, RaceTypeDetail } from "../types/race.ts";
import { toMyAttendance } from "./toMyAttendance.ts";

export function toRaceDetail(race: RaceResponse): RaceDetail {
  const description = race.description;
  return {
    id: race.id,
    name: race.name,
    version: race.version,
    raceDate: description.race_date,
    deadline: description.deadline,
    text: description.text ?? "",
    detail: toRaceTypeDetail(description),
    myAttendance: toMyAttendance(race.my_attendance),
    runningOrderConfirmed: race.running_order_confirmed,
    canSubmitLaps: race.can_submit_laps,
  };
}

// The backend validates that a description matches its type, so the fields
// of the type are present.
function toRaceTypeDetail(description: RaceResponse["description"]): RaceTypeDetail {
  if (description.type === "distance") {
    return {
      type: "distance",
      legDistancesKm: description.leg_distances_km ?? [],
      targetTotalTime: description.target?.total_time,
    };
  }
  return {
    type: "time",
    timeLimitMinutes: description.time_limit_minutes ?? 0,
    lapDistanceKm: description.lap_distance_km ?? 0,
    targetLaps: description.target?.laps,
  };
}

import type {
  RaceListItem,
  RaceListResponse,
  RaceListView,
  RaceSummaryResponse,
} from "../types/race.ts";
import { toMyAttendance } from "./toMyAttendance.ts";

// Lays out R1 by role (docs/design/screens.md). The backend already returns
// only the races the caller may see, newest registered first; the order is
// kept.
export function toRaceListView(list: RaceListResponse, isOrganizer: boolean): RaceListView {
  const races = list.races.map(toRaceListItem);
  const invitations = races.filter(
    (race) => race.myAttendance !== undefined && race.myAttendance !== "attending",
  );
  if (isOrganizer) {
    return { role: "organizer", races, invitations };
  }
  return {
    role: "member",
    myRaces: races.filter((race) => race.myAttendance === "attending"),
    invitations,
  };
}

function toRaceListItem(race: RaceSummaryResponse): RaceListItem {
  return {
    id: race.id,
    name: race.name,
    raceDate: race.race_date,
    myAttendance: toMyAttendance(race.my_attendance),
    runningOrderConfirmed: race.running_order_confirmed,
  };
}

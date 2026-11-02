import type { MyAttendance, RaceType } from "../types/race.ts";

export const attendanceLabels: Record<MyAttendance, string> = {
  attending: "Attending",
  undecided: "Undecided",
  not_answered: "Not answered",
};

export const raceTypeLabels: Record<RaceType, string> = {
  distance: "Distance",
  time: "Time",
};

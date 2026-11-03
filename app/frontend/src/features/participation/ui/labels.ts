import type { AttendanceAnswer } from "../types/participation.ts";

export const answerLabels: Record<AttendanceAnswer | "not_answered", string> = {
  attending: "Attending",
  not_attending: "Not attending",
  undecided: "Undecided",
  not_answered: "Not answered",
};

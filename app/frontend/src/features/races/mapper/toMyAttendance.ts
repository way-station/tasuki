import type { MyAttendance, RaceResponse } from "../types/race.ts";

// `my_attendance` is omitted when the caller is not invited, and null when
// they are invited but have not answered (docs/design/api.md).
export function toMyAttendance(value: RaceResponse["my_attendance"]): MyAttendance | undefined {
  if (value === undefined) {
    return undefined;
  }
  return value ?? "not_answered";
}

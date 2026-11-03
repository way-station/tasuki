import type { AttendanceRow, InvitationListResponse } from "../types/participation.ts";

// R5: invited members only, with "not_answered" for those without an answer.
export function toAttendanceRows(list: InvitationListResponse): AttendanceRow[] {
  return list.members
    .filter((member) => member.invited)
    .map((member) => ({
      rosterMemberId: member.roster_member_id,
      name: member.name,
      answer: member.attendance ?? "not_answered",
    }));
}

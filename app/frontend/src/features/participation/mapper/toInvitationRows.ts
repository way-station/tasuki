import type { InvitationListResponse, InvitationRow } from "../types/participation.ts";

// R4: every roster member who has not left, in the backend's order.
export function toInvitationRows(list: InvitationListResponse): InvitationRow[] {
  return list.members.map((member) => ({
    rosterMemberId: member.roster_member_id,
    name: member.name,
    invited: member.invited,
  }));
}

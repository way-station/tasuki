import type { RosterMemberListResponse, RosterMemberView } from "../types/roster.ts";

// M1. An invite link can be created only for members who have not left and
// have no user, approved or waiting (docs/design/api.md).
export function toRosterMemberViews(list: RosterMemberListResponse): RosterMemberView[] {
  return list.roster_members.map((member) => ({
    id: member.id,
    name: member.name,
    left: member.left,
    linkStatus: member.link_status,
    inviteLinkExpiresAt: member.invite_link_expires_at ?? undefined,
    canCreateInviteLink: !member.left && member.link_status === "none",
    canUnlink: member.link_status === "approved",
  }));
}

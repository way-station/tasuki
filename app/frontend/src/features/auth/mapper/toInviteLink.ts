import type { InviteLink, InviteLinkResponse } from "../types/auth.ts";

export function toInviteLink(inviteLink: InviteLinkResponse): InviteLink {
  return { teamName: inviteLink.team_name };
}

import type { MemberLinkListResponse, PendingLinkView } from "../types/roster.ts";

// M3.
export function toPendingLinkViews(list: MemberLinkListResponse): PendingLinkView[] {
  return list.links.map((link) => ({
    id: link.id,
    lineDisplayName: link.line_display_name,
    rosterMember: { id: link.roster_member.id, name: link.roster_member.name },
    createdAt: link.created_at,
  }));
}

import type { components } from "../../../api/schema.gen.ts";

// What the backend returns and takes (docs/design/api.md).
export type RosterMemberResponse = components["schemas"]["RosterMember"];
export type RosterMemberListResponse = components["schemas"]["RosterMemberList"];
export type RosterMemberCreateBody = components["schemas"]["RosterMemberCreate"];
export type RosterMemberDetailResponse = components["schemas"]["RosterMemberDetail"];
export type InviteLinkCreatedResponse = components["schemas"]["InviteLinkCreated"];
export type MemberLinkListResponse = components["schemas"]["MemberLinkList"];

// Whether a roster member has a user. A rejected link shows as "none".
export type LinkStatus = RosterMemberResponse["link_status"];

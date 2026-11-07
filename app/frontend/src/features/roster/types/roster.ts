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

// A roster member on M1, with what the organizer may do to them.
export type RosterMemberView = {
  id: string;
  name: string;
  left: boolean;
  linkStatus: LinkStatus;
  // RFC 3339. Only while an invite link is active.
  inviteLinkExpiresAt?: string;
  // Not for members who have left or have a user, approved or waiting: the
  // backend answers 409 state_conflict.
  canCreateInviteLink: boolean;
  canUnlink: boolean;
};

// An invite link just created on M1, to copy and send directly. Creating
// another one makes this one unusable.
export type InviteLinkView = {
  url: string;
  // RFC 3339
  expiresAt: string;
};

// The laps of a member in one race, on M2.
export type RaceLapsOfMemberView = {
  raceId: string;
  raceName: string;
  laps: { id: string; time: string }[];
};

// M2. View only.
export type MemberProfileView = {
  id: string;
  name: string;
  left: boolean;
  // MM:SS
  averageTime: string;
  lapCount: number;
  lapsByRace: RaceLapsOfMemberView[];
};

// A waiting link on M3: someone logged in with LINE through an invite link
// and waits for an organizer.
export type PendingLinkView = {
  id: string;
  lineDisplayName: string;
  rosterMember: { id: string; name: string };
  // RFC 3339: when they logged in.
  createdAt: string;
};

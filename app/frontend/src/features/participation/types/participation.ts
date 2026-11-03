import type { components } from "../../../api/schema.gen.ts";

// What the backend returns (docs/design/api.md).
export type InvitationListResponse = components["schemas"]["InvitationList"];
export type EntryListResponse = components["schemas"]["EntryList"];

// An answer a member can send.
export type AttendanceAnswer = components["schemas"]["AttendanceStatus"];

// A roster member who has not left.
export type MemberRef = {
  rosterMemberId: string;
  name: string;
};

// A row of R4: whether the member is invited.
export type InvitationRow = MemberRef & {
  invited: boolean;
};

// A row of R5: an invited member and their answer.
export type AttendanceRow = MemberRef & {
  answer: AttendanceAnswer | "not_answered";
};

// R6. The MVP allows one entry per race.
export type EntryEditor = {
  entry: { id: string; name: string; members: MemberRef[] } | null;
  // Members who answered attending and are not in the entry yet: only they
  // can be added.
  candidates: MemberRef[];
};

// The caller's answer as R2 shows it: "not attending" never reaches R2.
export type CurrentAnswer = "attending" | "undecided" | "not_answered";

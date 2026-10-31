import type { components } from "../../../api/schema.gen.ts";

// What the backend returns (docs/design/api.md).
export type MeResponse = components["schemas"]["Me"];
export type InviteLinkResponse = components["schemas"]["InviteLinkInfo"];

// The screen a user lands on (docs/design/screens.md): the team's screens, or
// A2 waiting, A3 rejected, A4 cannot use tasuki.
export type EntryScreen = "home" | "waiting" | "rejected" | "unavailable";

// The signed-in user.
export type CurrentUser = {
  // Drafts are kept per user (ADR-0013).
  userId: string;
  displayName: string;
  isOrganizer: boolean;
  // Set only when the user's link to a roster member is approved. Member
  // screens (L1, P1, own attendance) need it, organizers included.
  rosterMember: { id: string; name: string } | null;
  entryScreen: EntryScreen;
};

// The invite link before login (A1). The roster member's name is not shown.
export type InviteLink = {
  teamName: string;
};

// Why LINE Login did not finish, from `/auth/complete?error=<code>`.
// "unavailable" goes to A4; "failed" means LINE Login failed or was cancelled.
export type LoginError = "unavailable" | "failed";

// Where finishing a login at `/auth/complete` stands.
export type LoginState = "pending" | "done" | LoginError;

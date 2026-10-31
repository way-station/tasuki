import type { CurrentUser, EntryScreen, MeResponse } from "../types/auth.ts";

export function toCurrentUser(me: MeResponse): CurrentUser {
  return {
    userId: me.user_id,
    displayName: me.display_name,
    isOrganizer: me.is_organizer,
    rosterMember: me.roster_member,
    entryScreen: toEntryScreen(me),
  };
}

// Organizers use the organizer screens whatever their own link's status is
// (ADR-0006, docs/design/api.md `GET /me`).
function toEntryScreen(me: MeResponse): EntryScreen {
  if (me.is_organizer) {
    return "home";
  }
  switch (me.status) {
    case "approved":
      return "home";
    case "pending":
      return "waiting";
    case "rejected":
      return "rejected";
    case "none":
      return "unavailable";
  }
}

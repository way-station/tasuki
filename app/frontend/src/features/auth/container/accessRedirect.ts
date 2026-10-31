import type { CurrentUser } from "../types/auth.ts";

// Who may open a screen (docs/design/screens.md):
// - "team": every screen of the team, such as R1 and R2
// - "organizer": organizer screens (R3–R7, M1–M3)
// - "linked": member screens that need a roster member (L1, P1), organizers included
export type Access = "team" | "organizer" | "linked";

type Redirect = "/login" | "/waiting" | "/rejected" | "/unavailable" | "/";

// Returns where to send the user instead, or null when they may stay.
export function accessRedirect(
  user: CurrentUser | "unauthenticated",
  access: Access,
): Redirect | null {
  if (user === "unauthenticated") {
    return "/login";
  }
  switch (user.entryScreen) {
    case "waiting":
      return "/waiting";
    case "rejected":
      return "/rejected";
    case "unavailable":
      return "/unavailable";
    case "home":
      break;
  }
  if (access === "organizer" && !user.isOrganizer) {
    return "/";
  }
  if (access === "linked" && user.rosterMember === null) {
    return "/";
  }
  return null;
}

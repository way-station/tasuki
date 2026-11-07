// Every query key, in one place, so a mutation can invalidate every query it
// affects (ADR-0013).
export const queryKeys = {
  health: ["health"],
  me: ["me"],
  inviteLink: (token: string) => ["invite-links", token] as const,
  // A race is under the list's key, so invalidating races refreshes both.
  races: ["races"],
  race: (raceId: string) => ["races", raceId] as const,
  invitations: (raceId: string) => ["races", raceId, "invitations"] as const,
  entries: (raceId: string) => ["races", raceId, "entries"] as const,
  // A member's profile is under the roster's key, so invalidating the roster
  // refreshes both.
  rosterMembers: ["roster-members"],
  rosterMember: (rosterMemberId: string) => ["roster-members", rosterMemberId] as const,
  memberLinks: ["member-links"],
} as const;

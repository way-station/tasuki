import type {
  EntryEditor,
  EntryListResponse,
  InvitationListResponse,
} from "../types/participation.ts";

// R6 from the entries and the invitations. Only members who answered
// attending can be added (docs/design/api.md).
export function toEntryEditor(
  entries: EntryListResponse,
  invitations: InvitationListResponse,
): EntryEditor {
  const first = entries.entries[0];
  const entry =
    first === undefined
      ? null
      : {
          id: first.id,
          name: first.name,
          members: first.members.map((member) => ({
            rosterMemberId: member.roster_member_id,
            name: member.name,
          })),
        };
  const inEntry = new Set(entry?.members.map((member) => member.rosterMemberId));
  const candidates = invitations.members
    .filter((member) => member.attendance === "attending" && !inEntry.has(member.roster_member_id))
    .map((member) => ({ rosterMemberId: member.roster_member_id, name: member.name }));
  return { entry, candidates };
}

import type {
  EntryListResponse,
  InvitationListResponse,
  RaceManagementSummary,
  RunningOrderListResponse,
} from "../types/race.ts";

// R2's management cards from the invitations, the entries, and the running
// orders. Members who are not invited have no answer and are not counted.
export function toRaceManagementSummary(
  invitations: InvitationListResponse,
  entries: EntryListResponse,
  orders: RunningOrderListResponse,
): RaceManagementSummary {
  const invited = invitations.members.filter((member) => member.invited);
  const count = (answer: string | null) =>
    invited.filter((member) => member.attendance === answer).length;
  const entry = entries.entries[0];
  const order =
    entry === undefined
      ? undefined
      : orders.running_orders.find((saved) => saved.entry_id === entry.id);
  return {
    invitedCount: invited.length,
    attendance: {
      attending: count("attending"),
      undecided: count("undecided"),
      notAnswered: count(null),
      notAttending: count("not_attending"),
    },
    entry: entry === undefined ? null : { name: entry.name, memberCount: entry.members.length },
    runningOrder: order?.status ?? null,
  };
}

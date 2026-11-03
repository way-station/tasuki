import type { InvitationRow } from "../types/participation.ts";

type InvitationListViewProps = {
  rows: InvitationRow[];
  busy: boolean;
  onInvite: (row: InvitationRow) => void;
  onCancel: (row: InvitationRow) => void;
};

// R4 (docs/design/screens.md).
export function InvitationListView({ rows, busy, onInvite, onCancel }: InvitationListViewProps) {
  return (
    <section>
      <h1>Invitations</h1>
      {rows.length === 0 ? (
        <p>No roster members.</p>
      ) : (
        <ul>
          {rows.map((row) => (
            <li key={row.rosterMemberId}>
              {row.name} {row.invited ? "Invited" : "Not invited"}{" "}
              {row.invited ? (
                <button type="button" disabled={busy} onClick={() => onCancel(row)}>
                  Cancel {row.name}'s invitation
                </button>
              ) : (
                <button type="button" disabled={busy} onClick={() => onInvite(row)}>
                  Invite {row.name}
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

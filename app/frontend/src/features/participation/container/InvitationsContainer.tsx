import { useParams } from "@tanstack/react-router";
import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useCancelInvitation, useInvite } from "../hooks/mutations.ts";
import { useInvitationRows } from "../hooks/useInvitationRows.ts";
import type { InvitationRow } from "../types/participation.ts";
import { Confirmation } from "../ui/Confirmation.tsx";
import { InvitationListView } from "../ui/InvitationListView.tsx";
import { ParticipationMessage } from "../ui/ParticipationMessage.tsx";

// R4 (docs/design/screens.md).
export function InvitationsContainer() {
  const { raceId } = useParams({ from: "/team/races/$raceId" });
  const rows = useInvitationRows(raceId);
  const invite = useInvite(raceId);
  const cancelInvitation = useCancelInvitation(raceId);
  const { notify } = useNotifications();
  const [cancelling, setCancelling] = useState<InvitationRow | null>(null);

  if (rows.isPending) {
    return <ParticipationMessage title="Loading…" />;
  }
  if (rows.isError) {
    return <ParticipationMessage title="Something went wrong" message="Reload the page." />;
  }

  const inviteMember = async (row: InvitationRow) => {
    try {
      await invite.mutateAsync(row.rosterMemberId);
      notify("success", `${row.name} was invited.`);
    } catch (caught) {
      notify(
        "error",
        errorCode(caught) === "validation_failed"
          ? `${row.name} has left the team.`
          : `${row.name} could not be invited. Try again.`,
      );
    }
  };

  const cancel = async (row: InvitationRow) => {
    try {
      await cancelInvitation.mutateAsync(row.rosterMemberId);
      notify("success", `${row.name}'s invitation was cancelled.`);
    } catch {
      notify("error", `${row.name}'s invitation could not be cancelled. Try again.`);
    } finally {
      setCancelling(null);
    }
  };

  const busy = invite.isPending || cancelInvitation.isPending;
  return (
    <>
      <InvitationListView
        rows={rows.data}
        busy={busy}
        onInvite={(row) => void inviteMember(row)}
        onCancel={setCancelling}
      />
      {cancelling !== null && (
        <Confirmation
          title={`Cancel ${cancelling.name}'s invitation?`}
          messages={[
            `${cancelling.name} is also removed from attendance, the entry, and the running order. If the running order is confirmed, their lap times for this race are deleted and their average time is recalculated.`,
          ]}
          confirmLabel="Cancel invitation"
          busy={busy}
          onConfirm={() => void cancel(cancelling)}
          onCancel={() => setCancelling(null)}
        />
      )}
    </>
  );
}

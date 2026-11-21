import { useParams } from "@tanstack/react-router";
import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useCancelInvitation, useInvite } from "../hooks/mutations.ts";
import { useInvitationRows } from "../hooks/useInvitationRows.ts";
import type { InvitationRow } from "../types/participation.ts";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";
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
    return <ParticipationMessage title="読み込み中…" />;
  }
  if (rows.isError) {
    return (
      <ParticipationMessage
        title="うまくいきませんでした"
        message="ページを読み込み直してください。"
      />
    );
  }

  const inviteMember = async (row: InvitationRow) => {
    try {
      await invite.mutateAsync(row.rosterMemberId);
      notify("success", `${row.name} さんを誘いました。`);
    } catch (caught) {
      notify(
        "error",
        errorCode(caught) === "validation_failed"
          ? `${row.name} さんはもう退会しています。`
          : `${row.name} さんを誘えませんでした。もう一度お試しください。`,
      );
    }
  };

  const cancel = async (row: InvitationRow) => {
    try {
      await cancelInvitation.mutateAsync(row.rosterMemberId);
      notify("success", `${row.name} さんの招待を取り消しました。`);
    } catch {
      notify("error", `${row.name} さんの招待を取り消せませんでした。もう一度お試しください。`);
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
          title={`${cancelling.name} さんの招待を取り消しますか？`}
          messages={[
            "出欠、エントリー、走順からも外れます。走順が確定していれば、この大会のラップタイムは消え、平均タイムは計算し直されます。",
          ]}
          confirmLabel="取り消す"
          busyLabel="取り消し中…"
          tone="danger"
          busy={busy}
          onConfirm={() => void cancel(cancelling)}
          onCancel={() => setCancelling(null)}
        />
      )}
    </>
  );
}

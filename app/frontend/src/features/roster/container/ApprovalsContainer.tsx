import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";
import { useApproveLink, useRejectLink } from "../hooks/mutations.ts";
import { usePendingLinks } from "../hooks/usePendingLinks.ts";
import type { PendingLinkView } from "../types/roster.ts";
import { PendingLinkList } from "../ui/PendingLinkList.tsx";
import { RosterMessage } from "../ui/RosterMessage.tsx";

// M3 (docs/design/screens.md). Rejecting asks first; approving does not.
export function ApprovalsContainer() {
  const links = usePendingLinks();
  const approve = useApproveLink();
  const reject = useRejectLink();
  const { notify } = useNotifications();
  const [rejecting, setRejecting] = useState<PendingLinkView | null>(null);

  if (links.error !== null) {
    return (
      <RosterMessage title="うまくいきませんでした" message="ページを読み込み直してください。" />
    );
  }
  if (links.data === undefined) {
    return <RosterMessage title="読み込み中…" />;
  }

  // Another organizer may have decided first: the list is fetched again.
  const decide = async (link: PendingLinkView, approved: boolean) => {
    try {
      await (approved ? approve : reject).mutateAsync(link.id);
      notify(
        "success",
        approved
          ? `${link.rosterMember.name}（${link.lineDisplayName}）を承認しました。`
          : `${link.rosterMember.name}（${link.lineDisplayName}）を却下しました。理由はアプリの外で伝えてください。`,
      );
    } catch (caught) {
      switch (errorCode(caught)) {
        case "state_conflict":
        case "not_found":
          notify("error", "この紐付けはもう決まっていたので、最新を出しています。");
          return;
        default:
          notify("error", "紐付けを決められませんでした。もう一度お試しください。");
      }
    } finally {
      setRejecting(null);
    }
  };

  return (
    <>
      <PendingLinkList
        links={links.data}
        busy={approve.isPending || reject.isPending}
        onApprove={(link) => void decide(link, true)}
        onReject={setRejecting}
      />
      {rejecting !== null && (
        <Confirmation
          title={`${rejecting.rosterMember.name}（${rejecting.lineDisplayName}）を却下しますか？`}
          messages={["この人は tasuki を使えません。理由はアプリの外で伝えてください。"]}
          confirmLabel="却下する"
          busyLabel="却下中…"
          tone="danger"
          busy={reject.isPending}
          onConfirm={() => void decide(rejecting, false)}
          onCancel={() => setRejecting(null)}
        />
      )}
    </>
  );
}

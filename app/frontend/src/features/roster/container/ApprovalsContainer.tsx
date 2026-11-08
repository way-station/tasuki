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
    return <RosterMessage title="Something went wrong" message="Reload the page." />;
  }
  if (links.data === undefined) {
    return <RosterMessage title="Loading…" />;
  }

  // Another organizer may have decided first: the list is fetched again.
  const decide = async (link: PendingLinkView, approved: boolean) => {
    try {
      await (approved ? approve : reject).mutateAsync(link.id);
      notify(
        "success",
        approved
          ? `${link.lineDisplayName} is approved as ${link.rosterMember.name}.`
          : `${link.lineDisplayName} is rejected. Tell them outside the app.`,
      );
    } catch (caught) {
      switch (errorCode(caught)) {
        case "state_conflict":
        case "not_found":
          notify("error", "This link was already decided. The latest is shown.");
          return;
        default:
          notify("error", "The link could not be decided. Try again.");
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
          title={`Reject ${rejecting.lineDisplayName}?`}
          messages={[
            `${rejecting.lineDisplayName} is not linked to ${rejecting.rosterMember.name}. To join, they need a new invite link. Tell them outside the app.`,
          ]}
          confirmLabel="Reject"
          busy={reject.isPending}
          onConfirm={() => void decide(rejecting, false)}
          onCancel={() => setRejecting(null)}
        />
      )}
    </>
  );
}

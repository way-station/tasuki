import { useId } from "react";
import type { InviteLinkView } from "../types/roster.ts";
import { formatDateTime } from "./labels.ts";

type InviteLinkPanelProps = {
  name: string;
  link: InviteLinkView;
  onCopy: () => void;
  onClose: () => void;
};

// The invite link just created on M1, to copy and send to the member directly.
// The link is shown too, for browsers that cannot copy.
export function InviteLinkPanel({ name, link, onCopy, onClose }: InviteLinkPanelProps) {
  const id = useId();
  return (
    <section aria-label={`Invite link for ${name}`}>
      <h2>Invite link for {name}</h2>
      <label htmlFor={id}>Link</label>
      <input id={id} value={link.url} readOnly />
      <p>
        Valid until <time dateTime={link.expiresAt}>{formatDateTime(link.expiresAt)}</time>. Send it
        to {name} directly. Creating another link makes this one unusable.
      </p>
      <button type="button" onClick={onCopy}>
        Copy link
      </button>
      <button type="button" onClick={onClose}>
        Close
      </button>
    </section>
  );
}

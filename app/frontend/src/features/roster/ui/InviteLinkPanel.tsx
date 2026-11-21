import { useId } from "react";
import {
  fieldHelp,
  input,
  panel,
  primaryButton,
  secondaryButton,
  smallButton,
} from "../../../shared/ui/classes.ts";
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
    <section aria-label={`${name} さんの招待リンク`} className={panel}>
      <h2 className="text-headline font-semibold">{name} さんの招待リンク</h2>
      <div className="flex items-center gap-2">
        <label htmlFor={id} className="sr-only">
          招待リンク
        </label>
        <input id={id} className={input} value={link.url} readOnly />
        <button
          type="button"
          onClick={onCopy}
          className={`${primaryButton} ${smallButton} shrink-0`}
        >
          コピー
        </button>
      </div>
      <p className={fieldHelp}>
        <time dateTime={link.expiresAt}>{formatDateTime(link.expiresAt)}</time>{" "}
        まで使えます。本人に直接送ってください。作り直すと、このリンクは使えなくなります。
      </p>
      <button
        type="button"
        onClick={onClose}
        className={`${secondaryButton} ${smallButton} self-start`}
      >
        閉じる
      </button>
    </section>
  );
}

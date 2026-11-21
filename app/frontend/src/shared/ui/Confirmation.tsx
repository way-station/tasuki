import { useEffect, useId, useRef } from "react";
import {
  blockButton,
  dangerButton,
  primaryButton,
  secondaryButton,
  warningNote,
} from "./classes.ts";

type ConfirmationProps = {
  title: string;
  messages: string[];
  // A warning under the messages, in the warning colors.
  warning?: string;
  confirmLabel: string;
  // What the confirm button says while the request runs, for example 取り消し中….
  busyLabel: string;
  // Danger for actions that cannot be undone or that delete data
  // (docs/design/components.md, Buttons).
  tone: "danger" | "primary";
  busy: boolean;
  onConfirm: () => void;
  onCancel: () => void;
};

// Asks before an action with side effects the user cannot undo
// (docs/design/components.md, Confirmation dialog). The screen behind is
// dimmed and cannot be clicked, and focus moves into the dialog.
export function Confirmation({
  title,
  messages,
  warning,
  confirmLabel,
  busyLabel,
  tone,
  busy,
  onConfirm,
  onCancel,
}: ConfirmationProps) {
  const titleId = useId();
  const dialog = useRef<HTMLDivElement>(null);

  useEffect(() => {
    dialog.current?.focus();
  }, []);

  return (
    <div className="fixed inset-0 z-10 flex items-center justify-center bg-slate-900/40 p-4">
      <div
        ref={dialog}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        className="flex w-full max-w-80 flex-col gap-4 rounded-xl bg-white p-6"
      >
        <h2 id={titleId} className="text-headline font-semibold">
          {title}
        </h2>
        {messages.map((message) => (
          <p key={message} className="text-body">
            {message}
          </p>
        ))}
        {warning !== undefined && <p className={warningNote}>{warning}</p>}
        <div className="flex flex-col gap-2">
          <button
            type="button"
            onClick={onConfirm}
            disabled={busy}
            className={`${tone === "danger" ? dangerButton : primaryButton} ${blockButton}`}
          >
            {busy ? busyLabel : confirmLabel}
          </button>
          <button
            type="button"
            onClick={onCancel}
            disabled={busy}
            className={`${secondaryButton} ${blockButton}`}
          >
            やめる
          </button>
        </div>
      </div>
    </div>
  );
}

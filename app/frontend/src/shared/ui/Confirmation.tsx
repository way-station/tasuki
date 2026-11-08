import { useId } from "react";

type ConfirmationProps = {
  title: string;
  messages: string[];
  confirmLabel: string;
  busy: boolean;
  onConfirm: () => void;
  onCancel: () => void;
};

// Asks before an action with side effects the user cannot undo.
export function Confirmation({
  title,
  messages,
  confirmLabel,
  busy,
  onConfirm,
  onCancel,
}: ConfirmationProps) {
  const titleId = useId();
  return (
    <div role="alertdialog" aria-labelledby={titleId}>
      <h2 id={titleId}>{title}</h2>
      {messages.map((message) => (
        <p key={message}>{message}</p>
      ))}
      <button type="button" onClick={onConfirm} disabled={busy}>
        {confirmLabel}
      </button>
      <button type="button" onClick={onCancel} disabled={busy}>
        Cancel
      </button>
    </div>
  );
}

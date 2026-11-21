import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Confirmation } from "./Confirmation.tsx";

function renderConfirmation(busy: boolean) {
  const onConfirm = vi.fn();
  const onCancel = vi.fn();
  render(
    <Confirmation
      title="Taro さんの招待を取り消しますか？"
      messages={["出欠、エントリー、走順からも外れます。"]}
      warning="ラップタイムが消えます。"
      confirmLabel="取り消す"
      busyLabel="取り消し中…"
      tone="danger"
      busy={busy}
      onConfirm={onConfirm}
      onCancel={onCancel}
    />,
  );
  return { onConfirm, onCancel };
}

describe("Confirmation", () => {
  it("moves focus into the dialog and answers with either button", () => {
    const { onConfirm, onCancel } = renderConfirmation(false);

    const dialog = screen.getByRole("alertdialog", { name: "Taro さんの招待を取り消しますか？" });
    expect(dialog).toHaveFocus();
    expect(dialog).toHaveTextContent("ラップタイムが消えます。");
    fireEvent.click(screen.getByRole("button", { name: "取り消す" }));
    fireEvent.click(screen.getByRole("button", { name: "やめる" }));

    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("says what is happening and disables both buttons while busy", () => {
    renderConfirmation(true);

    expect(screen.getByRole("button", { name: "取り消し中…" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "やめる" })).toBeDisabled();
  });
});

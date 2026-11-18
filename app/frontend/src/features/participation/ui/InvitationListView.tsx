import type { InvitationRow } from "../types/participation.ts";
import {
  card,
  cards,
  content,
  dangerButton,
  group,
  groupHeading,
  primaryButton,
  smallButton,
} from "../../../shared/ui/classes.ts";

type InvitationListViewProps = {
  rows: InvitationRow[];
  busy: boolean;
  onInvite: (row: InvitationRow) => void;
  onCancel: (row: InvitationRow) => void;
};

// R4 (docs/design/screens.md): a card for each roster member who has not left.
export function InvitationListView({ rows, busy, onInvite, onCancel }: InvitationListViewProps) {
  return (
    <section className={content}>
      <h1 className="text-title1 font-bold">招待</h1>
      <section aria-label="名簿のメンバー" className={group}>
        <h2 className={groupHeading}>名簿のメンバー</h2>
        {rows.length === 0 ? (
          <p className="text-subheadline text-slate-600">まだ名簿にメンバーはいません</p>
        ) : (
          <ul className={cards}>
            {rows.map((row) => (
              <li key={row.rosterMemberId} className={card}>
                <span className="flex min-w-0 grow flex-col gap-0.5">
                  <span className="text-headline font-semibold">{row.name}</span>
                  <span className="text-subheadline text-slate-600">
                    {row.invited ? "招待済み" : "まだ誘っていません"}
                  </span>
                </span>
                {row.invited ? (
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => onCancel(row)}
                    aria-label={`${row.name} さんの招待を取り消す`}
                    className={`${dangerButton} ${smallButton} shrink-0`}
                  >
                    取り消す
                  </button>
                ) : (
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => onInvite(row)}
                    aria-label={`${row.name} さんを誘う`}
                    className={`${primaryButton} ${smallButton} shrink-0`}
                  >
                    誘う
                  </button>
                )}
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  );
}

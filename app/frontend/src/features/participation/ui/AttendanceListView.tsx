import type { AttendanceRow } from "../types/participation.ts";
import { answerColors, answerLabels } from "./labels.ts";
import { card, cards, content, group, groupHeading, label } from "../../../shared/ui/classes.ts";

type Answer = AttendanceRow["answer"];

// The order of the groups: who runs first, who never runs last.
const order: Answer[] = ["attending", "undecided", "not_answered", "not_attending"];

type AttendanceListViewProps = {
  rows: AttendanceRow[];
};

// R5 (docs/design/screens.md). Invited members grouped by their answer.
// View only: organizers cannot answer for a member.
export function AttendanceListView({ rows }: AttendanceListViewProps) {
  return (
    <section className={content}>
      <h1 className="text-title1 font-bold">出欠</h1>
      {rows.length === 0 ? (
        <p className="text-subheadline text-slate-600">まだ誰も招待していません</p>
      ) : (
        order.map((answer) => {
          const members = rows.filter((row) => row.answer === answer);
          if (members.length === 0) {
            return null;
          }
          const heading = `${answerLabels[answer]}（${members.length}）`;
          return (
            <section key={answer} aria-label={answerLabels[answer]} className={group}>
              <h2 className={groupHeading}>{heading}</h2>
              <ul className={cards}>
                {members.map((row) => (
                  <li key={row.rosterMemberId} className={card}>
                    <span className="min-w-0 grow text-headline font-semibold">{row.name}</span>
                    <span className={`${label} ${answerColors[answer]}`}>
                      {answerLabels[answer]}
                    </span>
                  </li>
                ))}
              </ul>
            </section>
          );
        })
      )}
    </section>
  );
}

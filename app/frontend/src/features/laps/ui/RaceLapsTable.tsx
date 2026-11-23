import {
  card,
  cards,
  cardStack,
  group,
  groupHeading,
  numbers,
} from "../../../shared/ui/classes.ts";
import type { RunnerLapsView } from "../types/lap.ts";
import { legsText } from "./legsText.ts";

type RaceLapsTableProps = {
  runners: RunnerLapsView[];
};

// The laps of every runner on R2, for organizers. View only. One card per
// runner with their legs, the mean of their laps, and a table of laps per km
// (wireframes/r2.html, after Strava's splits).
export function RaceLapsTable({ runners }: RaceLapsTableProps) {
  return (
    <section aria-label="走者のラップタイム" className={group}>
      <h2 className={groupHeading}>走者のラップタイム</h2>
      <ul className={cards}>
        {runners.map((runner) =>
          runner.laps.length === 0 ? (
            <li key={runner.rosterMemberId} className={card}>
              <span className="flex min-w-0 grow flex-col gap-0.5">
                <span className="text-headline font-semibold">{runner.name}</span>
                <span className="text-subheadline text-slate-600">
                  {[
                    ...(runner.legs.length > 0 ? [legsText(runner.legs, false)] : []),
                    "まだ入力されていません",
                  ].join("・")}
                </span>
              </span>
            </li>
          ) : (
            <li key={runner.rosterMemberId} className={`${card} ${cardStack}`}>
              <div className="flex items-start gap-3">
                <span className="flex min-w-0 grow flex-col gap-0.5">
                  <span className="text-headline font-semibold">{runner.name}</span>
                  {runner.legs.length > 0 && (
                    <span className="text-subheadline text-slate-600">
                      {legsText(runner.legs, false)}
                    </span>
                  )}
                </span>
                {runner.averageTime !== null && (
                  <span className={`shrink-0 text-subheadline text-slate-600 ${numbers}`}>
                    平均 {runner.averageTime}
                  </span>
                )}
              </div>
              <table className={`w-full border-collapse ${numbers}`}>
                <thead>
                  <tr>
                    <th
                      scope="col"
                      className="pb-1 text-left text-footnote font-normal text-slate-600"
                    >
                      km
                    </th>
                    <th
                      scope="col"
                      className="pb-1 text-right text-footnote font-normal text-slate-600"
                    >
                      ラップ
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {runner.laps.map((lap, index) => (
                    <tr key={lap.id}>
                      <td className="border-t border-slate-200 py-2">{index + 1}</td>
                      <td className="border-t border-slate-200 py-2 text-right">{lap.time}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </li>
          ),
        )}
      </ul>
    </section>
  );
}

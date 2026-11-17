import { card, cardStack, numbers } from "../../../shared/ui/classes.ts";
import type { AverageView } from "../types/lap.ts";

type AverageTextProps = {
  average: AverageView;
};

// The average time over all of a member's laps, and how many there are, as
// two cards of large numbers (P1 wireframe, after Strava's Workout Stats).
export function AverageText({ average }: AverageTextProps) {
  return (
    <dl className="grid grid-cols-2 gap-2">
      <div className={`${card} ${cardStack}`}>
        <dt className="text-subheadline text-slate-600">平均タイム（1km）</dt>
        <dd className={`text-title2 font-bold ${numbers}`}>{average.averageTime}</dd>
      </div>
      <div className={`${card} ${cardStack}`}>
        <dt className="text-subheadline text-slate-600">ラップ数</dt>
        <dd className={`text-title2 font-bold ${numbers}`}>{average.lapCount}</dd>
      </div>
    </dl>
  );
}

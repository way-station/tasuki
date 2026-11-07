import type { AverageView } from "../types/lap.ts";

type AverageTextProps = {
  average: AverageView;
};

// The average time over all of a member's laps, and how many there are.
export function AverageText({ average }: AverageTextProps) {
  return (
    <dl>
      <dt>Average time</dt>
      <dd>{average.averageTime} /km</dd>
      <dt>Laps</dt>
      <dd>{average.lapCount}</dd>
    </dl>
  );
}

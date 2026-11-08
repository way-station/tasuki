import type { RunnerLapsView } from "../types/lap.ts";

type RaceLapsTableProps = {
  runners: RunnerLapsView[];
};

// The laps of every runner on R2, for organizers. View only.
export function RaceLapsTable({ runners }: RaceLapsTableProps) {
  return (
    <section aria-label="Lap times">
      <h2>Lap times</h2>
      <table>
        <thead>
          <tr>
            <th scope="col">Runner</th>
            <th scope="col">Laps (per km)</th>
          </tr>
        </thead>
        <tbody>
          {runners.map((runner) => (
            <tr key={runner.rosterMemberId}>
              <td>{runner.name}</td>
              <td>
                {runner.laps.length === 0
                  ? "None yet"
                  : runner.laps.map((lap) => lap.time).join(", ")}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

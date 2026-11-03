import type { AttendanceRow } from "../types/participation.ts";
import { answerLabels } from "./labels.ts";

type AttendanceListViewProps = {
  rows: AttendanceRow[];
};

// R5 (docs/design/screens.md). View only: organizers cannot answer for a member.
export function AttendanceListView({ rows }: AttendanceListViewProps) {
  return (
    <section>
      <h1>Attendance</h1>
      {rows.length === 0 ? (
        <p>Nobody is invited yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th scope="col">Member</th>
              <th scope="col">Answer</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.rosterMemberId}>
                <td>{row.name}</td>
                <td>{answerLabels[row.answer]}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

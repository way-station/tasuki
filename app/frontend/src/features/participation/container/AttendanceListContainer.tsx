import { useParams } from "@tanstack/react-router";
import { useAttendanceRows } from "../hooks/useAttendanceRows.ts";
import { AttendanceListView } from "../ui/AttendanceListView.tsx";
import { ParticipationMessage } from "../ui/ParticipationMessage.tsx";

// R5 (docs/design/screens.md).
export function AttendanceListContainer() {
  const { raceId } = useParams({ from: "/team/races/$raceId" });
  const rows = useAttendanceRows(raceId);

  if (rows.isPending) {
    return <ParticipationMessage title="Loading…" />;
  }
  if (rows.isError) {
    return <ParticipationMessage title="Something went wrong" message="Reload the page." />;
  }
  return <AttendanceListView rows={rows.data} />;
}

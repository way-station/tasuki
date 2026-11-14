import { useNavigate, useParams } from "@tanstack/react-router";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useMe } from "../../auth/hooks/useMe.ts";
import { RaceLapsSectionContainer } from "../../laps/container/RaceLapsSectionContainer.tsx";
import { AttendanceAnswerContainer } from "../../participation/container/AttendanceAnswerContainer.tsx";
import { RunningOrderSectionContainer } from "../../running-orders/container/RunningOrderSectionContainer.tsx";
import { useRace } from "../hooks/useRace.ts";
import { RaceDetailView, type RaceSection } from "../ui/RaceDetailView.tsx";
import { RaceMessage } from "../ui/RaceMessage.tsx";

const sectionPaths = {
  edit: "/races/$raceId/edit",
  invitations: "/races/$raceId/invitations",
  attendance: "/races/$raceId/attendance",
  entry: "/races/$raceId/entry",
  "running-order": "/races/$raceId/running-order",
  laps: "/races/$raceId/laps",
} as const satisfies Record<RaceSection, string>;

// R2 (docs/design/screens.md).
export function RaceDetailContainer() {
  const { raceId } = useParams({ from: "/team/races/$raceId" });
  const me = useMe();
  const race = useRace(raceId);
  const navigate = useNavigate();

  if (race.isPending) {
    return <RaceMessage title="Loading…" />;
  }
  if (race.isError) {
    // A race the user may not see is not_found too (docs/design/api.md).
    if (errorCode(race.error) === "not_found") {
      return <RaceMessage title="Race not found" />;
    }
    return <RaceMessage title="Something went wrong" message="Reload the page." />;
  }
  const open = (section: RaceSection) =>
    void navigate({ to: sectionPaths[section], params: { raceId } });
  const myAttendance = race.data.myAttendance;
  const isOrganizer = me.data?.isOrganizer ?? false;
  return (
    <RaceDetailView
      race={race.data}
      isOrganizer={isOrganizer}
      onOpen={open}
      attendance={
        // "Not attending" cannot be changed (docs/design/api.md), so there is
        // nothing to answer.
        myAttendance === undefined || myAttendance === "not_attending" ? undefined : (
          <AttendanceAnswerContainer raceId={raceId} current={myAttendance} />
        )
      }
      runningOrder={
        <RunningOrderSectionContainer
          raceId={raceId}
          canSee={isOrganizer || myAttendance === "attending"}
        />
      }
      laps={<RaceLapsSectionContainer raceId={raceId} canSee={isOrganizer} />}
    />
  );
}

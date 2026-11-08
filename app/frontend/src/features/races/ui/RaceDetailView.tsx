import type { ReactNode } from "react";
import type { RaceDetail } from "../types/race.ts";
import { attendanceLabels, raceTypeLabels } from "./labels.ts";

// The screens R2 links to (docs/design/screens.md).
export type RaceSection =
  | "edit"
  | "invitations"
  | "attendance"
  | "entry"
  | "running-order"
  | "laps";

const organizerSections: { section: RaceSection; label: string }[] = [
  { section: "edit", label: "Edit race" },
  { section: "invitations", label: "Invitations" },
  { section: "attendance", label: "Attendance" },
  { section: "entry", label: "Entry" },
  { section: "running-order", label: "Running order" },
];

type RaceDetailViewProps = {
  race: RaceDetail;
  isOrganizer: boolean;
  onOpen: (section: RaceSection) => void;
  // Answering attendance, from the participation feature. Shown instead of the
  // answer alone when given.
  attendance?: ReactNode;
  // The runners' lap times for organizers, from the laps feature.
  laps?: ReactNode;
};

// R2 (docs/design/screens.md): the race and its description. Attendance, the
// running order, and lap times come with their features.
export function RaceDetailView({
  race,
  isOrganizer,
  onOpen,
  attendance,
  laps,
}: RaceDetailViewProps) {
  const detail = race.detail;
  return (
    <section>
      <h1>{race.name}</h1>
      <dl>
        <dt>Type</dt>
        <dd>{raceTypeLabels[detail.type]}</dd>
        <dt>Race date</dt>
        <dd>
          <time dateTime={race.raceDate}>{race.raceDate}</time>
        </dd>
        {race.deadline !== undefined && (
          <>
            <dt>Deadline</dt>
            <dd>
              <time dateTime={race.deadline}>{formatDateTime(race.deadline)}</time>
            </dd>
          </>
        )}
        {detail.type === "distance" ? (
          <>
            <dt>Legs</dt>
            <dd>
              <ol>
                {detail.legDistancesKm.map((km, index) => (
                  <li key={index}>{km} km</li>
                ))}
              </ol>
            </dd>
            {detail.targetTotalTime !== undefined && (
              <>
                <dt>Target time</dt>
                <dd>{detail.targetTotalTime}</dd>
              </>
            )}
          </>
        ) : (
          <>
            <dt>Time limit</dt>
            <dd>{detail.timeLimitMinutes} minutes</dd>
            <dt>Lap distance</dt>
            <dd>{detail.lapDistanceKm} km</dd>
            {detail.targetLaps !== undefined && (
              <>
                <dt>Target laps</dt>
                <dd>{detail.targetLaps}</dd>
              </>
            )}
          </>
        )}
        {race.text !== "" && (
          <>
            <dt>Notes</dt>
            <dd>{race.text}</dd>
          </>
        )}
        {race.myAttendance !== undefined && attendance === undefined && (
          <>
            <dt>My attendance</dt>
            <dd>{attendanceLabels[race.myAttendance]}</dd>
          </>
        )}
        {race.runningOrderConfirmed !== undefined && (
          <>
            <dt>Running order</dt>
            <dd>{race.runningOrderConfirmed ? "Confirmed" : "Not confirmed yet"}</dd>
          </>
        )}
      </dl>
      {attendance}
      {laps}
      {(isOrganizer || race.canSubmitLaps) && (
        <nav aria-label="Race">
          <ul>
            {isOrganizer &&
              organizerSections.map(({ section, label }) => (
                <li key={section}>
                  <button type="button" onClick={() => onOpen(section)}>
                    {label}
                  </button>
                </li>
              ))}
            {race.canSubmitLaps && (
              <li>
                <button type="button" onClick={() => onOpen("laps")}>
                  My lap times
                </button>
              </li>
            )}
          </ul>
        </nav>
      )}
    </section>
  );
}

function formatDateTime(rfc3339: string): string {
  return new Date(rfc3339).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

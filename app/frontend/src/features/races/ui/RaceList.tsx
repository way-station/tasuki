import type { RaceListItem, RaceListView } from "../types/race.ts";
import { attendanceLabels } from "./labels.ts";

type RaceListProps = {
  view: RaceListView;
  onOpen: (raceId: string) => void;
  onRegister: () => void;
};

// R1 (docs/design/screens.md).
export function RaceList({ view, onOpen, onRegister }: RaceListProps) {
  return (
    <section>
      <h1>Races</h1>
      {view.role === "organizer" ? (
        <>
          <button type="button" onClick={onRegister}>
            Register race
          </button>
          <RaceSection title="All races" races={view.races} onOpen={onOpen} showConfirmed />
        </>
      ) : (
        <RaceSection title="My races" races={view.myRaces} onOpen={onOpen} showConfirmed />
      )}
      {(view.role === "member" || view.invitations.length > 0) && (
        <RaceSection title="Invitations" races={view.invitations} onOpen={onOpen} showAnswer />
      )}
    </section>
  );
}

type RaceSectionProps = {
  title: string;
  races: RaceListItem[];
  onOpen: (raceId: string) => void;
  showConfirmed?: boolean;
  showAnswer?: boolean;
};

function RaceSection({ title, races, onOpen, showConfirmed, showAnswer }: RaceSectionProps) {
  return (
    <section aria-label={title}>
      <h2>{title}</h2>
      {races.length === 0 ? (
        <p>No races.</p>
      ) : (
        <ul>
          {races.map((race) => (
            <li key={race.id}>
              <button type="button" onClick={() => onOpen(race.id)}>
                {race.name}
              </button>{" "}
              <time dateTime={race.raceDate}>{race.raceDate}</time>
              {showConfirmed && race.runningOrderConfirmed && <span> Running order confirmed</span>}
              {showAnswer && race.myAttendance !== undefined && (
                <span> {attendanceLabels[race.myAttendance]}</span>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

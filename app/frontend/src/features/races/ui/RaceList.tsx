import {
  card,
  cards,
  confirmedLabel,
  content,
  group,
  groupHeading,
  label,
  neutralLabel,
  primaryButton,
  smallButton,
  waitingLabel,
} from "../../../shared/ui/classes.ts";
import type { MyAttendance, RaceListItem, RaceListView } from "../types/race.ts";
import { attendanceLabels, formatDate } from "./labels.ts";

type RaceListProps = {
  view: RaceListView;
  onOpen: (raceId: string) => void;
  onRegister: () => void;
};

// R1 (docs/design/screens.md, wireframes/r1.html).
export function RaceList({ view, onOpen, onRegister }: RaceListProps) {
  return (
    <section>
      <header className="flex items-end gap-3 border-b border-slate-200 bg-white p-4">
        <h1 className="grow text-large-title font-bold">大会</h1>
        {view.role === "organizer" && (
          <button type="button" onClick={onRegister} className={`${primaryButton} ${smallButton}`}>
            大会を登録
          </button>
        )}
      </header>
      <div className={content}>
        {view.role === "organizer" ? (
          <RaceSection title="すべての大会" races={view.races} onOpen={onOpen} showConfirmed />
        ) : (
          <RaceSection title="参加する大会" races={view.myRaces} onOpen={onOpen} showConfirmed />
        )}
        {(view.role === "member" || view.invitations.length > 0) && (
          <RaceSection title="招待" races={view.invitations} onOpen={onOpen} showAnswer />
        )}
      </div>
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
    <section aria-label={title} className={group}>
      <h2 className={groupHeading}>{title}</h2>
      {races.length === 0 ? (
        <p className="text-subheadline text-slate-600">まだありません</p>
      ) : (
        <ul className={cards}>
          {races.map((race) => (
            <li key={race.id}>
              <button
                type="button"
                onClick={() => onOpen(race.id)}
                className={`${card} w-full text-left`}
              >
                <span className="flex min-w-0 grow flex-col gap-0.5">
                  <span className="text-headline font-semibold">{race.name}</span>
                  <time dateTime={race.raceDate} className="text-subheadline text-slate-600">
                    {formatDate(race.raceDate)}
                  </time>
                </span>
                {showConfirmed && race.runningOrderConfirmed && (
                  <span className={`${label} ${confirmedLabel}`}>走順確定</span>
                )}
                {showAnswer && race.myAttendance !== undefined && (
                  <span className={`${label} ${answerLabelColors[race.myAttendance]}`}>
                    {attendanceLabels[race.myAttendance]}
                  </span>
                )}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

const answerLabelColors: Record<MyAttendance, string> = {
  attending: confirmedLabel,
  undecided: neutralLabel,
  not_attending: neutralLabel,
  not_answered: waitingLabel,
};

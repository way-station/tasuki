import type { ReactNode } from "react";
import {
  blockButton,
  card,
  cards,
  confirmedLabel,
  content,
  group,
  groupHeading,
  label,
  numbers,
  panel,
  primaryButton,
} from "../../../shared/ui/classes.ts";
import type { RaceDetail, RaceManagementSummary } from "../types/race.ts";
import { attendanceLabels, formatDate, formatDateTime, raceTypeLabels } from "./labels.ts";

// The screens R2 links to (docs/design/screens.md).
export type RaceSection =
  | "edit"
  | "invitations"
  | "attendance"
  | "entry"
  | "running-order"
  | "laps";

const organizerSections: { section: RaceSection; label: string }[] = [
  { section: "edit", label: "大会を編集" },
  { section: "invitations", label: "招待" },
  { section: "attendance", label: "出欠" },
  { section: "entry", label: "エントリー" },
  { section: "running-order", label: "走順" },
];

type RaceDetailViewProps = {
  race: RaceDetail;
  isOrganizer: boolean;
  // Organizers only. `undefined` while loading: the cards show no second line.
  summary?: RaceManagementSummary;
  onOpen: (section: RaceSection) => void;
  // Answering attendance, from the participation feature. Shown instead of the
  // answer alone when given.
  attendance?: ReactNode;
  // The running order and prediction, from the running-orders feature.
  runningOrder?: ReactNode;
  // The runners' lap times for organizers, from the laps feature.
  laps?: ReactNode;
};

// R2 (docs/design/screens.md, wireframes/r2.html): the race and its
// description. Attendance, the running order, and lap times come with their
// features. Organizers answer attendance under the description; members after
// the running order.
export function RaceDetailView({
  race,
  isOrganizer,
  summary,
  onOpen,
  attendance,
  runningOrder,
  laps,
}: RaceDetailViewProps) {
  const myLaps = race.canSubmitLaps && (
    <button
      type="button"
      onClick={() => onOpen("laps")}
      className={`${primaryButton} ${blockButton}`}
    >
      自分のラップタイム
    </button>
  );
  return (
    <section className={content}>
      <div className={group}>
        <h1 className="text-title1 font-bold">{race.name}</h1>
        {race.runningOrderConfirmed === true && (
          <p>
            <span className={`${label} ${confirmedLabel}`}>走順確定</span>
          </p>
        )}
      </div>
      <RaceFacts race={race} showAnswer={attendance === undefined} />
      {isOrganizer ? (
        <>
          {attendance}
          <nav aria-label="大会の管理" className={group}>
            <h2 className={groupHeading}>大会の管理</h2>
            <ul className={cards}>
              {organizerSections.map(({ section, label: sectionLabel }) => {
                const state = summary === undefined ? undefined : sectionState(section, summary);
                return (
                  <li key={section}>
                    <button
                      type="button"
                      onClick={() => onOpen(section)}
                      className={`${card} w-full text-left`}
                    >
                      <span className="flex min-w-0 grow flex-col gap-0.5">
                        <span className="text-headline font-semibold">{sectionLabel}</span>
                        {state !== undefined && (
                          <span className={`text-subheadline text-slate-600 ${numbers}`}>
                            {state}
                          </span>
                        )}
                      </span>
                      <span aria-hidden="true" className="text-slate-600">
                        ›
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          </nav>
          {runningOrder}
          {myLaps}
          {laps}
        </>
      ) : (
        <>
          {runningOrder}
          {myLaps}
          {attendance}
        </>
      )}
    </section>
  );
}

type RaceFactsProps = {
  race: RaceDetail;
  // The answer alone, when the attendance feature does not show its form.
  showAnswer: boolean;
};

// The description in a panel, in the order of the wireframe.
function RaceFacts({ race, showAnswer }: RaceFactsProps) {
  const detail = race.detail;
  const facts: { term: string; value: ReactNode }[] = [
    {
      term: "種別",
      value:
        detail.type === "distance"
          ? `${raceTypeLabels.distance}・${detail.legDistancesKm.length} 区`
          : raceTypeLabels.time,
    },
    { term: "開催日", value: <time dateTime={race.raceDate}>{formatDate(race.raceDate)}</time> },
  ];
  if (detail.type === "distance") {
    facts.push({
      term: "区の距離",
      value: detail.legDistancesKm.map((km) => `${km} km`).join("・"),
    });
    if (detail.targetTotalTime !== undefined) {
      facts.push({ term: "目標タイム", value: detail.targetTotalTime });
    }
  } else {
    facts.push(
      { term: "制限時間", value: `${detail.timeLimitMinutes} 分` },
      { term: "1周の距離", value: `${detail.lapDistanceKm} km` },
    );
    if (detail.targetLaps !== undefined) {
      facts.push({ term: "目標周回数", value: `${detail.targetLaps} 周` });
    }
  }
  if (race.deadline !== undefined) {
    facts.push({
      term: "締め切り",
      value: <time dateTime={race.deadline}>{formatDateTime(race.deadline)}</time>,
    });
  }
  if (showAnswer && race.myAttendance !== undefined) {
    facts.push({ term: "あなたの出欠", value: attendanceLabels[race.myAttendance] });
  }

  return (
    <section aria-label="説明" className={panel}>
      <dl className="grid grid-cols-[max-content_minmax(0,1fr)] gap-x-4 gap-y-2">
        {facts.map(({ term, value }) => (
          <div key={term} className="contents">
            <dt className="text-subheadline text-slate-600">{term}</dt>
            <dd className={numbers}>{value}</dd>
          </div>
        ))}
      </dl>
      {race.text !== "" && <p className="text-body whitespace-pre-wrap">{race.text}</p>}
    </section>
  );
}

// The second line of a management card (wireframes/r2.html). Editing the
// race has none.
function sectionState(section: RaceSection, summary: RaceManagementSummary): string | undefined {
  const { attending, undecided, notAnswered, notAttending } = summary.attendance;
  switch (section) {
    case "invitations":
      return summary.invitedCount === 0
        ? "まだ誰も招待していません"
        : `${summary.invitedCount} 人を招待`;
    case "attendance":
      return summary.invitedCount === 0
        ? "まだ誰も招待していません"
        : `参加 ${attending}・保留 ${undecided}・未回答 ${notAnswered}・不参加 ${notAttending}`;
    case "entry":
      return summary.entry === null
        ? "まだありません"
        : `${summary.entry.name}・${summary.entry.memberCount} 人`;
    case "running-order":
      if (summary.runningOrder === null) {
        return "まだありません";
      }
      return summary.runningOrder === "confirmed" ? "確定" : "下書き";
    default:
      return undefined;
  }
}

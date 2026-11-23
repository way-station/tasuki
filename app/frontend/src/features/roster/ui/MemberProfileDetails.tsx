import {
  card,
  cards,
  cardStack,
  content,
  group,
  groupHeading,
  label,
  neutralLabel,
  numbers,
} from "../../../shared/ui/classes.ts";
import { formatDate } from "../../races/ui/labels.ts";
import type { MemberProfileView } from "../types/roster.ts";

type MemberProfileDetailsProps = {
  profile: MemberProfileView;
};

// M2 (docs/design/screens.md, wireframes/m2.html). View only.
export function MemberProfileDetails({ profile }: MemberProfileDetailsProps) {
  return (
    <section className={content}>
      <div className={group}>
        <h1 className="text-title1 font-bold">{profile.name}</h1>
        {profile.left && (
          <p>
            <span className={`${label} ${neutralLabel}`}>退会</span>
          </p>
        )}
      </div>
      <dl className="grid grid-cols-2 gap-2">
        <div className={`${card} ${cardStack}`}>
          <dt className="text-subheadline text-slate-600">平均タイム（1km）</dt>
          <dd className={`text-title2 font-bold ${numbers}`}>{profile.averageTime}</dd>
        </div>
        <div className={`${card} ${cardStack}`}>
          <dt className="text-subheadline text-slate-600">ラップ数</dt>
          <dd className={`text-title2 font-bold ${numbers}`}>{profile.lapCount}</dd>
        </div>
      </dl>
      <section aria-label="大会ごとのラップタイム" className={group}>
        <h2 className={groupHeading}>大会ごとのラップタイム</h2>
        {profile.lapsByRace.length === 0 ? (
          <p className="text-subheadline text-slate-600">まだラップタイムはありません</p>
        ) : (
          <ul className={cards}>
            {profile.lapsByRace.map((race) => (
              <li key={race.raceId} className={`${card} ${cardStack}`}>
                <span className="flex flex-col gap-0.5">
                  <span className="text-headline font-semibold">{race.raceName}</span>
                  {(race.raceDate !== undefined || race.legNumbers.length > 0) && (
                    <span className={`text-subheadline text-slate-600 ${numbers}`}>
                      {raceLine(race.raceDate, race.legNumbers)}
                    </span>
                  )}
                </span>
                <table className={`w-full border-collapse ${numbers}`}>
                  <thead>
                    <tr>
                      <th
                        scope="col"
                        className="pb-1 text-left text-footnote font-normal text-slate-600"
                      >
                        km
                      </th>
                      <th
                        scope="col"
                        className="pb-1 text-right text-footnote font-normal text-slate-600"
                      >
                        ラップ
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {race.laps.map((lap, index) => (
                      <tr key={lap.id}>
                        <td className="border-t border-slate-200 py-2">{index + 1}</td>
                        <td className="border-t border-slate-200 py-2 text-right">{lap.time}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  );
}

// 2026年12月20日（日）・1区・4区 (wireframes/m2.html).
function raceLine(raceDate: string | undefined, legNumbers: number[]): string {
  const parts = legNumbers.map((legNumber) => `${legNumber}区`);
  if (raceDate !== undefined) {
    parts.unshift(formatDate(raceDate));
  }
  return parts.join("・");
}

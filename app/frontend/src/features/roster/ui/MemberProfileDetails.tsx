import type { MemberProfileView } from "../types/roster.ts";

type MemberProfileDetailsProps = {
  profile: MemberProfileView;
};

// M2 (docs/design/screens.md). View only.
export function MemberProfileDetails({ profile }: MemberProfileDetailsProps) {
  return (
    <section>
      <h1>{profile.name}</h1>
      {profile.left && <p>Left the team</p>}
      <dl>
        <dt>Average time</dt>
        <dd>{profile.averageTime} /km</dd>
        <dt>Laps</dt>
        <dd>{profile.lapCount}</dd>
      </dl>
      <h2>Lap times by race</h2>
      {profile.lapsByRace.length === 0 ? (
        <p>No lap times yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th scope="col">Race</th>
              <th scope="col">Laps (per km)</th>
            </tr>
          </thead>
          <tbody>
            {profile.lapsByRace.map((race) => (
              <tr key={race.raceId}>
                <td>{race.raceName}</td>
                <td>{race.laps.map((lap) => lap.time).join(", ")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

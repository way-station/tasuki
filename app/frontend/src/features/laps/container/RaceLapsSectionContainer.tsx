import { useRaceLaps } from "../hooks/useRaceLaps.ts";
import { useRunnerLegs } from "../hooks/useRunnerLegs.ts";
import { RaceLapsTable } from "../ui/RaceLapsTable.tsx";

type RaceLapsSectionContainerProps = {
  raceId: string;
  // Only organizers may ask (docs/design/screens.md).
  canSee: boolean;
};

// The lap times of the race's runners on R2, with each runner's legs from the
// running order. View only.
export function RaceLapsSectionContainer({ raceId, canSee }: RaceLapsSectionContainerProps) {
  const laps = useRaceLaps(raceId, canSee);
  const legsOf = useRunnerLegs(raceId, canSee);
  if (!canSee || laps.data === undefined || laps.data.length === 0) {
    return null;
  }
  const runners = laps.data.map((runner) => ({
    ...runner,
    legs: legsOf(runner.rosterMemberId),
  }));
  return <RaceLapsTable runners={runners} />;
}

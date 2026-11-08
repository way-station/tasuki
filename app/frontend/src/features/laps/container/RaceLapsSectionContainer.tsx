import { useRaceLaps } from "../hooks/useRaceLaps.ts";
import { RaceLapsTable } from "../ui/RaceLapsTable.tsx";

type RaceLapsSectionContainerProps = {
  raceId: string;
  // Only organizers may ask (docs/design/screens.md).
  canSee: boolean;
};

// The lap times of the race's runners on R2. View only.
export function RaceLapsSectionContainer({ raceId, canSee }: RaceLapsSectionContainerProps) {
  const laps = useRaceLaps(raceId, canSee);
  if (!canSee || laps.data === undefined || laps.data.length === 0) {
    return null;
  }
  return <RaceLapsTable runners={laps.data} />;
}

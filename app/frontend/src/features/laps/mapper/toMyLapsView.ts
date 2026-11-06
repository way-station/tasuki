import type { MyLapsResponse, MyLapsView } from "../types/lap.ts";

// L1, in the order the laps were submitted.
export function toMyLapsView(myLaps: MyLapsResponse): MyLapsView {
  return {
    laps: myLaps.laps.map((lap) => ({ id: lap.id, time: lap.time })),
    averageTime: myLaps.average_time,
    lapCount: myLaps.lap_count,
  };
}

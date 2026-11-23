import type {
  PredictionView,
  RunningOrderResponse,
  RunningOrderView,
} from "../types/runningOrder.ts";

// A saved running order for R2 and R7. An empty leg keeps `runner: null`, shown
// as "not decided".
export function toRunningOrderView(order: RunningOrderResponse): RunningOrderView {
  return {
    entryId: order.entry_id,
    entryName: order.entry_name,
    status: order.status,
    version: order.version,
    legs: order.legs.map((leg) => ({
      legNumber: leg.leg_number,
      distanceKm: leg.distance_km,
      runner:
        leg.runner === null
          ? null
          : { rosterMemberId: leg.runner.roster_member_id, name: leg.runner.name },
    })),
    prediction: toPredictionView(order),
  };
}

// The backend fills `total_time` for distance races and `laps` for time races.
// It says whether legs were left out; the legs say which.
function toPredictionView(order: RunningOrderResponse): PredictionView | null {
  const prediction = order.prediction;
  if (prediction === null) {
    return null;
  }
  const emptyLegNumbers = prediction.has_empty_legs
    ? order.legs.filter((leg) => leg.runner === null).map((leg) => leg.leg_number)
    : [];
  if (prediction.total_time !== null) {
    return { type: "distance", totalTime: prediction.total_time, emptyLegNumbers };
  }
  if (prediction.laps !== null) {
    return { type: "time", laps: prediction.laps, emptyLegNumbers };
  }
  return null;
}

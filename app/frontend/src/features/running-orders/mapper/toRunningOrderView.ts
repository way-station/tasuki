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
    prediction: toPredictionView(order.prediction),
  };
}

// The backend fills `total_time` for distance races and `laps` for time races.
function toPredictionView(prediction: RunningOrderResponse["prediction"]): PredictionView | null {
  if (prediction === null) {
    return null;
  }
  if (prediction.total_time !== null) {
    return {
      type: "distance",
      totalTime: prediction.total_time,
      hasEmptyLegs: prediction.has_empty_legs,
    };
  }
  if (prediction.laps !== null) {
    return { type: "time", laps: prediction.laps, hasEmptyLegs: prediction.has_empty_legs };
  }
  return null;
}

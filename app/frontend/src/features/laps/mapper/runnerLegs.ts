import type { OrderLeg, RunnerLeg } from "../types/lap.ts";

// The legs one roster member runs in the running order, in leg order.
export function runnerLegs(legs: OrderLeg[], rosterMemberId: string): RunnerLeg[] {
  return legs
    .filter((leg) => leg.runner?.rosterMemberId === rosterMemberId)
    .map((leg) =>
      leg.distanceKm === undefined
        ? { legNumber: leg.legNumber }
        : { legNumber: leg.legNumber, distanceKm: leg.distanceKm },
    );
}

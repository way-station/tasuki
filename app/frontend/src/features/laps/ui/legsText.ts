import type { RunnerLeg } from "../types/lap.ts";

// The legs as text: 2区・3.2 km for one leg, 1区・4区・10 km (the total) for
// several. Time races have no distance: 1区・4区.
export function legsText(legs: RunnerLeg[], withDistance: boolean): string {
  const names = legs.map((leg) => `${leg.legNumber}区`);
  const distances = legs.flatMap((leg) => (leg.distanceKm === undefined ? [] : [leg.distanceKm]));
  if (!withDistance || distances.length !== legs.length) {
    return names.join("・");
  }
  const total = Math.round(distances.reduce((sum, km) => sum + km, 0) * 100) / 100;
  return [...names, `${total} km`].join("・");
}

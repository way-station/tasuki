import type { components } from "../../../api/schema.gen.ts";

// What the backend returns and takes (docs/design/api.md).
export type RunningOrderListResponse = components["schemas"]["RunningOrderList"];
export type RunningOrderResponse = components["schemas"]["RunningOrder"];
export type RunningOrderSaveBody = components["schemas"]["RunningOrderSave"];

export type RunningOrderStatus = "draft" | "confirmed";

// A member of the entry who can run a leg.
export type Runner = {
  rosterMemberId: string;
  name: string;
};

// A leg as R2 and R7 show it.
export type LegView = {
  legNumber: number;
  // Distance races only.
  distanceKm?: number;
  // `null` for an empty leg, shown as "not decided".
  runner: Runner | null;
};

// Calculated by the backend from the saved running order; empty legs are left
// out.
export type PredictionView =
  | {
      type: "distance";
      // H:MM:SS
      totalTime: string;
      hasEmptyLegs: boolean;
    }
  | {
      type: "time";
      // Rounded down.
      laps: number;
      hasEmptyLegs: boolean;
    };

// A saved running order (R2, R7). The MVP allows one entry per race, so a race
// has at most one.
export type RunningOrderView = {
  entryId: string;
  entryName: string;
  status: RunningOrderStatus;
  // Organizers only.
  version?: number;
  legs: LegView[];
  // `null` while no leg has a runner.
  prediction: PredictionView | null;
};

// What the organizer edits on R7, kept as a draft (ADR-0013). Leg numbers are
// the positions: the first runner runs leg 1. `null` is an empty leg.
export type RunningOrderForm = {
  runnerIds: (string | null)[];
};

// How R7 numbers the legs: distance races have one leg per leg distance in the
// description; time races have as many as the organizer adds.
export type LegRule = { type: "distance"; distancesKm: number[] } | { type: "time" };

// R7. Runners are picked from the entry, so R6 comes first when the race has
// no entry yet.
export type RunningOrderEditor =
  | { entry: null }
  | {
      entry: { id: string; name: string };
      legRule: LegRule;
      // The entry's members.
      runners: Runner[];
      // `null` until the running order is saved for the first time.
      saved: RunningOrderView | null;
    };

import { describe, expect, it } from "vitest";
import type { RunningOrderResponse } from "../types/runningOrder.ts";
import { toRunningOrderView } from "./toRunningOrderView.ts";

const order: RunningOrderResponse = {
  entry_id: "e1",
  entry_name: "Team A",
  status: "confirmed",
  version: 2,
  legs: [
    { leg_number: 1, distance_km: 5, runner: { roster_member_id: "m1", name: "Taro" } },
    { leg_number: 2, distance_km: 3.2, runner: null },
  ],
  prediction: { total_time: "0:28:10", laps: null, has_empty_legs: true },
};

describe("toRunningOrderView", () => {
  it("keeps empty legs as not decided", () => {
    expect(toRunningOrderView(order)).toEqual({
      entryId: "e1",
      entryName: "Team A",
      status: "confirmed",
      version: 2,
      legs: [
        { legNumber: 1, distanceKm: 5, runner: { rosterMemberId: "m1", name: "Taro" } },
        { legNumber: 2, distanceKm: 3.2, runner: null },
      ],
      prediction: { type: "distance", totalTime: "0:28:10", hasEmptyLegs: true },
    });
  });

  it("reads the laps of a time race", () => {
    const view = toRunningOrderView({
      ...order,
      prediction: { total_time: null, laps: 25, has_empty_legs: false },
    });

    expect(view.prediction).toEqual({ type: "time", laps: 25, hasEmptyLegs: false });
  });

  it("has no prediction while no leg has a runner", () => {
    expect(toRunningOrderView({ ...order, prediction: null }).prediction).toBeNull();
  });

  it("has no version for a member", () => {
    const { version: _, ...forMember } = order;

    expect(toRunningOrderView(forMember).version).toBeUndefined();
  });
});

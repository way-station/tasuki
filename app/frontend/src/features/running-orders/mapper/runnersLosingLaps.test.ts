import { describe, expect, it } from "vitest";
import type { RunningOrderView } from "../types/runningOrder.ts";
import { runnersLosingLaps } from "./runnersLosingLaps.ts";

const taro = { rosterMemberId: "m1", name: "Taro" };
const jiro = { rosterMemberId: "m2", name: "Jiro" };

const confirmed: RunningOrderView = {
  entryId: "e1",
  entryName: "Team A",
  status: "confirmed",
  version: 4,
  legs: [
    { legNumber: 1, runner: taro },
    { legNumber: 2, runner: jiro },
    { legNumber: 3, runner: taro },
  ],
  prediction: null,
};

describe("runnersLosingLaps", () => {
  it("names a runner whose number of legs goes down", () => {
    expect(runnersLosingLaps(confirmed, { runnerIds: ["m1", "m2", "m2"] })).toEqual([taro]);
  });

  it("names a runner removed from every leg", () => {
    expect(runnersLosingLaps(confirmed, { runnerIds: ["m1", "m1", "m1"] })).toEqual([jiro]);
  });

  it("keeps the laps of a runner moved to another leg", () => {
    expect(runnersLosingLaps(confirmed, { runnerIds: ["m2", "m1", "m1"] })).toEqual([]);
  });

  it("names nobody for a draft, which has no laps", () => {
    expect(
      runnersLosingLaps({ ...confirmed, status: "draft" }, { runnerIds: [null, null, null] }),
    ).toEqual([]);
  });

  it("names nobody before the first save", () => {
    expect(runnersLosingLaps(null, { runnerIds: ["m1"] })).toEqual([]);
  });
});

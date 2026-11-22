import { describe, expect, it } from "vitest";
import { runnerLegs } from "./runnerLegs.ts";

const taro = { rosterMemberId: "m1" };
const jiro = { rosterMemberId: "m2" };

describe("runnerLegs", () => {
  it("gives the legs the runner runs, in order, with their distance", () => {
    expect(
      runnerLegs(
        [
          { legNumber: 1, distanceKm: 5, runner: taro },
          { legNumber: 2, distanceKm: 3.2, runner: jiro },
          { legNumber: 3, distanceKm: 3.2, runner: null },
          { legNumber: 4, distanceKm: 5, runner: taro },
        ],
        "m1",
      ),
    ).toEqual([
      { legNumber: 1, distanceKm: 5 },
      { legNumber: 4, distanceKm: 5 },
    ]);
  });

  it("leaves the distance out for time races", () => {
    expect(runnerLegs([{ legNumber: 2, runner: taro }], "m1")).toEqual([{ legNumber: 2 }]);
  });
});

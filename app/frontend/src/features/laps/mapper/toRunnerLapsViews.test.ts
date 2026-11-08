import { describe, expect, it } from "vitest";
import { toRunnerLapsViews } from "./toRunnerLapsViews.ts";

describe("toRunnerLapsViews", () => {
  it("lists each runner with their laps, including runners without any", () => {
    expect(
      toRunnerLapsViews({
        runners: [
          { roster_member_id: "m1", name: "Taro", laps: [{ id: "l1", time: "3:25" }] },
          { roster_member_id: "m2", name: "Jiro", laps: [] },
        ],
      }),
    ).toEqual([
      { rosterMemberId: "m1", name: "Taro", laps: [{ id: "l1", time: "3:25" }] },
      { rosterMemberId: "m2", name: "Jiro", laps: [] },
    ]);
  });
});

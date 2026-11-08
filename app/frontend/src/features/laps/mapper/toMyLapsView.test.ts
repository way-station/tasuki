import { describe, expect, it } from "vitest";
import { toMyLapsView } from "./toMyLapsView.ts";

describe("toMyLapsView", () => {
  it("keeps the laps in the order they were submitted, with the average", () => {
    expect(
      toMyLapsView({
        laps: [
          { id: "l1", time: "3:25" },
          { id: "l2", time: "3:20" },
        ],
        average_time: "3:28",
        lap_count: 8,
      }),
    ).toEqual({
      laps: [
        { id: "l1", time: "3:25" },
        { id: "l2", time: "3:20" },
      ],
      averageTime: "3:28",
      lapCount: 8,
    });
  });
});

import { describe, expect, it } from "vitest";
import { toProfileView } from "./toProfileView.ts";

describe("toProfileView", () => {
  it("shows the default average before any lap", () => {
    expect(toProfileView({ name: "Taro", average_time: "3:40", lap_count: 0 })).toEqual({
      name: "Taro",
      averageTime: "3:40",
      lapCount: 0,
    });
  });
});

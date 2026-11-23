import { describe, expect, it } from "vitest";
import { averageTime } from "./averageTime.ts";

const laps = (...times: string[]) => times.map((time, index) => ({ id: `l${index}`, time }));

describe("averageTime", () => {
  it("gives the mean of the laps as MM:SS", () => {
    expect(averageTime(laps("3:55", "3:58", "4:01"))).toBe("3:58");
  });

  it("rounds half a second up", () => {
    expect(averageTime(laps("3:20", "3:25"))).toBe("3:23");
  });

  it("gives null without laps", () => {
    expect(averageTime([])).toBeNull();
  });
});

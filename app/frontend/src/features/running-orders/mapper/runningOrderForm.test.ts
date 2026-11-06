import { describe, expect, it } from "vitest";
import type { RunningOrderView } from "../types/runningOrder.ts";
import {
  hasEveryRunner,
  isSaved,
  toRunningOrderBody,
  toRunningOrderForm,
} from "./runningOrderForm.ts";

const saved: RunningOrderView = {
  entryId: "e1",
  entryName: "Team A",
  status: "draft",
  version: 3,
  legs: [
    { legNumber: 1, runner: { rosterMemberId: "m1", name: "Taro" } },
    { legNumber: 2, runner: null },
  ],
  prediction: null,
};

describe("toRunningOrderForm", () => {
  it("starts a distance race with one empty leg per leg distance", () => {
    expect(toRunningOrderForm({ type: "distance", distancesKm: [5, 3.2, 4] }, null)).toEqual({
      runnerIds: [null, null, null],
    });
  });

  it("starts a time race with one empty leg", () => {
    expect(toRunningOrderForm({ type: "time" }, null)).toEqual({ runnerIds: [null] });
  });

  it("starts from the saved runners", () => {
    expect(toRunningOrderForm({ type: "time" }, saved)).toEqual({ runnerIds: ["m1", null] });
  });
});

describe("toRunningOrderBody", () => {
  it("numbers the legs from 1 and sends the version", () => {
    expect(toRunningOrderBody({ runnerIds: ["m1", null, "m1"] }, 3)).toEqual({
      version: 3,
      legs: [
        { leg_number: 1, roster_member_id: "m1" },
        { leg_number: 2, roster_member_id: null },
        { leg_number: 3, roster_member_id: "m1" },
      ],
    });
  });

  it("leaves the version out for the first draft", () => {
    expect(toRunningOrderBody({ runnerIds: [null] })).toEqual({
      legs: [{ leg_number: 1, roster_member_id: null }],
    });
  });
});

describe("isSaved", () => {
  it("is true when the form matches the saved runners", () => {
    expect(isSaved({ runnerIds: ["m1", null] }, saved)).toBe(true);
  });

  it("is false after a change, a new leg, or before the first save", () => {
    expect(isSaved({ runnerIds: ["m2", null] }, saved)).toBe(false);
    expect(isSaved({ runnerIds: ["m1", null, null] }, saved)).toBe(false);
    expect(isSaved({ runnerIds: ["m1", null] }, null)).toBe(false);
  });
});

describe("hasEveryRunner", () => {
  it("is true only when no leg is empty", () => {
    expect(hasEveryRunner({ runnerIds: ["m1", "m2"] })).toBe(true);
    expect(hasEveryRunner({ runnerIds: ["m1", null] })).toBe(false);
  });
});

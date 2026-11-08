import { describe, expect, it } from "vitest";
import type { RunningOrderView } from "../types/runningOrder.ts";
import { checkRunningOrder } from "./runningOrderChecks.ts";

const taro = { rosterMemberId: "m1", name: "Taro" };

function savedOrder(status: "draft" | "confirmed"): RunningOrderView {
  return {
    entryId: "e1",
    entryName: "Team A",
    status,
    version: 1,
    legs: [
      { legNumber: 1, runner: taro },
      { legNumber: 2, runner: taro },
    ],
    prediction: null,
  };
}

describe("checkRunningOrder", () => {
  it("allows confirming a saved draft with every leg filled", () => {
    expect(checkRunningOrder(savedOrder("draft"), { runnerIds: ["m1", "m1"] })).toEqual({
      unsaved: false,
      everyRunner: true,
      canConfirm: true,
      losingLaps: [],
    });
  });

  it("asks to save first when the form has changes", () => {
    const checks = checkRunningOrder(savedOrder("draft"), { runnerIds: ["m1", null] });

    expect(checks).toMatchObject({ unsaved: true, everyRunner: false, canConfirm: false });
  });

  it("cannot confirm before the first save", () => {
    expect(checkRunningOrder(null, { runnerIds: ["m1"] }).canConfirm).toBe(false);
  });

  it("cannot confirm again, and names who loses laps in a confirmed order", () => {
    const checks = checkRunningOrder(savedOrder("confirmed"), { runnerIds: ["m1", null] });

    expect(checks.canConfirm).toBe(false);
    expect(checks.losingLaps).toEqual([taro]);
  });
});

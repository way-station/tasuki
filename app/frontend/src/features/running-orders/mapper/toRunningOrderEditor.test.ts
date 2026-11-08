import { describe, expect, it } from "vitest";
import type { components } from "../../../api/schema.gen.ts";
import type { RunningOrderResponse } from "../types/runningOrder.ts";
import { toRunningOrderEditor } from "./toRunningOrderEditor.ts";

type Race = components["schemas"]["Race"];

const distanceRace: Race = {
  id: "race-1",
  name: "Hakone",
  description: { type: "distance", race_date: "2026-11-23", leg_distances_km: [5, 3.2] },
  can_submit_laps: false,
};
const timeRace: Race = {
  ...distanceRace,
  description: {
    type: "time",
    race_date: "2026-11-23",
    time_limit_minutes: 180,
    lap_distance_km: 2,
  },
};
const entries = {
  entries: [
    {
      id: "e1",
      name: "Team A",
      members: [
        { roster_member_id: "m1", name: "Taro" },
        { roster_member_id: "m2", name: "Jiro" },
      ],
    },
  ],
};
const saved: RunningOrderResponse = {
  entry_id: "e1",
  entry_name: "Team A",
  status: "draft",
  version: 1,
  legs: [{ leg_number: 1, runner: { roster_member_id: "m2", name: "Jiro" } }],
  prediction: null,
};

describe("toRunningOrderEditor", () => {
  it("needs the entry first", () => {
    expect(toRunningOrderEditor(distanceRace, { entries: [] }, { running_orders: [] })).toEqual({
      entry: null,
    });
  });

  it("offers the entry's members and numbers legs by the leg distances", () => {
    expect(toRunningOrderEditor(distanceRace, entries, { running_orders: [] })).toEqual({
      entry: { id: "e1", name: "Team A" },
      legRule: { type: "distance", distancesKm: [5, 3.2] },
      runners: [
        { rosterMemberId: "m1", name: "Taro" },
        { rosterMemberId: "m2", name: "Jiro" },
      ],
      saved: null,
    });
  });

  it("lets a time race have any number of legs", () => {
    const editor = toRunningOrderEditor(timeRace, entries, { running_orders: [] });

    expect(editor.entry === null ? undefined : editor.legRule).toEqual({ type: "time" });
  });

  it("includes the saved running order of the entry", () => {
    const editor = toRunningOrderEditor(distanceRace, entries, { running_orders: [saved] });

    expect(editor.entry === null ? undefined : editor.saved?.version).toBe(1);
  });
});

import { describe, expect, it } from "vitest";
import type {
  RaceResponse,
  RosterMemberDetailResponse,
  RunningOrderListResponse,
} from "../types/roster.ts";
import { toMemberProfileView } from "./toMemberProfileView.ts";

const detail: RosterMemberDetailResponse = {
  id: "m1",
  name: "Taro",
  left: true,
  average_time: "3:28",
  lap_count: 3,
  laps_by_race: [
    {
      race_id: "race-1",
      race_name: "Hakone",
      laps: [
        { id: "l1", time: "3:20" },
        { id: "l2", time: "3:30" },
      ],
    },
    { race_id: "race-2", race_name: "Ekiden", laps: [{ id: "l3", time: "3:35" }] },
  ],
};

describe("toMemberProfileView", () => {
  it("adds each race's date and the member's legs when they are loaded", () => {
    const race: RaceResponse = {
      id: "race-1",
      name: "Hakone",
      description: { type: "distance", race_date: "2026-12-20", leg_distances_km: [5, 3, 5] },
      can_submit_laps: false,
    };
    const orders: RunningOrderListResponse = {
      running_orders: [
        {
          entry_id: "e1",
          entry_name: "Team A",
          status: "confirmed",
          legs: [
            { leg_number: 1, distance_km: 5, runner: { roster_member_id: "m1", name: "Taro" } },
            { leg_number: 2, distance_km: 3, runner: { roster_member_id: "m2", name: "Jiro" } },
            { leg_number: 3, distance_km: 5, runner: { roster_member_id: "m1", name: "Taro" } },
          ],
          prediction: null,
        },
      ],
    };

    const view = toMemberProfileView(
      detail,
      new Map([["race-1", race]]),
      new Map([["race-1", orders]]),
    );

    expect(
      view.lapsByRace.map(({ raceId, raceDate, legNumbers }) => ({ raceId, raceDate, legNumbers })),
    ).toEqual([
      { raceId: "race-1", raceDate: "2026-12-20", legNumbers: [1, 3] },
      { raceId: "race-2", raceDate: undefined, legNumbers: [] },
    ]);
  });

  it("keeps the laps per race in order", () => {
    expect(toMemberProfileView(detail, new Map(), new Map())).toEqual({
      id: "m1",
      name: "Taro",
      left: true,
      averageTime: "3:28",
      lapCount: 3,
      lapsByRace: [
        {
          raceId: "race-1",
          raceName: "Hakone",
          legNumbers: [],
          laps: [
            { id: "l1", time: "3:20" },
            { id: "l2", time: "3:30" },
          ],
        },
        {
          raceId: "race-2",
          raceName: "Ekiden",
          legNumbers: [],
          laps: [{ id: "l3", time: "3:35" }],
        },
      ],
    });
  });
});

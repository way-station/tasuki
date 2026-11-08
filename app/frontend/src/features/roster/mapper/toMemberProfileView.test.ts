import { describe, expect, it } from "vitest";
import { toMemberProfileView } from "./toMemberProfileView.ts";

describe("toMemberProfileView", () => {
  it("keeps the laps per race in order", () => {
    expect(
      toMemberProfileView({
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
      }),
    ).toEqual({
      id: "m1",
      name: "Taro",
      left: true,
      averageTime: "3:28",
      lapCount: 3,
      lapsByRace: [
        {
          raceId: "race-1",
          raceName: "Hakone",
          laps: [
            { id: "l1", time: "3:20" },
            { id: "l2", time: "3:30" },
          ],
        },
        { raceId: "race-2", raceName: "Ekiden", laps: [{ id: "l3", time: "3:35" }] },
      ],
    });
  });
});

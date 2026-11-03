import { describe, expect, it } from "vitest";
import type { RaceResponse } from "../types/race.ts";
import { toRaceDetail } from "./toRaceDetail.ts";

describe("toRaceDetail", () => {
  it("converts a distance race as an organizer sees it", () => {
    const race: RaceResponse = {
      id: "race-1",
      name: "Hakone",
      version: 3,
      description: {
        type: "distance",
        race_date: "2026-11-23",
        deadline: "2026-11-01T18:00:00+09:00",
        text: "Meet at 7:00",
        leg_distances_km: [5, 3.2],
        target: { total_time: "2:45:00" },
      },
      running_order_confirmed: false,
      can_submit_laps: false,
    };

    expect(toRaceDetail(race)).toEqual({
      id: "race-1",
      name: "Hakone",
      version: 3,
      raceDate: "2026-11-23",
      deadline: "2026-11-01T18:00:00+09:00",
      text: "Meet at 7:00",
      detail: { type: "distance", legDistancesKm: [5, 3.2], targetTotalTime: "2:45:00" },
      myAttendance: undefined,
      runningOrderConfirmed: false,
      canSubmitLaps: false,
    });
  });

  it("converts a time race as an invited member sees it", () => {
    const race: RaceResponse = {
      id: "race-2",
      name: "Ekiden 24h",
      description: {
        type: "time",
        race_date: "2026-12-05",
        time_limit_minutes: 180,
        lap_distance_km: 2,
        target: { laps: 30 },
      },
      my_attendance: null,
      can_submit_laps: false,
    };

    const detail = toRaceDetail(race);

    expect(detail.detail).toEqual({
      type: "time",
      timeLimitMinutes: 180,
      lapDistanceKm: 2,
      targetLaps: 30,
    });
    expect(detail).toMatchObject({ version: undefined, deadline: undefined, text: "" });
    expect(detail.myAttendance).toBe("not_answered");
    expect(detail.runningOrderConfirmed).toBeUndefined();
  });
});

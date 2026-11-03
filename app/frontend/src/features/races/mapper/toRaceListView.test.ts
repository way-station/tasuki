import { describe, expect, it } from "vitest";
import type { RaceSummaryResponse } from "../types/race.ts";
import { toRaceListView } from "./toRaceListView.ts";

function summary(id: string, overrides: Partial<RaceSummaryResponse> = {}): RaceSummaryResponse {
  return { id, name: `Race ${id}`, race_date: "2026-11-23", ...overrides };
}

describe("toRaceListView", () => {
  it("converts a race", () => {
    const view = toRaceListView(
      { races: [summary("1", { my_attendance: "attending", running_order_confirmed: true })] },
      false,
    );

    expect(view).toEqual({
      role: "member",
      myRaces: [
        {
          id: "1",
          name: "Race 1",
          raceDate: "2026-11-23",
          myAttendance: "attending",
          runningOrderConfirmed: true,
        },
      ],
      invitations: [],
    });
  });

  it("splits a member's races into My races and Invitations, keeping the order", () => {
    const view = toRaceListView(
      {
        races: [
          summary("1", { my_attendance: null }),
          summary("2", { my_attendance: "attending" }),
          summary("3", { my_attendance: "undecided" }),
          summary("4", { my_attendance: "attending" }),
        ],
      },
      false,
    );

    expect(view.role).toBe("member");
    if (view.role !== "member") return;
    expect(view.myRaces.map((race) => race.id)).toEqual(["2", "4"]);
    expect(view.invitations.map((race) => [race.id, race.myAttendance])).toEqual([
      ["1", "not_answered"],
      ["3", "undecided"],
    ]);
  });

  it("shows an organizer every race, and their own invitations", () => {
    const view = toRaceListView(
      {
        races: [
          summary("1"),
          summary("2", { my_attendance: null }),
          summary("3", { my_attendance: "attending" }),
        ],
      },
      true,
    );

    expect(view.role).toBe("organizer");
    if (view.role !== "organizer") return;
    expect(view.races.map((race) => race.id)).toEqual(["1", "2", "3"]);
    expect(view.races[0].myAttendance).toBeUndefined();
    expect(view.invitations.map((race) => race.id)).toEqual(["2"]);
  });
});

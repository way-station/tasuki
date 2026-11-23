import { describe, expect, it } from "vitest";
import type { InvitationListResponse, RunningOrderListResponse } from "../types/race.ts";
import { toRaceManagementSummary } from "./toRaceManagementSummary.ts";

const invitations: InvitationListResponse = {
  members: [
    { roster_member_id: "m1", name: "Taro", invited: true, attendance: "attending" },
    { roster_member_id: "m2", name: "Jiro", invited: true, attendance: "undecided" },
    { roster_member_id: "m3", name: "Hanako", invited: true, attendance: null },
    { roster_member_id: "m4", name: "Saburo", invited: true, attendance: "not_attending" },
    { roster_member_id: "m5", name: "Shiro", invited: false, attendance: null },
  ],
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
const orders: RunningOrderListResponse = {
  running_orders: [
    {
      entry_id: "e1",
      entry_name: "Team A",
      status: "confirmed",
      version: 2,
      legs: [],
      prediction: null,
    },
  ],
};

describe("toRaceManagementSummary", () => {
  it("counts the invited members by answer, and takes the entry and its order", () => {
    expect(toRaceManagementSummary(invitations, entries, orders)).toEqual({
      invitedCount: 4,
      attendance: { attending: 1, undecided: 1, notAnswered: 1, notAttending: 1 },
      entry: { name: "Team A", memberCount: 2 },
      runningOrder: "confirmed",
    });
  });

  it("has no entry or running order before they are made", () => {
    const summary = toRaceManagementSummary(
      { members: [] },
      { entries: [] },
      { running_orders: [] },
    );

    expect(summary).toEqual({
      invitedCount: 0,
      attendance: { attending: 0, undecided: 0, notAnswered: 0, notAttending: 0 },
      entry: null,
      runningOrder: null,
    });
  });

  it("has no running order while the entry's order was never saved", () => {
    expect(toRaceManagementSummary(invitations, entries, { running_orders: [] }).runningOrder).toBe(
      null,
    );
  });
});

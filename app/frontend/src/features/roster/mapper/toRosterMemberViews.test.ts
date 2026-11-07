import { describe, expect, it } from "vitest";
import type { RosterMemberResponse } from "../types/roster.ts";
import { toRosterMemberViews } from "./toRosterMemberViews.ts";

function member(overrides: Partial<RosterMemberResponse>): RosterMemberResponse {
  return {
    id: "m1",
    name: "Taro",
    left: false,
    link_status: "none",
    invite_link_expires_at: null,
    ...overrides,
  };
}

describe("toRosterMemberViews", () => {
  it("shows the expiry of an active invite link", () => {
    expect(
      toRosterMemberViews({
        roster_members: [member({ invite_link_expires_at: "2026-11-14T09:00:00Z" })],
      }),
    ).toEqual([
      {
        id: "m1",
        name: "Taro",
        left: false,
        linkStatus: "none",
        inviteLinkExpiresAt: "2026-11-14T09:00:00Z",
        canCreateInviteLink: true,
        canUnlink: false,
      },
    ]);
  });

  it("allows an invite link only for members who stay and have no user", () => {
    const views = toRosterMemberViews({
      roster_members: [
        member({ id: "none" }),
        member({ id: "pending", link_status: "pending" }),
        member({ id: "approved", link_status: "approved" }),
        member({ id: "left", left: true }),
      ],
    });

    expect(views.map((view) => [view.id, view.canCreateInviteLink])).toEqual([
      ["none", true],
      ["pending", false],
      ["approved", false],
      ["left", false],
    ]);
  });

  it("allows unlinking an approved user only", () => {
    const views = toRosterMemberViews({
      roster_members: [
        member({ id: "pending", link_status: "pending" }),
        member({ id: "approved", link_status: "approved", left: true }),
      ],
    });

    expect(views.map((view) => [view.id, view.canUnlink])).toEqual([
      ["pending", false],
      ["approved", true],
    ]);
  });
});

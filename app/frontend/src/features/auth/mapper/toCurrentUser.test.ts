import { describe, expect, it } from "vitest";
import type { MeResponse } from "../types/auth.ts";
import { toCurrentUser } from "./toCurrentUser.ts";

function me(overrides: Partial<MeResponse>): MeResponse {
  return {
    display_name: "Taro",
    is_organizer: false,
    status: "none",
    roster_member: null,
    ...overrides,
  };
}

describe("toCurrentUser", () => {
  it("converts the fields", () => {
    expect(
      toCurrentUser(
        me({ status: "approved", roster_member: { id: "member-1", name: "Taro Yamada" } }),
      ),
    ).toEqual({
      displayName: "Taro",
      isOrganizer: false,
      rosterMember: { id: "member-1", name: "Taro Yamada" },
      entryScreen: "home",
    });
  });

  describe("member", () => {
    it.each([
      ["approved", "home"],
      ["pending", "waiting"],
      ["rejected", "rejected"],
      ["none", "unavailable"],
    ] as const)("with status %s lands on %s", (status, screen) => {
      expect(toCurrentUser(me({ status })).entryScreen).toBe(screen);
    });
  });

  describe("organizer", () => {
    it.each(["approved", "pending", "rejected", "none"] as const)(
      "with status %s lands on the team's screens",
      (status) => {
        expect(toCurrentUser(me({ is_organizer: true, status })).entryScreen).toBe("home");
      },
    );

    it("has no roster member until linked", () => {
      expect(toCurrentUser(me({ is_organizer: true, status: "none" })).rosterMember).toBeNull();
    });
  });
});

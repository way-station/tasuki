import { describe, expect, it } from "vitest";
import type { CurrentUser, EntryScreen } from "../types/auth.ts";
import { type Access, accessRedirect } from "./accessRedirect.ts";

function user(overrides: Partial<CurrentUser>): CurrentUser {
  return {
    displayName: "Taro",
    isOrganizer: false,
    rosterMember: null,
    entryScreen: "home",
    ...overrides,
  };
}

const linked = { id: "member-1", name: "Taro Yamada" };
const accesses: Access[] = ["team", "organizer", "linked"];

describe("accessRedirect", () => {
  it.each(accesses)("sends a user without a session to the login screen (%s)", (access) => {
    expect(accessRedirect("unauthenticated", access)).toBe("/login");
  });

  it.each([
    ["waiting", "/waiting"],
    ["rejected", "/rejected"],
    ["unavailable", "/unavailable"],
  ] as [EntryScreen, string][])(
    "sends a member who cannot use tasuki yet (%s) to %s",
    (entryScreen, redirect) => {
      for (const access of accesses) {
        expect(accessRedirect(user({ entryScreen }), access)).toBe(redirect);
      }
    },
  );

  describe("member", () => {
    const member = user({ rosterMember: linked });

    it("opens the team's screens and member screens", () => {
      expect(accessRedirect(member, "team")).toBeNull();
      expect(accessRedirect(member, "linked")).toBeNull();
    });

    it("goes home from organizer screens", () => {
      expect(accessRedirect(member, "organizer")).toBe("/");
    });
  });

  describe("organizer without a link", () => {
    const organizer = user({ isOrganizer: true });

    it("opens the team's screens and organizer screens", () => {
      expect(accessRedirect(organizer, "team")).toBeNull();
      expect(accessRedirect(organizer, "organizer")).toBeNull();
    });

    it("goes home from member screens", () => {
      expect(accessRedirect(organizer, "linked")).toBe("/");
    });
  });

  describe("organizer with a link", () => {
    const organizer = user({ isOrganizer: true, rosterMember: linked });

    it.each(accesses)("opens every screen (%s)", (access) => {
      expect(accessRedirect(organizer, access)).toBeNull();
    });
  });
});

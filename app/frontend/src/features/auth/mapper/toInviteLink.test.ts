import { describe, expect, it } from "vitest";
import { toInviteLink } from "./toInviteLink.ts";

describe("toInviteLink", () => {
  it("converts the team name", () => {
    expect(toInviteLink({ team_name: "Team" })).toEqual({ teamName: "Team" });
  });
});

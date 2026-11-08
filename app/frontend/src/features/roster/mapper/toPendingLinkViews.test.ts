import { describe, expect, it } from "vitest";
import { toPendingLinkViews } from "./toPendingLinkViews.ts";

describe("toPendingLinkViews", () => {
  it("shows who logged in, for which roster member, and when", () => {
    expect(
      toPendingLinkViews({
        links: [
          {
            id: "link-1",
            line_display_name: "taro_line",
            roster_member: { id: "m1", name: "Taro" },
            created_at: "2026-11-07T09:00:00Z",
          },
        ],
      }),
    ).toEqual([
      {
        id: "link-1",
        lineDisplayName: "taro_line",
        rosterMember: { id: "m1", name: "Taro" },
        createdAt: "2026-11-07T09:00:00Z",
      },
    ]);
  });
});

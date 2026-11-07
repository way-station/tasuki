import { describe, expect, it } from "vitest";
import { toInviteLinkView } from "./toInviteLinkView.ts";

describe("toInviteLinkView", () => {
  it("keeps the URL to copy and its expiry", () => {
    expect(
      toInviteLinkView({
        url: "https://tasuki.example/invite/abc",
        expires_at: "2026-11-14T09:00:00Z",
      }),
    ).toEqual({ url: "https://tasuki.example/invite/abc", expiresAt: "2026-11-14T09:00:00Z" });
  });
});

import { describe, expect, it } from "vitest";
import { toLoginError } from "./toLoginError.ts";

describe("toLoginError", () => {
  it.each([
    ["invite_unusable", "unavailable"],
    ["left", "unavailable"],
    ["login_failed", "failed"],
    ["something_new", "failed"],
  ] as const)("converts %s to %s", (code, error) => {
    expect(toLoginError(code)).toBe(error);
  });
});

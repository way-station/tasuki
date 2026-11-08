import { describe, expect, it } from "vitest";
import { emptyRosterMemberForm, toRosterMemberBody } from "./rosterMemberForm.ts";

describe("emptyRosterMemberForm", () => {
  it("starts with an empty name", () => {
    expect(emptyRosterMemberForm()).toEqual({ name: "" });
  });
});

describe("toRosterMemberBody", () => {
  it("sends the trimmed name", () => {
    expect(toRosterMemberBody({ name: " Taro " })).toEqual({ ok: true, body: { name: "Taro" } });
  });

  it("needs a name", () => {
    expect(toRosterMemberBody({ name: "  " })).toEqual({
      ok: false,
      errors: { name: "Enter a name." },
    });
  });
});

import { describe, expect, it } from "vitest";
import {
  newRosterMemberForm,
  RosterMemberFormInvalidError,
  rosterMemberFormErrors,
} from "./rosterMemberFormInvalid.ts";

describe("rosterMemberFormErrors", () => {
  it("reads the per-field errors of an invalid form only", () => {
    const errors = { name: "Enter a name." };

    expect(rosterMemberFormErrors(new RosterMemberFormInvalidError(errors))).toEqual(errors);
    expect(rosterMemberFormErrors(new Error("network"))).toBeUndefined();
  });
});

describe("newRosterMemberForm", () => {
  it("starts with an empty name", () => {
    expect(newRosterMemberForm()).toEqual({ name: "" });
  });
});

import { describe, expect, it } from "vitest";
import {
  LapsFormInvalidError,
  LapTimeInvalidError,
  lapsFormErrors,
  lapTimeError,
} from "./lapsFormInvalid.ts";

describe("lapsFormErrors", () => {
  it("reads the per-time errors of an invalid form only", () => {
    const errors = { "times.0": "1 km のタイムを 分:秒 で入れてください（例：3:25）。" };

    expect(lapsFormErrors(new LapsFormInvalidError(errors))).toEqual(errors);
    expect(lapsFormErrors(new Error("network"))).toBeUndefined();
  });
});

describe("lapTimeError", () => {
  it("reads the error of an invalid edited time only", () => {
    expect(lapTimeError(new LapTimeInvalidError("Enter M:SS."))).toBe("Enter M:SS.");
    expect(lapTimeError(new Error("network"))).toBeUndefined();
  });
});

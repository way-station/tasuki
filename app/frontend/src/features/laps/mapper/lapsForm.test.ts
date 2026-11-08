import { describe, expect, it } from "vitest";
import { emptyLapsForm, toLapEditBody, toLapsBody } from "./lapsForm.ts";

describe("emptyLapsForm", () => {
  it("starts with one empty lap", () => {
    expect(emptyLapsForm()).toEqual({ times: [""] });
  });
});

describe("toLapsBody", () => {
  it("sends the trimmed times together", () => {
    expect(toLapsBody({ times: [" 3:20", "10:05 "] })).toEqual({
      ok: true,
      body: { times: ["3:20", "10:05"] },
    });
  });

  it("reports each time that is not M:SS", () => {
    expect(toLapsBody({ times: ["3:20", "", "3:5", "3:60", "205"] })).toEqual({
      ok: false,
      errors: {
        "times.1": "Enter the time per km as M:SS, for example 3:25.",
        "times.2": "Enter the time per km as M:SS, for example 3:25.",
        "times.3": "Enter the time per km as M:SS, for example 3:25.",
        "times.4": "Enter the time per km as M:SS, for example 3:25.",
      },
    });
  });

  it("needs at least one time", () => {
    expect(toLapsBody({ times: [] })).toEqual({
      ok: false,
      errors: { times: "Add at least one lap time." },
    });
  });
});

describe("toLapEditBody", () => {
  it("sends a valid time", () => {
    expect(toLapEditBody(" 3:30 ")).toEqual({ ok: true, body: { time: "3:30" } });
  });

  it("rejects a time that is not M:SS", () => {
    expect(toLapEditBody("3.30")).toEqual({
      ok: false,
      error: "Enter the time per km as M:SS, for example 3:25.",
    });
  });
});

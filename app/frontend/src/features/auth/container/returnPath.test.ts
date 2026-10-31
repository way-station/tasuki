import { afterEach, describe, expect, it } from "vitest";
import { saveReturnPath, takeReturnPath } from "./returnPath.ts";

afterEach(() => {
  localStorage.clear();
});

describe("returnPath", () => {
  it("returns the saved path once", () => {
    saveReturnPath("/races/1/running-order?tab=draft");

    expect(takeReturnPath()).toBe("/races/1/running-order?tab=draft");
    expect(takeReturnPath()).toBe("/");
  });

  it("returns home when nothing was saved", () => {
    expect(takeReturnPath()).toBe("/");
  });

  it("forgets an earlier path when none is given", () => {
    saveReturnPath("/races/1");
    saveReturnPath(undefined);

    expect(takeReturnPath()).toBe("/");
  });

  it.each(["https://evil.example/", "//evil.example/", "/login", "/auth/complete?code=x"])(
    "never returns to %s",
    (path) => {
      saveReturnPath(path);

      expect(takeReturnPath()).toBe("/");
    },
  );
});

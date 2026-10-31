import { describe, expect, it } from "vitest";
import { toHealthView } from "./toHealthView.ts";

describe("toHealthView", () => {
  it("shows every store as reachable when the backend is healthy", () => {
    expect(toHealthView({ status: "ok", mongo: "ok", postgres: "ok" })).toEqual({
      healthy: true,
      stores: [
        { name: "MongoDB", reachable: true },
        { name: "PostgreSQL", reachable: true },
      ],
    });
  });

  it("shows which store is unreachable", () => {
    expect(toHealthView({ status: "error", mongo: "ok", postgres: "unreachable" })).toEqual({
      healthy: false,
      stores: [
        { name: "MongoDB", reachable: true },
        { name: "PostgreSQL", reachable: false },
      ],
    });
  });
});

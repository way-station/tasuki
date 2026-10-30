import { describe, expect, it } from "vitest";
import { ApiError } from "./ApiError.ts";
import { createApiClient, type Health } from "./client.ts";

function clientAnswering(status: number, body: string, contentType: string) {
  return createApiClient({
    baseUrl: "http://localhost",
    fetch: async () => new Response(body, { status, headers: { "Content-Type": contentType } }),
  });
}

const health: Health = { status: "ok", mongo: "ok", postgres: "ok" };
const unhealthy: Health = { status: "error", mongo: "ok", postgres: "unreachable" };

describe("createApiClient", () => {
  describe("getHealth", () => {
    it("returns the health", async () => {
      const client = clientAnswering(200, JSON.stringify(health), "application/json");

      await expect(client.getHealth()).resolves.toEqual(health);
    });

    it("returns the health for 503, which says which store is unreachable", async () => {
      const client = clientAnswering(503, JSON.stringify(unhealthy), "application/json");

      await expect(client.getHealth()).resolves.toEqual(unhealthy);
    });
  });

  describe("errors", () => {
    it("throws an ApiError with the code of a problem+json response", async () => {
      const problem = {
        type: "about:blank",
        title: "Unauthenticated",
        status: 401,
        code: "unauthenticated",
      };
      const client = clientAnswering(401, JSON.stringify(problem), "application/problem+json");

      const error = await client.getHealth().catch((caught: unknown) => caught);

      expect(error).toBeInstanceOf(ApiError);
      expect(error).toMatchObject({ status: 401, code: "unauthenticated", problem });
    });

    it("throws an ApiError without a code when something else answered", async () => {
      const client = clientAnswering(502, "<html>Bad Gateway</html>", "text/html");

      const error = await client.getHealth().catch((caught: unknown) => caught);

      expect(error).toBeInstanceOf(ApiError);
      expect(error).toMatchObject({ status: 502, code: undefined, problem: undefined });
    });
  });
});

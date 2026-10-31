import { describe, expect, it } from "vitest";
import { ApiError } from "./ApiError.ts";
import { createApiClient, type Health, type Me } from "./client.ts";

function clientAnswering(status: number, body: string, contentType: string) {
  return createApiClient({
    baseUrl: "http://localhost",
    fetch: async () => new Response(body, { status, headers: { "Content-Type": contentType } }),
  });
}

// Records each request, and answers every one with the given response.
function recordingClient(status: number, body?: unknown) {
  const requests: Request[] = [];
  const client = createApiClient({
    baseUrl: "http://localhost",
    fetch: async (input) => {
      requests.push(input as Request);
      if (body === undefined) {
        return new Response(null, { status });
      }
      const contentType = status < 400 ? "application/json" : "application/problem+json";
      return new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": contentType },
      });
    },
  });
  return { client, requests };
}

function problemBody(status: number, code: string) {
  return { type: "about:blank", title: code, status, code };
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

  describe("auth and joining", () => {
    it("checks an invite link", async () => {
      const { client, requests } = recordingClient(200, { team_name: "Team" });

      await expect(client.getInviteLink("abc")).resolves.toEqual({ team_name: "Team" });
      expect(requests[0].method).toBe("GET");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/invite-links/abc");
    });

    it("throws invite_unusable for an unusable invite link", async () => {
      const { client } = recordingClient(404, problemBody(404, "invite_unusable"));

      await expect(client.getInviteLink("abc")).rejects.toMatchObject({
        status: 404,
        code: "invite_unusable",
      });
    });

    it("starts LINE Login with the invite token", async () => {
      const { client, requests } = recordingClient(200, {
        authorization_url: "https://line.example/auth",
      });

      await expect(client.authorizeLine("abc")).resolves.toBe("https://line.example/auth");
      expect(requests[0].method).toBe("POST");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/auth/line/authorize");
      await expect(requests[0].json()).resolves.toEqual({ invite_token: "abc" });
    });

    it("starts LINE Login without an invite token", async () => {
      const { client, requests } = recordingClient(200, {
        authorization_url: "https://line.example/auth",
      });

      await client.authorizeLine();

      await expect(requests[0].json()).resolves.toEqual({});
    });

    it("exchanges the one-time code for a session", async () => {
      const { client, requests } = recordingClient(200, { expires_at: "2026-11-01T00:00:00Z" });

      await expect(client.createSession("one-time")).resolves.toBeUndefined();
      expect(requests[0].method).toBe("POST");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/session");
      await expect(requests[0].json()).resolves.toEqual({ code: "one-time" });
    });

    it("throws invalid_request for an unknown, expired, or used code", async () => {
      const { client } = recordingClient(400, problemBody(400, "invalid_request"));

      await expect(client.createSession("used")).rejects.toMatchObject({
        status: 400,
        code: "invalid_request",
      });
    });

    it("logs out", async () => {
      const { client, requests } = recordingClient(204);

      await expect(client.logout()).resolves.toBeUndefined();
      expect(requests[0].method).toBe("DELETE");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/session");
    });

    it("gets me", async () => {
      const me: Me = {
        user_id: "user-1",
        display_name: "Taro",
        is_organizer: false,
        status: "approved",
        roster_member: { id: "member-1", name: "Taro Yamada" },
      };
      const { client, requests } = recordingClient(200, me);

      await expect(client.getMe()).resolves.toEqual(me);
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/me");
    });

    it("throws unauthenticated without a session", async () => {
      const { client } = recordingClient(401, problemBody(401, "unauthenticated"));

      await expect(client.getMe()).rejects.toBeInstanceOf(ApiError);
      await expect(client.getMe()).rejects.toMatchObject({ code: "unauthenticated" });
    });
  });
});

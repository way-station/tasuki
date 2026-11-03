import { describe, expect, it } from "vitest";
import { ApiError } from "./ApiError.ts";
import { createApiClient, type Health, type Me, type Race, type RaceCreate } from "./client.ts";

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

  describe("races", () => {
    const race: Race = {
      id: "race-1",
      name: "Hakone",
      version: 3,
      description: { type: "distance", race_date: "2026-11-23", leg_distances_km: [5, 3.2] },
      can_submit_laps: false,
    };
    const body: RaceCreate = {
      name: "Hakone",
      description: {
        type: "distance",
        race_date: "2026-11-23",
        deadline: "2026-11-01T18:00:00+09:00",
        leg_distances_km: [5, 3.2],
      },
    };

    it("lists races", async () => {
      const list = { races: [{ id: "race-1", name: "Hakone", race_date: "2026-11-23" }] };
      const { client, requests } = recordingClient(200, list);

      await expect(client.listRaces()).resolves.toEqual(list);
      expect(requests[0].method).toBe("GET");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races");
    });

    it("gets a race", async () => {
      const { client, requests } = recordingClient(200, race);

      await expect(client.getRace("race-1")).resolves.toEqual(race);
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1");
    });

    it("throws not_found for a race the caller may not see", async () => {
      const { client } = recordingClient(404, problemBody(404, "not_found"));

      await expect(client.getRace("hidden")).rejects.toMatchObject({
        status: 404,
        code: "not_found",
      });
    });

    it("creates a race", async () => {
      const { client, requests } = recordingClient(201, race);

      await expect(client.createRace(body)).resolves.toEqual(race);
      expect(requests[0].method).toBe("POST");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races");
      await expect(requests[0].json()).resolves.toEqual(body);
    });

    it("updates a race with the version it read", async () => {
      const { client, requests } = recordingClient(200, race);

      await expect(client.updateRace("race-1", { ...body, version: 3 })).resolves.toEqual(race);
      expect(requests[0].method).toBe("PUT");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1");
      await expect(requests[0].json()).resolves.toEqual({ ...body, version: 3 });
    });

    it("throws version_conflict when someone else saved first", async () => {
      const { client } = recordingClient(409, problemBody(409, "version_conflict"));

      await expect(client.updateRace("race-1", { ...body, version: 2 })).rejects.toMatchObject({
        status: 409,
        code: "version_conflict",
      });
    });
  });

  describe("invitations, attendance, and entries", () => {
    it("lists invitations", async () => {
      const list = {
        members: [{ roster_member_id: "m1", name: "Taro", invited: true, attendance: null }],
      };
      const { client, requests } = recordingClient(200, list);

      await expect(client.listInvitations("race-1")).resolves.toEqual(list);
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1/invitations");
    });

    it("invites a member", async () => {
      const { client, requests } = recordingClient(204);

      await expect(client.invite("race-1", "m1")).resolves.toBeUndefined();
      expect(requests[0].method).toBe("PUT");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1/invitations/m1");
    });

    it("throws validation_failed when inviting a member who left", async () => {
      const { client } = recordingClient(422, problemBody(422, "validation_failed"));

      await expect(client.invite("race-1", "m1")).rejects.toMatchObject({
        code: "validation_failed",
      });
    });

    it("cancels an invitation", async () => {
      const { client, requests } = recordingClient(204);

      await expect(client.cancelInvitation("race-1", "m1")).resolves.toBeUndefined();
      expect(requests[0].method).toBe("DELETE");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1/invitations/m1");
    });

    it("answers attendance", async () => {
      const { client, requests } = recordingClient(200, { status: "attending" });

      await expect(client.answerAttendance("race-1", "attending")).resolves.toBeUndefined();
      expect(requests[0].method).toBe("PUT");
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1/attendance");
      await expect(requests[0].json()).resolves.toEqual({ status: "attending" });
    });

    it("throws not_found after answering not attending", async () => {
      const { client } = recordingClient(404, problemBody(404, "not_found"));

      await expect(client.answerAttendance("race-1", "undecided")).rejects.toMatchObject({
        code: "not_found",
      });
    });

    it("lists entries", async () => {
      const list = { entries: [{ id: "e1", name: "Team A", members: [] }] };
      const { client, requests } = recordingClient(200, list);

      await expect(client.listEntries("race-1")).resolves.toEqual(list);
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1/entries");
    });

    it("creates an entry", async () => {
      const entry = { id: "e1", name: "Team A", members: [] };
      const { client, requests } = recordingClient(201, entry);

      await expect(client.createEntry("race-1", "Team A")).resolves.toEqual(entry);
      expect(requests[0].method).toBe("POST");
      await expect(requests[0].json()).resolves.toEqual({ name: "Team A" });
    });

    it("throws state_conflict when the race already has an entry", async () => {
      const { client } = recordingClient(409, problemBody(409, "state_conflict"));

      await expect(client.createEntry("race-1", "Team B")).rejects.toMatchObject({
        code: "state_conflict",
      });
    });

    it("adds and removes an entry member", async () => {
      const { client, requests } = recordingClient(204);

      await client.addEntryMember("race-1", "e1", "m1");
      await client.removeEntryMember("race-1", "e1", "m1");

      expect(requests.map((request) => request.method)).toEqual(["PUT", "DELETE"]);
      expect(new URL(requests[0].url).pathname).toBe("/api/v1/races/race-1/entries/e1/members/m1");
      expect(new URL(requests[1].url).pathname).toBe("/api/v1/races/race-1/entries/e1/members/m1");
    });
  });
});

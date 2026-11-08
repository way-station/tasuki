import createClient from "openapi-fetch";
import { toApiError } from "./ApiError.ts";
import type { components, paths } from "./schema.gen.ts";

export type Health = components["schemas"]["Health"];
export type InviteLinkInfo = components["schemas"]["InviteLinkInfo"];
export type Me = components["schemas"]["Me"];
export type RaceList = components["schemas"]["RaceList"];
export type Race = components["schemas"]["Race"];
export type RaceCreate = components["schemas"]["RaceCreate"];
export type RaceUpdate = components["schemas"]["RaceUpdate"];
export type InvitationList = components["schemas"]["InvitationList"];
export type AttendanceStatus = components["schemas"]["AttendanceStatus"];
export type EntryList = components["schemas"]["EntryList"];
export type Entry = components["schemas"]["Entry"];
export type MyLaps = components["schemas"]["MyLaps"];
export type RaceLaps = components["schemas"]["RaceLaps"];
export type Profile = components["schemas"]["Profile"];

// Every call the frontend makes to the backend. Hooks depend on this
// interface, never on fetch, so tests can pass a fake (ADR-0012). Each feature
// adds the methods its hooks need. Methods throw ApiError for error responses.
export interface ApiClient {
  // Resolves for 503 too: the body says which store is unreachable.
  getHealth(): Promise<Health>;

  // Auth and joining (docs/design/api.md). The session is an HttpOnly cookie
  // the browser sends and stores; no method returns or takes a token (ADR-0014).
  getInviteLink(token: string): Promise<InviteLinkInfo>;
  // Returns the LINE authorization URL to navigate to.
  authorizeLine(inviteToken?: string): Promise<string>;
  createSession(code: string): Promise<void>;
  logout(): Promise<void>;
  getMe(): Promise<Me>;

  // Races (docs/design/api.md). What each caller gets back differs by role.
  listRaces(): Promise<RaceList>;
  getRace(raceId: string): Promise<Race>;
  createRace(body: RaceCreate): Promise<Race>;
  // Throws version_conflict when the race changed since `body.version` was read.
  updateRace(raceId: string, body: RaceUpdate): Promise<Race>;

  // Invitations, attendance, and entries (docs/design/api.md).
  listInvitations(raceId: string): Promise<InvitationList>;
  invite(raceId: string, rosterMemberId: string): Promise<void>;
  cancelInvitation(raceId: string, rosterMemberId: string): Promise<void>;
  answerAttendance(raceId: string, status: AttendanceStatus): Promise<void>;
  listEntries(raceId: string): Promise<EntryList>;
  // Throws state_conflict when the race already has an entry.
  createEntry(raceId: string, name: string): Promise<Entry>;
  addEntryMember(raceId: string, entryId: string, rosterMemberId: string): Promise<void>;
  removeEntryMember(raceId: string, entryId: string, rosterMemberId: string): Promise<void>;

  // Lap times and the profile (docs/design/api.md). The my-laps calls throw
  // not_found unless the caller is a runner in a confirmed running order of the
  // race; each resolves to the caller's laps and average after the change.
  getMyLaps(raceId: string): Promise<MyLaps>;
  submitLaps(raceId: string, times: string[]): Promise<MyLaps>;
  editLap(raceId: string, lapId: string, time: string): Promise<MyLaps>;
  deleteLap(raceId: string, lapId: string): Promise<MyLaps>;
  // Organizers only.
  listRaceLaps(raceId: string): Promise<RaceLaps>;
  getMyProfile(): Promise<Profile>;
}

type ApiClientOptions = {
  // Defaults to the page's origin: the frontend and the API share one site (ADR-0018).
  baseUrl?: string;
  fetch?: typeof fetch;
};

export function createApiClient(options: ApiClientOptions = {}): ApiClient {
  const client = createClient<paths>(options);

  return {
    async getHealth() {
      const { data, error, response } = await client.GET("/health");
      if (data !== undefined) {
        return data;
      }
      if (response.status === 503) {
        return error;
      }
      throw toApiError(response, error);
    },

    async getInviteLink(token) {
      return dataOrThrow(
        await client.GET("/api/v1/invite-links/{token}", { params: { path: { token } } }),
      );
    },

    async authorizeLine(inviteToken) {
      const body = inviteToken === undefined ? {} : { invite_token: inviteToken };
      const result = dataOrThrow(await client.POST("/api/v1/auth/line/authorize", { body }));
      return result.authorization_url;
    },

    async createSession(code) {
      dataOrThrow(await client.POST("/api/v1/session", { body: { code } }));
    },

    async logout() {
      dataOrThrow(await client.DELETE("/api/v1/session"));
    },

    async getMe() {
      return dataOrThrow(await client.GET("/api/v1/me"));
    },

    async listRaces() {
      return dataOrThrow(await client.GET("/api/v1/races"));
    },

    async getRace(raceId) {
      return dataOrThrow(
        await client.GET("/api/v1/races/{race_id}", { params: { path: { race_id: raceId } } }),
      );
    },

    async createRace(body) {
      return dataOrThrow(await client.POST("/api/v1/races", { body }));
    },

    async updateRace(raceId, body) {
      return dataOrThrow(
        await client.PUT("/api/v1/races/{race_id}", {
          params: { path: { race_id: raceId } },
          body,
        }),
      );
    },

    async listInvitations(raceId) {
      return dataOrThrow(
        await client.GET("/api/v1/races/{race_id}/invitations", {
          params: { path: { race_id: raceId } },
        }),
      );
    },

    async invite(raceId, rosterMemberId) {
      dataOrThrow(
        await client.PUT("/api/v1/races/{race_id}/invitations/{roster_member_id}", {
          params: { path: { race_id: raceId, roster_member_id: rosterMemberId } },
        }),
      );
    },

    async cancelInvitation(raceId, rosterMemberId) {
      dataOrThrow(
        await client.DELETE("/api/v1/races/{race_id}/invitations/{roster_member_id}", {
          params: { path: { race_id: raceId, roster_member_id: rosterMemberId } },
        }),
      );
    },

    async answerAttendance(raceId, status) {
      dataOrThrow(
        await client.PUT("/api/v1/races/{race_id}/attendance", {
          params: { path: { race_id: raceId } },
          body: { status },
        }),
      );
    },

    async listEntries(raceId) {
      return dataOrThrow(
        await client.GET("/api/v1/races/{race_id}/entries", {
          params: { path: { race_id: raceId } },
        }),
      );
    },

    async createEntry(raceId, name) {
      return dataOrThrow(
        await client.POST("/api/v1/races/{race_id}/entries", {
          params: { path: { race_id: raceId } },
          body: { name },
        }),
      );
    },

    async addEntryMember(raceId, entryId, rosterMemberId) {
      dataOrThrow(
        await client.PUT("/api/v1/races/{race_id}/entries/{entry_id}/members/{roster_member_id}", {
          params: {
            path: { race_id: raceId, entry_id: entryId, roster_member_id: rosterMemberId },
          },
        }),
      );
    },

    async removeEntryMember(raceId, entryId, rosterMemberId) {
      dataOrThrow(
        await client.DELETE(
          "/api/v1/races/{race_id}/entries/{entry_id}/members/{roster_member_id}",
          {
            params: {
              path: { race_id: raceId, entry_id: entryId, roster_member_id: rosterMemberId },
            },
          },
        ),
      );
    },

    async getMyLaps(raceId) {
      return dataOrThrow(
        await client.GET("/api/v1/races/{race_id}/my-laps", {
          params: { path: { race_id: raceId } },
        }),
      );
    },

    async submitLaps(raceId, times) {
      return dataOrThrow(
        await client.POST("/api/v1/races/{race_id}/my-laps", {
          params: { path: { race_id: raceId } },
          body: { times },
        }),
      );
    },

    async editLap(raceId, lapId, time) {
      return dataOrThrow(
        await client.PATCH("/api/v1/races/{race_id}/my-laps/{lap_id}", {
          params: { path: { race_id: raceId, lap_id: lapId } },
          body: { time },
        }),
      );
    },

    async deleteLap(raceId, lapId) {
      return dataOrThrow(
        await client.DELETE("/api/v1/races/{race_id}/my-laps/{lap_id}", {
          params: { path: { race_id: raceId, lap_id: lapId } },
        }),
      );
    },

    async listRaceLaps(raceId) {
      return dataOrThrow(
        await client.GET("/api/v1/races/{race_id}/laps", {
          params: { path: { race_id: raceId } },
        }),
      );
    },

    async getMyProfile() {
      return dataOrThrow(await client.GET("/api/v1/me/profile"));
    },
  };
}

type Result<T> = { data?: T; error?: unknown; response: Response };

function dataOrThrow<T>({ data, error, response }: Result<T>): T {
  if (!response.ok) {
    throw toApiError(response, error);
  }
  return data as T;
}

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
  };
}

type Result<T> = { data?: T; error?: unknown; response: Response };

function dataOrThrow<T>({ data, error, response }: Result<T>): T {
  if (!response.ok) {
    throw toApiError(response, error);
  }
  return data as T;
}

import createClient from "openapi-fetch";
import { toApiError } from "./ApiError.ts";
import type { components, paths } from "./schema.gen.ts";

export type Health = components["schemas"]["Health"];

// Every call the frontend makes to the backend. Hooks depend on this
// interface, never on fetch, so tests can pass a fake (ADR-0012). Each feature
// adds the methods its hooks need. Methods throw ApiError for error responses.
export interface ApiClient {
  // Resolves for 503 too: the body says which store is unreachable.
  getHealth(): Promise<Health>;
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
  };
}

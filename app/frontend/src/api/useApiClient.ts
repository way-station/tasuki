import { use } from "react";
import { ApiClientContext } from "./apiClientContext.ts";
import type { ApiClient } from "./client.ts";

export function useApiClient(): ApiClient {
  const client = use(ApiClientContext);
  if (client === null) {
    throw new Error("useApiClient must be used inside ApiClientProvider");
  }
  return client;
}

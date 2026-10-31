import type { ReactNode } from "react";
import { ApiClientContext } from "./apiClientContext.ts";
import type { ApiClient } from "./client.ts";

type ApiClientProviderProps = {
  client: ApiClient;
  children: ReactNode;
};

export function ApiClientProvider({ client, children }: ApiClientProviderProps) {
  return <ApiClientContext value={client}>{children}</ApiClientContext>;
}

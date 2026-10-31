import { type QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";
import { ApiClientProvider } from "../api/ApiClientProvider.tsx";
import type { ApiClient } from "../api/client.ts";
import { createTestQueryClient } from "./testQueryClient.ts";

type TestProvidersProps = {
  client: ApiClient;
  // Pass one to inspect the cache; createTestQueryClient() makes it.
  queryClient?: QueryClient;
  children: ReactNode;
};

// The providers the app has, with a fake ApiClient and a fresh cache per test.
export function TestProviders({ client, queryClient, children }: TestProvidersProps) {
  const [defaultQueryClient] = useState(createTestQueryClient);

  return (
    <QueryClientProvider client={queryClient ?? defaultQueryClient}>
      <ApiClientProvider client={client}>{children}</ApiClientProvider>
    </QueryClientProvider>
  );
}

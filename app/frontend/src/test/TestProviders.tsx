import { QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";
import { ApiClientProvider } from "../api/ApiClientProvider.tsx";
import type { ApiClient } from "../api/client.ts";
import { createQueryClient } from "../api/queryClient.ts";

type TestProvidersProps = {
  client: ApiClient;
  children: ReactNode;
};

// The providers the app has, with a fake ApiClient and a fresh cache per test.
// Retries keep the app's rules but do not wait, so tests stay fast.
export function TestProviders({ client, children }: TestProvidersProps) {
  const [queryClient] = useState(() => {
    const created = createQueryClient();
    created.setDefaultOptions({
      queries: { ...created.getDefaultOptions().queries, retryDelay: 0 },
    });
    return created;
  });

  return (
    <QueryClientProvider client={queryClient}>
      <ApiClientProvider client={client}>{children}</ApiClientProvider>
    </QueryClientProvider>
  );
}

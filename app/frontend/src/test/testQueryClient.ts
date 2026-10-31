import type { QueryClient } from "@tanstack/react-query";
import { createQueryClient } from "../api/queryClient.ts";

// The app's QueryClient, with retries that keep the app's rules but do not
// wait, so tests stay fast.
export function createTestQueryClient(): QueryClient {
  const queryClient = createQueryClient();
  queryClient.setDefaultOptions({
    queries: { ...queryClient.getDefaultOptions().queries, retryDelay: 0 },
  });
  return queryClient;
}

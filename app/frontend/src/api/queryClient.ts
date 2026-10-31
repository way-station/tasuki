import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
import { ApiError } from "./ApiError.ts";

// TanStack Query's default number of retries.
const maxRetries = 3;

// Requests that fail with a 4xx status would fail the same way again (ADR-0013).
function shouldRetry(failureCount: number, error: Error): boolean {
  if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
    return false;
  }
  return failureCount < maxRetries;
}

export function createQueryClient(): QueryClient {
  // A 401 means the session ended. Every query that holds data is removed, so
  // nothing the previous user could see stays, and the route guard (#113)
  // picks the next screen (ADR-0013). Handled here once, not in each screen.
  //
  // Queries without data stay. A screen fetches a removed query again at
  // once, so removing a query that keeps failing with 401 would fetch it
  // again and again.
  const clearOnUnauthenticated = (error: Error) => {
    if (error instanceof ApiError && error.status === 401) {
      queryClient.removeQueries({ predicate: (query) => query.state.data !== undefined });
    }
  };

  const queryClient = new QueryClient({
    queryCache: new QueryCache({ onError: clearOnUnauthenticated }),
    mutationCache: new MutationCache({ onError: clearOnUnauthenticated }),
    defaultOptions: {
      queries: {
        retry: shouldRetry,
        // Fetch again whenever a screen opens or the window regains focus, so
        // changes made by others show up right away.
        staleTime: 0,
      },
    },
  });
  return queryClient;
}

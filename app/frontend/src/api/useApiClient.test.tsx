import { QueryClientProvider, useQuery } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { ApiClientProvider } from "./ApiClientProvider.tsx";
import type { ApiClient, Health } from "./client.ts";
import { createFakeApiClient } from "./fakeApiClient.ts";
import { createQueryClient } from "./queryClient.ts";
import { queryKeys } from "./queryKeys.ts";
import { useApiClient } from "./useApiClient.ts";

function wrapperWith(client: ApiClient) {
  const queryClient = createQueryClient();
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <ApiClientProvider client={client}>{children}</ApiClientProvider>
    </QueryClientProvider>
  );
}

describe("useApiClient", () => {
  it("lets a hook call the backend through the client it is given", async () => {
    const health: Health = { status: "ok", mongo: "ok", postgres: "ok" };
    const client = createFakeApiClient({ getHealth: () => Promise.resolve(health) });

    const { result } = renderHook(
      () => {
        const api = useApiClient();
        return useQuery({ queryKey: queryKeys.health, queryFn: () => api.getHealth() });
      },
      { wrapper: wrapperWith(client) },
    );

    await waitFor(() => expect(result.current.data).toEqual(health));
  });

  it("fails outside ApiClientProvider", () => {
    expect(() => renderHook(() => useApiClient())).toThrow("inside ApiClientProvider");
  });
});

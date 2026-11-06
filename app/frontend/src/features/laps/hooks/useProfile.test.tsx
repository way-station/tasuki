import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { ApiClient } from "../../../api/client.ts";
import { createFakeApiClient } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useProfile } from "./useProfile.ts";

function renderWith<T>(hook: () => T, client: ApiClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={createTestQueryClient()}>
      {children}
    </TestProviders>
  );
  return renderHook(hook, { wrapper });
}

describe("useProfile", () => {
  it("gets the caller's profile", async () => {
    const { result } = renderWith(
      () => useProfile(),
      createFakeApiClient({
        getMyProfile: () => Promise.resolve({ name: "Taro", average_time: "3:40", lap_count: 0 }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ name: "Taro", averageTime: "3:40", lapCount: 0 });
  });
});

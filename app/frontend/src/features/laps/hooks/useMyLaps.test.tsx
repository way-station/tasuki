import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { ApiClient } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useMyLaps } from "./useMyLaps.ts";

function renderWith<T>(hook: () => T, client: ApiClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={createTestQueryClient()}>
      {children}
    </TestProviders>
  );
  return renderHook(hook, { wrapper });
}

describe("useMyLaps", () => {
  it("gets the caller's laps and average", async () => {
    const { result } = renderWith(
      () => useMyLaps("race-1"),
      createFakeApiClient({
        getMyLaps: () =>
          Promise.resolve({
            laps: [{ id: "l1", time: "3:25" }],
            average_time: "3:25",
            lap_count: 1,
          }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({
      laps: [{ id: "l1", time: "3:25" }],
      averageTime: "3:25",
      lapCount: 1,
    });
  });

  it("reports not_found when the caller is not a runner", async () => {
    const { result } = renderWith(
      () => useMyLaps("race-1"),
      createFakeApiClient({ getMyLaps: () => Promise.reject(problem(404, "not_found")) }),
    );

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "not_found" }));
  });
});

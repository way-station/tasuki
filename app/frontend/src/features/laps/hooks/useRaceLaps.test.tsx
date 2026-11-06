import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient } from "../../../api/client.ts";
import { createFakeApiClient } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useRaceLaps } from "./useRaceLaps.ts";

function renderWith<T>(hook: () => T, client: ApiClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={createTestQueryClient()}>
      {children}
    </TestProviders>
  );
  return renderHook(hook, { wrapper });
}

describe("useRaceLaps", () => {
  it("gets every runner's laps for an organizer", async () => {
    const { result } = renderWith(
      () => useRaceLaps("race-1", true),
      createFakeApiClient({
        listRaceLaps: () =>
          Promise.resolve({ runners: [{ roster_member_id: "m1", name: "Taro", laps: [] }] }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([{ rosterMemberId: "m1", name: "Taro", laps: [] }]);
  });

  it("asks nothing for a member", () => {
    const listRaceLaps = vi.fn(() => Promise.resolve({ runners: [] }));
    const { result } = renderWith(
      () => useRaceLaps("race-1", false),
      createFakeApiClient({ listRaceLaps }),
    );

    expect(result.current.fetchStatus).toBe("idle");
    expect(listRaceLaps).not.toHaveBeenCalled();
  });
});

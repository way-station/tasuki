import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, RunningOrder } from "../../../api/client.ts";
import { createFakeApiClient } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useRunningOrders } from "./useRunningOrders.ts";

function renderWith<T>(hook: () => T, client: ApiClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={createTestQueryClient()}>
      {children}
    </TestProviders>
  );
  return renderHook(hook, { wrapper });
}

const confirmed: RunningOrder = {
  entry_id: "e1",
  entry_name: "Team A",
  status: "confirmed",
  legs: [{ leg_number: 1, runner: { roster_member_id: "m1", name: "Taro" } }],
  prediction: { total_time: null, laps: 25, has_empty_legs: false },
};

describe("useRunningOrders", () => {
  it("lists the running orders the caller may see", async () => {
    const { result } = renderWith(
      () => useRunningOrders("race-1", true),
      createFakeApiClient({
        listRunningOrders: () => Promise.resolve({ running_orders: [confirmed] }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.map((order) => [order.status, order.prediction])).toEqual([
      ["confirmed", { type: "time", laps: 25, hasEmptyLegs: false }],
    ]);
  });

  it("asks nothing for a caller who may not see running orders", () => {
    const listRunningOrders = vi.fn(() => Promise.resolve({ running_orders: [] }));
    const { result } = renderWith(
      () => useRunningOrders("race-1", false),
      createFakeApiClient({ listRunningOrders }),
    );

    expect(result.current.fetchStatus).toBe("idle");
    expect(listRunningOrders).not.toHaveBeenCalled();
  });
});

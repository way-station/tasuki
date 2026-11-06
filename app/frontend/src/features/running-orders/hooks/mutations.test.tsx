import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, RunningOrder } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { queryKeys } from "../../../api/queryKeys.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useConfirmRunningOrder, useSaveRunningOrder } from "./mutations.ts";

// A cache holding the race list and the race's running orders, as on R7.
function cachedRace() {
  const queryClient = createTestQueryClient();
  queryClient.setQueryData(queryKeys.races, { races: [] });
  queryClient.setQueryData(queryKeys.runningOrders("race-1"), { running_orders: [] });
  return queryClient;
}

function invalidated(queryClient: ReturnType<typeof createTestQueryClient>) {
  return [queryKeys.races, queryKeys.runningOrders("race-1")].map(
    (key) => queryClient.getQueryState(key)?.isInvalidated ?? false,
  );
}

function renderWith<T>(hook: () => T, client: ApiClient, queryClient = cachedRace()) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={queryClient}>
      {children}
    </TestProviders>
  );
  return { ...renderHook(hook, { wrapper }), queryClient };
}

const order: RunningOrder = {
  entry_id: "e1",
  entry_name: "Team A",
  status: "draft",
  version: 2,
  legs: [{ leg_number: 1, runner: { roster_member_id: "m1", name: "Taro" } }],
  prediction: null,
};

describe("useSaveRunningOrder", () => {
  it("sends the first draft without a version and fetches the race again", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order));
    const { result, queryClient } = renderWith(
      () => useSaveRunningOrder("race-1"),
      createFakeApiClient({ saveRunningOrder }),
    );

    const saved = await result.current.mutateAsync({
      entryId: "e1",
      form: { runnerIds: ["m1"] },
      version: 0,
    });

    expect(saveRunningOrder).toHaveBeenCalledWith("race-1", "e1", {
      legs: [{ leg_number: 1, roster_member_id: "m1" }],
    });
    expect(saved.version).toBe(2);
    expect(invalidated(queryClient)).toEqual([true, true]);
  });

  it("sends the version the edit started from", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order));
    const { result } = renderWith(
      () => useSaveRunningOrder("race-1"),
      createFakeApiClient({ saveRunningOrder }),
    );

    await result.current.mutateAsync({ entryId: "e1", form: { runnerIds: [null] }, version: 1 });

    expect(saveRunningOrder).toHaveBeenCalledWith("race-1", "e1", {
      version: 1,
      legs: [{ leg_number: 1, roster_member_id: null }],
    });
  });

  it("fetches the running order again after version_conflict", async () => {
    const { result, queryClient } = renderWith(
      () => useSaveRunningOrder("race-1"),
      createFakeApiClient({
        saveRunningOrder: () => Promise.reject(problem(409, "version_conflict")),
      }),
    );

    result.current.mutate({ entryId: "e1", form: { runnerIds: ["m1"] }, version: 1 });

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "version_conflict" }));
    expect(invalidated(queryClient)).toEqual([true, true]);
  });
});

describe("useConfirmRunningOrder", () => {
  it("confirms with the version and fetches the race again", async () => {
    const confirmRunningOrder = vi.fn(() =>
      Promise.resolve({ ...order, status: "confirmed" as const }),
    );
    const { result, queryClient } = renderWith(
      () => useConfirmRunningOrder("race-1"),
      createFakeApiClient({ confirmRunningOrder }),
    );

    const confirmed = await result.current.mutateAsync({ entryId: "e1", version: 2 });

    expect(confirmRunningOrder).toHaveBeenCalledWith("race-1", "e1", 2);
    expect(confirmed.status).toBe("confirmed");
    expect(invalidated(queryClient)).toEqual([true, true]);
  });

  it.each([
    [422, "validation_failed"],
    [409, "state_conflict"],
  ] as const)("reports %i %s", async (status, code) => {
    const { result } = renderWith(
      () => useConfirmRunningOrder("race-1"),
      createFakeApiClient({ confirmRunningOrder: () => Promise.reject(problem(status, code)) }),
    );

    result.current.mutate({ entryId: "e1", version: 2 });

    await waitFor(() => expect(result.current.error).toMatchObject({ code }));
  });
});

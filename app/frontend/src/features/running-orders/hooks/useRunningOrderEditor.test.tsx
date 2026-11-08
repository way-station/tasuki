import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { ApiClient, Race, RunningOrder } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useRunningOrderEditor } from "./useRunningOrderEditor.ts";

function renderWith<T>(hook: () => T, client: ApiClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={createTestQueryClient()}>
      {children}
    </TestProviders>
  );
  return renderHook(hook, { wrapper });
}

const race: Race = {
  id: "race-1",
  name: "Hakone",
  version: 1,
  description: { type: "distance", race_date: "2026-11-23", leg_distances_km: [5, 3] },
  can_submit_laps: false,
};
const entries = {
  entries: [{ id: "e1", name: "Team A", members: [{ roster_member_id: "m1", name: "Taro" }] }],
};
const saved: RunningOrder = {
  entry_id: "e1",
  entry_name: "Team A",
  status: "draft",
  version: 4,
  legs: [
    { leg_number: 1, distance_km: 5, runner: { roster_member_id: "m1", name: "Taro" } },
    { leg_number: 2, distance_km: 3, runner: null },
  ],
  prediction: null,
};

function client(runningOrders: RunningOrder[], entryList = entries) {
  return createFakeApiClient({
    getRace: () => Promise.resolve(race),
    listEntries: () => Promise.resolve(entryList),
    listRunningOrders: () => Promise.resolve({ running_orders: runningOrders }),
  });
}

describe("useRunningOrderEditor", () => {
  it("starts the first draft from empty legs at version 0", async () => {
    const { result } = renderWith(() => useRunningOrderEditor("race-1"), client([]));

    await waitFor(() => expect(result.current.data).toBeDefined());
    expect(result.current.data?.source).toEqual({ value: { runnerIds: [null, null] }, version: 0 });
  });

  it("starts from the saved running order and its version", async () => {
    const { result } = renderWith(() => useRunningOrderEditor("race-1"), client([saved]));

    await waitFor(() => expect(result.current.data).toBeDefined());
    expect(result.current.data?.source).toEqual({ value: { runnerIds: ["m1", null] }, version: 4 });
  });

  it("has no form while the race has no entry", async () => {
    const { result } = renderWith(
      () => useRunningOrderEditor("race-1"),
      client([], { entries: [] }),
    );

    await waitFor(() => expect(result.current.data).toBeDefined());
    expect(result.current.data).toEqual({ editor: { entry: null }, source: undefined });
  });

  it("reports an error from any of the three queries", async () => {
    const { result } = renderWith(
      () => useRunningOrderEditor("race-1"),
      createFakeApiClient({
        getRace: () => Promise.resolve(race),
        listEntries: () => Promise.reject(problem(403, "forbidden")),
        listRunningOrders: () => Promise.resolve({ running_orders: [] }),
      }),
    );

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "forbidden" }));
  });
});

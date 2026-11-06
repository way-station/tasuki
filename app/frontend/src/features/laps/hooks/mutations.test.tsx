import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, MyLaps } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { queryKeys } from "../../../api/queryKeys.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { lapsFormErrors, lapTimeError } from "./lapsFormInvalid.ts";
import { useDeleteLap, useEditLap, useSubmitLaps } from "./mutations.ts";

// A cache holding what a change of laps affects.
function cachedLaps() {
  const queryClient = createTestQueryClient();
  queryClient.setQueryData(queryKeys.myLaps("race-1"), myLaps);
  queryClient.setQueryData(queryKeys.raceLaps("race-1"), { runners: [] });
  queryClient.setQueryData(queryKeys.profile, { name: "Taro", average_time: "3:40", lap_count: 0 });
  return queryClient;
}

function invalidated(queryClient: ReturnType<typeof createTestQueryClient>) {
  return [queryKeys.myLaps("race-1"), queryKeys.raceLaps("race-1"), queryKeys.profile].map(
    (key) => queryClient.getQueryState(key)?.isInvalidated ?? false,
  );
}

function renderWith<T>(hook: () => T, client: ApiClient) {
  const queryClient = cachedLaps();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={queryClient}>
      {children}
    </TestProviders>
  );
  return { ...renderHook(hook, { wrapper }), queryClient };
}

const myLaps: MyLaps = { laps: [{ id: "l1", time: "3:25" }], average_time: "3:25", lap_count: 1 };

describe("useSubmitLaps", () => {
  it("sends the times together and fetches the laps and profile again", async () => {
    const submitLaps = vi.fn(() => Promise.resolve(myLaps));
    const { result, queryClient } = renderWith(
      () => useSubmitLaps("race-1"),
      createFakeApiClient({ submitLaps }),
    );

    const laps = await result.current.mutateAsync({ times: ["3:25 ", "3:30"] });

    expect(submitLaps).toHaveBeenCalledWith("race-1", ["3:25", "3:30"]);
    expect(laps.averageTime).toBe("3:25");
    expect(invalidated(queryClient)).toEqual([true, true, true]);
  });

  it("sends nothing for an invalid time", async () => {
    const submitLaps = vi.fn(() => Promise.resolve(myLaps));
    const { result } = renderWith(
      () => useSubmitLaps("race-1"),
      createFakeApiClient({ submitLaps }),
    );

    result.current.mutate({ times: ["3:5"] });

    await waitFor(() => expect(lapsFormErrors(result.current.error)).toHaveProperty("times.0"));
    expect(submitLaps).not.toHaveBeenCalled();
  });

  it("reports not_found when the caller is no longer a runner", async () => {
    const { result, queryClient } = renderWith(
      () => useSubmitLaps("race-1"),
      createFakeApiClient({ submitLaps: () => Promise.reject(problem(404, "not_found")) }),
    );

    result.current.mutate({ times: ["3:25"] });

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "not_found" }));
    expect(invalidated(queryClient)).toEqual([true, true, true]);
  });
});

describe("useEditLap", () => {
  it("sends the edited time", async () => {
    const editLap = vi.fn(() => Promise.resolve(myLaps));
    const { result } = renderWith(() => useEditLap("race-1"), createFakeApiClient({ editLap }));

    await result.current.mutateAsync({ lapId: "l1", time: " 3:20" });

    expect(editLap).toHaveBeenCalledWith("race-1", "l1", "3:20");
  });

  it("sends nothing for an invalid time", async () => {
    const editLap = vi.fn(() => Promise.resolve(myLaps));
    const { result } = renderWith(() => useEditLap("race-1"), createFakeApiClient({ editLap }));

    result.current.mutate({ lapId: "l1", time: "3.20" });

    await waitFor(() => expect(lapTimeError(result.current.error)).toBeDefined());
    expect(editLap).not.toHaveBeenCalled();
  });
});

describe("useDeleteLap", () => {
  it("deletes the lap and fetches the laps and profile again", async () => {
    const deleteLap = vi.fn(() => Promise.resolve({ ...myLaps, laps: [], lap_count: 0 }));
    const { result, queryClient } = renderWith(
      () => useDeleteLap("race-1"),
      createFakeApiClient({ deleteLap }),
    );

    const laps = await result.current.mutateAsync("l1");

    expect(deleteLap).toHaveBeenCalledWith("race-1", "l1");
    expect(laps.lapCount).toBe(0);
    expect(invalidated(queryClient)).toEqual([true, true, true]);
  });
});

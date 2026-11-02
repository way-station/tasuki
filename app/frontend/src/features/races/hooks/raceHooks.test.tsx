import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, Race } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { queryKeys } from "../../../api/queryKeys.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import type { RaceForm } from "../types/race.ts";
import { raceFormErrors } from "./raceFormInvalid.ts";
import { useCreateRace } from "./useCreateRace.ts";
import { useRace } from "./useRace.ts";
import { useRaceFormSource } from "./useRaceFormSource.ts";
import { useRaces } from "./useRaces.ts";
import { useUpdateRace } from "./useUpdateRace.ts";

function renderWith<T>(hook: () => T, client: ApiClient, queryClient = createTestQueryClient()) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={queryClient}>
      {children}
    </TestProviders>
  );
  return { ...renderHook(hook, { wrapper }), queryClient };
}

// A cache holding the list and the race, as after opening R1 and R2.
function cachedRaces() {
  const queryClient = createTestQueryClient();
  queryClient.setQueryData(queryKeys.races, { races: [] });
  queryClient.setQueryData(queryKeys.race("race-1"), race);
  return queryClient;
}

function isInvalidated(
  queryClient: ReturnType<typeof createTestQueryClient>,
  key: readonly unknown[],
) {
  return queryClient.getQueryState(key)?.isInvalidated ?? false;
}

const race: Race = {
  id: "race-1",
  name: "Hakone",
  version: 3,
  description: {
    type: "distance",
    race_date: "2026-11-23",
    deadline: "2026-11-01T18:00:00+09:00",
    leg_distances_km: [5, 3.2],
  },
  can_submit_laps: false,
};

const validForm: RaceForm = {
  name: "Hakone",
  type: "distance",
  raceDate: "2026-11-23",
  deadline: "2026-11-01T18:00",
  text: "",
  legDistancesKm: ["5", "3.2"],
  targetTotalTime: "",
  timeLimitMinutes: "",
  lapDistanceKm: "",
  targetLaps: "",
};

describe("useRaces", () => {
  const list = {
    races: [
      {
        id: "race-1",
        name: "Hakone",
        race_date: "2026-11-23",
        my_attendance: "attending" as const,
      },
      { id: "race-2", name: "Ekiden", race_date: "2026-12-05", my_attendance: null },
    ],
  };

  it("lays out a member's races in My races and Invitations", async () => {
    const { result } = renderWith(
      () => useRaces(false),
      createFakeApiClient({ listRaces: () => Promise.resolve(list) }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toMatchObject({
      role: "member",
      myRaces: [{ id: "race-1" }],
      invitations: [{ id: "race-2", myAttendance: "not_answered" }],
    });
  });

  it("lays out every race for an organizer", async () => {
    const { result } = renderWith(
      () => useRaces(true),
      createFakeApiClient({ listRaces: () => Promise.resolve(list) }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toMatchObject({
      role: "organizer",
      races: [{ id: "race-1" }, { id: "race-2" }],
    });
  });
});

describe("useRace", () => {
  it("returns the race", async () => {
    const getRace = vi.fn(() => Promise.resolve(race));
    const { result } = renderWith(() => useRace("race-1"), createFakeApiClient({ getRace }));

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toMatchObject({ id: "race-1", name: "Hakone", version: 3 });
    expect(getRace).toHaveBeenCalledWith("race-1");
  });

  it("fails with not_found once, without retrying, for a race the caller may not see", async () => {
    const getRace = vi.fn(() => Promise.reject(problem(404, "not_found")));
    const { result } = renderWith(() => useRace("hidden"), createFakeApiClient({ getRace }));

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toMatchObject({ code: "not_found" });
    expect(getRace).toHaveBeenCalledTimes(1);
  });
});

describe("useCreateRace", () => {
  it("registers the race from the form, resolves to it, and fetches the list again", async () => {
    const createRace = vi.fn(() => Promise.resolve(race));
    const { result, queryClient } = renderWith(
      useCreateRace,
      createFakeApiClient({ createRace }),
      cachedRaces(),
    );

    await expect(result.current.mutateAsync(validForm)).resolves.toMatchObject({ id: "race-1" });
    expect(createRace).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Hakone",
        description: expect.objectContaining({ leg_distances_km: [5, 3.2] }),
      }),
    );
    expect(isInvalidated(queryClient, queryKeys.races)).toBe(true);
  });

  it("refuses an invalid form without sending it", async () => {
    const createRace = vi.fn(() => Promise.resolve(race));
    const { result } = renderWith(useCreateRace, createFakeApiClient({ createRace }));

    const error = await result.current
      .mutateAsync({ ...validForm, name: "" })
      .catch((caught: unknown) => caught);

    expect(raceFormErrors(error)).toEqual({ name: "Enter a name." });
    expect(createRace).not.toHaveBeenCalled();
  });

  it("fails with validation_failed and keeps the list", async () => {
    const { result, queryClient } = renderWith(
      useCreateRace,
      createFakeApiClient({ createRace: () => Promise.reject(problem(422, "validation_failed")) }),
      cachedRaces(),
    );

    await expect(result.current.mutateAsync(validForm)).rejects.toMatchObject({
      code: "validation_failed",
    });
    expect(isInvalidated(queryClient, queryKeys.races)).toBe(false);
  });
});

describe("useUpdateRace", () => {
  it("saves the race with the version it started from, and fetches races again", async () => {
    const updateRace = vi.fn(() => Promise.resolve({ ...race, version: 4 }));
    const { result, queryClient } = renderWith(
      () => useUpdateRace("race-1"),
      createFakeApiClient({ updateRace }),
      cachedRaces(),
    );

    await expect(
      result.current.mutateAsync({ form: validForm, version: 3 }),
    ).resolves.toMatchObject({ version: 4 });
    expect(updateRace).toHaveBeenCalledWith("race-1", expect.objectContaining({ version: 3 }));
    expect(isInvalidated(queryClient, queryKeys.races)).toBe(true);
    expect(isInvalidated(queryClient, queryKeys.race("race-1"))).toBe(true);
  });

  it("refuses an invalid form without sending it", async () => {
    const updateRace = vi.fn(() => Promise.resolve(race));
    const { result } = renderWith(
      () => useUpdateRace("race-1"),
      createFakeApiClient({ updateRace }),
    );

    const error = await result.current
      .mutateAsync({ form: { ...validForm, deadline: "" }, version: 3 })
      .catch((caught: unknown) => caught);

    expect(raceFormErrors(error)).toEqual({ deadline: "Enter the deadline." });
    expect(updateRace).not.toHaveBeenCalled();
  });

  it("fetches the race again when someone else saved first", async () => {
    const { result, queryClient } = renderWith(
      () => useUpdateRace("race-1"),
      createFakeApiClient({ updateRace: () => Promise.reject(problem(409, "version_conflict")) }),
      cachedRaces(),
    );

    await expect(result.current.mutateAsync({ form: validForm, version: 2 })).rejects.toMatchObject(
      {
        code: "version_conflict",
      },
    );
    await waitFor(() => expect(isInvalidated(queryClient, queryKeys.race("race-1"))).toBe(true));
  });

  it("keeps the race for other errors", async () => {
    const { result, queryClient } = renderWith(
      () => useUpdateRace("race-1"),
      createFakeApiClient({ updateRace: () => Promise.reject(problem(422, "validation_failed")) }),
      cachedRaces(),
    );

    await expect(result.current.mutateAsync({ form: validForm, version: 3 })).rejects.toMatchObject(
      {
        code: "validation_failed",
      },
    );
    expect(isInvalidated(queryClient, queryKeys.race("race-1"))).toBe(false);
  });
});

describe("useRaceFormSource", () => {
  it("starts registering from an empty form without fetching", () => {
    const getRace = vi.fn(() => Promise.resolve(race));
    const { result } = renderWith(
      () => useRaceFormSource(undefined),
      createFakeApiClient({ getRace }),
    );

    expect(result.current.source).toMatchObject({
      version: 0,
      value: { name: "", legDistancesKm: [""] },
    });
    expect(getRace).not.toHaveBeenCalled();
  });

  it("starts editing from the saved race and its version", async () => {
    const { result } = renderWith(
      () => useRaceFormSource("race-1"),
      createFakeApiClient({ getRace: () => Promise.resolve(race) }),
    );

    await waitFor(() => expect(result.current.source).toBeDefined());
    expect(result.current.source).toMatchObject({
      version: 3,
      value: { name: "Hakone", type: "distance", legDistancesKm: ["5", "3.2"] },
    });
  });

  it("fails with not_found for a race that does not exist", async () => {
    const { result } = renderWith(
      () => useRaceFormSource("missing"),
      createFakeApiClient({ getRace: () => Promise.reject(problem(404, "not_found")) }),
    );

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "not_found" }));
    expect(result.current.source).toBeUndefined();
  });
});

import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { render, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ApiClient, Me, MyLaps, Race } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { createAppRouter } from "../../../app/router.tsx";
import { TestProviders } from "../../../test/TestProviders.tsx";

afterEach(() => {
  localStorage.clear();
});

const member: Me = {
  user_id: "user-1",
  display_name: "Taro",
  is_organizer: false,
  status: "approved",
  roster_member: { id: "m1", name: "Taro" },
};
const organizer: Me = { ...member, is_organizer: true, roster_member: null };

function renderAt(path: string, me: Me, stubs: Partial<ApiClient>) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  const client = createFakeApiClient({ getMe: () => Promise.resolve(me), ...stubs });
  render(
    <TestProviders client={client}>
      <RouterProvider router={router} />
    </TestProviders>,
  );
  return router;
}

const race: Race = {
  id: "race-1",
  name: "Hakone",
  description: { type: "distance", race_date: "2026-11-23", leg_distances_km: [5, 3] },
  can_submit_laps: true,
};

function myLaps(times: string[], averageTime = "3:28"): MyLaps {
  return {
    laps: times.map((time, index) => ({ id: `lap-${index + 1}`, time })),
    average_time: averageTime,
    lap_count: times.length,
  };
}

describe("R2 lap times", () => {
  it("shows an organizer every runner's laps", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve({ ...race, version: 1, can_submit_laps: false }),
      listRaceLaps: () =>
        Promise.resolve({
          runners: [
            {
              roster_member_id: "m1",
              name: "Taro",
              laps: [
                { id: "lap-1", time: "3:20" },
                { id: "lap-2", time: "3:25" },
              ],
            },
            { roster_member_id: "m2", name: "Jiro", laps: [] },
          ],
        }),
    });

    const section = await screen.findByRole("region", { name: "Lap times" });
    expect(
      within(section)
        .getAllByRole("row")
        .map((row) => row.textContent),
    ).toEqual(["RunnerLaps (per km)", "Taro3:20, 3:25", "JiroNone yet"]);
  });

  it("asks nothing for a member", async () => {
    const listRaceLaps = vi.fn(() => Promise.resolve({ runners: [] }));
    renderAt("/races/race-1", member, { getRace: () => Promise.resolve(race), listRaceLaps });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(listRaceLaps).not.toHaveBeenCalled();
  });
});

describe("L1 my lap times", () => {
  it("shows the laps in the order they were submitted, with the average", async () => {
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps(["3:20", "3:25"], "3:23")),
    });

    const list = await screen.findByRole("list", { name: "Submitted lap times" });
    expect(
      within(list)
        .getAllByRole("listitem")
        .map((item) => item.firstChild?.textContent),
    ).toEqual(["Lap 1: 3:20", "Lap 2: 3:25"]);
    expect(screen.getByText("3:23 /km")).toBeInTheDocument();
  });

  it("tells a member who does not run in a confirmed order", async () => {
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.reject(problem(404, "not_found")),
    });

    expect(
      await screen.findByRole("heading", { name: "Lap times are not open" }),
    ).toBeInTheDocument();
  });
});

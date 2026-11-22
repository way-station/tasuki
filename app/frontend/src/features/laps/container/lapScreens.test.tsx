import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
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

// A confirmed running order of `race`: leg 1 (5 km) and leg 2 (3 km).
function confirmedOrder(runners: ({ roster_member_id: string; name: string } | null)[]) {
  return {
    running_orders: [
      {
        entry_id: "e1",
        entry_name: "Team A",
        status: "confirmed" as const,
        legs: runners.map((runner, index) => ({
          leg_number: index + 1,
          distance_km: [5, 3][index],
          runner,
        })),
        prediction: null,
      },
    ],
  };
}

function newLap(number: number) {
  return screen.getByLabelText(`${number} km`);
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

    const section = await screen.findByRole("region", { name: "走者のラップタイム" });
    const [taro, jiro] = within(section).getAllByRole("listitem");
    expect(
      within(taro)
        .getAllByRole("row")
        .map((row) => row.textContent),
    ).toEqual(["kmラップ", "13:20", "23:25"]);
    expect(jiro).toHaveTextContent("Jiroまだ入力されていません");
  });

  it("shows each runner's legs and the mean of their laps", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve({ ...race, version: 1, can_submit_laps: false }),
      listRunningOrders: () =>
        Promise.resolve(
          confirmedOrder([
            { roster_member_id: "m1", name: "Taro" },
            { roster_member_id: "m2", name: "Jiro" },
          ]),
        ),
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

    const section = await screen.findByRole("region", { name: "走者のラップタイム" });
    const [taro, jiro] = within(section).getAllByRole("listitem");
    expect(await within(taro).findByText("1区")).toBeInTheDocument();
    expect(taro).toHaveTextContent("平均 3:23");
    expect(jiro).toHaveTextContent("Jiro2区・まだ入力されていません");
  });

  it("asks nothing for a member", async () => {
    const listRaceLaps = vi.fn(() => Promise.resolve({ runners: [] }));
    renderAt("/races/race-1", member, { getRace: () => Promise.resolve(race), listRaceLaps });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(listRaceLaps).not.toHaveBeenCalled();
  });
});

describe("L1 my lap times", () => {
  it("shows the member's legs and their distance under the title", async () => {
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps([])),
      listRunningOrders: () =>
        Promise.resolve(confirmedOrder([null, { roster_member_id: "m1", name: "Taro" }])),
    });

    expect(await screen.findByText("2区・3 km")).toBeInTheDocument();
  });

  it("shows the laps in the order they were submitted, one per km", async () => {
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps(["3:20", "3:25"], "3:23")),
    });

    const list = await screen.findByRole("list", { name: "送ったラップタイム" });
    expect(
      within(list)
        .getAllByRole("listitem")
        .map((item) => item.firstChild?.textContent),
    ).toEqual(["1 km3:20", "2 km3:25"]);
  });

  it("tells a member who does not run in a confirmed order", async () => {
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.reject(problem(404, "not_found")),
    });

    expect(
      await screen.findByRole("heading", { name: "ラップタイムはまだ入れられません" }),
    ).toBeInTheDocument();
  });

  it("submits several lap times together and starts an empty form again", async () => {
    let saved = myLaps([]);
    const submitLaps = vi.fn((_raceId: string, times: string[]) => {
      saved = myLaps(times);
      return Promise.resolve(saved);
    });
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(saved),
      submitLaps,
    });

    fireEvent.change(await screen.findByLabelText("1 km"), { target: { value: "3:20" } });
    fireEvent.click(screen.getByRole("button", { name: "行を足す" }));
    fireEvent.change(newLap(2), { target: { value: " 3:25 " } });
    fireEvent.click(screen.getByRole("button", { name: "送る" }));

    expect(await screen.findByText("ラップタイムを足しました。")).toBeInTheDocument();
    expect(submitLaps).toHaveBeenCalledWith("race-1", ["3:20", "3:25"]);
    // The two laps are in the list now, so the empty form starts at 3 km.
    await waitFor(() => expect(newLap(3)).toHaveValue(""));
    expect(screen.queryByLabelText("4 km")).not.toBeInTheDocument();
    expect(await screen.findByText("3:25")).toBeInTheDocument();
  });

  it("shows the error of each invalid time and sends nothing", async () => {
    const submitLaps = vi.fn();
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps([])),
      submitLaps,
    });

    fireEvent.change(await screen.findByLabelText("1 km"), { target: { value: "325" } });
    fireEvent.click(screen.getByRole("button", { name: "送る" }));

    expect(
      await screen.findByText("1 km のタイムを 分:秒 で入れてください（例：3:25）。", {
        selector: "p[id]",
      }),
    ).toBeInTheDocument();
    expect(newLap(1)).toHaveAttribute("aria-invalid", "true");
    expect(submitLaps).not.toHaveBeenCalled();
  });

  it("edits one lap in place", async () => {
    const editLap = vi.fn(() => Promise.resolve(myLaps(["3:18"])));
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps(["3:20"])),
      editLap,
    });

    fireEvent.click(await screen.findByRole("button", { name: "1 km のラップタイムを直す" }));
    fireEvent.change(screen.getByLabelText("1 km"), { target: { value: "3:18" } });
    fireEvent.click(screen.getByRole("button", { name: "1 km のラップタイムを保存" }));

    expect(await screen.findByText("ラップタイムを保存しました。")).toBeInTheDocument();
    expect(editLap).toHaveBeenCalledWith("race-1", "lap-1", "3:18");
  });

  it("deletes one lap", async () => {
    const deleteLap = vi.fn(() => Promise.resolve(myLaps([])));
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps(["3:20"])),
      deleteLap,
    });

    fireEvent.click(await screen.findByRole("button", { name: "1 km のラップタイムを消す" }));

    expect(await screen.findByText("ラップタイムを消しました。")).toBeInTheDocument();
    expect(deleteLap).toHaveBeenCalledWith("race-1", "lap-1");
  });

  it("tells the member when the lap is already gone", async () => {
    renderAt("/races/race-1/laps", member, {
      getMyLaps: () => Promise.resolve(myLaps(["3:20"])),
      deleteLap: () => Promise.reject(problem(404, "not_found")),
    });

    fireEvent.click(await screen.findByRole("button", { name: "1 km のラップタイムを消す" }));

    expect(
      await screen.findByText("ラップタイムがほかの画面で変えられていたので、最新を出しています。"),
    ).toBeInTheDocument();
  });

  it("keeps the lap times being added as a draft after reloading", async () => {
    const stubs = { getMyLaps: () => Promise.resolve(myLaps([])) };
    renderAt("/races/race-1/laps", member, stubs);
    fireEvent.change(await screen.findByLabelText("1 km"), { target: { value: "3:20" } });

    document.body.innerHTML = "";
    renderAt("/races/race-1/laps", member, stubs);

    await waitFor(() => expect(newLap(1)).toHaveValue("3:20"));
  });
});

describe("P1 my profile", () => {
  it("shows the member's name, average time, and lap count", async () => {
    renderAt("/me", member, {
      getMyProfile: () => Promise.resolve({ name: "Taro", average_time: "3:28", lap_count: 8 }),
    });

    expect(await screen.findByRole("heading", { name: "Taro" })).toBeInTheDocument();
    expect(screen.getByText("3:28")).toBeInTheDocument();
    expect(screen.getByText("8")).toBeInTheDocument();
  });
});

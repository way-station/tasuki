import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ApiClient, Entry, Me, Race, RunningOrder } from "../../../api/client.ts";
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

const distanceRace: Race = {
  id: "race-1",
  name: "Hakone",
  version: 1,
  description: {
    type: "distance",
    race_date: "2026-11-23",
    deadline: "2026-11-01T18:00:00+09:00",
    leg_distances_km: [5, 3],
  },
  can_submit_laps: false,
};
const timeRace: Race = {
  ...distanceRace,
  description: {
    type: "time",
    race_date: "2026-11-23",
    deadline: "2026-11-01T18:00:00+09:00",
    time_limit_minutes: 180,
    lap_distance_km: 2,
  },
};
const entry: Entry = {
  id: "e1",
  name: "Team A",
  members: [
    { roster_member_id: "m1", name: "Taro" },
    { roster_member_id: "m2", name: "Jiro" },
  ],
};
const taro = { roster_member_id: "m1", name: "Taro" };
const jiro = { roster_member_id: "m2", name: "Jiro" };

function order(status: "draft" | "confirmed", runners: (typeof taro | null)[]): RunningOrder {
  return {
    entry_id: "e1",
    entry_name: "Team A",
    status,
    version: 2,
    legs: runners.map((runner, index) => ({
      leg_number: index + 1,
      distance_km: [5, 3][index],
      runner,
    })),
    prediction: { total_time: "0:28:10", laps: null, has_empty_legs: runners.includes(null) },
  };
}

function editorStubs(race: Race, orders: RunningOrder[]): Partial<ApiClient> {
  return {
    getRace: () => Promise.resolve(race),
    listEntries: () => Promise.resolve({ entries: [entry] }),
    listRunningOrders: () => Promise.resolve({ running_orders: orders }),
  };
}

function leg(number: number) {
  return screen.getByLabelText(`${number}区の走者`);
}

describe("R2 running order", () => {
  it("shows an organizer the draft with its prediction", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve(distanceRace),
      listRunningOrders: () => Promise.resolve({ running_orders: [order("draft", [taro, null])] }),
    });

    const section = await screen.findByRole("region", { name: "Team A の走順" });
    expect(section).toHaveTextContent("下書き");
    expect(
      within(section)
        .getAllByRole("row")
        .map((row) => row.textContent),
    ).toEqual(["区走者距離", "1区Taro5 km", "2区未定3 km"]);
    expect(section).toHaveTextContent("予想タイム0:28:10空きの区を除いた予想です。");
  });

  it("shows a member who answered attending the confirmed order", async () => {
    renderAt("/races/race-1", member, {
      getRace: () =>
        Promise.resolve({ ...distanceRace, version: undefined, my_attendance: "attending" }),
      listRunningOrders: () =>
        Promise.resolve({ running_orders: [order("confirmed", [taro, jiro])] }),
    });

    expect(await screen.findByRole("region", { name: "Team A の走順" })).toHaveTextContent(
      "走順確定",
    );
  });

  it("asks nothing for a member who has not answered attending", async () => {
    const listRunningOrders = vi.fn(() => Promise.resolve({ running_orders: [] }));
    renderAt("/races/race-1", member, {
      getRace: () =>
        Promise.resolve({ ...distanceRace, version: undefined, my_attendance: "undecided" }),
      listRunningOrders,
    });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(listRunningOrders).not.toHaveBeenCalled();
  });
});

describe("R7 running order editor", () => {
  it("sends the organizer to the entry when there is none", async () => {
    const router = renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, []),
      listEntries: () => Promise.resolve({ entries: [] }),
    });

    fireEvent.click(await screen.findByRole("button", { name: "エントリーへ" }));

    await waitFor(() => expect(router.state.location.pathname).toBe("/races/race-1/entry"));
  });

  it("starts with one empty leg per leg distance and saves the first draft without a version", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order("draft", [taro, null])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, []),
      saveRunningOrder,
    });

    expect(await screen.findByText("Team A")).toBeInTheDocument();
    expect(screen.queryByText("下書き")).not.toBeInTheDocument();
    expect(leg(1)).toHaveValue("");
    expect(leg(2)).toHaveValue("");
    expect(screen.queryByRole("button", { name: "区を足す" })).not.toBeInTheDocument();
    fireEvent.change(leg(1), { target: { value: "m1" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("走順を保存しました。")).toBeInTheDocument();
    expect(saveRunningOrder).toHaveBeenCalledWith("race-1", "e1", {
      legs: [
        { leg_number: 1, roster_member_id: "m1" },
        { leg_number: 2, roster_member_id: null },
      ],
    });
  });

  it("adds and removes legs for a time race", async () => {
    renderAt("/races/race-1/running-order", organizer, editorStubs(timeRace, []));

    fireEvent.click(await screen.findByRole("button", { name: "区を足す" }));
    expect(leg(2)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "1区を消す" }));
    expect(screen.queryByLabelText("2区の走者")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "1区を消す" })).not.toBeInTheDocument();
  });

  it("asks to save before confirming, and notes the prediction is of the saved order", async () => {
    renderAt(
      "/races/race-1/running-order",
      organizer,
      editorStubs(distanceRace, [order("draft", [taro, jiro])]),
    );

    expect(await screen.findByRole("button", { name: "確定する" })).toBeEnabled();
    fireEvent.change(leg(2), { target: { value: "m1" } });

    expect(screen.getByRole("button", { name: "確定する" })).toBeDisabled();
    expect(screen.getByText("確定するには、変更を保存してください。")).toBeInTheDocument();
    expect(
      screen.getByText("保存していない変更があります。保存すると予想が更新されます。"),
    ).toBeInTheDocument();
  });

  it("cannot confirm while a leg is empty", async () => {
    renderAt(
      "/races/race-1/running-order",
      organizer,
      editorStubs(distanceRace, [order("draft", [taro, null])]),
    );

    expect(await screen.findByRole("button", { name: "確定する" })).toBeDisabled();
    expect(
      screen.getByText("確定するには、すべての区に走者を入れてください。"),
    ).toBeInTheDocument();
  });

  it("confirms a saved draft with every leg filled", async () => {
    const confirmRunningOrder = vi.fn(() => Promise.resolve(order("confirmed", [taro, jiro])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, [order("draft", [taro, jiro])]),
      confirmRunningOrder,
    });

    fireEvent.click(await screen.findByRole("button", { name: "確定する" }));

    expect(
      await screen.findByText("走順を確定しました。参加するメンバーに見えるようになりました。"),
    ).toBeInTheDocument();
    expect(confirmRunningOrder).toHaveBeenCalledWith("race-1", "e1", 2);
  });

  it("names the runners who lose their laps before saving a confirmed order", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order("confirmed", [taro, taro])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, [order("confirmed", [taro, jiro])]),
      saveRunningOrder,
    });

    fireEvent.change(await screen.findByLabelText("2区の走者"), { target: { value: "m1" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    const dialog = screen.getByRole("alertdialog", { name: "確定した走順を変えますか？" });
    expect(dialog).toHaveTextContent("ラップタイムが消える人：Jiro。");
    expect(saveRunningOrder).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "保存する" }));

    await waitFor(() =>
      expect(saveRunningOrder).toHaveBeenCalledWith("race-1", "e1", {
        version: 2,
        legs: [
          { leg_number: 1, roster_member_id: "m1" },
          { leg_number: 2, roster_member_id: "m1" },
        ],
      }),
    );
  });

  it("saves a confirmed order at once when nobody loses laps", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order("confirmed", [jiro, taro])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, [order("confirmed", [taro, jiro])]),
      saveRunningOrder,
    });

    fireEvent.change(await screen.findByLabelText("1区の走者"), { target: { value: "m2" } });
    fireEvent.change(leg(2), { target: { value: "m1" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("走順を保存しました。")).toBeInTheDocument();
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("keeps the draft after a conflict, and loads the latest when asked", async () => {
    let saved = order("draft", [taro, null]);
    renderAt("/races/race-1/running-order", organizer, {
      getRace: () => Promise.resolve(distanceRace),
      listEntries: () => Promise.resolve({ entries: [entry] }),
      listRunningOrders: () => Promise.resolve({ running_orders: [saved] }),
      saveRunningOrder: () => {
        saved = { ...order("draft", [jiro, jiro]), version: 3 };
        return Promise.reject(problem(409, "version_conflict"));
      },
    });
    fireEvent.change(await screen.findByLabelText("2区の走者"), { target: { value: "m1" } });

    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("ほかの人が先に保存しました。")).toBeInTheDocument();
    expect(leg(2)).toHaveValue("m1");

    fireEvent.click(screen.getByRole("button", { name: "最新を読み込む（自分の変更を捨てる）" }));

    await waitFor(() => expect(leg(1)).toHaveValue("m2"));
  });

  it("keeps the input as a draft after reloading", async () => {
    renderAt("/races/race-1/running-order", organizer, editorStubs(distanceRace, []));
    fireEvent.change(await screen.findByLabelText("1区の走者"), { target: { value: "m2" } });

    document.body.innerHTML = "";
    renderAt("/races/race-1/running-order", organizer, editorStubs(distanceRace, []));

    await waitFor(() => expect(leg(1)).toHaveValue("m2"));
  });
});

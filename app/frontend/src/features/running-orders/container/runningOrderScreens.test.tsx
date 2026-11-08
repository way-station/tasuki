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
  return screen.getByLabelText(new RegExp(`^Leg ${number}\\b`));
}

describe("R2 running order", () => {
  it("shows an organizer the draft with its prediction", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve(distanceRace),
      listRunningOrders: () => Promise.resolve({ running_orders: [order("draft", [taro, null])] }),
    });

    const section = await screen.findByRole("region", { name: "Running order of Team A" });
    expect(section).toHaveTextContent("Draft");
    expect(
      within(section)
        .getAllByRole("row")
        .map((row) => row.textContent),
    ).toEqual(["LegDistanceRunner", "15 kmTaro", "23 kmNot decided"]);
    expect(section).toHaveTextContent("Prediction: 0:28:10 (some legs are empty and left out)");
  });

  it("shows a member who answered attending the confirmed order", async () => {
    renderAt("/races/race-1", member, {
      getRace: () =>
        Promise.resolve({ ...distanceRace, version: undefined, my_attendance: "attending" }),
      listRunningOrders: () =>
        Promise.resolve({ running_orders: [order("confirmed", [taro, jiro])] }),
    });

    expect(
      await screen.findByRole("region", { name: "Running order of Team A" }),
    ).toHaveTextContent("Confirmed");
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

    fireEvent.click(await screen.findByRole("button", { name: "Go to the entry" }));

    await waitFor(() => expect(router.state.location.pathname).toBe("/races/race-1/entry"));
  });

  it("starts with one empty leg per leg distance and saves the first draft without a version", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order("draft", [taro, null])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, []),
      saveRunningOrder,
    });

    expect(await screen.findByText("Team A · Not saved yet")).toBeInTheDocument();
    expect(leg(1)).toHaveValue("");
    expect(leg(2)).toHaveValue("");
    expect(screen.queryByRole("button", { name: "Add leg" })).not.toBeInTheDocument();
    fireEvent.change(leg(1), { target: { value: "m1" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Running order saved.")).toBeInTheDocument();
    expect(saveRunningOrder).toHaveBeenCalledWith("race-1", "e1", {
      legs: [
        { leg_number: 1, roster_member_id: "m1" },
        { leg_number: 2, roster_member_id: null },
      ],
    });
  });

  it("adds and removes legs for a time race", async () => {
    renderAt("/races/race-1/running-order", organizer, editorStubs(timeRace, []));

    fireEvent.click(await screen.findByRole("button", { name: "Add leg" }));
    expect(leg(2)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Remove leg 1" }));
    expect(screen.queryByLabelText(/^Leg 2\b/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Remove leg 1" })).not.toBeInTheDocument();
  });

  it("asks to save before confirming, and notes the prediction is of the saved order", async () => {
    renderAt(
      "/races/race-1/running-order",
      organizer,
      editorStubs(distanceRace, [order("draft", [taro, jiro])]),
    );

    expect(await screen.findByRole("button", { name: "Confirm" })).toBeEnabled();
    fireEvent.change(leg(2), { target: { value: "m1" } });

    expect(screen.getByRole("button", { name: "Confirm" })).toBeDisabled();
    expect(screen.getByText("Save your changes before confirming.")).toBeInTheDocument();
    expect(
      screen.getByText("This is the prediction of the saved order. Saving updates it."),
    ).toBeInTheDocument();
  });

  it("cannot confirm while a leg is empty", async () => {
    renderAt(
      "/races/race-1/running-order",
      organizer,
      editorStubs(distanceRace, [order("draft", [taro, null])]),
    );

    expect(await screen.findByRole("button", { name: "Confirm" })).toBeDisabled();
    expect(screen.getByText("Every leg needs a runner before confirming.")).toBeInTheDocument();
  });

  it("confirms a saved draft with every leg filled", async () => {
    const confirmRunningOrder = vi.fn(() => Promise.resolve(order("confirmed", [taro, jiro])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, [order("draft", [taro, jiro])]),
      confirmRunningOrder,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Confirm" }));

    expect(
      await screen.findByText("Running order confirmed. Attending members can see it now."),
    ).toBeInTheDocument();
    expect(confirmRunningOrder).toHaveBeenCalledWith("race-1", "e1", 2);
  });

  it("names the runners who lose their laps before saving a confirmed order", async () => {
    const saveRunningOrder = vi.fn(() => Promise.resolve(order("confirmed", [taro, taro])));
    renderAt("/races/race-1/running-order", organizer, {
      ...editorStubs(distanceRace, [order("confirmed", [taro, jiro])]),
      saveRunningOrder,
    });

    fireEvent.change(await screen.findByLabelText(/^Leg 2\b/), { target: { value: "m1" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    const dialog = screen.getByRole("alertdialog", { name: "Save and delete lap times?" });
    expect(dialog).toHaveTextContent("Jiro will run fewer legs");
    expect(saveRunningOrder).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "Save" }));

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

    fireEvent.change(await screen.findByLabelText(/^Leg 1\b/), { target: { value: "m2" } });
    fireEvent.change(leg(2), { target: { value: "m1" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Running order saved.")).toBeInTheDocument();
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
    fireEvent.change(await screen.findByLabelText(/^Leg 2\b/), { target: { value: "m1" } });

    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText(/Your changes are kept here\./)).toBeInTheDocument();
    expect(leg(2)).toHaveValue("m1");

    fireEvent.click(screen.getByRole("button", { name: "Load the latest (discard my changes)" }));

    await waitFor(() => expect(leg(1)).toHaveValue("m2"));
  });

  it("keeps the input as a draft after reloading", async () => {
    renderAt("/races/race-1/running-order", organizer, editorStubs(distanceRace, []));
    fireEvent.change(await screen.findByLabelText(/^Leg 1\b/), { target: { value: "m2" } });

    document.body.innerHTML = "";
    renderAt("/races/race-1/running-order", organizer, editorStubs(distanceRace, []));

    await waitFor(() => expect(leg(1)).toHaveValue("m2"));
  });
});

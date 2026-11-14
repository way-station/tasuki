import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ApiClient, Me, Race } from "../../../api/client.ts";
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
  roster_member: { id: "member-1", name: "Taro Yamada" },
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
  version: 3,
  description: {
    type: "distance",
    race_date: "2026-11-23",
    deadline: "2026-11-01T18:00:00+09:00",
    text: "Meet at 7:00",
    leg_distances_km: [5, 3.2],
  },
  running_order_confirmed: false,
  can_submit_laps: false,
};

// What a member gets for the same race: no version, no deadline.
const memberRace: Race = {
  ...distanceRace,
  version: undefined,
  description: { ...distanceRace.description, deadline: undefined },
  my_attendance: "attending",
};

describe("R1 race list", () => {
  const races = {
    races: [
      {
        id: "race-1",
        name: "Hakone",
        race_date: "2026-11-23",
        my_attendance: "attending" as const,
        running_order_confirmed: true,
      },
      { id: "race-2", name: "Ekiden", race_date: "2026-12-05", my_attendance: null },
    ],
  };

  it("shows a member My races and Invitations", async () => {
    renderAt("/", member, { listRaces: () => Promise.resolve(races) });

    const myRaces = await screen.findByRole("region", { name: "My races" });
    expect(within(myRaces).getByRole("button", { name: "Hakone" })).toBeInTheDocument();
    expect(myRaces).toHaveTextContent("Running order confirmed");
    const invitations = screen.getByRole("region", { name: "Invitations" });
    expect(within(invitations).getByRole("button", { name: "Ekiden" })).toBeInTheDocument();
    expect(invitations).toHaveTextContent("Not answered");
    expect(screen.queryByRole("button", { name: "Register race" })).not.toBeInTheDocument();
  });

  it("shows an organizer every race and the register button", async () => {
    renderAt("/", organizer, {
      listRaces: () =>
        Promise.resolve({ races: [{ id: "race-1", name: "Hakone", race_date: "2026-11-23" }] }),
    });

    const all = await screen.findByRole("region", { name: "All races" });
    expect(within(all).getByRole("button", { name: "Hakone" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Register race" })).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Invitations" })).not.toBeInTheDocument();
  });

  it("opens a race", async () => {
    const router = renderAt("/", member, {
      listRaces: () => Promise.resolve(races),
      getRace: () => Promise.resolve(memberRace),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Hakone" }));

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/races/race-1");
  });
});

describe("R2 race detail", () => {
  it("shows a member the description without the deadline or organizer actions", async () => {
    renderAt("/races/race-1", member, { getRace: () => Promise.resolve(memberRace) });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(screen.getByText("Meet at 7:00")).toBeInTheDocument();
    expect(screen.getByText("3.2 km")).toBeInTheDocument();
    const attendance = screen.getByRole("group", { name: "My attendance" });
    expect(within(attendance).getByRole("button", { name: "Attending" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    expect(screen.queryByText("Deadline")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Edit race" })).not.toBeInTheDocument();
  });

  it("shows a runner the way to their lap times", async () => {
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve({ ...memberRace, can_submit_laps: true }),
    });

    expect(await screen.findByRole("button", { name: "My lap times" })).toBeInTheDocument();
  });

  it("shows an organizer the deadline and the organizer screens", async () => {
    const router = renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve(distanceRace),
    });

    expect(await screen.findByText("Deadline")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Edit race" }));

    expect(await screen.findByRole("heading", { name: "Edit race" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/races/race-1/edit");
  });

  it("shows a linked organizer who is not attending no answer to change", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve({ ...distanceRace, my_attendance: "not_attending" }),
    });

    expect(await screen.findByText("Deadline")).toBeInTheDocument();
    expect(screen.queryByRole("group", { name: "My attendance" })).not.toBeInTheDocument();
  });

  it("says a race is not found when the user may not see it", async () => {
    renderAt("/races/hidden", member, {
      getRace: () => Promise.reject(problem(404, "not_found")),
    });

    expect(await screen.findByRole("heading", { name: "Race not found" })).toBeInTheDocument();
  });
});

describe("R3 register race", () => {
  function fillRequired() {
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Hakone" } });
    fireEvent.change(screen.getByLabelText("Race date"), { target: { value: "2026-11-23" } });
    fireEvent.change(screen.getByLabelText("Deadline"), { target: { value: "2026-11-01T18:00" } });
    fireEvent.change(screen.getByLabelText("Leg 1"), { target: { value: "5" } });
  }

  it("registers the race and opens it", async () => {
    const createRace = vi.fn(() => Promise.resolve(distanceRace));
    const router = renderAt("/races/new", organizer, {
      createRace,
      getRace: () => Promise.resolve(distanceRace),
    });
    await screen.findByRole("heading", { name: "Register race" });

    fillRequired();
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/races/race-1");
    expect(screen.getByText("Race registered.")).toBeInTheDocument();
    expect(createRace).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Hakone",
        description: expect.objectContaining({ leg_distances_km: [5] }),
      }),
    );
  });

  it("shows what is missing without sending", async () => {
    const createRace = vi.fn(() => Promise.resolve(distanceRace));
    renderAt("/races/new", organizer, { createRace });
    await screen.findByRole("heading", { name: "Register race" });

    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Enter a name.")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveAttribute("aria-invalid", "true");
    expect(createRace).not.toHaveBeenCalled();
  });

  it("shows the time race fields and keeps the distance input when switching back", async () => {
    renderAt("/races/new", organizer, {});
    await screen.findByRole("heading", { name: "Register race" });
    fireEvent.change(screen.getByLabelText("Leg 1"), { target: { value: "5" } });

    fireEvent.click(screen.getByLabelText("Time"));
    expect(screen.getByLabelText("Time limit (minutes)")).toBeInTheDocument();
    expect(screen.queryByLabelText("Leg 1")).not.toBeInTheDocument();

    fireEvent.click(screen.getByLabelText("Distance"));
    expect(screen.getByLabelText("Leg 1")).toHaveValue("5");
  });

  it("adds and removes legs", async () => {
    renderAt("/races/new", organizer, {});
    await screen.findByRole("heading", { name: "Register race" });

    fireEvent.click(screen.getByRole("button", { name: "Add leg" }));
    expect(screen.getByLabelText("Leg 2")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Remove leg 1" }));
    expect(screen.queryByLabelText("Leg 2")).not.toBeInTheDocument();
  });

  it("keeps the input as a draft after reloading", async () => {
    renderAt("/races/new", organizer, {});
    await screen.findByRole("heading", { name: "Register race" });
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Hakone" } });

    document.body.innerHTML = "";
    renderAt("/races/new", organizer, {});

    expect(await screen.findByLabelText("Name")).toHaveValue("Hakone");
  });
});

describe("R3 edit race", () => {
  it("starts from the saved race and saves with its version", async () => {
    const updateRace = vi.fn(() => Promise.resolve({ ...distanceRace, version: 4 }));
    renderAt("/races/race-1/edit", organizer, {
      getRace: () => Promise.resolve(distanceRace),
      updateRace,
    });

    expect(await screen.findByLabelText("Name")).toHaveValue("Hakone");
    expect(screen.getByLabelText("Leg 2")).toHaveValue("3.2");
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Hakone 2026" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Race saved.")).toBeInTheDocument();
    expect(updateRace).toHaveBeenCalledWith(
      "race-1",
      expect.objectContaining({ name: "Hakone 2026", version: 3 }),
    );
  });

  it("keeps the draft after a conflict, and loads the latest when asked", async () => {
    let saved = distanceRace;
    renderAt("/races/race-1/edit", organizer, {
      getRace: () => Promise.resolve(saved),
      updateRace: () => {
        saved = { ...distanceRace, name: "Hakone (by someone else)", version: 4 };
        return Promise.reject(problem(409, "version_conflict"));
      },
    });
    fireEvent.change(await screen.findByLabelText("Name"), { target: { value: "Mine" } });

    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    const alert = await screen.findByText(/Someone else saved this race first\. Your changes/);
    expect(alert).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("Mine");

    fireEvent.click(screen.getByRole("button", { name: "Load the latest (discard my changes)" }));

    await waitFor(() =>
      expect(screen.getByLabelText("Name")).toHaveValue("Hakone (by someone else)"),
    );
  });
});

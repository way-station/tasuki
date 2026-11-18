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

    const myRaces = await screen.findByRole("region", { name: "参加する大会" });
    expect(within(myRaces).getByRole("button", { name: /^Hakone/ })).toBeInTheDocument();
    expect(myRaces).toHaveTextContent("走順確定");
    const invitations = screen.getByRole("region", { name: "招待" });
    expect(within(invitations).getByRole("button", { name: /^Ekiden/ })).toBeInTheDocument();
    expect(invitations).toHaveTextContent("未回答");
    expect(screen.queryByRole("button", { name: "大会を登録" })).not.toBeInTheDocument();
  });

  it("shows an organizer every race and the register button", async () => {
    renderAt("/", organizer, {
      listRaces: () =>
        Promise.resolve({ races: [{ id: "race-1", name: "Hakone", race_date: "2026-11-23" }] }),
    });

    const all = await screen.findByRole("region", { name: "すべての大会" });
    expect(within(all).getByRole("button", { name: /^Hakone/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "大会を登録" })).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "招待" })).not.toBeInTheDocument();
  });

  it("opens a race", async () => {
    const router = renderAt("/", member, {
      listRaces: () => Promise.resolve(races),
      getRace: () => Promise.resolve(memberRace),
    });

    fireEvent.click(await screen.findByRole("button", { name: /^Hakone/ }));

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/races/race-1");
  });
});

describe("R2 race detail", () => {
  it("shows a member the description without the deadline or organizer actions", async () => {
    renderAt("/races/race-1", member, { getRace: () => Promise.resolve(memberRace) });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(screen.getByText("Meet at 7:00")).toBeInTheDocument();
    expect(screen.getByText("5 km・3.2 km")).toBeInTheDocument();
    const attendance = screen.getByRole("group", { name: "あなたの出欠" });
    expect(within(attendance).getByRole("radio", { name: "参加" })).toBeChecked();
    expect(screen.queryByText("締め切り")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "大会を編集" })).not.toBeInTheDocument();
  });

  it("shows a runner the way to their lap times", async () => {
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve({ ...memberRace, can_submit_laps: true }),
    });

    expect(await screen.findByRole("button", { name: "自分のラップタイム" })).toBeInTheDocument();
  });

  it("shows an organizer the deadline and the organizer screens", async () => {
    const router = renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve(distanceRace),
    });

    expect(await screen.findByText("締め切り")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "大会を編集" }));

    expect(await screen.findByRole("heading", { name: "Edit race" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/races/race-1/edit");
  });

  it("shows a linked organizer who is not attending no answer to change", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve({ ...distanceRace, my_attendance: "not_attending" }),
    });

    expect(await screen.findByText("締め切り")).toBeInTheDocument();
    expect(screen.queryByRole("group", { name: "あなたの出欠" })).not.toBeInTheDocument();
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
    fireEvent.change(screen.getByLabelText("大会名"), { target: { value: "Hakone" } });
    fireEvent.change(screen.getByLabelText("開催日"), { target: { value: "2026-11-23" } });
    fireEvent.change(screen.getByLabelText("締め切り"), { target: { value: "2026-11-01T18:00" } });
    fireEvent.change(screen.getByLabelText("1区の距離"), { target: { value: "5" } });
  }

  it("registers the race and opens it", async () => {
    const createRace = vi.fn(() => Promise.resolve(distanceRace));
    const router = renderAt("/races/new", organizer, {
      createRace,
      getRace: () => Promise.resolve(distanceRace),
    });
    await screen.findByRole("heading", { name: "Register race" });

    fillRequired();
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

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

    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("Enter a name.")).toBeInTheDocument();
    expect(screen.getByLabelText("大会名")).toHaveAttribute("aria-invalid", "true");
    expect(createRace).not.toHaveBeenCalled();
  });

  it("shows the time race fields and keeps the distance input when switching back", async () => {
    renderAt("/races/new", organizer, {});
    await screen.findByRole("heading", { name: "Register race" });
    fireEvent.change(screen.getByLabelText("1区の距離"), { target: { value: "5" } });

    fireEvent.click(screen.getByLabelText("時間制"));
    expect(screen.getByLabelText("制限時間")).toBeInTheDocument();
    expect(screen.queryByLabelText("1区の距離")).not.toBeInTheDocument();

    fireEvent.click(screen.getByLabelText("距離制"));
    expect(screen.getByLabelText("1区の距離")).toHaveValue("5");
  });

  it("adds and removes legs", async () => {
    renderAt("/races/new", organizer, {});
    await screen.findByRole("heading", { name: "Register race" });

    fireEvent.click(screen.getByRole("button", { name: "区を足す" }));
    expect(screen.getByLabelText("2区の距離")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "1区を消す" }));
    expect(screen.queryByLabelText("2区の距離")).not.toBeInTheDocument();
  });

  it("keeps the input as a draft after reloading", async () => {
    renderAt("/races/new", organizer, {});
    await screen.findByRole("heading", { name: "Register race" });
    fireEvent.change(screen.getByLabelText("大会名"), { target: { value: "Hakone" } });

    document.body.innerHTML = "";
    renderAt("/races/new", organizer, {});

    expect(await screen.findByLabelText("大会名")).toHaveValue("Hakone");
  });
});

describe("R3 edit race", () => {
  it("starts from the saved race and saves with its version", async () => {
    const updateRace = vi.fn(() => Promise.resolve({ ...distanceRace, version: 4 }));
    renderAt("/races/race-1/edit", organizer, {
      getRace: () => Promise.resolve(distanceRace),
      updateRace,
    });

    expect(await screen.findByLabelText("大会名")).toHaveValue("Hakone");
    expect(screen.getByLabelText("2区の距離")).toHaveValue("3.2");
    fireEvent.change(screen.getByLabelText("大会名"), { target: { value: "Hakone 2026" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

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
    fireEvent.change(await screen.findByLabelText("大会名"), { target: { value: "Mine" } });

    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    const alert = await screen.findByText("ほかの人が先に保存しました。");
    expect(alert).toBeInTheDocument();
    expect(screen.getByLabelText("大会名")).toHaveValue("Mine");

    fireEvent.click(screen.getByRole("button", { name: "最新を読み込む（自分の変更を捨てる）" }));

    await waitFor(() =>
      expect(screen.getByLabelText("大会名")).toHaveValue("Hakone (by someone else)"),
    );
  });
});

import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { ApiClient } from "../api/client.ts";
import { createFakeApiClient } from "../api/fakeApiClient.ts";
import { TestProviders } from "../test/TestProviders.tsx";
import { createAppRouter } from "./router.tsx";

function renderAt(path: string, stubs: Partial<ApiClient> = {}) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  // An organizer linked to a roster member may open every screen. What other
  // callers see is tested with the route guard (features/auth/container).
  const client = createFakeApiClient({
    listRaces: () => Promise.resolve({ races: [] }),
    listInvitations: () => Promise.resolve({ members: [] }),
    listEntries: () => Promise.resolve({ entries: [] }),
    listRunningOrders: () => Promise.resolve({ running_orders: [] }),
    listRaceLaps: () => Promise.resolve({ runners: [] }),
    getMyLaps: () => Promise.resolve({ laps: [], average_time: "3:40", lap_count: 0 }),
    getMyProfile: () =>
      Promise.resolve({ name: "Taro Yamada", average_time: "3:40", lap_count: 0 }),
    listRosterMembers: () => Promise.resolve({ roster_members: [] }),
    getRosterMember: (rosterMemberId) =>
      Promise.resolve({
        id: rosterMemberId,
        name: "Jiro",
        left: false,
        average_time: "3:40",
        lap_count: 0,
        laps_by_race: [],
      }),
    listPendingMemberLinks: () => Promise.resolve({ links: [] }),
    getRace: (raceId) =>
      Promise.resolve({
        id: raceId,
        name: "Hakone",
        version: 1,
        description: {
          type: "distance",
          race_date: "2026-11-23",
          deadline: "2026-11-01T18:00:00+09:00",
          leg_distances_km: [5],
        },
        can_submit_laps: false,
      }),
    getMe: () =>
      Promise.resolve({
        user_id: "user-1",
        display_name: "Taro",
        is_organizer: true,
        status: "approved",
        roster_member: { id: "member-1", name: "Taro Yamada" },
      }),
    getInviteLink: () => Promise.resolve({ team_name: "Team" }),
    ...stubs,
  });
  render(
    <TestProviders client={client}>
      <RouterProvider router={router} />
    </TestProviders>,
  );
}

describe("createAppRouter", () => {
  it.each([
    ["/invite/abc", "Team"],
    ["/login", "ログイン"],
    ["/auth/complete", "ログイン"],
    ["/waiting", "幹事の承認を待っています"],
    ["/rejected", "却下されました"],
    ["/unavailable", "tasuki を利用できません"],
    ["/", "大会"],
    ["/races/new", "大会を登録"],
    ["/races/123", "Hakone"],
    ["/races/123/edit", "大会を編集"],
    ["/races/123/invitations", "招待"],
    ["/races/123/attendance", "出欠"],
    ["/races/123/entry", "エントリー"],
    ["/races/123/running-order", "走順"],
    ["/races/123/laps", "自分のラップタイム"],
    ["/me", "Taro Yamada"],
    ["/roster", "名簿"],
    ["/roster/123", "Jiro"],
    ["/approvals", "承認"],
  ])("shows %s as %s", async (path, heading) => {
    renderAt(path);

    expect(await screen.findByRole("heading", { name: heading })).toBeInTheDocument();
  });

  it.each(["/unknown", "/races/123/unknown", "/roster/123/unknown"])(
    "shows the not-found page for %s",
    async (path) => {
      renderAt(path);

      expect(
        await screen.findByRole("heading", { name: "ページが見つかりません" }),
      ).toBeInTheDocument();
    },
  );

  it("shows every screen inside the common layout", async () => {
    renderAt("/races/123");

    const heading = await screen.findByRole("heading", { name: "Hakone" });
    expect(screen.getByRole("main")).toContainElement(heading);
  });

  it("shows the not-found page inside the common layout", async () => {
    renderAt("/unknown");

    const heading = await screen.findByRole("heading", { name: "ページが見つかりません" });
    expect(screen.getByRole("main")).toContainElement(heading);
  });
});

describe("page frame", () => {
  async function tabs() {
    const bar = within(await screen.findByRole("navigation", { name: "メイン" }));
    return bar.getAllByRole("link");
  }

  it("gives an organizer linked to a roster member four tabs, with the current one marked", async () => {
    renderAt("/roster");

    const links = await tabs();
    expect(links.map((link) => link.textContent)).toEqual(["大会", "名簿", "承認", "自分"]);
    expect(links[1]).toHaveAttribute("aria-current", "page");
    expect(screen.queryByRole("navigation", { name: "戻る" })).not.toBeInTheDocument();
  });

  it("goes to the screen of a tab", async () => {
    renderAt("/roster");

    fireEvent.click((await tabs())[3]);

    expect(await screen.findByRole("heading", { name: "Taro Yamada" })).toBeInTheDocument();
  });

  it("shows the number of waiting links on 承認", async () => {
    const link = {
      id: "link-1",
      line_display_name: "taro_line",
      roster_member: { id: "m1", name: "Taro" },
      created_at: "2026-11-07T09:00:00Z",
    };
    renderAt("/", {
      listPendingMemberLinks: () => Promise.resolve({ links: [link, { ...link, id: "link-2" }] }),
    });

    expect(await screen.findByRole("link", { name: "承認 2 件" })).toBeInTheDocument();
  });

  it("goes back from the race to the race list", async () => {
    renderAt("/races/123");

    const back = within(await screen.findByRole("navigation", { name: "戻る" }));
    fireEvent.click(back.getByRole("link", { name: "大会" }));

    expect(await screen.findByRole("heading", { name: "大会" })).toBeInTheDocument();
  });

  it("names the race on the way back from a race's screen", async () => {
    renderAt("/races/123/entry");

    const back = within(await screen.findByRole("navigation", { name: "戻る" }));
    expect(await back.findByRole("link", { name: "Hakone" })).toHaveAttribute("href", "/races/123");
  });

  it("goes back from a member's profile to the roster", async () => {
    renderAt("/roster/123");

    const back = within(await screen.findByRole("navigation", { name: "戻る" }));
    expect(back.getByRole("link", { name: "名簿" })).toHaveAttribute("href", "/roster");
  });

  it("has no frame on the joining and access screens", async () => {
    renderAt("/login");

    await screen.findByRole("heading", { name: "ログイン" });
    expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
  });
});

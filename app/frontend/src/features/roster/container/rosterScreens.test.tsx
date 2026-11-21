import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ApiClient, Me, RosterMember } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { createAppRouter } from "../../../app/router.tsx";
import { TestProviders } from "../../../test/TestProviders.tsx";

afterEach(() => {
  vi.unstubAllGlobals();
});

const organizer: Me = {
  user_id: "user-1",
  display_name: "Hanako",
  is_organizer: true,
  status: "approved",
  roster_member: null,
};

function renderAt(path: string, stubs: Partial<ApiClient>) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  const client = createFakeApiClient({
    getMe: () => Promise.resolve(organizer),
    ...stubs,
  });
  render(
    <TestProviders client={client}>
      <RouterProvider router={router} />
    </TestProviders>,
  );
  return router;
}

function member(overrides: Partial<RosterMember>): RosterMember {
  return {
    id: "m1",
    name: "Taro",
    left: false,
    link_status: "none",
    invite_link_expires_at: null,
    ...overrides,
  };
}

function rosterOf(...members: RosterMember[]): Partial<ApiClient> {
  return { listRosterMembers: () => Promise.resolve({ roster_members: members }) };
}

describe("M1 roster", () => {
  it("shows each member with their status and user", async () => {
    renderAt(
      "/roster",
      rosterOf(
        member({ id: "m1", name: "Taro", link_status: "approved" }),
        member({ id: "m2", name: "Jiro", left: true }),
      ),
    );

    const [taro, jiro] = within(
      await screen.findByRole("region", { name: "メンバー" }),
    ).getAllByRole("listitem");
    expect(taro).toHaveTextContent("承認済み");
    expect(jiro).toHaveTextContent("退会");
    expect(screen.getByRole("button", { name: "Taro さんの紐付けを解除" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Jiro さんの退会を取り消す" })).toBeInTheDocument();
  });

  it("offers an invite link only to members who stay and have no user", async () => {
    renderAt(
      "/roster",
      rosterOf(
        member({ id: "m1", name: "Taro" }),
        member({ id: "m2", name: "Jiro", link_status: "pending" }),
        member({ id: "m3", name: "Saburo", left: true }),
      ),
    );

    expect(
      await screen.findByRole("button", { name: "Taro さんの招待リンクを作る" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^Jiro さんの招待リンク/ }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^Saburo さんの招待リンク/ }),
    ).not.toBeInTheDocument();
  });

  it("adds a member by name", async () => {
    const createRosterMember = vi.fn(() => Promise.resolve(member({})));
    renderAt("/roster", { ...rosterOf(), createRosterMember });

    fireEvent.change(await screen.findByLabelText("メンバーを追加（名前だけ）"), {
      target: { value: " Taro " },
    });
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    expect(await screen.findByText("名簿にメンバーを足しました。")).toBeInTheDocument();
    expect(createRosterMember).toHaveBeenCalledWith("Taro");
    expect(screen.getByLabelText("メンバーを追加（名前だけ）")).toHaveValue("");
  });

  it("shows an error for an empty name and sends nothing", async () => {
    const createRosterMember = vi.fn();
    renderAt("/roster", { ...rosterOf(), createRosterMember });

    fireEvent.click(await screen.findByRole("button", { name: "追加" }));

    expect(await screen.findByText("名前を入れてください。")).toBeInTheDocument();
    expect(createRosterMember).not.toHaveBeenCalled();
  });

  it("creates an invite link and copies it", async () => {
    const writeText = vi.fn(() => Promise.resolve());
    vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
    renderAt("/roster", {
      ...rosterOf(member({})),
      createInviteLink: () =>
        Promise.resolve({
          url: "https://tasuki.example/invite/abc",
          expires_at: "2026-11-14T09:00:00Z",
        }),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんの招待リンクを作る" }));
    const panel = await screen.findByRole("region", { name: "Taro さんの招待リンク" });
    expect(within(panel).getByLabelText("招待リンク")).toHaveValue(
      "https://tasuki.example/invite/abc",
    );
    fireEvent.click(within(panel).getByRole("button", { name: "コピー" }));

    expect(await screen.findByText("招待リンクをコピーしました。")).toBeInTheDocument();
    expect(writeText).toHaveBeenCalledWith("https://tasuki.example/invite/abc");
  });

  it("tells the organizer to copy by hand when the browser cannot", async () => {
    vi.stubGlobal("navigator", {
      ...navigator,
      clipboard: { writeText: () => Promise.reject(new Error("denied")) },
    });
    renderAt("/roster", {
      ...rosterOf(member({})),
      createInviteLink: () =>
        Promise.resolve({
          url: "https://tasuki.example/invite/abc",
          expires_at: "2026-11-14T09:00:00Z",
        }),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんの招待リンクを作る" }));
    fireEvent.click(await screen.findByRole("button", { name: "コピー" }));

    expect(
      await screen.findByText("リンクをコピーできませんでした。選んで手でコピーしてください。"),
    ).toBeInTheDocument();
  });

  it("asks before marking a member as left", async () => {
    const setRosterMemberLeft = vi.fn(() => Promise.resolve(member({ left: true })));
    renderAt("/roster", { ...rosterOf(member({})), setRosterMemberLeft });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんを退会にする" }));
    const dialog = screen.getByRole("alertdialog", { name: "Taro さんを退会にしますか？" });
    expect(setRosterMemberLeft).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "退会にする" }));

    expect(await screen.findByText("Taro さんを退会にしました。")).toBeInTheDocument();
    expect(setRosterMemberLeft).toHaveBeenCalledWith("m1", true);
  });

  it("undoes left at once", async () => {
    const setRosterMemberLeft = vi.fn(() => Promise.resolve(member({})));
    renderAt("/roster", { ...rosterOf(member({ left: true })), setRosterMemberLeft });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんの退会を取り消す" }));

    expect(await screen.findByText("Taro さんを戻しました。")).toBeInTheDocument();
    expect(setRosterMemberLeft).toHaveBeenCalledWith("m1", false);
  });

  it("asks before unlinking a user", async () => {
    const unlinkRosterMember = vi.fn(() => Promise.resolve());
    renderAt("/roster", {
      ...rosterOf(member({ link_status: "approved" })),
      unlinkRosterMember,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんの紐付けを解除" }));
    const dialog = screen.getByRole("alertdialog", { name: "Taro さんの紐付けを解除しますか？" });
    fireEvent.click(within(dialog).getByRole("button", { name: "解除する" }));

    expect(await screen.findByText("Taro さんの紐付けを解除しました。")).toBeInTheDocument();
    expect(unlinkRosterMember).toHaveBeenCalledWith("m1");
  });

  it("tells the organizer when the roster changed in the meantime", async () => {
    renderAt("/roster", {
      ...rosterOf(member({ link_status: "approved" })),
      unlinkRosterMember: () => Promise.reject(problem(409, "state_conflict")),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんの紐付けを解除" }));
    fireEvent.click(screen.getByRole("button", { name: "解除する" }));

    expect(
      await screen.findByText("名簿がほかの人に変えられていたので、最新を出しています。"),
    ).toBeInTheDocument();
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("opens a member's profile", async () => {
    const router = renderAt("/roster", rosterOf(member({})));

    fireEvent.click(await screen.findByRole("button", { name: "Taroユーザーなし" }));

    await waitFor(() => expect(router.state.location.pathname).toBe("/roster/m1"));
  });
});

describe("M2 member profile", () => {
  it("shows the laps per race", async () => {
    renderAt("/roster/m1", {
      getRosterMember: () =>
        Promise.resolve({
          id: "m1",
          name: "Taro",
          left: false,
          average_time: "3:28",
          lap_count: 3,
          laps_by_race: [
            {
              race_id: "race-1",
              race_name: "Hakone",
              laps: [
                { id: "l1", time: "3:20" },
                { id: "l2", time: "3:30" },
              ],
            },
          ],
        }),
    });

    expect(await screen.findByRole("heading", { name: "Taro" })).toBeInTheDocument();
    expect(screen.getByText("3:28")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "大会ごとのラップタイム" }))
        .getAllByRole("row")
        .map((row) => row.textContent),
    ).toEqual(["kmラップ", "13:20", "23:30"]);
    expect(screen.getByText("Hakone")).toBeInTheDocument();
  });

  it("tells the organizer about an unknown member", async () => {
    renderAt("/roster/unknown", {
      getRosterMember: () => Promise.reject(problem(404, "not_found")),
    });

    expect(
      await screen.findByRole("heading", { name: "名簿のメンバーが見つかりません" }),
    ).toBeInTheDocument();
  });
});

describe("M3 approvals", () => {
  const waiting = {
    links: [
      {
        id: "link-1",
        line_display_name: "taro_line",
        roster_member: { id: "m1", name: "Taro" },
        created_at: "2026-11-07T09:00:00Z",
      },
    ],
  };

  it("approves a waiting link at once", async () => {
    const approveMemberLink = vi.fn(() => Promise.resolve());
    renderAt("/approvals", {
      listPendingMemberLinks: () => Promise.resolve(waiting),
      approveMemberLink,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro（taro_line）を承認" }));

    expect(await screen.findByText("Taro（taro_line）を承認しました。")).toBeInTheDocument();
    expect(approveMemberLink).toHaveBeenCalledWith("link-1");
  });

  it("asks before rejecting a waiting link", async () => {
    const rejectMemberLink = vi.fn(() => Promise.resolve());
    renderAt("/approvals", {
      listPendingMemberLinks: () => Promise.resolve(waiting),
      rejectMemberLink,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro（taro_line）を却下" }));
    const dialog = screen.getByRole("alertdialog", { name: "Taro（taro_line）を却下しますか？" });
    expect(rejectMemberLink).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "却下する" }));

    expect(
      await screen.findByText(
        "Taro（taro_line）を却下しました。理由はアプリの外で伝えてください。",
      ),
    ).toBeInTheDocument();
    expect(rejectMemberLink).toHaveBeenCalledWith("link-1");
  });

  it("tells the organizer when another organizer decided first", async () => {
    renderAt("/approvals", {
      listPendingMemberLinks: () => Promise.resolve(waiting),
      approveMemberLink: () => Promise.reject(problem(409, "state_conflict")),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro（taro_line）を承認" }));

    expect(
      await screen.findByText("この紐付けはもう決まっていたので、最新を出しています。"),
    ).toBeInTheDocument();
  });

  it("says when nobody is waiting", async () => {
    renderAt("/approvals", { listPendingMemberLinks: () => Promise.resolve({ links: [] }) });

    expect(await screen.findByText("承認待ちはありません")).toBeInTheDocument();
  });
});

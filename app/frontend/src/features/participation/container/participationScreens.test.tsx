import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, Entry, InvitationList, Me, Race } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { createAppRouter } from "../../../app/router.tsx";
import { TestProviders } from "../../../test/TestProviders.tsx";

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
  const client = createFakeApiClient({
    getMe: () => Promise.resolve(me),
    listRaces: () => Promise.resolve({ races: [] }),
    ...stubs,
  });
  render(
    <TestProviders client={client}>
      <RouterProvider router={router} />
    </TestProviders>,
  );
  return router;
}

function raceWith(myAttendance: Race["my_attendance"]): Race {
  return {
    id: "race-1",
    name: "Hakone",
    description: { type: "distance", race_date: "2026-11-23", leg_distances_km: [5] },
    my_attendance: myAttendance,
    can_submit_laps: false,
  };
}

async function attendanceGroup() {
  return within(await screen.findByRole("group", { name: "あなたの出欠" }));
}

describe("R2 attendance answer", () => {
  it("sends an answer that needs no warning at once", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith(null)),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("radio", { name: "参加" }));

    expect(await screen.findByText("回答を保存しました。")).toBeInTheDocument();
    expect(answerAttendance).toHaveBeenCalledWith("race-1", "attending");
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("warns about the entry before leaving attending, and sends nothing when cancelled", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith("attending")),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("radio", { name: "保留" }));

    const dialog = screen.getByRole("alertdialog", { name: "保留にしますか？" });
    expect(dialog).toHaveTextContent("エントリーに入っていれば、エントリーと走順から外れます");
    fireEvent.click(within(dialog).getByRole("button", { name: "やめる" }));
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    expect(answerAttendance).not.toHaveBeenCalled();
  });

  it("sends the change after the member confirms", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith("attending")),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("radio", { name: "保留" }));
    fireEvent.click(screen.getByRole("button", { name: "保留にする" }));

    await waitFor(() => expect(answerAttendance).toHaveBeenCalledWith("race-1", "undecided"));
  });

  it("says not attending is final, and goes home after it", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    const router = renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith(null)),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("radio", { name: "不参加" }));

    const dialog = screen.getByRole("alertdialog", { name: "不参加にしますか？" });
    expect(dialog).toHaveTextContent("回答は変えられません");
    expect(dialog).not.toHaveTextContent("エントリーに入っていれば");
    fireEvent.click(within(dialog).getByRole("button", { name: "不参加にする" }));

    expect(await screen.findByRole("heading", { name: "大会" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/");
    expect(answerAttendance).toHaveBeenCalledWith("race-1", "not_attending");
  });

  it("gives both warnings when leaving attending for not attending", async () => {
    renderAt("/races/race-1", member, { getRace: () => Promise.resolve(raceWith("attending")) });

    fireEvent.click((await attendanceGroup()).getByRole("radio", { name: "不参加" }));

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent("回答は変えられません");
    expect(dialog).toHaveTextContent("エントリーに入っていれば");
  });

  it("tells the member when they can no longer answer", async () => {
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith(null)),
      answerAttendance: () => Promise.reject(problem(404, "not_found")),
    });

    fireEvent.click((await attendanceGroup()).getByRole("radio", { name: "参加" }));

    expect(await screen.findByText("この大会にはもう回答できません。")).toBeInTheDocument();
  });

  it("shows no answer to an organizer who is not invited", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve({ ...raceWith(null), my_attendance: undefined, version: 1 }),
    });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(screen.queryByRole("group", { name: "あなたの出欠" })).not.toBeInTheDocument();
  });
});

const invitations: InvitationList = {
  members: [
    { roster_member_id: "m1", name: "Taro", invited: true, attendance: "attending" },
    { roster_member_id: "m2", name: "Jiro", invited: false, attendance: null },
    { roster_member_id: "m3", name: "Hanako", invited: true, attendance: null },
  ],
};

describe("R4 invitations", () => {
  it("invites a member", async () => {
    const invite = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/invitations", organizer, {
      listInvitations: () => Promise.resolve(invitations),
      invite,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Jiro さんを誘う" }));

    expect(await screen.findByText("Jiro さんを誘いました。")).toBeInTheDocument();
    expect(invite).toHaveBeenCalledWith("race-1", "m2");
  });

  it("tells the organizer when the member has left", async () => {
    renderAt("/races/race-1/invitations", organizer, {
      listInvitations: () => Promise.resolve(invitations),
      invite: () => Promise.reject(problem(422, "validation_failed")),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Jiro さんを誘う" }));

    expect(await screen.findByText("Jiro さんはもう退会しています。")).toBeInTheDocument();
  });

  it("warns before cancelling an invitation", async () => {
    const cancelInvitation = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/invitations", organizer, {
      listInvitations: () => Promise.resolve(invitations),
      cancelInvitation,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Taro さんの招待を取り消す" }));

    const dialog = screen.getByRole("alertdialog", { name: "Taro さんの招待を取り消しますか？" });
    expect(dialog).toHaveTextContent("出欠、エントリー、走順からも外れます");
    expect(cancelInvitation).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "取り消す" }));

    expect(await screen.findByText("Taro さんの招待を取り消しました。")).toBeInTheDocument();
    expect(cancelInvitation).toHaveBeenCalledWith("race-1", "m1");
  });
});

describe("R5 attendance list", () => {
  it("groups the invited members by their answer", async () => {
    renderAt("/races/race-1/attendance", organizer, {
      listInvitations: () => Promise.resolve(invitations),
    });

    const attending = await screen.findByRole("region", { name: "○ 参加" });
    expect(within(attending).getByRole("heading")).toHaveTextContent("○ 参加（1）");
    expect(within(attending).getByText("Taro")).toBeInTheDocument();
    const notAnswered = screen.getByRole("region", { name: "未回答" });
    expect(within(notAnswered).getByText("Hanako")).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "△ 保留" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});

describe("R6 entry", () => {
  const entry: Entry = {
    id: "e1",
    name: "Team A",
    members: [{ roster_member_id: "m4", name: "Saburo" }],
  };

  it("creates the entry with a name", async () => {
    const createEntry = vi.fn(() => Promise.resolve(entry));
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [] }),
      listInvitations: () => Promise.resolve(invitations),
      createEntry,
    });

    fireEvent.click(await screen.findByRole("button", { name: "エントリーを作る" }));
    expect(screen.getByText("名前を入れてください。")).toBeInTheDocument();
    expect(createEntry).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText("エントリーの名前"), { target: { value: "Team A" } });
    fireEvent.click(screen.getByRole("button", { name: "エントリーを作る" }));

    expect(await screen.findByText("エントリーを作りました。")).toBeInTheDocument();
    expect(createEntry).toHaveBeenCalledWith("race-1", "Team A");
  });

  it("tells the organizer when the race already has an entry", async () => {
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [] }),
      listInvitations: () => Promise.resolve(invitations),
      createEntry: () => Promise.reject(problem(409, "state_conflict")),
    });

    fireEvent.change(await screen.findByLabelText("エントリーの名前"), {
      target: { value: "Team B" },
    });
    fireEvent.click(screen.getByRole("button", { name: "エントリーを作る" }));

    expect(await screen.findByText("この大会にはもうエントリーがあります。")).toBeInTheDocument();
  });

  it("offers only attending members, and adds one", async () => {
    const addEntryMember = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [entry] }),
      listInvitations: () => Promise.resolve(invitations),
      addEntryMember,
    });

    const attending = within(await screen.findByRole("region", { name: "入れられるメンバー" }));
    expect(
      attending.queryByRole("button", { name: "Hanako さんを入れる" }),
    ).not.toBeInTheDocument();
    fireEvent.click(attending.getByRole("button", { name: "Taro さんを入れる" }));

    await waitFor(() => expect(addEntryMember).toHaveBeenCalledWith("race-1", "e1", "m1"));
  });

  it("warns before removing a member", async () => {
    const removeEntryMember = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [entry] }),
      listInvitations: () => Promise.resolve(invitations),
      removeEntryMember,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Saburo さんを外す" }));

    const dialog = screen.getByRole("alertdialog", {
      name: "Saburo さんをエントリーから外しますか？",
    });
    expect(dialog).toHaveTextContent("走順からも外れます");
    fireEvent.click(within(dialog).getByRole("button", { name: "外す" }));

    await waitFor(() => expect(removeEntryMember).toHaveBeenCalledWith("race-1", "e1", "m4"));
  });
});

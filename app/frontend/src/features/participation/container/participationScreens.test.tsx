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
  return within(await screen.findByRole("group", { name: "My attendance" }));
}

describe("R2 attendance answer", () => {
  it("sends an answer that needs no warning at once", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith(null)),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("button", { name: "Attending" }));

    expect(await screen.findByText("Your answer was saved.")).toBeInTheDocument();
    expect(answerAttendance).toHaveBeenCalledWith("race-1", "attending");
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("warns about the entry before leaving attending, and sends nothing when cancelled", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith("attending")),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("button", { name: "Undecided" }));

    const dialog = screen.getByRole("alertdialog", { name: "Change your answer?" });
    expect(dialog).toHaveTextContent("If you are in the entry, you are removed from it");
    fireEvent.click(within(dialog).getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    expect(answerAttendance).not.toHaveBeenCalled();
  });

  it("sends the change after the member confirms", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith("attending")),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("button", { name: "Undecided" }));
    fireEvent.click(screen.getByRole("button", { name: "Change answer" }));

    await waitFor(() => expect(answerAttendance).toHaveBeenCalledWith("race-1", "undecided"));
  });

  it("says not attending is final, and goes home after it", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    const router = renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith(null)),
      answerAttendance,
    });

    fireEvent.click((await attendanceGroup()).getByRole("button", { name: "Not attending" }));

    const dialog = screen.getByRole("alertdialog", { name: "Answer not attending?" });
    expect(dialog).toHaveTextContent("Not attending is final");
    expect(dialog).not.toHaveTextContent("If you are in the entry");
    fireEvent.click(within(dialog).getByRole("button", { name: "Answer not attending" }));

    expect(await screen.findByRole("heading", { name: "Races" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/");
    expect(answerAttendance).toHaveBeenCalledWith("race-1", "not_attending");
  });

  it("gives both warnings when leaving attending for not attending", async () => {
    renderAt("/races/race-1", member, { getRace: () => Promise.resolve(raceWith("attending")) });

    fireEvent.click((await attendanceGroup()).getByRole("button", { name: "Not attending" }));

    const dialog = screen.getByRole("alertdialog");
    expect(dialog).toHaveTextContent("Not attending is final");
    expect(dialog).toHaveTextContent("If you are in the entry");
  });

  it("tells the member when they can no longer answer", async () => {
    renderAt("/races/race-1", member, {
      getRace: () => Promise.resolve(raceWith(null)),
      answerAttendance: () => Promise.reject(problem(404, "not_found")),
    });

    fireEvent.click((await attendanceGroup()).getByRole("button", { name: "Attending" }));

    expect(await screen.findByText("You can no longer answer for this race.")).toBeInTheDocument();
  });

  it("shows no answer to an organizer who is not invited", async () => {
    renderAt("/races/race-1", organizer, {
      getRace: () => Promise.resolve({ ...raceWith(null), my_attendance: undefined, version: 1 }),
    });

    expect(await screen.findByRole("heading", { name: "Hakone" })).toBeInTheDocument();
    expect(screen.queryByRole("group", { name: "My attendance" })).not.toBeInTheDocument();
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

    fireEvent.click(await screen.findByRole("button", { name: "Invite Jiro" }));

    expect(await screen.findByText("Jiro was invited.")).toBeInTheDocument();
    expect(invite).toHaveBeenCalledWith("race-1", "m2");
  });

  it("tells the organizer when the member has left", async () => {
    renderAt("/races/race-1/invitations", organizer, {
      listInvitations: () => Promise.resolve(invitations),
      invite: () => Promise.reject(problem(422, "validation_failed")),
    });

    fireEvent.click(await screen.findByRole("button", { name: "Invite Jiro" }));

    expect(await screen.findByText("Jiro has left the team.")).toBeInTheDocument();
  });

  it("warns before cancelling an invitation", async () => {
    const cancelInvitation = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/invitations", organizer, {
      listInvitations: () => Promise.resolve(invitations),
      cancelInvitation,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Cancel Taro's invitation" }));

    const dialog = screen.getByRole("alertdialog", { name: "Cancel Taro's invitation?" });
    expect(dialog).toHaveTextContent(
      "also removed from attendance, the entry, and the running order",
    );
    expect(cancelInvitation).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole("button", { name: "Cancel invitation" }));

    expect(await screen.findByText("Taro's invitation was cancelled.")).toBeInTheDocument();
    expect(cancelInvitation).toHaveBeenCalledWith("race-1", "m1");
  });
});

describe("R5 attendance list", () => {
  it("shows each invited member's answer", async () => {
    renderAt("/races/race-1/attendance", organizer, {
      listInvitations: () => Promise.resolve(invitations),
    });

    const rows = await screen.findAllByRole("row");
    expect(rows.map((row) => row.textContent)).toEqual([
      "MemberAnswer",
      "TaroAttending",
      "HanakoNot answered",
    ]);
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

    fireEvent.click(await screen.findByRole("button", { name: "Create entry" }));
    expect(screen.getByText("Enter a name.")).toBeInTheDocument();
    expect(createEntry).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText("Entry name"), { target: { value: "Team A" } });
    fireEvent.click(screen.getByRole("button", { name: "Create entry" }));

    expect(await screen.findByText("The entry was created.")).toBeInTheDocument();
    expect(createEntry).toHaveBeenCalledWith("race-1", "Team A");
  });

  it("tells the organizer when the race already has an entry", async () => {
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [] }),
      listInvitations: () => Promise.resolve(invitations),
      createEntry: () => Promise.reject(problem(409, "state_conflict")),
    });

    fireEvent.change(await screen.findByLabelText("Entry name"), { target: { value: "Team B" } });
    fireEvent.click(screen.getByRole("button", { name: "Create entry" }));

    expect(await screen.findByText("This race already has an entry.")).toBeInTheDocument();
  });

  it("offers only attending members, and adds one", async () => {
    const addEntryMember = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [entry] }),
      listInvitations: () => Promise.resolve(invitations),
      addEntryMember,
    });

    const attending = within(await screen.findByRole("region", { name: "Attending members" }));
    expect(attending.queryByRole("button", { name: "Add Hanako" })).not.toBeInTheDocument();
    fireEvent.click(attending.getByRole("button", { name: "Add Taro" }));

    await waitFor(() => expect(addEntryMember).toHaveBeenCalledWith("race-1", "e1", "m1"));
  });

  it("warns before removing a member", async () => {
    const removeEntryMember = vi.fn(() => Promise.resolve());
    renderAt("/races/race-1/entry", organizer, {
      listEntries: () => Promise.resolve({ entries: [entry] }),
      listInvitations: () => Promise.resolve(invitations),
      removeEntryMember,
    });

    fireEvent.click(await screen.findByRole("button", { name: "Remove Saburo" }));

    const dialog = screen.getByRole("alertdialog", { name: "Remove Saburo from the entry?" });
    expect(dialog).toHaveTextContent("also removed from the running order");
    fireEvent.click(within(dialog).getByRole("button", { name: "Remove" }));

    await waitFor(() => expect(removeEntryMember).toHaveBeenCalledWith("race-1", "e1", "m4"));
  });
});

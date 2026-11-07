import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ApiClient, Me, RosterMember } from "../../../api/client.ts";
import { createFakeApiClient } from "../../../api/fakeApiClient.ts";
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

    const rows = await screen.findAllByRole("row");
    expect(rows.slice(1).map((row) => within(row).getAllByRole("cell")[2].textContent)).toEqual([
      "Approved",
      "No user",
    ]);
    expect(within(rows[2]).getAllByRole("cell")[1]).toHaveTextContent("Left");
    expect(screen.getByRole("button", { name: "Unlink Taro" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Undo left for Jiro" })).toBeInTheDocument();
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
      await screen.findByRole("button", { name: "Create an invite link for Taro" }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /invite link for Jiro/ })).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /invite link for Saburo/ }),
    ).not.toBeInTheDocument();
  });

  it("adds a member by name", async () => {
    const createRosterMember = vi.fn(() => Promise.resolve(member({})));
    renderAt("/roster", { ...rosterOf(), createRosterMember });

    fireEvent.change(await screen.findByLabelText("Name"), { target: { value: " Taro " } });
    fireEvent.click(screen.getByRole("button", { name: "Add" }));

    expect(await screen.findByText("Roster member added.")).toBeInTheDocument();
    expect(createRosterMember).toHaveBeenCalledWith("Taro");
    expect(screen.getByLabelText("Name")).toHaveValue("");
  });

  it("shows an error for an empty name and sends nothing", async () => {
    const createRosterMember = vi.fn();
    renderAt("/roster", { ...rosterOf(), createRosterMember });

    fireEvent.click(await screen.findByRole("button", { name: "Add" }));

    expect(await screen.findByText("Enter a name.")).toBeInTheDocument();
    expect(createRosterMember).not.toHaveBeenCalled();
  });
});

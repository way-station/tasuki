import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { createFakeApiClient } from "../api/fakeApiClient.ts";
import { TestProviders } from "../test/TestProviders.tsx";
import { createAppRouter } from "./router.tsx";

function renderAt(path: string) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  // An organizer linked to a roster member may open every screen. What other
  // callers see is tested with the route guard (features/auth/container).
  const client = createFakeApiClient({
    listRaces: () => Promise.resolve({ races: [] }),
    listInvitations: () => Promise.resolve({ members: [] }),
    listEntries: () => Promise.resolve({ entries: [] }),
    listRaceLaps: () => Promise.resolve({ runners: [] }),
    getMyLaps: () => Promise.resolve({ laps: [], average_time: "3:40", lap_count: 0 }),
    getMyProfile: () =>
      Promise.resolve({ name: "Taro Yamada", average_time: "3:40", lap_count: 0 }),
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
    ["/login", "Log in"],
    ["/auth/complete", "Log in"],
    ["/waiting", "Waiting for approval"],
    ["/rejected", "You were rejected"],
    ["/unavailable", "tasuki cannot be used"],
    ["/", "Races"],
    ["/races/new", "Register race"],
    ["/races/123", "Hakone"],
    ["/races/123/edit", "Edit race"],
    ["/races/123/invitations", "Invitations"],
    ["/races/123/attendance", "Attendance"],
    ["/races/123/entry", "Entry"],
    ["/races/123/running-order", "R7 Running order editor"],
    ["/races/123/laps", "My lap times"],
    ["/me", "Taro Yamada"],
    ["/roster", "M1 Roster"],
    ["/roster/123", "M2 Member profile"],
    ["/approvals", "M3 Approvals"],
  ])("shows %s as %s", async (path, heading) => {
    renderAt(path);

    expect(await screen.findByRole("heading", { name: heading })).toBeInTheDocument();
  });

  it.each(["/unknown", "/races/123/unknown", "/roster/123/unknown"])(
    "shows the not-found page for %s",
    async (path) => {
      renderAt(path);

      expect(await screen.findByRole("heading", { name: "Page not found" })).toBeInTheDocument();
    },
  );

  it("shows every screen inside the common layout", async () => {
    renderAt("/races/123");

    const heading = await screen.findByRole("heading", { name: "Hakone" });
    expect(screen.getByRole("banner")).toHaveTextContent("tasuki");
    expect(screen.getByRole("main")).toContainElement(heading);
  });

  it("shows the not-found page inside the common layout", async () => {
    renderAt("/unknown");

    const heading = await screen.findByRole("heading", { name: "Page not found" });
    expect(screen.getByRole("main")).toContainElement(heading);
  });
});

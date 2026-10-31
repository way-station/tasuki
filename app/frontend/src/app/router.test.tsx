import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { createFakeApiClient } from "../api/fakeApiClient.ts";
import { TestProviders } from "../test/TestProviders.tsx";
import { createAppRouter } from "./router.tsx";

function renderAt(path: string) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  const client = createFakeApiClient({
    getHealth: () => Promise.resolve({ status: "ok", mongo: "ok", postgres: "ok" }),
  });
  render(
    <TestProviders client={client}>
      <RouterProvider router={router} />
    </TestProviders>,
  );
}

describe("createAppRouter", () => {
  it.each([
    ["/invite/abc", "A1 Invite link landing"],
    ["/auth/complete?code=abc", "Completing login"],
    ["/waiting", "A2 Waiting for approval"],
    ["/rejected", "A3 Rejected"],
    ["/unavailable", "A4 Cannot use tasuki"],
    ["/", "R1 Race list"],
    ["/races/new", "R3 Register race"],
    ["/races/123", "R2 Race detail"],
    ["/races/123/edit", "R3 Edit race"],
    ["/races/123/invitations", "R4 Race invitations"],
    ["/races/123/attendance", "R5 Attendance list"],
    ["/races/123/entry", "R6 Entry"],
    ["/races/123/running-order", "R7 Running order editor"],
    ["/races/123/laps", "L1 My lap times"],
    ["/me", "P1 My profile"],
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

    const heading = await screen.findByRole("heading", { name: "R2 Race detail" });
    expect(screen.getByRole("banner")).toHaveTextContent("tasuki");
    expect(screen.getByRole("main")).toContainElement(heading);
  });

  it("shows the backend health on the home page", async () => {
    renderAt("/");

    expect(await screen.findByRole("region", { name: "Backend health" })).toBeInTheDocument();
  });

  it("shows the not-found page inside the common layout", async () => {
    renderAt("/unknown");

    const heading = await screen.findByRole("heading", { name: "Page not found" });
    expect(screen.getByRole("main")).toContainElement(heading);
  });
});

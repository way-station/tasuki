import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";
import { AppRoutes } from "./routes.tsx";

function renderAt(path: string) {
  render(
    <MemoryRouter initialEntries={[path]}>
      <AppRoutes />
    </MemoryRouter>,
  );
}

describe("AppRoutes", () => {
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
  ])("shows %s as %s", (path, heading) => {
    renderAt(path);

    expect(screen.getByRole("heading", { name: heading })).toBeInTheDocument();
  });

  it.each(["/unknown", "/races/123/unknown", "/roster/123/unknown"])(
    "shows the not-found page for %s",
    (path) => {
      renderAt(path);

      expect(screen.getByRole("heading", { name: "Page not found" })).toBeInTheDocument();
    },
  );

  it("shows every screen inside the common layout", () => {
    renderAt("/races/123");

    expect(screen.getByRole("banner")).toHaveTextContent("tasuki");
    expect(screen.getByRole("main")).toContainElement(
      screen.getByRole("heading", { name: "R2 Race detail" }),
    );
  });
});
